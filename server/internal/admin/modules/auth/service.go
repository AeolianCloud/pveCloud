package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/shared/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
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
	adminCaptchaLimitedAction  = "admin.captcha.limited"
	adminLoginLimitedAction    = "admin.login.limited"
	adminAuditObjectAuth       = "admin_auth"
	adminLoginFailureLimit     = int64(5)
	adminLoginFailureWindowMin = 15
	adminCaptchaTTLSeconds     = 120
	adminCaptchaRateLimit      = int64(30)
	adminCaptchaRateWindow     = time.Minute
)

/**
 * AdminAuthService 处理管理端登录、会话和 token 签发。
 */
type AdminAuthService struct {
	db           *gorm.DB
	redis        *cache.Redis
	cfg          bootstrap.JWTConfig
	auditService *AdminAuditService
}

/**
 * NewAdminAuthService 创建管理端认证服务。
 *
 * @param db 数据库连接
 * @param redis Redis 访问器
 * @param cfg JWT 配置
 * @param auditService 后台审计服务
 * @return *AdminAuthService 管理端认证服务
 */
func NewAdminAuthService(db *gorm.DB, redis *cache.Redis, cfg bootstrap.JWTConfig, auditService *AdminAuditService) *AdminAuthService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &AdminAuthService{db: db, redis: redis, cfg: cfg, auditService: auditService}
}

/**
 * Captcha 生成管理员登录验证码并把答案写入 Redis 短 TTL。
 *
 * @param ctx 请求上下文
 * @param clientIP 客户端 IP，用于限制验证码获取频率
 * @return admin.LoginCaptchaResponse 验证码图片和标识
 * @return error 生成失败原因
 */
func (s *AdminAuthService) Captcha(ctx context.Context, clientIP string) (admindto.LoginCaptchaResponse, error) {
	if err := s.ensureCaptchaAllowed(ctx, clientIP); err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}

	code, err := randomCaptchaCode(4)
	if err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}
	captchaID, err := newAdminCaptchaID()
	if err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}

	key := s.loginCaptchaRedisKey(captchaID)
	if err := s.redis.Client().Set(ctx, key, hashText(code), adminCaptchaTTLSeconds*time.Second).Err(); err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}

	return admindto.LoginCaptchaResponse{
		CaptchaID: captchaID,
		Image:     captchaImageDataURL(code),
		ExpiresIn: adminCaptchaTTLSeconds,
	}, nil
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
	if err := s.verifyLoginCaptcha(ctx, req.CaptchaID, req.CaptchaCode); err != nil {
		return admindto.LoginResponse{}, err
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

		return s.recordAudit(ctx, tx, &admin.ID, adminLoginSuccessAction, result.Session.SessionID, clientIP, userAgent, "登录成功")
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
		Admin:           support.AdminSummary(admin),
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		Menus:           support.VisibleAdminMenus(permissionCodes),
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
		return s.recordAudit(ctx, tx, &adminID, adminLogoutAction, sessionID, clientIP, userAgent, "退出登录")
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
		return s.recordAudit(ctx, tx, &adminID, adminRefreshAction, result.Session.SessionID, clientIP, userAgent, "刷新登录会话")
	}); err != nil {
		return admindto.LoginResponse{}, err
	}

	return result, nil
}

func (s *AdminAuthService) issueSession(ctx context.Context, tx *gorm.DB, admin models.AdminUser, clientIP string, userAgent string) (admindto.LoginResponse, error) {
	roleIDs, err := support.RoleIDs(ctx, tx, admin.ID)
	if err != nil {
		return admindto.LoginResponse{}, err
	}
	permissionCodes, err := support.PermissionCodes(ctx, tx, admin.ID)
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
		LastSeenIP: textutil.StringPtr(clientIP),
		UserAgent:  textutil.StringPtr(textutil.TrimTo(userAgent, 500)),
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
		Admin:           support.AdminSummary(admin),
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		Session:         support.SessionSummary(session),
	}, nil
}

func (s *AdminAuthService) ensureLoginAllowed(ctx context.Context, identifier string, clientIP string) error {
	count, err := s.redis.Client().Get(ctx, s.loginFailureRedisKey(identifier, clientIP)).Int64()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return err
	}
	if count >= adminLoginFailureLimit {
		_ = s.recordRiskAudit(
			ctx,
			s.db,
			nil,
			adminLoginLimitedAction,
			loginThrottleObjectID(identifier),
			"high",
			"登录失败次数达到限制",
			clientIP,
			"",
			"登录失败次数过多",
		)
		return apperrors.ErrTooManyRequests.WithMessage("登录失败次数过多，请 15 分钟后再试")
	}
	return nil
}

func (s *AdminAuthService) recordLoginFailure(ctx context.Context, adminID *uint64, identifier string, clientIP string, userAgent string, remark string) error {
	if err := s.increaseLoginFailures(ctx, identifier, clientIP); err != nil {
		return err
	}
	return s.recordAudit(ctx, s.db, adminID, adminLoginFailureAction, loginThrottleObjectID(identifier), clientIP, userAgent, remark)
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

func (s *AdminAuthService) ensureCaptchaAllowed(ctx context.Context, clientIP string) error {
	key := s.redis.Key("admin", "login_captcha_rate", hashText(clientIP))
	count, err := s.redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err := s.redis.Client().Expire(ctx, key, adminCaptchaRateWindow).Err(); err != nil {
			return err
		}
	}
	if count > adminCaptchaRateLimit {
		_ = s.recordRiskAudit(
			ctx,
			s.db,
			nil,
			adminCaptchaLimitedAction,
			"captcha_"+hashText(clientIP)[:32],
			"medium",
			"验证码获取频率达到限制",
			clientIP,
			"",
			"验证码获取过于频繁",
		)
		return apperrors.ErrTooManyRequests.WithMessage("验证码获取过于频繁，请稍后再试")
	}
	return nil
}

func (s *AdminAuthService) verifyLoginCaptcha(ctx context.Context, captchaID string, captchaCode string) error {
	captchaID = strings.TrimSpace(captchaID)
	captchaCode = strings.TrimSpace(captchaCode)
	if captchaID == "" || captchaCode == "" {
		return apperrors.ErrValidation.WithMessage("请输入验证码")
	}

	key := s.loginCaptchaRedisKey(captchaID)
	expected, err := s.redis.Client().GetDel(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return apperrors.ErrValidation.WithMessage("验证码已过期，请重新获取")
	}
	if err != nil {
		return err
	}
	if !strings.EqualFold(expected, hashText(captchaCode)) {
		return apperrors.ErrValidation.WithMessage("验证码错误，请重新输入")
	}
	return nil
}

func (s *AdminAuthService) loginFailureRedisKey(identifier string, clientIP string) string {
	return s.redis.Key("admin", "login_fail", hashText(clientIP), hashText(identifier))
}

func (s *AdminAuthService) loginCaptchaRedisKey(captchaID string) string {
	return s.redis.Key("admin", "login_captcha", captchaID)
}

func (s *AdminAuthService) recordAudit(ctx context.Context, db *gorm.DB, adminID *uint64, action string, objectID string, clientIP string, userAgent string, remark string) error {
	return s.auditService.Record(ctx, db, AdminAuditWriteInput{
		AdminID:    adminID,
		Action:     action,
		ObjectType: adminAuditObjectAuth,
		ObjectID:   objectID,
		IP:         clientIP,
		UserAgent:  userAgent,
		Remark:     remark,
	})
}

func (s *AdminAuthService) recordRiskAudit(ctx context.Context, db *gorm.DB, adminID *uint64, action string, objectID string, riskLevel string, riskReason string, clientIP string, userAgent string, remark string) error {
	return s.auditService.RecordRisk(ctx, db, AdminRiskWriteInput{
		AdminAuditWriteInput: AdminAuditWriteInput{
			AdminID:    adminID,
			Action:     action,
			ObjectType: adminAuditObjectAuth,
			ObjectID:   objectID,
			IP:         clientIP,
			UserAgent:  userAgent,
			Remark:     remark,
		},
		RiskLevel:  riskLevel,
		RiskReason: riskReason,
	})
}

func newAdminSessionID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "adm_" + hex.EncodeToString(bytes[:]), nil
}

func newAdminCaptchaID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "adm_captcha_" + hex.EncodeToString(bytes[:]), nil
}

func randomCaptchaCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var builder strings.Builder
	builder.Grow(length)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		builder.WriteByte(alphabet[index.Int64()])
	}
	return builder.String(), nil
}

func captchaImageDataURL(code string) string {
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="132" height="44" viewBox="0 0 132 44"><rect width="132" height="44" rx="8" fill="#f4f7ff"/><path d="M8 32 C30 10, 52 40, 76 18 S110 36, 124 12" fill="none" stroke="#9db6ff" stroke-width="2" opacity=".65"/><path d="M15 12 L118 35" stroke="#d2dcff" stroke-width="2" opacity=".75"/><text x="66" y="29" text-anchor="middle" font-family="Consolas, Menlo, monospace" font-size="24" font-weight="700" letter-spacing="5" fill="#2458d9">%s</text></svg>`,
		code,
	)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

func loginThrottleObjectID(identifier string) string {
	return "login_" + hashText(identifier)[:32]
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}
