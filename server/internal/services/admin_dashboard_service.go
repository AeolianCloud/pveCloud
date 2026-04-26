package services

import (
	"context"
	"errors"

	"gorm.io/gorm"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/models"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
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
	if admin.Status != adminStatusActive {
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
			Menus:           VisibleAdminMenus(permissionCodes),
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

	queries := []metricQuery{
		{key: "active_users", title: "活跃用户", table: "users", where: "status = 'active' AND deleted_at IS NULL", unit: "人"},
		{key: "pending_orders", title: "待支付订单", table: "orders", where: "status = 'pending'", unit: "单"},
		{key: "running_instances", title: "运行中实例", table: "instances", where: "status = 'running' AND deleted_at IS NULL", unit: "台"},
		{key: "open_tickets", title: "待处理工单", table: "tickets", where: "status IN ('open', 'pending_admin')", unit: "单"},
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

func VisibleAdminMenus(permissionCodes []string) []admindto.MenuItem {
	permissionSet := make(map[string]struct{}, len(permissionCodes))
	for _, code := range permissionCodes {
		permissionSet[code] = struct{}{}
	}

	menus := adminMenuCatalog()
	visible := make([]admindto.MenuItem, 0, len(menus))
	for _, menu := range menus {
		if menu.PermissionCode == nil {
			visible = append(visible, menu)
			continue
		}
		if _, ok := permissionSet[*menu.PermissionCode]; ok {
			visible = append(visible, menu)
		}
	}
	return visible
}

func adminMenuCatalog() []admindto.MenuItem {
	return []admindto.MenuItem{
		menuItem("dashboard", "控制台", "/dashboard", "layout-dashboard", "dashboard:view"),
	}
}

func menuItem(key string, title string, path string, icon string, permissionCode string) admindto.MenuItem {
	return admindto.MenuItem{
		Key:            key,
		Title:          title,
		Path:           path,
		Icon:           &icon,
		PermissionCode: &permissionCode,
	}
}
