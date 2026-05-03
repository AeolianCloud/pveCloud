package webuser

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	webUserObjectType          = "web_user"
	webUserSessionObjectType   = "web_user_session"
	webUserCreateAction        = "web.user.create"
	webUserUpdateAction        = "web.user.update"
	webUserPasswordResetAction = "web.user.password_reset"
	webUserSessionRevokeAction = "web.user_session.revoke"
	webUserSessionRevokeReason = "admin_revoke"
	webUserSessionExpireReason = "expired"
)

/**
 * WebUserService 处理用户端账号和会话管理。
 */
type WebUserService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

/**
 * NewWebUserService 创建用户端账号管理服务。
 */
func NewWebUserService(db *gorm.DB, auditService *AdminAuditService) *WebUserService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &WebUserService{db: db, auditService: auditService}
}

/**
 * Users 分页查询用户端账号。
 */
func (s *WebUserService) Users(ctx context.Context, query admindto.WebUserListQuery) (admindto.PageResponse[admindto.WebUserItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Model(&models.User{})
	db = applyWebUserFilters(db, query)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.WebUserItem]{}, err
	}

	var users []models.User
	if err := db.Order("id DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&users).Error; err != nil {
		return admindto.PageResponse[admindto.WebUserItem]{}, err
	}
	items := make([]admindto.WebUserItem, 0, len(users))
	for _, user := range users {
		items = append(items, webUserItem(user))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

/**
 * CreateUser 创建用户端账号。
 */
func (s *WebUserService) CreateUser(ctx context.Context, operatorID uint64, req admindto.WebUserCreateRequest) (admindto.WebUserItem, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	if err := s.ensureUserUnique(ctx, 0, username, email); err != nil {
		return admindto.WebUserItem{}, err
	}
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return admindto.WebUserItem{}, err
	}

	var created models.User
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		created = models.User{
			Username:     username,
			Email:        email,
			PasswordHash: passwordHash,
			DisplayName:  textutil.NormalizeOptionalString(req.DisplayName),
			Status:       strings.TrimSpace(req.Status),
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     webUserCreateAction,
			ObjectType: webUserObjectType,
			ObjectID:   textutil.Uint64String(created.ID),
			AfterData:  webUserAuditSnapshot(created),
			Remark:     "创建 Web 用户",
		})
	}); err != nil {
		return admindto.WebUserItem{}, err
	}
	return webUserItem(created), nil
}

/**
 * UserDetail 查询用户端账号详情。
 */
func (s *WebUserService) UserDetail(ctx context.Context, id uint64) (admindto.WebUserItem, error) {
	user, err := s.findUser(ctx, s.db, id)
	if err != nil {
		return admindto.WebUserItem{}, err
	}
	return webUserItem(user), nil
}

/**
 * UpdateUser 更新用户端账号。
 */
func (s *WebUserService) UpdateUser(ctx context.Context, operatorID uint64, id uint64, req admindto.WebUserUpdateRequest) (admindto.WebUserItem, error) {
	var email string
	if req.Email != nil {
		email = strings.TrimSpace(*req.Email)
		if err := s.ensureUserUnique(ctx, id, "", email); err != nil {
			return admindto.WebUserItem{}, err
		}
	}

	var updated models.User
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := s.findUserForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		updates := map[string]any{}
		if req.Email != nil {
			updates["email"] = email
		}
		if req.DisplayName != nil {
			updates["display_name"] = textutil.NormalizeOptionalString(req.DisplayName)
		}
		if req.Status != nil {
			updates["status"] = strings.TrimSpace(*req.Status)
		}
		if len(updates) > 0 {
			if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     webUserUpdateAction,
			ObjectType: webUserObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: webUserAuditSnapshot(current),
			AfterData:  webUserAuditSnapshot(updated),
			Remark:     "更新 Web 用户",
		})
	}); err != nil {
		return admindto.WebUserItem{}, err
	}
	return webUserItem(updated), nil
}

/**
 * ResetPassword 重置用户端账号密码。
 */
func (s *WebUserService) ResetPassword(ctx context.Context, operatorID uint64, id uint64, req admindto.WebUserPasswordRequest) error {
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := s.findUserForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     webUserPasswordResetAction,
			ObjectType: webUserObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: map[string]any{"id": current.ID, "username": current.Username},
			AfterData:  map[string]any{"id": current.ID, "username": current.Username, "password_reset": true},
			Remark:     "重置 Web 用户密码",
		})
	})
}

/**
 * Sessions 分页查询用户端登录会话。
 */
func (s *WebUserService) Sessions(ctx context.Context, query admindto.WebUserSessionListQuery) (admindto.PageResponse[admindto.WebUserSessionItem], error) {
	if err := s.expireStaleSessions(ctx); err != nil {
		return admindto.PageResponse[admindto.WebUserSessionItem]{}, err
	}
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Table("user_sessions AS sessions").Joins("JOIN users ON users.id = sessions.user_id")
	db = applyWebUserSessionFilters(db, query)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.WebUserSessionItem]{}, err
	}

	var rows []webUserSessionListRow
	err := db.Select(`
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
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error
	if err != nil {
		return admindto.PageResponse[admindto.WebUserSessionItem]{}, err
	}
	items := make([]admindto.WebUserSessionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, webUserSessionItem(row))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

/**
 * RevokeSession 吊销用户端登录会话。
 */
func (s *WebUserService) RevokeSession(ctx context.Context, operatorID uint64, targetSessionID string) error {
	targetSessionID = strings.TrimSpace(targetSessionID)
	if targetSessionID == "" {
		return apperrors.ErrValidation.WithMessage("会话 ID 不能为空")
	}
	if err := s.expireStaleSessions(ctx); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session models.UserSession
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("session_id = ?", targetSessionID).First(&session).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("用户会话不存在")
		}
		if err != nil {
			return err
		}
		if session.Status != "active" {
			return apperrors.ErrConflict.WithMessage("会话已不是活跃状态")
		}
		now := time.Now()
		before := webUserSessionAuditSnapshot(session)
		after := session
		after.Status = "revoked"
		after.RevokedAt = &now
		after.RevokeReason = textutil.StringPtr(webUserSessionRevokeReason)
		if err := tx.Model(&models.UserSession{}).Where("id = ?", session.ID).Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    now,
			"revoke_reason": webUserSessionRevokeReason,
		}).Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     webUserSessionRevokeAction,
			ObjectType: webUserSessionObjectType,
			ObjectID:   session.SessionID,
			BeforeData: before,
			AfterData:  webUserSessionAuditSnapshot(after),
			Remark:     "吊销 Web 用户会话",
		})
	})
}

func applyWebUserFilters(db *gorm.DB, query admindto.WebUserListQuery) *gorm.DB {
	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", keyword, keyword, keyword)
	}
	if query.Status != "" {
		db = db.Where("status = ?", strings.TrimSpace(query.Status))
	}
	return db
}

func applyWebUserSessionFilters(db *gorm.DB, query admindto.WebUserSessionListQuery) *gorm.DB {
	if query.UserID > 0 {
		db = db.Where("sessions.user_id = ?", query.UserID)
	}
	if query.Status != "" {
		db = db.Where("sessions.status = ?", strings.TrimSpace(query.Status))
	}
	if from, ok := parseDateTime(query.DateFrom); ok {
		db = db.Where("sessions.issued_at >= ?", from)
	}
	if to, ok := parseDateTime(query.DateTo); ok {
		db = db.Where("sessions.issued_at <= ?", to)
	}
	return db
}

func (s *WebUserService) ensureUserUnique(ctx context.Context, excludeID uint64, username string, email string) error {
	query := s.db.WithContext(ctx).Model(&models.User{})
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if username != "" && email != "" {
		query = query.Where("username = ? OR email = ?", strings.TrimSpace(username), strings.TrimSpace(email))
	} else if username != "" {
		query = query.Where("username = ?", strings.TrimSpace(username))
	} else if email != "" {
		query = query.Where("email = ?", strings.TrimSpace(email))
	} else {
		return nil
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("用户账号或邮箱已存在")
	}
	return nil
}

func (s *WebUserService) findUser(ctx context.Context, db *gorm.DB, id uint64) (models.User, error) {
	var user models.User
	err := db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, apperrors.ErrNotFound.WithMessage("用户不存在")
	}
	return user, err
}

func (s *WebUserService) findUserForUpdate(ctx context.Context, db *gorm.DB, id uint64) (models.User, error) {
	var user models.User
	err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, apperrors.ErrNotFound.WithMessage("用户不存在")
	}
	return user, err
}

func (s *WebUserService) expireStaleSessions(ctx context.Context) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("status = ? AND expires_at <= ?", "active", now).
		Updates(map[string]any{"status": "expired", "revoked_at": now, "revoke_reason": webUserSessionExpireReason}).Error
}

func webUserItem(user models.User) admindto.WebUserItem {
	return admindto.WebUserItem{ID: user.ID, Username: user.Username, Email: user.Email, DisplayName: user.DisplayName, Status: user.Status, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
}

func webUserAuditSnapshot(user models.User) map[string]any {
	return map[string]any{"id": user.ID, "username": user.Username, "email": user.Email, "display_name": user.DisplayName, "status": user.Status}
}

func webUserSessionAuditSnapshot(session models.UserSession) map[string]any {
	return map[string]any{"session_id": session.SessionID, "user_id": session.UserID, "status": session.Status, "issued_at": session.IssuedAt, "expires_at": session.ExpiresAt, "revoked_at": session.RevokedAt, "revoke_reason": session.RevokeReason}
}

func webUserSessionItem(row webUserSessionListRow) admindto.WebUserSessionItem {
	return admindto.WebUserSessionItem{
		SessionID: row.SessionID,
		User:      admindto.WebUserItem{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName, Status: row.UserStatus, CreatedAt: row.UserCreatedAt, UpdatedAt: row.UserUpdatedAt},
		Status:    row.Status, IssuedAt: row.IssuedAt, ExpiresAt: row.ExpiresAt, LastSeenAt: row.LastSeenAt, LastSeenIP: row.LastSeenIP, UserAgent: row.UserAgent, RevokedAt: row.RevokedAt, RevokeReason: row.RevokeReason, CreatedAt: row.CreatedAt,
	}
}

func parseDateTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

type webUserSessionListRow struct {
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
