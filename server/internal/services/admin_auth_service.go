package services

import (
	"context"
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/models"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/pkg/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/password"
)

const adminStatusActive = "active"

/**
 * AdminAuthService 处理管理端登录和 token 签发。
 */
type AdminAuthService struct {
	db  *gorm.DB
	cfg bootstrap.JWTConfig
}

/**
 * NewAdminAuthService 创建管理端认证服务。
 *
 * @param db 数据库连接
 * @param cfg JWT 配置
 * @return *AdminAuthService 管理端认证服务
 */
func NewAdminAuthService(db *gorm.DB, cfg bootstrap.JWTConfig) *AdminAuthService {
	return &AdminAuthService{db: db, cfg: cfg}
}

/**
 * Login 校验管理员账号密码并签发管理端 JWT。
 *
 * @param ctx 请求上下文
 * @param req 登录请求
 * @param clientIP 客户端 IP，用于记录最后登录来源
 * @return admin.LoginResponse 登录响应
 * @return error 登录失败原因
 */
func (s *AdminAuthService) Login(ctx context.Context, req admindto.LoginRequest, clientIP string) (admindto.LoginResponse, error) {
	identifier := strings.TrimSpace(req.Username)
	if identifier == "" {
		return admindto.LoginResponse{}, apperrors.ErrValidation.WithMessage("管理员账号不能为空")
	}

	var admin models.AdminUser
	err := s.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("username = ? OR email = ?", identifier, identifier).
		First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("管理员账号或密码错误")
	}
	if err != nil {
		return admindto.LoginResponse{}, err
	}
	if admin.Status != adminStatusActive {
		return admindto.LoginResponse{}, apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
	}
	if !password.Verify(admin.PasswordHash, req.Password) {
		return admindto.LoginResponse{}, apperrors.ErrUnauthorized.WithMessage("管理员账号或密码错误")
	}

	roleIDs, err := s.roleIDs(ctx, admin.ID)
	if err != nil {
		return admindto.LoginResponse{}, err
	}
	permissionCodes, err := s.permissionCodes(ctx, admin.ID)
	if err != nil {
		return admindto.LoginResponse{}, err
	}

	ttl := time.Duration(s.cfg.AdminExpireMinutes) * time.Minute
	claims := jwtpkg.Claims{
		TokenType:       "admin",
		AdminID:         admin.ID,
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    s.cfg.AdminIssuer,
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	token, err := jwtpkg.Sign(claims, s.cfg.AdminSecret)
	if err != nil {
		return admindto.LoginResponse{}, err
	}

	now := time.Now()
	update := map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": clientIP,
	}
	if err := s.db.WithContext(ctx).Model(&models.AdminUser{}).Where("id = ?", admin.ID).Updates(update).Error; err != nil {
		return admindto.LoginResponse{}, err
	}

	return admindto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(ttl.Seconds()),
		Admin: admindto.AdminSummary{
			ID:          admin.ID,
			Username:    admin.Username,
			Email:       admin.Email,
			DisplayName: admin.DisplayName,
			Status:      admin.Status,
		},
		RoleIDs:         roleIDs,
		PermissionCodes: permissionCodes,
	}, nil
}

func (s *AdminAuthService) roleIDs(ctx context.Context, adminID uint64) ([]uint64, error) {
	var roleIDs []uint64
	err := s.db.WithContext(ctx).
		Table("admin_user_roles").
		Select("admin_roles.id").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", adminStatusActive).
		Order("admin_roles.id ASC").
		Scan(&roleIDs).Error
	return roleIDs, err
}

func (s *AdminAuthService) permissionCodes(ctx context.Context, adminID uint64) ([]string, error) {
	var codes []string
	err := s.db.WithContext(ctx).
		Table("admin_user_roles").
		Distinct("admin_permissions.code").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Joins("JOIN admin_role_permissions ON admin_role_permissions.role_id = admin_roles.id").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", adminStatusActive).
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}
