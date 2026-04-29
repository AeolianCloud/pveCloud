package dashboard

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
)

/**
 * AdminDashboardService 处理管理端首页数据聚合。
 */
type AdminDashboardService struct {
	db *gorm.DB
}

/**
 * NewAdminDashboardService 创建管理端首页服务。
 *
 * @param db 数据库连接
 * @return *AdminDashboardService 管理端首页服务
 */
func NewAdminDashboardService(db *gorm.DB) *AdminDashboardService {
	return &AdminDashboardService{db: db}
}

/**
 * Get 聚合当前管理员首页所需数据。
 *
 * @param ctx 请求上下文
 * @param adminID 管理员 ID
 * @param roleIDs 当前 token 中的角色 ID
 * @param permissionCodes 当前 token 中的权限码
 * @return admin.DashboardResponse 首页响应数据
 * @return error 聚合失败原因
 */
func (s *AdminDashboardService) Get(ctx context.Context, adminID uint64, roleIDs []uint64, permissionCodes []string, session admindto.SessionSummary) (admindto.DashboardResponse, error) {
	var admin models.AdminUser
	err := s.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id = ?", adminID).
		First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.DashboardResponse{}, apperrors.ErrUnauthorized
	}
	if err != nil {
		return admindto.DashboardResponse{}, err
	}
	if admin.Status != support.AdminStatusActive {
		return admindto.DashboardResponse{}, apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
	}

	metrics, err := s.metrics(ctx)
	if err != nil {
		return admindto.DashboardResponse{}, err
	}

	return admindto.DashboardResponse{
		AuthStateResponse: admindto.AuthStateResponse{
			Admin: admindto.AdminSummary{
				ID:          admin.ID,
				Username:    admin.Username,
				Email:       admin.Email,
				DisplayName: admin.DisplayName,
				Status:      admin.Status,
			},
			RoleIDs:         roleIDs,
			PermissionCodes: permissionCodes,
			Menus:           support.VisibleAdminMenus(permissionCodes),
			Session:         session,
		},
		Metrics: metrics,
	}, nil
}

func (s *AdminDashboardService) metrics(ctx context.Context) ([]admindto.DashboardMetric, error) {
	type metricQuery struct {
		key   string
		title string
		table string
		where string
		unit  string
	}

	today := time.Now().Format("2006-01-02")
	queries := []metricQuery{
		{key: "active_admins", title: "启用管理员", table: "admin_users", where: "status = 'active' AND deleted_at IS NULL", unit: "人"},
		{key: "active_roles", title: "启用角色", table: "admin_roles", where: "status = 'active'", unit: "个"},
		{key: "active_sessions", title: "活跃会话", table: "admin_sessions", where: "status = 'active' AND expires_at > NOW(3)", unit: "个"},
		{key: "risk_logs_today", title: "今日高危操作", table: "admin_risk_logs", where: "created_at >= '" + today + "'", unit: "次"},
	}

	metrics := make([]admindto.DashboardMetric, 0, len(queries))
	for _, query := range queries {
		var value int64
		if err := s.db.WithContext(ctx).Table(query.table).Where(query.where).Count(&value).Error; err != nil {
			return nil, err
		}
		unit := query.unit
		metrics = append(metrics, admindto.DashboardMetric{
			Key:   query.key,
			Title: query.title,
			Value: value,
			Unit:  &unit,
		})
	}

	return metrics, nil
}
