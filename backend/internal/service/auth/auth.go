// internal/service/auth/auth.go
// 认证业务逻辑：登录校验、Token 签发、登录日志记录、获取当前用户。
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/session"
)

// Service 认证服务。
type Service struct {
	db          *gorm.DB
	jwtSecret   string
	expireHours int
	// refreshExpireHours Refresh Token 有效期（小时）。
	// Refresh Token 用于刷新 Access Token，同时也用于“会话是否仍有效”的服务端判断。
	refreshExpireHours int
	// sessStore 会话存储（按你的要求使用 Redis）。
	sessStore session.Store
}

// New 创建认证服务实例。
func New(db *gorm.DB, jwtSecret string, expireHours, refreshExpireHours int, sessStore session.Store) *Service {
	// 兼容配置缺省：refresh_expire_hours 未配置时给一个合理默认值（7 天）
	if refreshExpireHours <= 0 {
		refreshExpireHours = 168
	}
	return &Service{
		db:                 db,
		jwtSecret:          jwtSecret,
		expireHours:        expireHours,
		refreshExpireHours: refreshExpireHours,
		sessStore:          sessStore,
	}
}

// LoginResult 登录成功返回的数据。
type LoginResult struct {
	// Token Access Token（JWT），用于访问受保护接口
	Token string `json:"token"`
	// RefreshToken Refresh Token（JWT），用于刷新 Access Token
	RefreshToken string           `json:"refresh_token"`
	User         *model.AdminUser `json:"user"`
}

// RefreshResult 刷新成功返回的数据。
type RefreshResult struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginOption 登录时附带的请求上下文信息，用于写登录日志。
type LoginOption struct {
	IP        string
	UserAgent string
}

// Login 验证用户名密码，成功后签发 JWT，并记录登录日志。
func (s *Service) Login(username, password string, opt LoginOption) (*LoginResult, error) {
	var user model.AdminUser
	// Preload Roles.Permissions，确保登录响应即携带完整权限数据，前端无需二次请求
	err := s.db.Preload("Roles.Permissions").Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.writeLog(0, username, opt, 0, "用户不存在")
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.writeLog(user.ID, username, opt, 0, "密码错误")
		return nil, ErrPasswordWrong
	}

	// 账号禁用检查
	if user.Status == 0 {
		s.writeLog(user.ID, username, opt, 0, "账号已被禁用")
		return nil, ErrUserDisabled
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", now)
	user.LastLoginAt = &now

	// 取第一个角色名作为 JWT role 字段（兼容中间件）
	roleName := ""
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Name
	}

	// 创建会话（Redis）：用于退出登录立即失效 + Refresh Token 刷新
	expiresAt := time.Now().Add(time.Duration(s.refreshExpireHours) * time.Hour)
	sid, refreshJTI, err := s.sessStore.Create(user.ID, expiresAt, session.LoginMeta{
		IP:        opt.IP,
		UserAgent: opt.UserAgent,
	})
	if err != nil {
		return nil, err
	}

	// 签发 Access Token + Refresh Token
	accessToken, err := s.signAccessToken(user.ID, user.Username, roleName, sid)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.signRefreshToken(user.ID, sid, refreshJTI)
	if err != nil {
		return nil, err
	}

	// 记录成功日志
	s.writeLog(user.ID, username, opt, 1, "")

	return &LoginResult{Token: accessToken, RefreshToken: refreshToken, User: &user}, nil
}

// GetUserByID 根据 ID 查询管理员信息（含角色），供 /profile 接口使用。
func (s *Service) GetUserByID(id uint) (*model.AdminUser, error) {
	var user model.AdminUser
	err := s.db.Preload("Roles.Permissions").First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// Refresh 使用 Refresh Token 刷新 Access Token，并旋转 Refresh Token。
// 安全策略：Refresh Token rotation
// - 每次 refresh 都会签发新的 Refresh Token，并更新 session.refresh_jti
// - 旧 Refresh Token 将立刻失效，避免泄露后被长期复用
func (s *Service) Refresh(refreshToken string, opt LoginOption) (*RefreshResult, error) {
	claims := &refreshClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid || claims.Type != "refresh" || claims.SessionID == 0 || claims.UserID == 0 {
		return nil, ErrRefreshTokenInvalid
	}

	// 查询会话（Redis），校验是否撤销/过期，且 refresh_jti 必须匹配（rotation）
	sess, err := s.sessStore.Get(claims.SessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	if sess.AdminUserID != claims.UserID {
		return nil, ErrRefreshTokenInvalid
	}
	if sess.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil, ErrSessionExpired
	}
	if sess.RefreshJTI != claims.ID {
		return nil, ErrRefreshTokenInvalid
	}

	// refresh rotation：原子更新 refresh_jti，旧 refresh token 将立刻失效
	newRefreshJTI := newJTI()
	if err := s.sessStore.RotateRefreshJTI(claims.SessionID, claims.ID, newRefreshJTI, session.LoginMeta{
		IP:        opt.IP,
		UserAgent: opt.UserAgent,
	}); err != nil {
		// refresh_jti 不匹配通常意味着旧 token 被重放/并发刷新，统一按 token 无效处理
		return nil, ErrRefreshTokenInvalid
	}

	// 取最新用户信息（主要用于 role 字段兼容）
	var user model.AdminUser
	if err := s.db.Preload("Roles").First(&user, claims.UserID).Error; err != nil {
		return nil, err
	}
	roleName := ""
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Name
	}

	accessToken, err := s.signAccessToken(user.ID, user.Username, roleName, claims.SessionID)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := s.signRefreshToken(user.ID, claims.SessionID, newRefreshJTI)
	if err != nil {
		return nil, err
	}

	return &RefreshResult{Token: accessToken, RefreshToken: newRefreshToken}, nil
}

// Logout 撤销当前会话，使该会话下的 Access/Refresh Token 立即失效。
func (s *Service) Logout(sessionID uint) error {
	if sessionID == 0 {
		return ErrSessionNotFound
	}
	if err := s.sessStore.Revoke(sessionID); err != nil {
		// 退出登录按幂等语义处理：会话已过期/已删除时也视为成功
		if errors.Is(err, session.ErrNotFound) {
			return nil
		}
		// Redis 等基础设施异常仍然向上返回，便于排查
		return err
	}
	return nil
}

// signAccessToken 签发 Access Token（JWT），写入 user_id、username、role 和 session_id。
func (s *Service) signAccessToken(userID uint, username, role string, sessionID uint) (string, error) {
	claims := middleware.Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newJTI(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// refreshClaims Refresh Token 的 claims。
// 注意：refresh token 只用于刷新，不用于访问业务接口。
type refreshClaims struct {
	UserID    uint   `json:"user_id"`
	SessionID uint   `json:"session_id"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

// signRefreshToken 签发 Refresh Token（JWT）。
func (s *Service) signRefreshToken(userID uint, sessionID uint, refreshJTI string) (string, error) {
	claims := refreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		Type:      "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			// refresh_jti 做 rotation 校验：必须与 session.refresh_jti 一致
			ID:        refreshJTI,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.refreshExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// writeLog 写登录日志，失败时静默处理，不影响主流程。
func (s *Service) writeLog(userID uint, username string, opt LoginOption, status int8, remark string) {
	log := model.AdminLoginLog{
		AdminUserID: userID,
		Username:    username,
		IP:          opt.IP,
		UserAgent:   opt.UserAgent,
		Status:      status,
		Remark:      remark,
	}
	s.db.Create(&log)
}

// newJTI 生成一个随机的 Token ID（jti），用于标识一次签发的 token。
// 说明：
// - 使用 crypto/rand 生成 16 字节随机数，再编码为 32 位 hex
// - 不引入第三方 uuid 依赖，保持项目依赖简单
func newJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand 理论上不应失败；极端情况下 fallback 到时间戳字符串，避免直接 panic
		return fmt.Sprintf("jti_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// 业务错误定义
var (
	ErrUserNotFound  = errors.New("用户不存在")
	ErrPasswordWrong = errors.New("密码错误")
	ErrUserDisabled  = errors.New("账号已被禁用")

	ErrRefreshTokenInvalid = errors.New("refresh token 无效")
	ErrSessionNotFound     = errors.New("会话不存在")
	ErrSessionExpired      = errors.New("会话已过期")
	ErrSessionRevoked      = errors.New("会话已撤销")
)
