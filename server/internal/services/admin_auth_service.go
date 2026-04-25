package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/models"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/cache"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/pkg/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/password"
)

const (
	adminStatusActive          = "active"
	adminSessionStatusActive   = "active"
	adminSessionStatusRevoked  = "revoked"
	adminRevokeReasonLogout    = "logout"
	adminRevokeReasonRefresh   = "refresh"
	adminLoginFailureAction    = "admin.login.failed"
	adminLoginSuccessAction    = "admin.login.success"
	adminLogoutAction          = "admin.logout"
	adminRefreshAction         = "admin.refresh"
	adminAuditObjectAuth       = "admin_auth"
	adminLoginFailureLimit     = int64(5)
	adminLoginFailureWindowMin = 15
)

/**
 * AdminAuthService 处理管理端登录、会话和 token 签发。
 */
type AdminAuthService struct {
	db    *gorm.DB
	redis *cache.Redis
	cfg   bootstrap.JWTConfig
}

/**
 * NewAdminAuthService 创建管理端认证服务。
 *
 * @param db 数据库连接
 * @param redis Redis 访问器
 * @param cfg JWT 配置
 * @return *AdminAuthService 管理端认证服务
 */
func NewAdminAuthService(db *gorm.DB, redis *cache.Redis, cfg bootstrap.JWTConfig) *AdminAuthService {
	return &AdminAuthService{db: db, redis: redis, cfg: cfg}
}

/**
 * Login 校验管理员账号密码并签发管理端 JWT。
 *
 * @param ctx 请求上下文
 * @param req 登录请求
 * @param clientIP 客户端 IP，用于记录最后登录来源
 * @param userAgent 浏览器 User-Agent
 * @return admin.LoginResponse 登录响应
 * @return error 登录失败原因
 */
func (s *AdminAuthService) Login(ctx context.Context, req admindto.LoginRequest, clientIP string, userAgent string) (admindto.LoginResponse, error) {
	identifier := strings.ToLower(strings.TrimSpace(req.Username))
	if identifier == "" {
		return admindto.LoginResponse{}, apperrors.ErrValidation.WithMessage("管理员账号不能为空")
	}
	if err := s.ensureLoginAllowed(ctx, identifier, clientIP); err != nil {
		return admindto.LoginResponse{}, err
	}

	var admin models.AdminUser
	err := s.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("username = ? OR email = ?", identifier, identifier).
		First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if recordErr := s.recordLoginFailure(ctx, nil, identifier, clientIP, userAgent, "账号或密码错误"); recordErr != nil {
			return admindto.LoginResponse{}, recordErr
		}
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("管理员账号或密码错误")
	}
	if err != nil {
		return admindto.LoginResponse{}, err
	}
	if admin.Status != adminStatusActive {
		if recordErr := s.recordLoginFailure(ctx, &admin.ID, identifier, clientIP, userAgent, "账号已禁用"); recordErr != nil {
			return admindto.LoginResponse{}, recordErr
		}
		return admindto.LoginResponse{}, apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
	}
	if !password.Verify(admin.PasswordHash, req.Password) {
		if recordErr := s.recordLoginFailure(ctx, &admin.ID, identifier, clientIP, userAgent, "账号或密码错误"); recordErr != nil {
			return admindto.LoginResponse{}, recordErr
		}
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("管理员账号或密码错误")
	}
	if err := s.clearLoginFailures(ctx, identifier, clientIP); err != nil {
		return admindto.LoginResponse{}, err
	}

	var result admindto.LoginResponse
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		issued, issueErr := s.issueSession(ctx, tx, admin, clientIP, userAgent)
		if issueErr != nil {
			return issueErr
		}
		result = issued

		now := time.Now()
		if err := tx.Model(&models.AdminUser{}).
			Where("id = ?", admin.ID).
			Updates(map[string]interface{}{
				"last_login_at": now,
				"last_login_ip": clientIP,
			}).Error; err != nil {
			return err
		}

		return s.createAudit(ctx, tx, &admin.ID, adminLoginSuccessAction, result.Session.SessionID, clientIP, userAgent, "登录成功")
	}); err != nil {
		return admindto.LoginResponse{}, err
	}

	return result, nil
}

/**
 * Me 返回当前管理端认证态。
 *
 * @param admin 当前管理员
 * @param roleIDs 当前角色 ID
 * @param permissionCodes 当前权限码
 * @param session 当前会话摘要
 * @return admin.AuthStateResponse 认证态响应
 */
func (s *AdminAuthService) Me(admin models.AdminUser, roleIDs []uint64, permissionCodes []string, session admindto.SessionSummary) admindto.AuthStateResponse {
	return admindto.AuthStateResponse{
		Admin:           adminSummary(admin),
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		Menus:           VisibleAdminMenus(permissionCodes),
		Session:         session,
	}
}

/**
 * Logout 吊销当前管理端会话。
 *
 * @param ctx 请求上下文
 * @param adminID 管理员 ID
 * @param sessionID 会话 ID
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return error 吊销失败原因
 */
func (s *AdminAuthService) Logout(ctx context.Context, adminID uint64, sessionID string, clientIP string, userAgent string) error {
	if strings.TrimSpace(sessionID) == "" {
		return apperrors.ErrUnauthorized
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		reason := adminRevokeReasonLogout
		if err := tx.Model(&models.AdminSession{}).
			Where("session_id = ? AND admin_id = ? AND status = ?", sessionID, adminID, adminSessionStatusActive).
			Updates(map[string]interface{}{
				"status":        adminSessionStatusRevoked,
				"revoked_at":    now,
				"revoke_reason": reason,
			}).Error; err != nil {
			return err
		}
		return s.createAudit(ctx, tx, &adminID, adminLogoutAction, sessionID, clientIP, userAgent, "退出登录")
	})
}

/**
 * Refresh 轮换当前管理端 token。
 *
 * @param ctx 请求上下文
 * @param adminID 管理员 ID
 * @param sessionID 旧会话 ID
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return admin.LoginResponse 新 token 响应
 * @return error 刷新失败原因
 */
func (s *AdminAuthService) Refresh(ctx context.Context, adminID uint64, sessionID string, clientIP string, userAgent string) (admindto.LoginResponse, error) {
	if strings.TrimSpace(sessionID) == "" {
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized
	}

	var result admindto.LoginResponse
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var oldSession models.AdminSession
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("session_id = ? AND admin_id = ? AND status = ?", sessionID, adminID, adminSessionStatusActive).
			First(&oldSession).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !oldSession.ExpiresAt.After(time.Now()) {
			return apperrors.ErrUnauthorized
		}

		var admin models.AdminUser
		err = tx.Where("deleted_at IS NULL").Where("id = ?", adminID).First(&admin).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if admin.Status != adminStatusActive {
			return apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
		}

		now := time.Now()
		reason := adminRevokeReasonRefresh
		if err := tx.Model(&models.AdminSession{}).
			Where("id = ?", oldSession.ID).
			Updates(map[string]interface{}{
				"status":        adminSessionStatusRevoked,
				"revoked_at":    now,
				"revoke_reason": reason,
			}).Error; err != nil {
			return err
		}

		issued, issueErr := s.issueSession(ctx, tx, admin, clientIP, userAgent)
		if issueErr != nil {
			return issueErr
		}
		result = issued
		return s.createAudit(ctx, tx, &adminID, adminRefreshAction, result.Session.SessionID, clientIP, userAgent, "刷新登录会话")
	}); err != nil {
		return admindto.LoginResponse{}, err
	}

	return result, nil
}

func (s *AdminAuthService) issueSession(ctx context.Context, tx *gorm.DB, admin models.AdminUser, clientIP string, userAgent string) (admindto.LoginResponse, error) {
	roleIDs, err := roleIDs(ctx, tx, admin.ID)
	if err != nil {
		return admindto.LoginResponse{}, err
	}
	permissionCodes, err := permissionCodes(ctx, tx, admin.ID)
	if err != nil {
		return admindto.LoginResponse{}, err
	}

	sessionID, err := newAdminSessionID()
	if err != nil {
		return admindto.LoginResponse{}, err
	}

	now := time.Now()
	ttl := time.Duration(s.cfg.AdminExpireMinutes) * time.Minute
	expiresAt := now.Add(ttl)
	session := models.AdminSession{
		SessionID:  sessionID,
		AdminID:    admin.ID,
		Status:     adminSessionStatusActive,
		IssuedAt:   now,
		ExpiresAt:  expiresAt,
		LastSeenAt: &now,
		LastSeenIP: stringPtr(clientIP),
		UserAgent:  stringPtr(trimTo(userAgent, 500)),
	}
	if err := tx.Create(&session).Error; err != nil {
		return admindto.LoginResponse{}, err
	}

	claims := jwtpkg.Claims{
		TokenType:       "admin",
		AdminID:         admin.ID,
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        sessionID,
			Issuer:    s.cfg.AdminIssuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwtpkg.Sign(claims, s.cfg.AdminSecret)
	if err != nil {
		return admindto.LoginResponse{}, err
	}

	return admindto.LoginResponse{
		AccessToken:     token,
		TokenType:       "Bearer",
		ExpiresIn:       int64(ttl.Seconds()),
		Admin:           adminSummary(admin),
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		Session:         sessionSummary(session),
	}, nil
}

func (s *AdminAuthService) ensureLoginAllowed(ctx context.Context, identifier string, clientIP string) error {
	count, err := s.redis.Client().Get(ctx, s.loginFailureRedisKey(identifier, clientIP)).Int64()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return err
	}
	if count >= adminLoginFailureLimit {
		return apperrors.ErrTooManyRequests.WithMessage("登录失败次数过多，请 15 分钟后再试")
	}
	return nil
}

func (s *AdminAuthService) recordLoginFailure(ctx context.Context, adminID *uint64, identifier string, clientIP string, userAgent string, remark string) error {
	if err := s.increaseLoginFailures(ctx, identifier, clientIP); err != nil {
		return err
	}
	return s.createAudit(ctx, s.db, adminID, adminLoginFailureAction, loginThrottleObjectID(identifier), clientIP, userAgent, remark)
}

func (s *AdminAuthService) increaseLoginFailures(ctx context.Context, identifier string, clientIP string) error {
	key := s.loginFailureRedisKey(identifier, clientIP)
	count, err := s.redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		return s.redis.Client().Expire(ctx, key, adminLoginFailureWindowMin*time.Minute).Err()
	}
	return nil
}

func (s *AdminAuthService) clearLoginFailures(ctx context.Context, identifier string, clientIP string) error {
	return s.redis.Client().Del(ctx, s.loginFailureRedisKey(identifier, clientIP)).Err()
}

func (s *AdminAuthService) loginFailureRedisKey(identifier string, clientIP string) string {
	return s.redis.Key("admin", "login_fail", hashText(clientIP), hashText(identifier))
}

func (s *AdminAuthService) createAudit(ctx context.Context, db *gorm.DB, adminID *uint64, action string, objectID string, clientIP string, userAgent string, remark string) error {
	log := models.AdminAuditLog{
		AdminID:    adminID,
		Action:     action,
		ObjectType: adminAuditObjectAuth,
		ObjectID:   stringPtr(objectID),
		IP:         stringPtr(clientIP),
		UserAgent:  stringPtr(trimTo(userAgent, 500)),
		Remark:     stringPtr(remark),
	}
	return db.WithContext(ctx).Create(&log).Error
}

func roleIDs(ctx context.Context, db *gorm.DB, adminID uint64) ([]uint64, error) {
	var roleIDs []uint64
	err := db.WithContext(ctx).
		Table("admin_user_roles").
		Select("admin_roles.id").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", adminStatusActive).
		Order("admin_roles.id ASC").
		Scan(&roleIDs).Error
	return roleIDs, err
}

func permissionCodes(ctx context.Context, db *gorm.DB, adminID uint64) ([]string, error) {
	var codes []string
	err := db.WithContext(ctx).
		Table("admin_user_roles").
		Distinct("admin_permissions.code").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Joins("JOIN admin_role_permissions ON admin_role_permissions.role_id = admin_roles.id").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", adminStatusActive).
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func adminSummary(admin models.AdminUser) admindto.AdminSummary {
	return admindto.AdminSummary{
		ID:          admin.ID,
		Username:    admin.Username,
		Email:       admin.Email,
		DisplayName: admin.DisplayName,
		Status:      admin.Status,
	}
}

func sessionSummary(session models.AdminSession) admindto.SessionSummary {
	return admindto.SessionSummary{
		SessionID: session.SessionID,
		IssuedAt:  session.IssuedAt,
		ExpiresAt: session.ExpiresAt,
	}
}

func newAdminSessionID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "adm_" + hex.EncodeToString(bytes[:]), nil
}

func loginThrottleObjectID(identifier string) string {
	return "login_" + hashText(identifier)[:32]
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func trimTo(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
