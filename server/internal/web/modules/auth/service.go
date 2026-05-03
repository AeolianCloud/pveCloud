package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/shared/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	websupport "github.com/AeolianCloud/pveCloud/server/internal/web/support"
)

const (
	userStatusActive         = "active"
	userSessionStatusActive  = "active"
	userSessionStatusRevoked = "revoked"
	revokeReasonLogout       = "logout"
	revokeReasonRefresh      = "refresh"
)

/**
 * UserAuthService 处理用户端登录、会话和 token 签发。
 */
type UserAuthService struct {
	db  *gorm.DB
	cfg bootstrap.JWTConfig
}

/**
 * NewUserAuthService 创建用户端认证服务。
 */
func NewUserAuthService(db *gorm.DB, cfg bootstrap.JWTConfig) *UserAuthService {
	return &UserAuthService{db: db, cfg: cfg}
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
