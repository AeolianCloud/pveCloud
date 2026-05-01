package audit

import (
	"strings"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
)

func TestAuditJSONPtrMasksSensitiveFields(t *testing.T) {
	raw := auditJSONPtr(map[string]any{
		"username": "root",
		"password": "secret-password",
		"profile": map[string]any{
			"access_token": "secret-token",
			"display":      "Root Admin",
		},
		"items": []any{
			map[string]any{"captcha_code": "1234"},
		},
	})

	if raw == nil {
		t.Fatal("expected masked json")
	}
	if strings.Contains(*raw, "secret-password") || strings.Contains(*raw, "secret-token") || strings.Contains(*raw, "1234") {
		t.Fatalf("expected sensitive values masked, got %s", *raw)
	}
	if !strings.Contains(*raw, adminAuditMaskedValue) || !strings.Contains(*raw, "Root Admin") {
		t.Fatalf("expected masked marker and non-sensitive value, got %s", *raw)
	}
}

func TestAuditLogItemHidesSensitiveDataWithoutPermission(t *testing.T) {
	beforeData := `{"password":"secret-password","display_name":"Root Admin"}`
	row := auditLogRow{
		AdminAuditLog: models.AdminAuditLog{
			BeforeData: &beforeData,
		},
	}

	item := row.auditItem(false)
	if item.BeforeData != nil || item.AfterData != nil || item.UserAgent != nil {
		t.Fatal("expected sensitive fields hidden without permission")
	}
}

func TestAuditLogItemMasksLegacySensitiveDataWithPermission(t *testing.T) {
	beforeData := `{"password":"secret-password","display_name":"Root Admin"}`
	row := auditLogRow{
		AdminAuditLog: models.AdminAuditLog{
			BeforeData: &beforeData,
		},
	}

	item := row.auditItem(true)
	if item.BeforeData == nil {
		t.Fatal("expected masked before_data with permission")
	}
	if strings.Contains(*item.BeforeData, "secret-password") || !strings.Contains(*item.BeforeData, adminAuditMaskedValue) {
		t.Fatalf("expected legacy sensitive value masked, got %s", *item.BeforeData)
	}
}

func TestBuildAdminAuditLogUsesRequestContext(t *testing.T) {
	adminID := uint64(42)
	ctx := WithRequestContext(t.Context(), RequestContext{
		AdminID:          &adminID,
		AdminUsername:    "root",
		AdminDisplayName: "Root Admin",
		SessionID:        "adm_session",
		RequestID:        "req_123",
		RequestMethod:    "PATCH",
		RequestPath:      "/admin-api/system-configs/1",
		IP:               "127.0.0.1",
		UserAgent:        "test-agent",
	})

	audit, err := buildAdminAuditLog(ctx, AdminAuditWriteInput{
		Action:     "system.config.update",
		ObjectType: "system_config",
		ObjectID:   "1",
	})
	if err != nil {
		t.Fatalf("expected audit log model, got %v", err)
	}
	if audit.AdminID == nil || *audit.AdminID != adminID {
		t.Fatalf("expected admin id from context, got %#v", audit.AdminID)
	}
	if audit.RequestID == nil || *audit.RequestID != "req_123" || audit.RequestPath == nil || *audit.RequestPath != "/admin-api/system-configs/1" {
		t.Fatalf("expected request context copied, got %#v", audit)
	}
}
