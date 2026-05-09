package webuser

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	mysqluser "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/user"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

const (
	webUserObjectType          = "web_user"
	webUserSessionObjectType   = "web_user_session"
	webUserCreateAction        = "web.user.create"
	webUserUpdateAction        = "web.user.update"
	webUserPasswordResetAction = "web.user.password_reset"
	webUserSessionRevokeAction = "web.user_session.revoke"
	webUserSessionRevokeReason = "admin_revoke"
	webUserPasswordResetReason = "admin_password_reset"
	webUserSessionExpireReason = "expired"
)

/**
 * WebUserService 处理用户端账号和会话管理。
 */
type WebUserService struct {
	db           *gorm.DB
	users        *mysqluser.Repository
	auditService *AdminAuditService
}

/**
 * NewWebUserService 创建用户端账号管理服务。
 */
func NewWebUserService(db *gorm.DB, auditService *AdminAuditService) *WebUserService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &WebUserService{db: db, users: mysqluser.NewRepository(db), auditService: auditService}
}

/**
 * Users 分页查询用户端账号。
 */
func (s *WebUserService) Users(ctx context.Context, query admindto.WebUserListQuery) (admindto.PageResponse[admindto.WebUserItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	users, total, err := s.users.Users(ctx, mysqluser.UserListFilters{
		Keyword: query.Keyword,
		Status:  query.Status,
	}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.WebUserItem]{}, err
	}
	items := make([]admindto.WebUserItem, 0, len(users))
	for _, user := range users {
		items = append(items, webUserItem(user))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
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

	var created mysqluser.User
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		created = mysqluser.User{
			Username:     username,
			Email:        email,
			PasswordHash: passwordHash,
			DisplayName:  textutil.NormalizeOptionalString(req.DisplayName),
			Status:       strings.TrimSpace(req.Status),
		}
		if err := s.users.CreateUser(ctx, tx, &created); err != nil {
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

	var updated mysqluser.User
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
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
		if err := s.users.UpdateUser(ctx, tx, id, updates); err != nil {
			return err
		}
		updated, err = s.users.FindUserByID(ctx, tx, id)
		if err != nil {
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
	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findUserForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.users.UpdateUserPasswordHash(ctx, tx, id, passwordHash); err != nil {
			return err
		}
		now := time.Now()
		if err := s.users.RevokeActiveUserSessionsByUserID(ctx, tx, id, now, webUserPasswordResetReason, "active", "revoked"); err != nil {
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
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	var dateFrom *time.Time
	if from, ok := parseDateTime(query.DateFrom); ok {
		dateFrom = &from
	}
	var dateTo *time.Time
	if to, ok := parseDateTime(query.DateTo); ok {
		dateTo = &to
	}
	rows, total, err := s.users.UserSessionRows(ctx, mysqluser.UserSessionListFilters{
		UserID:   query.UserID,
		Status:   query.Status,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.WebUserSessionItem]{}, err
	}
	items := make([]admindto.WebUserSessionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, webUserSessionItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
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
	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		session, err := s.users.FindUserSessionBySessionIDForUpdate(ctx, tx, targetSessionID)
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
		if err := s.users.UpdateUserSessionState(ctx, tx, session.ID, "revoked", now, webUserSessionRevokeReason); err != nil {
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

func (s *WebUserService) ensureUserUnique(ctx context.Context, excludeID uint64, username string, email string) error {
	count, err := s.users.CountUsersByIdentity(ctx, excludeID, username, email)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("用户账号或邮箱已存在")
	}
	return nil
}

func (s *WebUserService) findUser(ctx context.Context, db *gorm.DB, id uint64) (mysqluser.User, error) {
	user, err := s.users.FindUserByID(ctx, db, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqluser.User{}, apperrors.ErrNotFound.WithMessage("用户不存在")
	}
	return user, err
}

func (s *WebUserService) findUserForUpdate(ctx context.Context, db *gorm.DB, id uint64) (mysqluser.User, error) {
	user, err := s.users.FindUserByIDForUpdate(ctx, db, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqluser.User{}, apperrors.ErrNotFound.WithMessage("用户不存在")
	}
	return user, err
}

func (s *WebUserService) expireStaleSessions(ctx context.Context) error {
	now := time.Now()
	return s.users.ExpireStaleUserSessions(ctx, nil, now, "active", "expired", webUserSessionExpireReason)
}

func webUserItem(user mysqluser.User) admindto.WebUserItem {
	return admindto.WebUserItem{ID: user.ID, Username: user.Username, Email: user.Email, DisplayName: user.DisplayName, Status: user.Status, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
}

func webUserAuditSnapshot(user mysqluser.User) map[string]any {
	return map[string]any{"id": user.ID, "username": user.Username, "email": user.Email, "display_name": user.DisplayName, "status": user.Status}
}

func webUserSessionAuditSnapshot(session mysqluser.UserSession) map[string]any {
	return map[string]any{"session_id": session.SessionID, "user_id": session.UserID, "status": session.Status, "issued_at": session.IssuedAt, "expires_at": session.ExpiresAt, "revoked_at": session.RevokedAt, "revoke_reason": session.RevokeReason}
}

func webUserSessionItem(row mysqluser.UserSessionListRow) admindto.WebUserSessionItem {
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
