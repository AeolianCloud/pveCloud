package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	domainiam "github.com/AeolianCloud/pveCloud/server/internal/domain/iam"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqliam "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/iam"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	sharedcaptcha "github.com/AeolianCloud/pveCloud/server/internal/shared/captcha"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/shared/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
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
	cfg          config.JWTConfig
	iam          *mysqliam.Repository
	auditService *AdminAuditService
}

/**
 * AuthenticatedAdmin 表示管理端鉴权成功后的请求身份。
 */
type AuthenticatedAdmin struct {
	Claims          jwtpkg.Claims
	Admin           mysqliam.AdminUser
	Session         admindto.SessionSummary
	RoleIDs         []uint64
	PermissionCodes []string
	RequestContext  requestcontext.RequestContext
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
func NewAdminAuthService(db *gorm.DB, redis *cache.Redis, cfg config.JWTConfig, auditService *AdminAuditService) *AdminAuthService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &AdminAuthService{
		db:           db,
		redis:        redis,
		cfg:          cfg,
		iam:          mysqliam.NewRepository(db),
		auditService: auditService,
	}
}

/**
 * Captcha 生成管理员登录验证码并把答案写入 Redis 短 TTL。
 *
 * @param ctx 请求上下文
 * @return admin.LoginCaptchaResponse 验证码图片和标识
 * @return error 生成失败原因
 */
func (s *AdminAuthService) Captcha(ctx context.Context) (admindto.LoginCaptchaResponse, error) {
	request := requestcontext.RequestContextFrom(ctx)
	clientIP := request.IP
	if err := s.ensureCaptchaAllowed(ctx, clientIP); err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}

	code, err := sharedcaptcha.RandomCode(sharedcaptcha.DefaultCodeLength)
	if err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}
	captchaID, err := newAdminCaptchaID()
	if err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}

	key := s.loginCaptchaRedisKey(captchaID)
	if err := s.redis.Client().Set(ctx, key, sharedcaptcha.HashText(code), adminCaptchaTTLSeconds*time.Second).Err(); err != nil {
		return admindto.LoginCaptchaResponse{}, err
	}

	return admindto.LoginCaptchaResponse{
		CaptchaID: captchaID,
		Image:     sharedcaptcha.ImageDataURL(code),
		ExpiresIn: adminCaptchaTTLSeconds,
	}, nil
}

/**
 * Authenticate 校验管理端 JWT、会话、账号状态和当前 RBAC。
 */
func (s *AdminAuthService) Authenticate(ctx context.Context, tokenString string, clientIP string, userAgent string) (AuthenticatedAdmin, error) {
	claims, err := jwtpkg.Parse(tokenString, s.cfg.AdminSecret)
	if err != nil || claims.TokenType != "admin" || claims.AdminID == 0 || claims.Issuer != s.cfg.AdminIssuer || strings.TrimSpace(claims.ID) == "" {
		return AuthenticatedAdmin{}, apperrors.ErrUnauthorized
	}

	now := time.Now()
	session, err := s.iam.FindAdminSessionByID(ctx, claims.ID, claims.AdminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthenticatedAdmin{}, apperrors.ErrUnauthorized
	}
	if err != nil {
		return AuthenticatedAdmin{}, err
	}
	if session.Status != domainiam.SessionStatusActive || !session.ExpiresAt.After(now) {
		if session.Status == domainiam.SessionStatusActive && !session.ExpiresAt.After(now) {
			_ = s.iam.UpdateAdminSessionState(ctx, nil, session.ID, domainiam.SessionStatusExpired, now, domainiam.RevokeReasonExpired)
		}
		return AuthenticatedAdmin{}, apperrors.ErrUnauthorized
	}

	admin, err := s.iam.FindAdminByID(ctx, claims.AdminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthenticatedAdmin{}, apperrors.ErrUnauthorized
	}
	if err != nil {
		return AuthenticatedAdmin{}, err
	}
	if admin.Status != domainiam.AdminStatusActive {
		return AuthenticatedAdmin{}, apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
	}

	roleIDs, err := adminsupport.RoleIDs(ctx, s.db, admin.ID)
	if err != nil {
		return AuthenticatedAdmin{}, err
	}
	permissionCodes, err := adminsupport.PermissionCodes(ctx, s.db, admin.ID)
	if err != nil {
		return AuthenticatedAdmin{}, err
	}

	_ = s.iam.TouchAdminSession(ctx, session.ID, now, clientIP, textutil.TrimTo(userAgent, 500))

	return AuthenticatedAdmin{
		Claims:          *claims,
		Admin:           admin,
		Session:         adminsupport.SessionSummary(session),
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		RequestContext: requestcontext.RequestContext{
			AdminID:          &admin.ID,
			AdminUsername:    admin.Username,
			AdminDisplayName: admin.DisplayName,
			SessionID:        session.SessionID,
		},
	}, nil
}

/**
 * Login 校验管理员账号密码并签发管理端 JWT。
 *
 * @param ctx 请求上下文
 * @param req 登录请求
 * @return admin.LoginResponse 登录响应
 * @return error 登录失败原因
 */
func (s *AdminAuthService) Login(ctx context.Context, req admindto.LoginRequest) (admindto.LoginResponse, error) {
	request := requestcontext.RequestContextFrom(ctx)
	clientIP := request.IP
	userAgent := request.UserAgent
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

	admin, err := s.iam.FindAdminByAccount(ctx, identifier)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if recordErr := s.recordLoginFailure(ctx, nil, identifier, clientIP, userAgent, "账号或密码错误"); recordErr != nil {
			return admindto.LoginResponse{}, recordErr
		}
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("管理员账号或密码错误")
	}
	if err != nil {
		return admindto.LoginResponse{}, err
	}
	if admin.Status != domainiam.AdminStatusActive {
		if recordErr := s.recordLoginFailure(ctx, nil, identifier, clientIP, userAgent, "账号已禁用"); recordErr != nil {
			return admindto.LoginResponse{}, recordErr
		}
		return admindto.LoginResponse{}, apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
	}
	if !password.Verify(admin.PasswordHash, req.Password) {
		if recordErr := s.recordLoginFailure(ctx, nil, identifier, clientIP, userAgent, "账号或密码错误"); recordErr != nil {
			return admindto.LoginResponse{}, recordErr
		}
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("管理员账号或密码错误")
	}
	if err := s.clearLoginFailures(ctx, identifier, clientIP); err != nil {
		return admindto.LoginResponse{}, err
	}

	var result admindto.LoginResponse
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		issued, issueErr := s.issueSession(ctx, tx, admin, clientIP, userAgent)
		if issueErr != nil {
			return issueErr
		}
		result = issued

		now := time.Now()
		if err := s.iam.UpdateAdminLastLogin(ctx, tx, admin.ID, now, clientIP); err != nil {
			return err
		}

		return s.recordAudit(ctx, tx, &admin.ID, adminLoginSuccessAction, result.Session.SessionID, "登录成功")
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
func (s *AdminAuthService) Me(ctx context.Context, admin mysqliam.AdminUser, roleIDs []uint64, permissionCodes []string, session admindto.SessionSummary) (admindto.AuthStateResponse, error) {
	menus, err := adminsupport.VisibleAdminMenus(ctx, s.db, permissionCodes)
	if err != nil {
		return admindto.AuthStateResponse{}, err
	}
	return admindto.AuthStateResponse{
		Admin:           adminsupport.AdminSummary(admin),
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		Menus:           menus,
		Session:         session,
	}, nil
}

/**
 * Logout 吊销当前管理端会话。
 *
 * @param ctx 请求上下文
 * @param adminID 管理员 ID
 * @param sessionID 会话 ID
 * @return error 吊销失败原因
 */
func (s *AdminAuthService) Logout(ctx context.Context, adminID uint64, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return apperrors.ErrUnauthorized
	}

	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		now := time.Now()
		reason := adminRevokeReasonLogout
		if err := s.iam.RevokeActiveAdminSessionBySessionID(ctx, tx, sessionID, adminID, now, reason); err != nil {
			return err
		}
		return s.recordAudit(ctx, tx, &adminID, adminLogoutAction, sessionID, "退出登录")
	})
}

/**
 * Refresh 轮换当前管理端 token。
 *
 * @param ctx 请求上下文
 * @param adminID 管理员 ID
 * @param sessionID 旧会话 ID
 * @return admin.LoginResponse 新 token 响应
 * @return error 刷新失败原因
 */
func (s *AdminAuthService) Refresh(ctx context.Context, adminID uint64, sessionID string) (admindto.LoginResponse, error) {
	request := requestcontext.RequestContextFrom(ctx)
	clientIP := request.IP
	userAgent := request.UserAgent
	if strings.TrimSpace(sessionID) == "" {
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized
	}

	var result admindto.LoginResponse
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		oldSession, err := s.iam.FindActiveAdminSessionBySessionIDForUpdate(ctx, tx, sessionID, adminID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !oldSession.ExpiresAt.After(time.Now()) {
			return apperrors.ErrUnauthorized
		}

		admin, err := s.iam.FindAdminUserByID(ctx, tx, adminID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if admin.Status != domainiam.AdminStatusActive {
			return apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
		}

		now := time.Now()
		reason := adminRevokeReasonRefresh
		if err := s.iam.UpdateAdminSessionState(ctx, tx, oldSession.ID, adminSessionStatusRevoked, now, reason); err != nil {
			return err
		}

		issued, issueErr := s.issueSession(ctx, tx, admin, clientIP, userAgent)
		if issueErr != nil {
			return issueErr
		}
		result = issued
		return s.recordAudit(ctx, tx, &adminID, adminRefreshAction, result.Session.SessionID, "刷新登录会话")
	}); err != nil {
		return admindto.LoginResponse{}, err
	}

	return result, nil
}

func (s *AdminAuthService) issueSession(ctx context.Context, tx *gorm.DB, admin mysqliam.AdminUser, clientIP string, userAgent string) (admindto.LoginResponse, error) {
	roleIDs, err := adminsupport.RoleIDs(ctx, tx, admin.ID)
	if err != nil {
		return admindto.LoginResponse{}, err
	}
	permissionCodes, err := adminsupport.PermissionCodes(ctx, tx, admin.ID)
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
	session := mysqliam.AdminSession{
		SessionID:  sessionID,
		AdminID:    admin.ID,
		Status:     domainiam.SessionStatusActive,
		IssuedAt:   now,
		ExpiresAt:  expiresAt,
		LastSeenAt: &now,
		LastSeenIP: textutil.StringPtr(clientIP),
		UserAgent:  textutil.StringPtr(textutil.TrimTo(userAgent, 500)),
	}
	if err := s.iam.CreateAdminSession(ctx, tx, &session); err != nil {
		return admindto.LoginResponse{}, err
	}
	menus, err := adminsupport.VisibleAdminMenus(ctx, tx, permissionCodes)
	if err != nil {
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
		Admin:           adminsupport.AdminSummary(admin),
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		Menus:           menus,
		Session:         adminsupport.SessionSummary(session),
	}, nil
}

func (s *AdminAuthService) ensureLoginAllowed(ctx context.Context, identifier string, clientIP string) error {
	count, err := s.redis.Client().Get(ctx, s.loginFailureRedisKey(identifier, clientIP)).Int64()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return err
	}
	if count >= adminLoginFailureLimit {
		_ = s.recordAudit(
			ctx,
			s.db,
			nil,
			adminLoginLimitedAction,
			loginThrottleObjectID(identifier),
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
	return s.recordAudit(ctx, s.db, adminID, adminLoginFailureAction, loginThrottleObjectID(identifier), remark)
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
	key := s.redis.Key("admin", "login_captcha_rate", sharedcaptcha.HashText(clientIP))
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
		_ = s.recordAudit(
			ctx,
			s.db,
			nil,
			adminCaptchaLimitedAction,
			"captcha_"+sharedcaptcha.HashText(clientIP)[:32],
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
	if !strings.EqualFold(expected, sharedcaptcha.HashText(captchaCode)) {
		return apperrors.ErrValidation.WithMessage("验证码错误，请重新输入")
	}
	return nil
}

func (s *AdminAuthService) loginFailureRedisKey(identifier string, clientIP string) string {
	return s.redis.Key("admin", "login_fail", sharedcaptcha.HashText(clientIP), sharedcaptcha.HashText(identifier))
}

func (s *AdminAuthService) loginCaptchaRedisKey(captchaID string) string {
	return s.redis.Key("admin", "login_captcha", captchaID)
}

func (s *AdminAuthService) recordAudit(ctx context.Context, db *gorm.DB, adminID *uint64, action string, objectID string, remark string) error {
	return s.auditService.Record(ctx, db, AdminAuditWriteInput{
		AdminID:    adminID,
		Action:     action,
		ObjectType: adminAuditObjectAuth,
		ObjectID:   objectID,
		Remark:     remark,
	})
}

func newAdminSessionID() (string, error) {
	return sharedcaptcha.NewID("adm_")
}

func newAdminCaptchaID() (string, error) {
	return sharedcaptcha.NewID("adm_captcha_")
}

func loginThrottleObjectID(identifier string) string {
	return "login_" + sharedcaptcha.HashText(identifier)[:32]
}
