package user

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

type UserListFilters struct {
	Keyword string
	Status  string
}

type UserSessionListFilters struct {
	UserID   uint64
	Status   string
	DateFrom *time.Time
	DateTo   *time.Time
}

type UserSessionListRow struct {
	SessionID     string     `gorm:"column:session_id"`
	UserID        uint64     `gorm:"column:user_id"`
	Username      string     `gorm:"column:username"`
	Email         string     `gorm:"column:email"`
	DisplayName   *string    `gorm:"column:display_name"`
	UserStatus    string     `gorm:"column:user_status"`
	UserCreatedAt time.Time  `gorm:"column:user_created_at"`
	UserUpdatedAt time.Time  `gorm:"column:user_updated_at"`
	Status        string     `gorm:"column:status"`
	IssuedAt      time.Time  `gorm:"column:issued_at"`
	ExpiresAt     time.Time  `gorm:"column:expires_at"`
	LastSeenAt    *time.Time `gorm:"column:last_seen_at"`
	LastSeenIP    *string    `gorm:"column:last_seen_ip"`
	UserAgent     *string    `gorm:"column:user_agent"`
	RevokedAt     *time.Time `gorm:"column:revoked_at"`
	RevokeReason  *string    `gorm:"column:revoke_reason"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Users(ctx context.Context, filters UserListFilters, limit int, offset int) ([]User, int64, error) {
	query := r.applyUserListFilters(r.db.WithContext(ctx).Model(&User{}), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []User
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *Repository) UserSessionRows(ctx context.Context, filters UserSessionListFilters, limit int, offset int) ([]UserSessionListRow, int64, error) {
	query := r.applyUserSessionListFilters(r.db.WithContext(ctx).
		Table("user_sessions AS sessions").
		Joins("JOIN users ON users.id = sessions.user_id"), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []UserSessionListRow
	if err := query.Select(`
		sessions.session_id,
		users.id AS user_id,
		users.username,
		users.email,
		users.display_name,
		users.status AS user_status,
		users.created_at AS user_created_at,
		users.updated_at AS user_updated_at,
		sessions.status,
		sessions.issued_at,
		sessions.expires_at,
		sessions.last_seen_at,
		sessions.last_seen_ip,
		sessions.user_agent,
		sessions.revoked_at,
		sessions.revoke_reason,
		sessions.created_at
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

func (r *Repository) CreateUser(ctx context.Context, db *gorm.DB, user *User) error {
	return r.queryDB(db).WithContext(ctx).Create(user).Error
}

func (r *Repository) FindUserByID(ctx context.Context, db *gorm.DB, id uint64) (User, error) {
	var user User
	err := r.queryDB(db).WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (r *Repository) FindUserByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (User, error) {
	var user User
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&user).Error
	return user, err
}

func (r *Repository) FindUserByAccount(ctx context.Context, account string) (User, error) {
	account = strings.TrimSpace(account)
	var user User
	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", account, account).
		First(&user).Error
	return user, err
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", strings.TrimSpace(email)).First(&user).Error
	return user, err
}

func (r *Repository) CountUsersByIdentity(ctx context.Context, excludeID uint64, username string, email string) (int64, error) {
	query := r.db.WithContext(ctx).Model(&User{})
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if username != "" && email != "" {
		query = query.Where("username = ? OR email = ?", username, email)
	} else if username != "" {
		query = query.Where("username = ?", username)
	} else if email != "" {
		query = query.Where("email = ?", email)
	} else {
		return 0, nil
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) EmailExistsForOtherUser(ctx context.Context, db *gorm.DB, email string, userID uint64) (bool, error) {
	var existing User
	err := r.queryDB(db).WithContext(ctx).
		Where("email = ? AND id <> ?", strings.TrimSpace(email), userID).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) UpdateUser(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) UpdateUserPasswordHash(ctx context.Context, db *gorm.DB, id uint64, passwordHash string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *Repository) CreateUserSession(ctx context.Context, db *gorm.DB, session *UserSession) error {
	return r.queryDB(db).WithContext(ctx).Create(session).Error
}

func (r *Repository) FindUserSessionBySessionID(ctx context.Context, sessionID string, userID uint64) (UserSession, error) {
	var session UserSession
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error
	return session, err
}

func (r *Repository) FindUserSessionBySessionIDForUpdate(ctx context.Context, db *gorm.DB, sessionID string) (UserSession, error) {
	var session UserSession
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ?", sessionID).
		First(&session).Error
	return session, err
}

func (r *Repository) FindActiveUserSessionForUpdate(ctx context.Context, db *gorm.DB, sessionID string, userID uint64, activeStatus string) (UserSession, error) {
	var session UserSession
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, activeStatus).
		First(&session).Error
	return session, err
}

func (r *Repository) UpdateUserSessionState(ctx context.Context, db *gorm.DB, id uint64, status string, revokedAt time.Time, reason string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&UserSession{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        status,
			"revoked_at":    revokedAt,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) RevokeActiveUserSessionBySessionID(ctx context.Context, db *gorm.DB, sessionID string, userID uint64, now time.Time, reason string, activeStatus string, revokedStatus string) (int64, error) {
	result := r.queryDB(db).WithContext(ctx).
		Model(&UserSession{}).
		Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, activeStatus).
		Updates(map[string]any{
			"status":        revokedStatus,
			"revoked_at":    now,
			"revoke_reason": reason,
		})
	return result.RowsAffected, result.Error
}

func (r *Repository) RevokeActiveUserSessionsByUserID(ctx context.Context, db *gorm.DB, userID uint64, now time.Time, reason string, activeStatus string, revokedStatus string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&UserSession{}).
		Where("user_id = ? AND status = ?", userID, activeStatus).
		Updates(map[string]any{
			"status":        revokedStatus,
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) RevokeOtherActiveUserSessions(ctx context.Context, db *gorm.DB, userID uint64, excludedSessionID string, now time.Time, reason string, activeStatus string, revokedStatus string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&UserSession{}).
		Where("user_id = ? AND status = ? AND session_id <> ?", userID, activeStatus, excludedSessionID).
		Updates(map[string]any{
			"status":        revokedStatus,
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) ExpireStaleUserSessions(ctx context.Context, db *gorm.DB, now time.Time, activeStatus string, expiredStatus string, reason string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&UserSession{}).
		Where("status = ? AND expires_at <= ?", activeStatus, now).
		Updates(map[string]any{
			"status":        expiredStatus,
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error
}

func (r *Repository) TouchUserSession(ctx context.Context, id uint64, now time.Time, clientIP string, userAgent string) error {
	return r.db.WithContext(ctx).
		Model(&UserSession{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_seen_at": now,
			"last_seen_ip": clientIP,
			"user_agent":   userAgent,
		}).Error
}

func (r *Repository) RevokeActivePasswordResetTokensByUserID(ctx context.Context, db *gorm.DB, userID uint64, activeStatus string, revokedStatus string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&UserPasswordResetToken{}).
		Where("user_id = ? AND status = ?", userID, activeStatus).
		Update("status", revokedStatus).Error
}

func (r *Repository) CreatePasswordResetToken(ctx context.Context, db *gorm.DB, token *UserPasswordResetToken) error {
	return r.queryDB(db).WithContext(ctx).Create(token).Error
}

func (r *Repository) RevokeActivePasswordResetTokenByHash(ctx context.Context, tokenHash string, activeStatus string, revokedStatus string) error {
	return r.db.WithContext(ctx).
		Model(&UserPasswordResetToken{}).
		Where("token_hash = ? AND status = ?", tokenHash, activeStatus).
		Update("status", revokedStatus).Error
}

func (r *Repository) FindPasswordResetTokenByHashForUpdate(ctx context.Context, db *gorm.DB, tokenHash string) (UserPasswordResetToken, error) {
	var token UserPasswordResetToken
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", tokenHash).
		First(&token).Error
	return token, err
}

func (r *Repository) UpdatePasswordResetTokenState(ctx context.Context, db *gorm.DB, id uint64, status string, usedAt *time.Time) error {
	updates := map[string]any{"status": status}
	if usedAt != nil {
		updates["used_at"] = *usedAt
	}
	return r.queryDB(db).WithContext(ctx).
		Model(&UserPasswordResetToken{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) applyUserListFilters(db *gorm.DB, filters UserListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", keyword, keyword, keyword)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	return db
}

func (r *Repository) applyUserSessionListFilters(db *gorm.DB, filters UserSessionListFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("sessions.user_id = ?", filters.UserID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("sessions.status = ?", strings.TrimSpace(filters.Status))
	}
	if filters.DateFrom != nil {
		db = db.Where("sessions.issued_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		db = db.Where("sessions.issued_at <= ?", *filters.DateTo)
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
