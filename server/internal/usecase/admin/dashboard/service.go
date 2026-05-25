package dashboard

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	mysqldashboard "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/dashboard"
	mysqliam "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/iam"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

/**
 * AdminDashboardService 处理管理端首页数据聚合。
 */
type AdminDashboardService struct {
	db        *gorm.DB
	iam       *mysqliam.Repository
	dashboard *mysqldashboard.Repository
}

/**
 * NewAdminDashboardService 创建管理端首页服务。
 *
 * @param db 数据库连接
 * @return *AdminDashboardService 管理端首页服务
 */
func NewAdminDashboardService(db *gorm.DB) *AdminDashboardService {
	return &AdminDashboardService{
		db:        db,
		iam:       mysqliam.NewRepository(db),
		dashboard: mysqldashboard.NewRepository(db),
	}
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
	admin, err := s.iam.FindAdminUserByID(ctx, nil, adminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.DashboardResponse{}, apperrors.ErrUnauthorized
	}
	if err != nil {
		return admindto.DashboardResponse{}, err
	}
	if admin.Status != adminsupport.AdminStatusActive {
		return admindto.DashboardResponse{}, apperrors.ErrForbidden.WithMessage("管理员账号已被禁用")
	}

	metrics, err := s.metrics(ctx)
	if err != nil {
		return admindto.DashboardResponse{}, err
	}
	businessMetrics, err := s.businessMetrics(ctx)
	if err != nil {
		return admindto.DashboardResponse{}, err
	}
	menus, err := adminsupport.VisibleAdminMenus(ctx, s.db, permissionCodes)
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
			Menus:           menus,
			Session:         session,
		},
		Metrics:         metrics,
		BusinessMetrics: businessMetrics,
	}, nil
}

func (s *AdminDashboardService) metrics(ctx context.Context) ([]admindto.DashboardMetric, error) {
	type metricQuery struct {
		key   string
		title string
		unit  string
		count func(context.Context) (int64, error)
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	queries := []metricQuery{
		{key: "active_admins", title: "启用管理员", unit: "人", count: s.dashboard.CountActiveAdmins},
		{key: "active_roles", title: "启用角色", unit: "个", count: s.dashboard.CountActiveRoles},
		{key: "active_sessions", title: "活跃会话", unit: "个", count: func(ctx context.Context) (int64, error) {
			return s.dashboard.CountActiveSessions(ctx, now)
		}},
		{key: "audit_logs_today", title: "今日操作日志", unit: "次", count: func(ctx context.Context) (int64, error) {
			return s.dashboard.CountAuditLogsSince(ctx, todayStart)
		}},
	}

	metrics := make([]admindto.DashboardMetric, 0, len(queries))
	for _, query := range queries {
		value, err := query.count(ctx)
		if err != nil {
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

func (s *AdminDashboardService) businessMetrics(ctx context.Context) ([]admindto.DashboardBusinessMetric, error) {
	type businessMetricQuery struct {
		key              string
		title            string
		unit             string
		description      string
		targetPath       string
		targetPermission string
		severity         string
		count            func(context.Context) (int64, error)
	}

	queries := []businessMetricQuery{
		{
			key:              "pending_orders",
			title:            "待处理订单",
			unit:             "单",
			description:      "等待运营处理、交付或人工确认的订单",
			targetPath:       "/orders",
			targetPermission: "page.orders",
			severity:         "warning",
			count:            s.dashboard.CountPendingOrders,
		},
		{
			key:              "order_errors",
			title:            "交付异常订单",
			unit:             "单",
			description:      "真实支付后自动交付失败、需要人工处理的订单",
			targetPath:       "/orders",
			targetPermission: "page.orders",
			severity:         "error",
			count:            s.dashboard.CountOrderErrors,
		},
		{
			key:              "instance_errors",
			title:            "异常实例",
			unit:             "台",
			description:      "最近一次创建、同步或操作失败的实例",
			targetPath:       "/instances",
			targetPermission: "page.instances",
			severity:         "error",
			count:            s.dashboard.CountInstanceErrors,
		},
		{
			key:              "failed_async_tasks",
			title:            "失败异步任务",
			unit:             "个",
			description:      "Worker 已达到失败状态、可由管理端排查或重试的任务",
			targetPath:       "/async-tasks",
			targetPermission: "page.async-tasks",
			severity:         "error",
			count:            s.dashboard.CountFailedAsyncTasks,
		},
		{
			key:              "pending_tickets",
			title:            "待处理工单",
			unit:             "张",
			description:      "用户提交或回复后等待后台处理的工单",
			targetPath:       "/tickets",
			targetPermission: "page.tickets",
			severity:         "warning",
			count:            s.dashboard.CountPendingTickets,
		},
		{
			key:              "invoice_todo",
			title:            "待处理发票",
			unit:             "张",
			description:      "用户已提交或运营处理中、尚未完成开票的申请",
			targetPath:       "/invoices",
			targetPermission: "page.invoices",
			severity:         "warning",
			count:            s.dashboard.CountInvoiceTodo,
		},
		{
			key:              "payment_exceptions",
			title:            "支付异常",
			unit:             "笔",
			description:      "失败支付和处理中或失败退款的合计",
			targetPath:       "/payments",
			targetPermission: "page.payments",
			severity:         "error",
			count:            s.dashboard.CountPaymentExceptions,
		},
	}

	metrics := make([]admindto.DashboardBusinessMetric, 0, len(queries))
	for _, query := range queries {
		value, err := query.count(ctx)
		if err != nil {
			return nil, err
		}
		unit := query.unit
		targetPath := query.targetPath
		targetPermission := query.targetPermission
		metrics = append(metrics, admindto.DashboardBusinessMetric{
			Key:              query.key,
			Title:            query.title,
			Value:            value,
			Unit:             &unit,
			Description:      query.description,
			TargetPath:       &targetPath,
			TargetPermission: &targetPermission,
			Severity:         query.severity,
		})
	}

	return metrics, nil
}
