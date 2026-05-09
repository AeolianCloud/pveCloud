package iam

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

type AdminSessionListFilters struct {
	Keyword string
	Status  string
}

type AdminSessionListRow struct {
	SessionID        string     `gorm:"column:session_id"`
	AdminID          uint64     `gorm:"column:admin_id"`
	AdminUsername    string     `gorm:"column:admin_username"`
	AdminDisplayName string     `gorm:"column:admin_display_name"`
	AdminEmail       *string    `gorm:"column:admin_email"`
	Status           string     `gorm:"column:status"`
	IssuedAt         time.Time  `gorm:"column:issued_at"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	LastSeenAt       *time.Time `gorm:"column:last_seen_at"`
	LastSeenIP       *string    `gorm:"column:last_seen_ip"`
	UserAgent        *string    `gorm:"column:user_agent"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	RevokeReason     *string    `gorm:"column:revoke_reason"`
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AdminSessionRows(ctx context.Context, filters AdminSessionListFilters, limit int, offset int) ([]AdminSessionListRow, int64, error) {
	query := r.applyAdminSessionListFilters(r.db.WithContext(ctx).
		Table("admin_sessions AS sessions").
		Joins("JOIN admin_users ON admin_users.id = sessions.admin_id").
		Where("admin_users.deleted_at IS NULL"), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []AdminSessionListRow
	if err := query.
		Select(`
			sessions.session_id,
			sessions.admin_id,
			admin_users.username AS admin_username,
			admin_users.display_name AS admin_display_name,
			admin_users.email AS admin_email,
			sessions.status,
			sessions.issued_at,
			sessions.expires_at,
			sessions.last_seen_at,
			sessions.last_seen_ip,
			sessions.user_agent,
			sessions.revoked_at,
			sessions.revoke_reason
		`).
		Order("CASE sessions.status WHEN 'active' THEN 0 WHEN 'revoked' THEN 1 ELSE 2 END").
		Order("sessions.issued_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) FindAdminSessionBySessionIDForUpdate(ctx context.Context, db *gorm.DB, sessionID string) (AdminSession, error) {
	var session AdminSession
	err := r.queryDB(db).
		WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ?", sessionID).
		First(&session).Error
	return session, err
}

func (r *Repository) FindActiveAdminSessionBySessionIDForUpdate(ctx context.Context, db *gorm.DB, sessionID string, adminID uint64) (AdminSession, error) {
	var session AdminSession
	err := r.queryDB(db).
		WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ? AND admin_id = ? AND status = ?", sessionID, adminID, "active").
		First(&session).Error
	return session, err
}

func (r *Repository) FindAdminSessionByID(ctx context.Context, sessionID string, adminID uint64) (AdminSession, error) {
	var session AdminSession
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND admin_id = ?", sessionID, adminID).
		First(&session).Error
	return session, err
}

func (r *Repository) FindAdminByID(ctx context.Context, adminID uint64) (AdminUser, error) {
	var admin AdminUser
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id = ?", adminID).
		First(&admin).Error
	return admin, err
}

func (r *Repository) AdminRoleIDs(ctx context.Context, db *gorm.DB, adminID uint64, activeStatus string) ([]uint64, error) {
	var roleIDs []uint64
	err := r.queryDB(db).
		WithContext(ctx).
		Table("admin_user_roles").
		Select("admin_roles.id").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", activeStatus).
		Order("admin_roles.id ASC").
		Scan(&roleIDs).Error
	return roleIDs, err
}

func (r *Repository) AdminPermissionCodes(ctx context.Context, db *gorm.DB, adminID uint64, activeStatus string) ([]string, error) {
	var codes []string
	err := r.queryDB(db).
		WithContext(ctx).
		Table("admin_user_roles").
		Distinct("admin_permissions.code").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Joins("JOIN admin_role_permissions ON admin_role_permissions.role_id = admin_roles.id").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", activeStatus).
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func (r *Repository) CreateAdminSession(ctx context.Context, db *gorm.DB, session *AdminSession) error {
	return r.queryDB(db).WithContext(ctx).Create(session).Error
}

func (r *Repository) UpdateAdminSessionState(ctx context.Context, db *gorm.DB, id uint64, status string, revokedAt time.Time, reason string) error {
	return r.queryDB(db).
		WithContext(ctx).
		Model(&AdminSession{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        status,
			"revoked_at":    revokedAt,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) RevokeActiveAdminSessionBySessionID(ctx context.Context, db *gorm.DB, sessionID string, adminID uint64, now time.Time, reason string) error {
	return r.queryDB(db).
		WithContext(ctx).
		Model(&AdminSession{}).
		Where("session_id = ? AND admin_id = ? AND status = ?", sessionID, adminID, "active").
		Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) RevokeActiveAdminSessionsByAdminID(ctx context.Context, db *gorm.DB, adminID uint64, now time.Time, reason string) error {
	return r.queryDB(db).
		WithContext(ctx).
		Model(&AdminSession{}).
		Where("admin_id = ? AND status = ?", adminID, "active").
		Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) ActiveAdminSessions(ctx context.Context, adminID uint64, now time.Time, limit int) ([]AdminSession, error) {
	var sessions []AdminSession
	if err := r.db.WithContext(ctx).
		Where("admin_id = ? AND status = ?", adminID, "active").
		Where("expires_at > ?", now).
		Order("issued_at DESC").
		Limit(limit).
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *Repository) TouchAdminSession(ctx context.Context, id uint64, now time.Time, clientIP string, userAgent string) error {
	return r.db.WithContext(ctx).
		Model(&AdminSession{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_seen_at": now,
			"last_seen_ip": clientIP,
			"user_agent":   userAgent,
		}).Error
}

func (r *Repository) ExpireStaleAdminSessions(ctx context.Context, db *gorm.DB, now time.Time, activeStatus string, expiredStatus string, reason string) error {
	return r.queryDB(db).
		WithContext(ctx).
		Model(&AdminSession{}).
		Where("status = ? AND expires_at <= ?", activeStatus, now).
		Updates(map[string]any{
			"status":        expiredStatus,
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) applyAdminSessionListFilters(db *gorm.DB, filters AdminSessionListFilters) *gorm.DB {
	if filters.Keyword != "" {
		keyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		db = db.Where(
			"sessions.session_id LIKE ? OR admin_users.username LIKE ? OR admin_users.display_name LIKE ? OR sessions.last_seen_ip LIKE ?",
			keyword,
			keyword,
			keyword,
			keyword,
		)
	}
	if filters.Status != "" {
		db = db.Where("sessions.status = ?", strings.TrimSpace(filters.Status))
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
