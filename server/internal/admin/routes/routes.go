package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/middleware"
	adminrole "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/admin_role"
	adminsession "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/admin_session"
	adminuser "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/admin_user"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/auth"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/dashboard"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/system"
	systemconfig "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/system_config"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
)

/**
 * RegisterAdminRoutes 注册管理端 API 路由。
 *
 * @param group 管理端路由分组
 * @param app API 应用依赖容器
 */
func RegisterAdminRoutes(group *gin.RouterGroup, app *bootstrap.App) {
	systemHandler := system.NewSystemHandler()
	auditService := audit.NewAdminAuditService(app.DB)
	auditHandler := audit.NewAdminAuditHandler(auditService, middleware.CurrentAdminPermissionCodes)
	authService := auth.NewAdminAuthService(app.DB, app.Redis, app.Config.JWT, auditService)
	authHandler := auth.NewAuthHandler(authService)
	dashboardService := dashboard.NewAdminDashboardService(app.DB)
	dashboardHandler := dashboard.NewDashboardHandler(dashboardService)
	adminUserService := adminuser.NewAdminUserService(app.DB, auditService)
	adminUserHandler := adminuser.NewAdminUserHandler(adminUserService)
	adminRoleService := adminrole.NewAdminRoleService(app.DB, auditService)
	adminRoleHandler := adminrole.NewAdminRoleHandler(adminRoleService)
	adminSessionService := adminsession.NewAdminSessionService(app.DB, auditService)
	adminSessionHandler := adminsession.NewAdminSessionHandler(adminSessionService)
	systemConfigService := systemconfig.NewSystemConfigService(app.DB, auditService)
	systemConfigHandler := systemconfig.NewSystemConfigHandler(systemConfigService)

	admin := group.Group("")
	admin.Use(middleware.AdminAuditContext())
	admin.GET("/ping", systemHandler.Ping)
	admin.GET("/auth/captcha", authHandler.Captcha)
	admin.POST("/auth/login", authHandler.Login)

	protected := admin.Group("")
	protected.Use(middleware.AdminAuth(app.Config.JWT, app.DB))
	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.POST("/auth/refresh", authHandler.Refresh)
	protected.GET("/dashboard", middleware.AdminPermission("page.dashboard"), dashboardHandler.Show)
	protected.GET("/admin-users", middleware.AdminPermission("page.system-settings.admin-users"), adminUserHandler.List)
	protected.POST("/admin-users", middleware.AdminPermission("admin-user:create"), adminUserHandler.Create)
	protected.GET("/admin-users/:id", middleware.AdminPermission("page.system-settings.admin-users"), adminUserHandler.Detail)
	protected.PATCH("/admin-users/:id", middleware.AdminPermission("admin-user:update"), adminUserHandler.Update)
	protected.POST("/admin-users/:id/password", middleware.AdminPermission("admin-user:password-reset"), adminUserHandler.ResetPassword)
	protected.GET("/admin-roles", middleware.AdminPermission("page.system-settings.admin-roles"), adminRoleHandler.Roles)
	protected.POST("/admin-roles", middleware.AdminPermission("admin-role:create"), adminRoleHandler.CreateRole)
	protected.GET("/admin-roles/:id", middleware.AdminPermission("page.system-settings.admin-roles"), adminRoleHandler.RoleDetail)
	protected.PATCH("/admin-roles/:id", middleware.AdminPermission("admin-role:update"), adminRoleHandler.UpdateRole)
	protected.GET("/admin-permissions", middleware.AdminPermission("page.system-settings.admin-roles"), adminRoleHandler.Permissions)
	protected.GET("/admin-sessions", middleware.AdminPermission("page.system-settings.admin-sessions"), adminSessionHandler.List)
	protected.PATCH("/admin-sessions/:session_id", middleware.AdminPermission("admin-session:revoke"), adminSessionHandler.Update)
	protected.GET("/audit-logs", middleware.AdminPermission("page.system-settings.audit-logs"), auditHandler.Logs)
	protected.GET("/system-configs", middleware.AdminPermission("page.system-settings.config"), systemConfigHandler.Configs)
	protected.PATCH("/system-configs/:id", middleware.AdminPermission("system-config:update"), systemConfigHandler.Update)
}
