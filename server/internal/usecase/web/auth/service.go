package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	gojwt "github.com/golang-jwt/jwt/v5"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	domainuser "github.com/AeolianCloud/pveCloud/server/internal/domain/user"
	"github.com/AeolianCloud/pveCloud/server/internal/integration/mail"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlsystemconfig "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/systemconfig"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	mysqluser "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/user"
	sharedcaptcha "github.com/AeolianCloud/pveCloud/server/internal/shared/captcha"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/shared/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	websupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/support"
)

const (
	userStatusActive          = "active"
	userSessionStatusActive   = "active"
	userSessionStatusRevoked  = "revoked"
	userResetStatusActive     = "active"
	userResetStatusUsed       = "used"
	userResetStatusRevoked    = "revoked"
	revokeReasonLogout        = "logout"
	revokeReasonRefresh       = "refresh"
	revokeReasonPasswordReset = "password_reset"
	passwordResetTTL          = 30 * time.Minute
	userCaptchaRateLimit      = int64(30)
	userCaptchaRateWindow     = time.Minute
	userLoginFailureLimit     = int64(5)
	userLoginFailureWindow    = 15 * time.Minute
	userPasswordResetLimit    = int64(3)
	userPasswordResetWindow   = 15 * time.Minute
)

type userCaptchaScene struct {
	configKey       string
	redisKeySegment string
	rateKeySegment  string
	disabledMessage string
}

var (
	loginCaptchaScene = userCaptchaScene{
		configKey:       "web.auth.login_captcha_enabled",
		redisKeySegment: "login_captcha",
		rateKeySegment:  "login_captcha_rate",
		disabledMessage: "登录验证码未开启",
	}
	registerCaptchaScene = userCaptchaScene{
		configKey:       "web.auth.register_captcha_enabled",
		redisKeySegment: "register_captcha",
		rateKeySegment:  "register_captcha_rate",
		disabledMessage: "注册验证码未开启",
	}
	passwordResetRequestCaptchaScene = userCaptchaScene{
		configKey:       "web.auth.password_reset_request_captcha_enabled",
		redisKeySegment: "password_reset_request_captcha",
		rateKeySegment:  "password_reset_request_captcha_rate",
		disabledMessage: "忘记密码验证码未开启",
	}
	passwordResetConfirmCaptchaScene = userCaptchaScene{
		configKey:       "web.auth.password_reset_confirm_captcha_enabled",
		redisKeySegment: "password_reset_confirm_captcha",
		rateKeySegment:  "password_reset_confirm_captcha_rate",
		disabledMessage: "重置密码验证码未开启",
	}
)

/**
 * UserAuthService 处理用户端登录、会话和 token 签发。
 */
type UserAuthService struct {
	db      *gorm.DB
	redis   *cache.Redis
	cfg     config.JWTConfig
	mail    *mail.Sender
	webURL  string
	configs *mysqlsystemconfig.Repository
	users   *mysqluser.Repository
}

/**
 * AuthenticatedUser 表示用户端鉴权成功后的请求身份。
 */
type AuthenticatedUser struct {
	Claims  jwtpkg.Claims
	User    mysqluser.User
	Session webdto.SessionSummary
}

/**
 * NewUserAuthService 创建用户端认证服务。
 */
func NewUserAuthService(db *gorm.DB, redis *cache.Redis, cfg config.JWTConfig, mailCfg config.MailConfig) *UserAuthService {
	return &UserAuthService{
		db:      db,
		redis:   redis,
		cfg:     cfg,
		mail:    mail.NewSender(mailCfg),
		webURL:  mailCfg.PasswordResetURLBase,
		configs: mysqlsystemconfig.NewRepository(db),
		users:   mysqluser.NewRepository(db),
	}
}

/**
 * LoginCaptcha 生成登录验证码。
 */
func (s *UserAuthService) LoginCaptcha(ctx context.Context) (webdto.CaptchaResponse, error) {
	return s.generateCaptcha(ctx, loginCaptchaScene)
}

/**
 * RegisterCaptcha 生成注册验证码。
 */
func (s *UserAuthService) RegisterCaptcha(ctx context.Context) (webdto.CaptchaResponse, error) {
	return s.generateCaptcha(ctx, registerCaptchaScene)
}

/**
 * PasswordResetRequestCaptcha 生成忘记密码申请验证码。
 */
func (s *UserAuthService) PasswordResetRequestCaptcha(ctx context.Context) (webdto.CaptchaResponse, error) {
	return s.generateCaptcha(ctx, passwordResetRequestCaptchaScene)
}

/**
 * PasswordResetConfirmCaptcha 生成重置密码确认验证码。
 */
func (s *UserAuthService) PasswordResetConfirmCaptcha(ctx context.Context) (webdto.CaptchaResponse, error) {
	return s.generateCaptcha(ctx, passwordResetConfirmCaptchaScene)
}

/**
 * Authenticate 校验用户端 JWT、会话和用户状态。
 */
func (s *UserAuthService) Authenticate(ctx context.Context, tokenString string, clientIP string, userAgent string) (AuthenticatedUser, error) {
	claims, err := jwtpkg.Parse(tokenString, s.cfg.UserSecret)
	if err != nil || claims.TokenType != "user" || claims.UserID == 0 || claims.Issuer != s.cfg.UserIssuer || strings.TrimSpace(claims.ID) == "" {
		return AuthenticatedUser{}, apperrors.ErrUnauthorized
	}

	now := time.Now()
	session, err := s.users.FindUserSessionBySessionID(ctx, claims.ID, claims.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthenticatedUser{}, apperrors.ErrUnauthorized
	}
	if err != nil {
		return AuthenticatedUser{}, err
	}
	if !domainuser.IsSessionActiveAt(session.Status, session.ExpiresAt, now) {
		if domainuser.ShouldExpireSession(session.Status, session.ExpiresAt, now) {
			_ = s.users.UpdateUserSessionState(ctx, nil, session.ID, domainuser.SessionStatusExpired, now, "expired")
		}
		return AuthenticatedUser{}, apperrors.ErrUnauthorized
	}

	user, err := s.users.FindUserByID(ctx, nil, claims.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthenticatedUser{}, apperrors.ErrUnauthorized
	}
	if err != nil {
		return AuthenticatedUser{}, err
	}
	if !domainuser.IsActive(user.Status) {
		return AuthenticatedUser{}, apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
	}

	_ = s.users.TouchUserSession(ctx, session.ID, now, clientIP, textutil.TrimTo(userAgent, 500))

	return AuthenticatedUser{
		Claims:  *claims,
		User:    user,
		Session: websupport.SessionSummary(session),
	}, nil
}

/**
 * Login 校验用户账号密码并签发用户端 JWT。
 */
func (s *UserAuthService) Login(ctx context.Context, req webdto.LoginRequest) (webdto.LoginResponse, error) {
	request := requestcontext.RequestContextFrom(ctx)
	account := strings.ToLower(strings.TrimSpace(req.Account))
	if account == "" {
		return webdto.LoginResponse{}, apperrors.ErrValidation.WithMessage("账号不能为空")
	}
	if err := s.verifyCaptchaIfEnabled(ctx, loginCaptchaScene, req.CaptchaID, req.CaptchaCode); err != nil {
		return webdto.LoginResponse{}, err
	}
	if err := s.ensureLoginAllowed(ctx, request.IP, account); err != nil {
		return webdto.LoginResponse{}, err
	}

	user, err := s.users.FindUserByAccount(ctx, account)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if recordErr := s.recordLoginFailure(ctx, request.IP, account); recordErr != nil {
			return webdto.LoginResponse{}, recordErr
		}
		return webdto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("账号或密码错误")
	}
	if err != nil {
		return webdto.LoginResponse{}, err
	}
	if !domainuser.IsActive(user.Status) {
		return webdto.LoginResponse{}, apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
	}
	if !password.Verify(user.PasswordHash, req.Password) {
		if recordErr := s.recordLoginFailure(ctx, request.IP, account); recordErr != nil {
			return webdto.LoginResponse{}, recordErr
		}
		return webdto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("账号或密码错误")
	}
	if err := s.clearLoginFailures(ctx, request.IP, account); err != nil {
		return webdto.LoginResponse{}, err
	}

	var result webdto.LoginResponse
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		issued, issueErr := s.issueSession(ctx, tx, user, request.IP, request.UserAgent)
		if issueErr != nil {
			return issueErr
		}
		result = issued
		return nil
	}); err != nil {
		return webdto.LoginResponse{}, err
	}
	return result, nil
}

/**
 * Register 创建用户端账号并签发登录会话。
 */
func (s *UserAuthService) Register(ctx context.Context, req webdto.RegisterRequest) (webdto.LoginResponse, error) {
	request := requestcontext.RequestContextFrom(ctx)
	username := strings.ToLower(strings.TrimSpace(req.Username))
	email := strings.ToLower(strings.TrimSpace(req.Email))
	displayName := trimOptional(req.DisplayName)
	if username == "" || email == "" {
		return webdto.LoginResponse{}, apperrors.ErrValidation.WithMessage("用户名和邮箱不能为空")
	}
	if err := s.verifyCaptchaIfEnabled(ctx, registerCaptchaScene, req.CaptchaID, req.CaptchaCode); err != nil {
		return webdto.LoginResponse{}, err
	}

	var result webdto.LoginResponse
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		hash, err := password.Hash(req.Password)
		if err != nil {
			return err
		}
		user := mysqluser.User{
			Username:     username,
			Email:        email,
			PasswordHash: hash,
			DisplayName:  displayName,
			Status:       domainuser.StatusActive,
		}
		if err := s.users.CreateUser(ctx, tx, &user); err != nil {
			if isDuplicateEntry(err) {
				return apperrors.ErrConflict.WithMessage("用户名或邮箱已被使用")
			}
			return err
		}

		issued, issueErr := s.issueSession(ctx, tx, user, request.IP, request.UserAgent)
		if issueErr != nil {
			return issueErr
		}
		result = issued
		return nil
	}); err != nil {
		return webdto.LoginResponse{}, err
	}
	return result, nil
}

/**
 * Me 返回当前用户端认证态。
 */
func (s *UserAuthService) Me(user mysqluser.User, session webdto.SessionSummary) webdto.AuthStateResponse {
	return webdto.AuthStateResponse{User: websupport.UserSummary(user), Session: session}
}

/**
 * Logout 吊销当前用户端会话。
 */
func (s *UserAuthService) Logout(ctx context.Context, userID uint64, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return apperrors.ErrUnauthorized
	}
	now := time.Now()
	reason := revokeReasonLogout
	rows, err := s.users.RevokeActiveUserSessionBySessionID(ctx, nil, sessionID, userID, now, reason, domainuser.SessionStatusActive, domainuser.SessionStatusRevoked)
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperrors.ErrUnauthorized.WithMessage("会话已失效")
	}
	return nil
}

/**
 * Refresh 轮换当前用户端 token。
 */
func (s *UserAuthService) Refresh(ctx context.Context, userID uint64, sessionID string) (webdto.LoginResponse, error) {
	request := requestcontext.RequestContextFrom(ctx)
	if strings.TrimSpace(sessionID) == "" {
		return webdto.LoginResponse{}, apperrors.ErrUnauthorized
	}

	var result webdto.LoginResponse
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		oldSession, err := s.users.FindActiveUserSessionForUpdate(ctx, tx, sessionID, userID, domainuser.SessionStatusActive)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !domainuser.IsSessionActiveAt(oldSession.Status, oldSession.ExpiresAt, time.Now()) {
			return apperrors.ErrUnauthorized
		}

		user, err := s.users.FindUserByID(ctx, tx, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !domainuser.IsActive(user.Status) {
			return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
		}

		now := time.Now()
		reason := revokeReasonRefresh
		if err := s.users.UpdateUserSessionState(ctx, tx, oldSession.ID, domainuser.SessionStatusRevoked, now, reason); err != nil {
			return err
		}

		issued, issueErr := s.issueSession(ctx, tx, user, request.IP, request.UserAgent)
		if issueErr != nil {
			return issueErr
		}
		result = issued
		return nil
	}); err != nil {
		return webdto.LoginResponse{}, err
	}
	return result, nil
}

/**
 * RequestPasswordReset 创建一次性密码重置 token 并发送邮件。
 */
func (s *UserAuthService) RequestPasswordReset(ctx context.Context, req webdto.PasswordResetRequest) error {
	if err := s.verifyCaptchaIfEnabled(ctx, passwordResetRequestCaptchaScene, req.CaptchaID, req.CaptchaCode); err != nil {
		return err
	}
	if s.mail == nil || !s.mail.Enabled() {
		return apperrors.ErrInternal.WithMessage("密码找回服务暂不可用，请稍后再试")
	}

	request := requestcontext.RequestContextFrom(ctx)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := s.ensurePasswordResetAllowed(ctx, request.IP, email); err != nil {
		return err
	}

	user, err := s.users.FindUserByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !domainuser.IsActive(user.Status) {
		return nil
	}

	token, tokenHash, err := newPasswordResetToken()
	if err != nil {
		return err
	}
	now := time.Now()
	reset := mysqluser.UserPasswordResetToken{
		UserID:      user.ID,
		TokenHash:   tokenHash,
		Status:      domainuser.PasswordResetStatusActive,
		ExpiresAt:   now.Add(passwordResetTTL),
		RequestedIP: textutil.StringPtr(request.IP),
		UserAgent:   textutil.StringPtr(textutil.TrimTo(request.UserAgent, 500)),
	}
	shouldSend := false
	sendEmail := user.Email
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		lockedUser, err := s.users.FindUserByIDForUpdate(ctx, tx, user.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if !domainuser.IsActive(lockedUser.Status) {
			return nil
		}
		if err := s.users.RevokeActivePasswordResetTokensByUserID(ctx, tx, lockedUser.ID, domainuser.PasswordResetStatusActive, domainuser.PasswordResetStatusRevoked); err != nil {
			return err
		}
		if err := s.users.CreatePasswordResetToken(ctx, tx, &reset); err != nil {
			return err
		}
		sendEmail = lockedUser.Email
		shouldSend = true
		return nil
	}); err != nil {
		return err
	}
	if !shouldSend {
		return nil
	}

	if err := s.mail.SendPasswordReset(sendEmail, s.resetURL(token)); err != nil {
		_ = s.users.RevokeActivePasswordResetTokenByHash(ctx, tokenHash, domainuser.PasswordResetStatusActive, domainuser.PasswordResetStatusRevoked)
		return err
	}
	return nil
}

/**
 * ConfirmPasswordReset 使用一次性 token 重置密码。
 */
func (s *UserAuthService) ConfirmPasswordReset(ctx context.Context, req webdto.PasswordResetConfirmRequest) error {
	if err := s.verifyCaptchaIfEnabled(ctx, passwordResetConfirmCaptchaScene, req.CaptchaID, req.CaptchaCode); err != nil {
		return err
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return apperrors.ErrValidation.WithMessage("密码重置 token 不能为空")
	}
	tokenHash := hashPasswordResetToken(token)
	now := time.Now()

	var userDisabled bool
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		reset, err := s.users.FindPasswordResetTokenByHashForUpdate(ctx, tx, tokenHash)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrConflict.WithMessage("密码重置链接无效或已过期")
		}
		if err != nil {
			return err
		}
		if !domainuser.IsPasswordResetTokenUsable(reset.Status, reset.ExpiresAt, now) {
			return apperrors.ErrConflict.WithMessage("密码重置链接无效或已过期")
		}

		user, err := s.users.FindUserByIDForUpdate(ctx, tx, reset.UserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrConflict.WithMessage("密码重置链接无效或已过期")
		}
		if err != nil {
			return err
		}
		if !domainuser.IsActive(user.Status) {
			if err := s.users.UpdatePasswordResetTokenState(ctx, tx, reset.ID, domainuser.PasswordResetStatusRevoked, nil); err != nil {
				return err
			}
			userDisabled = true
			return nil
		}

		hash, err := password.Hash(req.Password)
		if err != nil {
			return err
		}
		if err := s.users.UpdateUserPasswordHash(ctx, tx, user.ID, hash); err != nil {
			return err
		}
		if err := s.users.UpdatePasswordResetTokenState(ctx, tx, reset.ID, domainuser.PasswordResetStatusUsed, &now); err != nil {
			return err
		}
		reason := revokeReasonPasswordReset
		return s.users.RevokeActiveUserSessionsByUserID(ctx, tx, reset.UserID, now, reason, domainuser.SessionStatusActive, domainuser.SessionStatusRevoked)
	}); err != nil {
		return err
	}
	if userDisabled {
		return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
	}
	return nil
}

func (s *UserAuthService) issueSession(ctx context.Context, tx *gorm.DB, user mysqluser.User, clientIP string, userAgent string) (webdto.LoginResponse, error) {
	sessionID, err := newUserSessionID()
	if err != nil {
		return webdto.LoginResponse{}, err
	}
	now := time.Now()
	ttl := time.Duration(s.cfg.UserExpireMinutes) * time.Minute
	expiresAt := now.Add(ttl)
	session := mysqluser.UserSession{
		SessionID:  sessionID,
		UserID:     user.ID,
		Status:     domainuser.SessionStatusActive,
		IssuedAt:   now,
		ExpiresAt:  expiresAt,
		LastSeenAt: &now,
		LastSeenIP: textutil.StringPtr(clientIP),
		UserAgent:  textutil.StringPtr(textutil.TrimTo(userAgent, 500)),
	}
	if err := s.users.CreateUserSession(ctx, tx, &session); err != nil {
		return webdto.LoginResponse{}, err
	}

	claims := jwtpkg.Claims{
		TokenType: "user",
		UserID:    user.ID,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        sessionID,
			Issuer:    s.cfg.UserIssuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwtpkg.Sign(claims, s.cfg.UserSecret)
	if err != nil {
		return webdto.LoginResponse{}, err
	}

	return webdto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(ttl.Seconds()),
		User:        websupport.UserSummary(user),
		Session:     websupport.SessionSummary(session),
	}, nil
}

func newUserSessionID() (string, error) {
	return sharedcaptcha.NewID("usr_")
}

func newPasswordResetToken() (string, string, error) {
	token, err := sharedcaptcha.RandomHex(32)
	if err != nil {
		return "", "", err
	}
	return token, hashPasswordResetToken(token), nil
}

func hashPasswordResetToken(token string) string {
	return sharedcaptcha.HashText(token)
}

func (s *UserAuthService) resetURL(token string) string {
	base := strings.TrimSpace(s.webURL)
	if base == "" {
		base = "http://localhost:5174/reset-password"
	}
	separator := "?"
	if strings.Contains(base, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%stoken=%s", base, separator, url.QueryEscape(token))
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (s *UserAuthService) generateCaptcha(ctx context.Context, scene userCaptchaScene) (webdto.CaptchaResponse, error) {
	enabled, err := s.isCaptchaEnabled(ctx, scene.configKey)
	if err != nil {
		return webdto.CaptchaResponse{}, err
	}
	if !enabled {
		return webdto.CaptchaResponse{}, apperrors.ErrForbidden.WithMessage(scene.disabledMessage)
	}
	if s.redis == nil {
		return webdto.CaptchaResponse{}, apperrors.ErrInternal.WithMessage("验证码服务暂不可用，请稍后再试")
	}

	request := requestcontext.RequestContextFrom(ctx)
	if err := s.ensureCaptchaAllowed(ctx, scene, request.IP); err != nil {
		return webdto.CaptchaResponse{}, err
	}

	code, err := sharedcaptcha.RandomCode(sharedcaptcha.DefaultCodeLength)
	if err != nil {
		return webdto.CaptchaResponse{}, err
	}
	captchaID, err := sharedcaptcha.NewID("web_captcha_")
	if err != nil {
		return webdto.CaptchaResponse{}, err
	}

	key := s.redis.Key("web", scene.redisKeySegment, captchaID)
	if err := s.redis.Client().Set(ctx, key, sharedcaptcha.HashText(code), time.Duration(sharedcaptcha.DefaultTTLSeconds)*time.Second).Err(); err != nil {
		return webdto.CaptchaResponse{}, err
	}

	return webdto.CaptchaResponse{
		CaptchaID: captchaID,
		Image:     sharedcaptcha.ImageDataURL(code),
		ExpiresIn: sharedcaptcha.DefaultTTLSeconds,
	}, nil
}

func (s *UserAuthService) verifyCaptchaIfEnabled(ctx context.Context, scene userCaptchaScene, captchaID string, captchaCode string) error {
	enabled, err := s.isCaptchaEnabled(ctx, scene.configKey)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	if s.redis == nil {
		return apperrors.ErrInternal.WithMessage("验证码服务暂不可用，请稍后再试")
	}

	captchaID = strings.TrimSpace(captchaID)
	captchaCode = strings.TrimSpace(captchaCode)
	if captchaID == "" || captchaCode == "" {
		return apperrors.ErrValidation.WithMessage("请输入验证码")
	}

	key := s.redis.Key("web", scene.redisKeySegment, captchaID)
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

func (s *UserAuthService) ensureCaptchaAllowed(ctx context.Context, scene userCaptchaScene, clientIP string) error {
	if s.redis == nil {
		return apperrors.ErrInternal.WithMessage("验证码服务暂不可用，请稍后再试")
	}

	key := s.redis.Key("web", scene.rateKeySegment, sharedcaptcha.HashText(clientIP))
	count, err := s.redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err := s.redis.Client().Expire(ctx, key, userCaptchaRateWindow).Err(); err != nil {
			return err
		}
	}
	if count > userCaptchaRateLimit {
		return apperrors.ErrTooManyRequests.WithMessage("验证码获取过于频繁，请稍后再试")
	}
	return nil
}

func (s *UserAuthService) ensureLoginAllowed(ctx context.Context, clientIP string, account string) error {
	if s.redis == nil {
		return apperrors.ErrInternal.WithMessage("登录服务暂不可用，请稍后再试")
	}
	count, err := s.redis.Client().Get(ctx, s.loginFailureRedisKey(clientIP, account)).Int64()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return err
	}
	if count >= userLoginFailureLimit {
		return apperrors.ErrTooManyRequests.WithMessage("登录失败次数过多，请稍后再试")
	}
	return nil
}

func (s *UserAuthService) recordLoginFailure(ctx context.Context, clientIP string, account string) error {
	if s.redis == nil {
		return apperrors.ErrInternal.WithMessage("登录服务暂不可用，请稍后再试")
	}
	key := s.loginFailureRedisKey(clientIP, account)
	count, err := s.redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		return s.redis.Client().Expire(ctx, key, userLoginFailureWindow).Err()
	}
	return nil
}

func (s *UserAuthService) clearLoginFailures(ctx context.Context, clientIP string, account string) error {
	if s.redis == nil {
		return nil
	}
	return s.redis.Client().Del(ctx, s.loginFailureRedisKey(clientIP, account)).Err()
}

func (s *UserAuthService) ensurePasswordResetAllowed(ctx context.Context, clientIP string, email string) error {
	if s.redis == nil {
		return apperrors.ErrInternal.WithMessage("密码找回服务暂不可用，请稍后再试")
	}
	key := s.passwordResetRateRedisKey(clientIP, email)
	count, err := s.redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err := s.redis.Client().Expire(ctx, key, userPasswordResetWindow).Err(); err != nil {
			return err
		}
	}
	if count > userPasswordResetLimit {
		return apperrors.ErrTooManyRequests.WithMessage("密码找回请求过于频繁，请稍后再试")
	}
	return nil
}

func (s *UserAuthService) loginFailureRedisKey(clientIP string, account string) string {
	return s.redis.Key("web", "login_fail", sharedcaptcha.HashText(clientIP), sharedcaptcha.HashText(account))
}

func (s *UserAuthService) passwordResetRateRedisKey(clientIP string, email string) string {
	return s.redis.Key("web", "password_reset_request", sharedcaptcha.HashText(clientIP), sharedcaptcha.HashText(email))
}

func (s *UserAuthService) isCaptchaEnabled(ctx context.Context, configKey string) (bool, error) {
	value, _, err := s.configs.ValueByKey(ctx, configKey)
	if err != nil {
		return false, err
	}
	return parseBoolConfig(value), nil
}

func parseBoolConfig(value *string) bool {
	if value == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*value), "true")
}
