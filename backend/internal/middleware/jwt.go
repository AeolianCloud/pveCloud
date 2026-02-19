package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 承载 JWT 中的业务身份信息。
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// JWTManager 统一管理 token 生成、验证与刷新。
type JWTManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTManager 创建 JWT 管理器。
func NewJWTManager(secret string, accessHours int, refreshDays int) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		accessTTL:  time.Duration(accessHours) * time.Hour,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

// AccessTTL 返回 access token 的有效期。
func (j *JWTManager) AccessTTL() time.Duration {
	return j.accessTTL
}

// RefreshTTL 返回 refresh token 的有效期。
func (j *JWTManager) RefreshTTL() time.Duration {
	return j.refreshTTL
}

// GenerateTokenPair 生成 access token 与 refresh token。
func (j *JWTManager) GenerateTokenPair(userID uint, email, role string) (string, string, error) {
	accessToken, refreshToken, _, _, err := j.GenerateTokenPairWithClaims(userID, email, role)
	return accessToken, refreshToken, err
}

// GenerateTokenPairWithClaims 生成 token 对并返回 claims，便于调用方做会话持久化。
func (j *JWTManager) GenerateTokenPairWithClaims(userID uint, email, role string) (string, string, *Claims, *Claims, error) {
	now := time.Now()
	accessID, err := newTokenID()
	if err != nil {
		return "", "", nil, nil, err
	}
	refreshID, err := newTokenID()
	if err != nil {
		return "", "", nil, nil, err
	}

	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(j.secret)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("sign access token: %w", err)
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(j.secret)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("sign refresh token: %w", err)
	}
	return accessToken, refreshToken, accessClaims, refreshClaims, nil
}

// GenerateAccessTokenByClaims 基于 refresh token 里的用户身份生成新的 access token。
func (j *JWTManager) GenerateAccessTokenByClaims(claims *Claims) (string, *Claims, error) {
	now := time.Now()
	accessID, err := newTokenID()
	if err != nil {
		return "", nil, err
	}
	newClaims := &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims).SignedString(j.secret)
	if err != nil {
		return "", nil, fmt.Errorf("sign refreshed access token: %w", err)
	}
	return accessToken, newClaims, nil
}

// ParseToken 校验并解析 token。
func (j *JWTManager) ParseToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// RefreshAccessToken 使用 refresh token 生成新的 access token。
func (j *JWTManager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := j.ParseToken(refreshToken)
	if err != nil {
		return "", err
	}
	if claims.Type != "refresh" {
		return "", errors.New("token type is not refresh")
	}
	newAccess, _, err := j.GenerateAccessTokenByClaims(claims)
	if err != nil {
		return "", err
	}
	return newAccess, nil
}

func newTokenID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token id: %w", err)
	}
	return hex.EncodeToString(raw), nil
}
