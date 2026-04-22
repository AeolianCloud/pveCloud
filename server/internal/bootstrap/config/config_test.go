package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
)

func TestLoadFromReadsNestedConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	writeConfig(t, path, `
app_env: local
public_api_addr: ":8080"
admin_api_addr: ":8081"
worker_addr: ":8082"
mysql_dsn: root:root@tcp(127.0.0.1:3306)/pvecloud?parseTime=true&loc=Local
redis_addr: 127.0.0.1:6379
jwt_web_secret: change-me-web
jwt_admin_secret: change-me-admin
payment:
  provider: mock
  callback_base_url: http://127.0.0.1:8080
  notify_path: /payments/callback
  merchant_id: demo-merchant
  merchant_secret: demo-merchant-secret
resource:
  provider: mock
  api_endpoint: http://127.0.0.1:9000
  api_token: demo-resource-token
worker:
  poll_interval: 5s
  batch_size: 20
`)

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Payment.Provider != "mock" {
		t.Fatalf("expected payment provider to load, got %q", cfg.Payment.Provider)
	}
	if cfg.Resource.APIEndpoint != "http://127.0.0.1:9000" {
		t.Fatalf("expected resource api endpoint to load, got %q", cfg.Resource.APIEndpoint)
	}
	if cfg.Worker.PollInterval != "5s" {
		t.Fatalf("expected worker poll interval to load, got %q", cfg.Worker.PollInterval)
	}
	if cfg.Worker.BatchSize != 20 {
		t.Fatalf("expected worker batch size to load, got %d", cfg.Worker.BatchSize)
	}
}

func TestLoadFromAllowsOmittedNestedConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	writeConfig(t, path, `
app_env: local
public_api_addr: ":8080"
admin_api_addr: ":8081"
worker_addr: ":8082"
mysql_dsn: root:root@tcp(127.0.0.1:3306)/pvecloud?parseTime=true&loc=Local
redis_addr: 127.0.0.1:6379
jwt_web_secret: change-me-web
jwt_admin_secret: change-me-admin
	`)

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Payment.Provider != "" || cfg.Resource.Provider != "" || cfg.Worker.BatchSize != 0 {
		t.Fatalf("expected omitted nested sections to remain zero-valued, got %+v", cfg)
	}
}

func TestValidateForServicePublicAPIRejectsMissingPaymentProvider(t *testing.T) {
	cfg := baseConfig()
	cfg.Payment = config.PaymentConfig{
		CallbackBaseURL: "http://127.0.0.1:8080",
		NotifyPath:      "/payments/callback",
		MerchantID:      "demo-merchant",
		MerchantSecret:  "demo-merchant-secret",
	}

	err := cfg.ValidateForService("public-api")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "payment.provider") {
		t.Fatalf("expected missing payment.provider error, got %v", err)
	}
}

func TestValidateForServiceWorkerRejectsNestedConfigIssues(t *testing.T) {
	t.Run("resource", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Resource = config.ResourceConfig{
			APIEndpoint: "http://127.0.0.1:9000",
			APIToken:    "demo-resource-token",
		}
		cfg.Worker = config.WorkerConfig{
			PollInterval: "5s",
			BatchSize:    1,
		}

		err := cfg.ValidateForService("worker")
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Error(), "resource.provider") {
			t.Fatalf("expected missing resource.provider error, got %v", err)
		}
	})

	t.Run("worker", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Resource = config.ResourceConfig{
			Provider:    "mock",
			APIEndpoint: "http://127.0.0.1:9000",
			APIToken:    "demo-resource-token",
		}
		cfg.Worker = config.WorkerConfig{
			PollInterval: "invalid-duration",
			BatchSize:    0,
		}

		err := cfg.ValidateForService("worker")
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Error(), "worker.poll_interval") && !strings.Contains(err.Error(), "worker.batch_size") {
			t.Fatalf("expected worker validation error, got %v", err)
		}
	})
}

func TestValidateForServiceAdminAPIOmitsNestedValidation(t *testing.T) {
	cfg := baseConfig()

	if err := cfg.ValidateForService("admin-api"); err != nil {
		t.Fatalf("validate admin-api: %v", err)
	}
}

func baseConfig() config.Config {
	return config.Config{
		AppEnv:         "local",
		PublicAPIAddr:  ":8080",
		AdminAPIAddr:   ":8081",
		WorkerAddr:     ":8082",
		MySQLDSN:       "root:root@tcp(127.0.0.1:3306)/pvecloud?parseTime=true&loc=Local",
		RedisAddr:      "127.0.0.1:6379",
		JWTWebSecret:   "change-me-web",
		JWTAdminSecret: "change-me-admin",
	}
}

func writeConfig(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}
