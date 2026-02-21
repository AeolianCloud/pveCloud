// internal/service/auth/auth.go
// 认证业务逻辑：登录校验、Token 签发、登录日志记录、获取当前用户。
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/model"
)

// Service 认证服务。
type Service struct {
	db          *gorm.DB
	jwtSecret   string
	expireHours int
}

// New 创建认证服务实例。
func New(db *gorm.DB, jwtSecret string, expireHours int) *Service {
	return &Service{db: db, jwtSecret: jwtSecret, expireHours: expireHours}
}

// LoginResult 登录成功返回的数据。
type LoginResult struct {
	Token string           `json:"token"`
	User  *model.AdminUser `json:"user"`
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

	// 签发 JWT
	token, err := s.signToken(user.ID, user.Username, roleName)
	if err != nil {
		return nil, err
	}

	// 记录成功日志
	s.writeLog(user.ID, username, opt, 1, "")

	return &LoginResult{Token: token, User: &user}, nil
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

// signToken 签发 JWT，写入 user_id、username 和 role。
func (s *Service) signToken(userID uint, username, role string) (string, error) {
	claims := middleware.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.expireHours) * time.Hour)),
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

// 业务错误定义
var (
	ErrUserNotFound  = errors.New("用户不存在")
	ErrPasswordWrong = errors.New("密码错误")
	ErrUserDisabled  = errors.New("账号已被禁用")
)
