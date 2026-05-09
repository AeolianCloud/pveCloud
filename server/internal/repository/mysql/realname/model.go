package realname

import "time"

/**
 * UserRealNameApplication 映射用户实名申请表。
 */
type UserRealNameApplication struct {
	ID                     uint64     `gorm:"column:id;primaryKey"`
	ApplicationNo          string     `gorm:"column:application_no"`
	UserID                 uint64     `gorm:"column:user_id"`
	RealName               string     `gorm:"column:real_name"`
	IDType                 string     `gorm:"column:id_type"`
	IDNumberDigest         *string    `gorm:"column:id_number_digest"`
	IDNumberDigestVersion  *string    `gorm:"column:id_number_digest_version"`
	IDNumberMasked         string     `gorm:"column:id_number_masked"`
	VerificationProvider   *string    `gorm:"column:verification_provider"`
	ProviderApplicationID  *string    `gorm:"column:provider_application_id"`
	ProviderStatus         *string    `gorm:"column:provider_status"`
	ProviderResultCode     *string    `gorm:"column:provider_result_code"`
	ProviderResultMessage  *string    `gorm:"column:provider_result_message"`
	ProviderStartedAt      *time.Time `gorm:"column:provider_started_at"`
	ProviderFinishedAt     *time.Time `gorm:"column:provider_finished_at"`
	ProviderResponseDigest *string    `gorm:"column:provider_response_digest"`
	ProviderTraceID        *string    `gorm:"column:provider_trace_id"`
	Status                 string     `gorm:"column:status"`
	RejectReason           *string    `gorm:"column:reject_reason"`
	SubmitAttempt          uint       `gorm:"column:submit_attempt"`
	CreatedAt              time.Time  `gorm:"column:created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at"`
}

/**
 * TableName 返回用户实名申请表名。
 */
func (UserRealNameApplication) TableName() string {
	return "user_real_name_applications"
}
