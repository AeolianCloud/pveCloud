package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadConfigAppliesConfiguredTimezone(t *testing.T) {
	oldLocal := time.Local
	t.Cleanup(func() {
		time.Local = oldLocal
	})

	path := writeTestConfig(t, "UTC")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.App.Timezone != "UTC" {
		t.Fatalf("cfg.App.Timezone = %q, want UTC", cfg.App.Timezone)
	}
	if time.Local.String() != "UTC" {
		t.Fatalf("time.Local = %q, want UTC", time.Local.String())
	}
}

func TestLoadConfigRejectsInvalidTimezone(t *testing.T) {
	path := writeTestConfig(t, "Not/AZone")
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want invalid timezone error")
	}
	if !strings.Contains(err.Error(), "app.timezone") {
		t.Fatalf("LoadConfig() error = %q, want app.timezone error", err.Error())
	}
}

func writeTestConfig(t *testing.T, timezone string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := "app:\n  timezone: " + timezone + "\njwt:\n  user_secret: test_user_secret_32_chars_minimum\n  admin_secret: test_admin_secret_32_chars_minimum\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
