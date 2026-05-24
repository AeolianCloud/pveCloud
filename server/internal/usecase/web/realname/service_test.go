package realname

import (
	"context"
	"testing"

	"gorm.io/gorm"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

func TestSubmitManualRealNameMasksIDNumberAndBlocksPendingDuplicate(t *testing.T) {
	db := openRealNameDB(t)
	seedRealNameUserAndConfig(t, db)

	service := NewRealNameService(db, nil)
	req := webdto.RealNameSubmitRequest{
		RealName: "Test User",
		IDType:   "id_card",
		IDNumber: "110101199001011234",
		Provider: "manual",
	}

	result, err := service.Submit(context.Background(), 11, req)
	if err != nil {
		t.Fatalf("submit manual real name: %v", err)
	}
	if result.ProviderAction.Provider != providerManual || result.ProviderAction.ActionType != "manual_review" || result.ProviderAction.RedirectURL != "" {
		t.Fatalf("manual submit should return manual review action, got %#v", result.ProviderAction)
	}
	if result.Application.IDNumberMasked != "1101**********1234" {
		t.Fatalf("application should expose only masked ID number, got %q", result.Application.IDNumberMasked)
	}

	var row struct {
		Status                string  `gorm:"column:status"`
		IDNumberDigest        *string `gorm:"column:id_number_digest"`
		IDNumberDigestVersion *string `gorm:"column:id_number_digest_version"`
		IDNumberMasked        string  `gorm:"column:id_number_masked"`
		VerificationProvider  string  `gorm:"column:verification_provider"`
	}
	if err := db.Table("user_real_name_applications").Where("user_id = ?", 11).Take(&row).Error; err != nil {
		t.Fatalf("load real name application: %v", err)
	}
	if row.Status != statusPending || row.VerificationProvider != providerManual {
		t.Fatalf("manual application should stay pending/manual, got %#v", row)
	}
	if row.IDNumberDigest != nil || row.IDNumberDigestVersion != nil {
		t.Fatalf("manual fallback without digest secret should not invent a digest, got digest=%v version=%v", row.IDNumberDigest, row.IDNumberDigestVersion)
	}
	if row.IDNumberMasked != "1101**********1234" {
		t.Fatalf("database should keep masked ID number only, got %q", row.IDNumberMasked)
	}

	_, err = service.Submit(context.Background(), 11, req)
	if apperrors.From(err).Code != apperrors.ErrConflict.Code {
		t.Fatalf("duplicate pending submit should conflict, got %v", err)
	}

	var applicationCount int64
	if err := db.Table("user_real_name_applications").Where("user_id = ?", 11).Count(&applicationCount).Error; err != nil {
		t.Fatalf("count real name applications: %v", err)
	}
	if applicationCount != 1 {
		t.Fatalf("duplicate pending submit must not create another application, got %d", applicationCount)
	}
	var logCount int64
	if err := db.Table("user_business_logs").Where("user_id = ? AND action = ? AND object_type = ?", 11, "real_name.submit", "real_name_application").Count(&logCount).Error; err != nil {
		t.Fatalf("count user business logs: %v", err)
	}
	if logCount != 1 {
		t.Fatalf("successful manual submit should write one business log, got %d", logCount)
	}
}

func openRealNameDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, realNameUsersSchema, realNameSystemConfigsSchema, realNameApplicationsSchema, realNameUserBusinessLogsSchema)
	return db
}

func seedRealNameUserAndConfig(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`INSERT INTO users (id, username, email, password_hash, status) VALUES (?, ?, ?, ?, ?)`, 11, "real-user", "real@example.com", "hash", "active").Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}
	statements := []string{
		`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES ('real_name.enabled', 'true', 'bool', '实名设置', 0)`,
		`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES ('real_name.allowed_providers', 'manual', 'string', '实名设置', 0)`,
		`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES ('real_name.default_provider', 'manual', 'string', '实名设置', 0)`,
		`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES ('real_name.manual_review_enabled', 'true', 'bool', '实名设置', 0)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed real name config with %q: %v", statement, err)
		}
	}
}

const realNameUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(191) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const realNameSystemConfigsSchema = `
CREATE TABLE system_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  config_key VARCHAR(191) NOT NULL,
  config_value TEXT NULL,
  value_type VARCHAR(32) NOT NULL DEFAULT 'string',
  group_name VARCHAR(64) NOT NULL DEFAULT '',
  is_secret TINYINT(1) NOT NULL DEFAULT 0,
  description VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_system_configs_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const realNameApplicationsSchema = `
CREATE TABLE user_real_name_applications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  application_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  real_name VARCHAR(64) NOT NULL,
  id_type VARCHAR(32) NOT NULL,
  id_number_digest VARCHAR(128) NULL,
  id_number_digest_version VARCHAR(32) NULL,
  id_number_masked VARCHAR(64) NOT NULL,
  verification_provider VARCHAR(32) NULL,
  provider_application_id VARCHAR(128) NULL,
  provider_status VARCHAR(32) NULL,
  provider_result_code VARCHAR(64) NULL,
  provider_result_message VARCHAR(255) NULL,
  provider_started_at DATETIME(3) NULL,
  provider_finished_at DATETIME(3) NULL,
  provider_response_digest VARCHAR(128) NULL,
  provider_trace_id VARCHAR(128) NULL,
  status VARCHAR(32) NOT NULL,
  reject_reason VARCHAR(255) NULL,
  submit_attempt INT UNSIGNED NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_real_name_application_no (application_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const realNameUserBusinessLogsSchema = `
CREATE TABLE user_business_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  username VARCHAR(64) NULL,
  email VARCHAR(191) NULL,
  request_id VARCHAR(64) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  module VARCHAR(64) NOT NULL,
  action VARCHAR(96) NOT NULL,
  object_type VARCHAR(64) NOT NULL,
  object_id VARCHAR(128) NULL,
  summary VARCHAR(500) NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
