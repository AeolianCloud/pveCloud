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
	"gorm.io/gorm/clause"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/mail"
	sharedcaptcha "github.com/AeolianCloud/pveCloud/server/internal/shared/captcha"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/shared/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	websupport "github.com/AeolianCloud/pveCloud/server/internal/web/support"
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
	db     *gorm.DB
	redis  *cache.Redis
	cfg    bootstrap.JWTConfig
	mail   *mail.Sender
	webURL string
}

/**
 * NewUserAuthService 创建用户端认证服务。
 */
func NewUserAuthService(db *gorm.DB, redis *cache.Redis, cfg bootstrap.JWTConfig, mailCfg bootstrap.MailConfig) *UserAuthService {
	return &UserAuthService{
		db:     db,
		redis:  redis,
		cfg:    cfg,
		mail:   mail.NewSender(mailCfg),
		webURL: mailCfg.PasswordResetURLBase,
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
 * Login 校验用户账号密码并签发用户端 JWT。
 */
func (s *UserAuthService) Login(ctx context.Context, req webdto.LoginRequest) (webdto.LoginResponse, error) {
	request := audit.RequestContextFrom(ctx)
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

	var user models.User
	err := s.db.WithContext(ctx).
		Where("username = ? OR email = ?", account, account).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if recordErr := s.recordLoginFailure(ctx, request.IP, account); recordErr != nil {
			return webdto.LoginResponse{}, recordErr
		}
		return webdto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("账号或密码错误")
	}
	if err != nil {
		return webdto.LoginResponse{}, err
	}
	if user.Status != userStatusActive {
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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	request := audit.RequestContextFrom(ctx)
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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		hash, err := password.Hash(req.Password)
		if err != nil {
			return err
		}
		user := models.User{
			Username:     username,
			Email:        email,
			PasswordHash: hash,
			DisplayName:  displayName,
			Status:       userStatusActive,
		}
		if err := tx.Create(&user).Error; err != nil {
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
func (s *UserAuthService) Me(user models.User, session webdto.SessionSummary) webdto.AuthStateResponse {
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
	result := s.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, userSessionStatusActive).
		Updates(map[string]interface{}{
			"status":        userSessionStatusRevoked,
			"revoked_at":    now,
			"revoke_reason": reason,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUnauthorized.WithMessage("会话已失效")
	}
	return nil
}

/**
 * Refresh 轮换当前用户端 token。
 */
func (s *UserAuthService) Refresh(ctx context.Context, userID uint64, sessionID string) (webdto.LoginResponse, error) {
	request := audit.RequestContextFrom(ctx)
	if strings.TrimSpace(sessionID) == "" {
		return webdto.LoginResponse{}, apperrors.ErrUnauthorized
	}

	var result webdto.LoginResponse
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var oldSession models.UserSession
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, userSessionStatusActive).
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

		var user models.User
		err = tx.Where("id = ?", userID).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if user.Status != userStatusActive {
			return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
		}

		now := time.Now()
		reason := revokeReasonRefresh
		if err := tx.Model(&models.UserSession{}).
			Where("id = ?", oldSession.ID).
			Updates(map[string]interface{}{
				"status":        userSessionStatusRevoked,
				"revoked_at":    now,
				"revoke_reason": reason,
			}).Error; err != nil {
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

	request := audit.RequestContextFrom(ctx)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := s.ensurePasswordResetAllowed(ctx, request.IP, email); err != nil {
		return err
	}

	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.Status != userStatusActive {
		return nil
	}

	token, tokenHash, err := newPasswordResetToken()
	if err != nil {
		return err
	}
	now := time.Now()
	reset := models.UserPasswordResetToken{
		UserID:      user.ID,
		TokenHash:   tokenHash,
		Status:      userResetStatusActive,
		ExpiresAt:   now.Add(passwordResetTTL),
		RequestedIP: textutil.StringPtr(request.IP),
		UserAgent:   textutil.StringPtr(textutil.TrimTo(request.UserAgent, 500)),
	}
	shouldSend := false
	sendEmail := user.Email
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lockedUser models.User
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", user.ID).First(&lockedUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if lockedUser.Status != userStatusActive {
			return nil
		}
		if err := tx.Model(&models.UserPasswordResetToken{}).
			Where("user_id = ? AND status = ?", lockedUser.ID, userResetStatusActive).
			Updates(map[string]interface{}{"status": userResetStatusRevoked}).Error; err != nil {
			return err
		}
		if err := tx.Create(&reset).Error; err != nil {
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
		_ = s.db.WithContext(ctx).Model(&models.UserPasswordResetToken{}).
			Where("token_hash = ? AND status = ?", tokenHash, userResetStatusActive).
			Update("status", userResetStatusRevoked).Error
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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reset models.UserPasswordResetToken
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ?", tokenHash).
			First(&reset).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrConflict.WithMessage("密码重置链接无效或已过期")
		}
		if err != nil {
			return err
		}
		if reset.Status != userResetStatusActive || !reset.ExpiresAt.After(now) {
			return apperrors.ErrConflict.WithMessage("密码重置链接无效或已过期")
		}

		var user models.User
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", reset.UserID).
			First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrConflict.WithMessage("密码重置链接无效或已过期")
		}
		if err != nil {
			return err
		}
		if user.Status != userStatusActive {
			if err := tx.Model(&models.UserPasswordResetToken{}).
				Where("id = ?", reset.ID).
				Update("status", userResetStatusRevoked).Error; err != nil {
				return err
			}
			userDisabled = true
			return nil
		}

		hash, err := password.Hash(req.Password)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).
			Where("id = ?", user.ID).
			Update("password_hash", hash).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.UserPasswordResetToken{}).
			Where("id = ?", reset.ID).
			Updates(map[string]interface{}{"status": userResetStatusUsed, "used_at": now}).Error; err != nil {
			return err
		}
		reason := revokeReasonPasswordReset
		return tx.Model(&models.UserSession{}).
			Where("user_id = ? AND status = ?", reset.UserID, userSessionStatusActive).
			Updates(map[string]interface{}{"status": userSessionStatusRevoked, "revoked_at": now, "revoke_reason": reason}).Error
	}); err != nil {
		return err
	}
	if userDisabled {
		return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
	}
	return nil
}

func (s *UserAuthService) issueSession(ctx context.Context, tx *gorm.DB, user models.User, clientIP string, userAgent string) (webdto.LoginResponse, error) {
	sessionID, err := newUserSessionID()
	if err != nil {
		return webdto.LoginResponse{}, err
	}
	now := time.Now()
	ttl := time.Duration(s.cfg.UserExpireMinutes) * time.Minute
	expiresAt := now.Add(ttl)
	session := models.UserSession{
		SessionID:  sessionID,
		UserID:     user.ID,
		Status:     userSessionStatusActive,
		IssuedAt:   now,
		ExpiresAt:  expiresAt,
		LastSeenAt: &now,
		LastSeenIP: textutil.StringPtr(clientIP),
		UserAgent:  textutil.StringPtr(textutil.TrimTo(userAgent, 500)),
	}
	if err := tx.Create(&session).Error; err != nil {
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

	request := audit.RequestContextFrom(ctx)
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
	var config models.SystemConfig
	err := s.db.WithContext(ctx).
		Select("config_value").
		Where("config_key = ? AND is_secret = 0", configKey).
		First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return parseBoolConfig(config.ConfigValue), nil
}

func parseBoolConfig(value *string) bool {
	if value == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*value), "true")
}
