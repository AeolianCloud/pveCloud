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

func TestBuildAdminRiskLogLinksAuditLog(t *testing.T) {
	risk, err := buildAdminRiskLog(AdminRiskWriteInput{
		AdminAuditWriteInput: AdminAuditWriteInput{
			Action:     "admin.login.limited",
			ObjectType: "admin_auth",
			ObjectID:   "login_hash",
		},
		RiskLevel:  "high",
		RiskReason: "登录失败次数达到限制",
	}, 42)
	if err != nil {
		t.Fatalf("expected risk log model, got %v", err)
	}
	if risk.AuditLogID == nil || *risk.AuditLogID != 42 {
		t.Fatalf("expected audit log id 42, got %#v", risk.AuditLogID)
	}
	if risk.RiskLevel != "high" || risk.RiskReason == "" {
		t.Fatalf("expected risk metadata copied, got %#v", risk)
	}
}
