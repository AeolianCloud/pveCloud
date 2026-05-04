package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	gojwt "github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/mail"
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
)

/**
 * UserAuthService 处理用户端登录、会话和 token 签发。
 */
type UserAuthService struct {
	db     *gorm.DB
	cfg    bootstrap.JWTConfig
	mail   *mail.Sender
	webURL string
}

/**
 * NewUserAuthService 创建用户端认证服务。
 */
func NewUserAuthService(db *gorm.DB, cfg bootstrap.JWTConfig, mailCfg bootstrap.MailConfig) *UserAuthService {
	return &UserAuthService{
		db:     db,
		cfg:    cfg,
		mail:   mail.NewSender(mailCfg),
		webURL: mailCfg.PasswordResetURLBase,
	}
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

	var user models.User
	err := s.db.WithContext(ctx).
		Where("username = ? OR email = ?", account, account).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("账号或密码错误")
	}
	if err != nil {
		return webdto.LoginResponse{}, err
	}
	if user.Status != userStatusActive {
		return webdto.LoginResponse{}, apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
	}
	if !password.Verify(user.PasswordHash, req.Password) {
		return webdto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("账号或密码错误")
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
	return s.db.WithContext(ctx).Model(&models.UserSession{}).
		Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, userSessionStatusActive).
		Updates(map[string]interface{}{
			"status":        userSessionStatusRevoked,
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error
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
	if s.mail == nil || !s.mail.Enabled() {
		return apperrors.ErrInternal.WithMessage("密码找回服务暂不可用，请稍后再试")
	}

	request := audit.RequestContextFrom(ctx)
	email := strings.ToLower(strings.TrimSpace(req.Email))

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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.UserPasswordResetToken{}).
			Where("user_id = ? AND status = ?", user.ID, userResetStatusActive).
			Updates(map[string]interface{}{"status": userResetStatusRevoked}).Error; err != nil {
			return err
		}
		return tx.Create(&reset).Error
	}); err != nil {
		return err
	}

	if err := s.mail.SendPasswordReset(user.Email, s.resetURL(token)); err != nil {
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
	tokenHash := hashPasswordResetToken(strings.TrimSpace(req.Token))
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
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "usr_" + hex.EncodeToString(bytes[:]), nil
}

func newPasswordResetToken() (string, string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(bytes[:])
	return token, hashPasswordResetToken(token), nil
}

func hashPasswordResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
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
