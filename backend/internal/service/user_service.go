package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/repository"
	"pvecloud/backend/internal/security"
)

var (
	errEmailRegistered          = errors.New("邮箱已被注册")
	errInvalidCredentials       = errors.New("邮箱或密码错误")
	errWeakPassword             = errors.New("密码强度不足")
	errAccountTemporarilyLocked = errors.New("账号已锁定，请30分钟后重试")
	errRefreshTokenRevoked      = errors.New("refresh token 已失效，请重新登录")
)

// UserService 负责注册、登录、刷新 token、登出和强制下线等用户认证逻辑。
type UserService struct {
	userRepo   *repository.UserRepository
	jwt        *middleware.JWTManager
	tokenStore *security.TokenStore
}

// NewUserService 创建用户服务。
func NewUserService(userRepo *repository.UserRepository, jwt *middleware.JWTManager, tokenStore *security.TokenStore) *UserService {
	return &UserService{userRepo: userRepo, jwt: jwt, tokenStore: tokenStore}
}

// Register 完成邮箱唯一校验、密码强度校验、密码哈希并创建用户。
func (s *UserService) Register(ctx context.Context, email, password string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !passwordStrong(password) {
		return errWeakPassword
	}

	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return errEmailRegistered
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{Email: email, PasswordHash: string(hash), Role: "user", Status: "active"}
	return s.userRepo.Create(ctx, user)
}

// Login 校验账号和密码，处理失败计数与锁定逻辑，并返回 token 对。
func (s *UserService) Login(ctx context.Context, email, password string) (string, string, *model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", nil, errInvalidCredentials
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return "", "", nil, errAccountTemporarilyLocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		user.LoginFailedCount += 1
		if user.LoginFailedCount >= 5 {
			lockedUntil := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockedUntil
			user.LoginFailedCount = 0
		}
		_ = s.userRepo.Update(ctx, user)
		return "", "", nil, errInvalidCredentials
	}

	user.LoginFailedCount = 0
	user.LockedUntil = nil
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", "", nil, err
	}

	access, refresh, _, refreshClaims, err := s.jwt.GenerateTokenPairWithClaims(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", nil, err
	}

	if s.tokenStore != nil {
		if err := s.tokenStore.SaveRefreshToken(ctx, user.ID, refreshClaims.ID, s.jwt.RefreshTTL()); err != nil {
			return "", "", nil, err
		}
	}

	return access, refresh, user, nil
}

// RefreshAccessToken 基于 refresh token 校验会话有效性后签发新 access token。
func (s *UserService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.jwt.ParseToken(refreshToken)
	if err != nil {
		return "", err
	}
	if claims.Type != "refresh" {
		return "", errors.New("token type is not refresh")
	}

	if s.tokenStore != nil && s.tokenStore.Enabled() {
		valid, err := s.tokenStore.IsRefreshTokenValid(ctx, claims.UserID, claims.ID)
		if err != nil {
			return "", err
		}
		if !valid {
			return "", errRefreshTokenRevoked
		}
		forcedAt, err := s.tokenStore.GetForceLogoutTime(ctx, claims.UserID)
		if err != nil {
			return "", err
		}
		if !forcedAt.IsZero() && claims.IssuedAt != nil && claims.IssuedAt.Time.Before(forcedAt) {
			return "", errRefreshTokenRevoked
		}
	}

	newAccess, _, err := s.jwt.GenerateAccessTokenByClaims(claims)
	if err != nil {
		return "", err
	}
	return newAccess, nil
}

// Logout 将 access token 加入黑名单，并可选撤销 refresh token。
func (s *UserService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	if s.tokenStore == nil || !s.tokenStore.Enabled() {
		return nil
	}

	if accessToken != "" {
		claims, err := s.jwt.ParseToken(accessToken)
		if err == nil && claims.Type == "access" {
			ttl := time.Until(claims.ExpiresAt.Time)
			_ = s.tokenStore.BlacklistAccessToken(ctx, claims.ID, ttl)
		}
	}

	if refreshToken != "" {
		claims, err := s.jwt.ParseToken(refreshToken)
		if err == nil && claims.Type == "refresh" {
			_ = s.tokenStore.RevokeRefreshToken(ctx, claims.UserID, claims.ID)
		}
	}

	return nil
}

// ForceLogout 将用户标记为强制下线，历史 token 立即失效。
func (s *UserService) ForceLogout(ctx context.Context, userID uint) error {
	if s.tokenStore == nil {
		return nil
	}
	return s.tokenStore.ForceLogoutUser(ctx, userID)
}

func passwordStrong(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasLetter && hasDigit
}
