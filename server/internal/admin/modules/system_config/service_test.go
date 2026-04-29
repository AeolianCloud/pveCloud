package systemconfig

import (
	"strings"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
)

func TestSystemConfigItemHidesSecretValue(t *testing.T) {
	value := "secret"
	item := systemConfigItem(models.SystemConfig{
		ID:          1,
		ConfigKey:   "payment.secret",
		ConfigValue: &value,
		ValueType:   "string",
		GroupName:   "payment",
		IsSecret:    true,
		UpdatedAt:   time.Now(),
	})

	if item.ConfigValue != nil {
		t.Fatalf("expected secret config value hidden, got %q", *item.ConfigValue)
	}
	if !item.HasValue {
		t.Fatal("expected has_value true for stored secret")
	}
}

func TestSystemConfigAuditSnapshotMasksSecretValue(t *testing.T) {
	value := "secret"
	snapshot := systemConfigAuditSnapshot(models.SystemConfig{
		ID:          1,
		ConfigKey:   "payment.secret",
		ConfigValue: &value,
		IsSecret:    true,
	})

	raw, ok := snapshot["config_value"].(*string)
	if !ok || raw == nil {
		t.Fatalf("expected masked config_value pointer, got %#v", snapshot["config_value"])
	}
	if strings.Contains(*raw, "secret") || *raw != adminAuditMaskedValue {
		t.Fatalf("expected masked value, got %q", *raw)
	}
}
