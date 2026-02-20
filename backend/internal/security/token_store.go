package security

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore 负责基于 Redis 管理 refresh token、access token 黑名单和强制下线时间戳。
type TokenStore struct {
	client *redis.Client
}

// NewTokenStore 创建 TokenStore；当 Redis 不可用时可传入 nil，所有操作退化为 no-op。
func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

// Enabled 返回当前 TokenStore 是否可用。
func (s *TokenStore) Enabled() bool {
	return s != nil && s.client != nil
}

// SaveRefreshToken 保存 refresh token 的 tokenID，用于后续刷新和撤销校验。
func (s *TokenStore) SaveRefreshToken(ctx context.Context, userID uint, tokenID string, ttl time.Duration) error {
	if !s.Enabled() {
		return nil
	}
	if tokenID == "" {
		return errors.New("refresh token id is empty")
	}
	return s.client.Set(ctx, s.refreshTokenKey(userID, tokenID), "1", ttl).Err()
}

// IsRefreshTokenValid 判断 refresh token 是否仍在有效会话白名单中。
func (s *TokenStore) IsRefreshTokenValid(ctx context.Context, userID uint, tokenID string) (bool, error) {
	if !s.Enabled() {
		return true, nil
	}
	if tokenID == "" {
		return false, nil
	}
	exists, err := s.client.Exists(ctx, s.refreshTokenKey(userID, tokenID)).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// RevokeRefreshToken 将 refresh token 从白名单中移除。
func (s *TokenStore) RevokeRefreshToken(ctx context.Context, userID uint, tokenID string) error {
	if !s.Enabled() {
		return nil
	}
	if tokenID == "" {
		return nil
	}
	return s.client.Del(ctx, s.refreshTokenKey(userID, tokenID)).Err()
}

// BlacklistAccessToken 将 access token 的 tokenID 拉黑到过期时间，防止继续使用。
func (s *TokenStore) BlacklistAccessToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if !s.Enabled() {
		return nil
	}
	if tokenID == "" {
		return errors.New("access token id is empty")
	}
	if ttl <= 0 {
		return nil
	}
	return s.client.Set(ctx, s.blacklistKey(tokenID), "1", ttl).Err()
}

// IsAccessTokenBlacklisted 检查 access token 是否已在黑名单中。
func (s *TokenStore) IsAccessTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if !s.Enabled() {
		return false, nil
	}
	if tokenID == "" {
		return false, nil
	}
	exists, err := s.client.Exists(ctx, s.blacklistKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// ForceLogoutUser 写入用户强制下线时间点，所有该时间之前签发的 token 都将失效。
func (s *TokenStore) ForceLogoutUser(ctx context.Context, userID uint) error {
	if !s.Enabled() {
		return nil
	}
	now := strconv.FormatInt(time.Now().Unix(), 10)
	return s.client.Set(ctx, s.forceLogoutKey(userID), now, 0).Err()
}

// GetForceLogoutTime 读取用户强制下线时间；未设置时返回零值时间。
func (s *TokenStore) GetForceLogoutTime(ctx context.Context, userID uint) (time.Time, error) {
	if !s.Enabled() {
		return time.Time{}, nil
	}
	value, err := s.client.Get(ctx, s.forceLogoutKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	unixSeconds, convErr := strconv.ParseInt(value, 10, 64)
	if convErr != nil {
		return time.Time{}, fmt.Errorf("parse force logout timestamp: %w", convErr)
	}
	return time.Unix(unixSeconds, 0), nil
}

// SaveEmailVerificationToken 保存邮箱验证 token 到 Redis，value 为 userID。
func (s *TokenStore) SaveEmailVerificationToken(ctx context.Context, token string, userID uint, ttl time.Duration) error {
	if !s.Enabled() {
		return nil
	}
	if token == "" {
		return errors.New("email verification token is empty")
	}
	return s.client.Set(ctx, s.emailVerifyKey(token), strconv.FormatUint(uint64(userID), 10), ttl).Err()
}

// ConsumeEmailVerificationToken 消费邮箱验证 token，成功返回 userID 并删除 token。
func (s *TokenStore) ConsumeEmailVerificationToken(ctx context.Context, token string) (uint, error) {
	if !s.Enabled() {
		return 0, errors.New("token store is not enabled")
	}
	if token == "" {
		return 0, errors.New("email verification token is empty")
	}
	key := s.emailVerifyKey(token)
	value, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, errors.New("email verification token not found or expired")
		}
		return 0, err
	}
	if delErr := s.client.Del(ctx, key).Err(); delErr != nil {
		return 0, delErr
	}
	parsed, convErr := strconv.ParseUint(value, 10, 64)
	if convErr != nil {
		return 0, fmt.Errorf("parse email verification user id: %w", convErr)
	}
	return uint(parsed), nil
}

func (s *TokenStore) refreshTokenKey(userID uint, tokenID string) string {
	return fmt.Sprintf("auth:refresh:%d:%s", userID, tokenID)
}

func (s *TokenStore) blacklistKey(tokenID string) string {
	return fmt.Sprintf("auth:blacklist:%s", tokenID)
}

func (s *TokenStore) forceLogoutKey(userID uint) string {
	return fmt.Sprintf("auth:force-logout:%d", userID)
}

func (s *TokenStore) emailVerifyKey(token string) string {
	return fmt.Sprintf("auth:email-verify:%s", token)
}
