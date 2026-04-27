package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/api/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/middleware"
	"github.com/AeolianCloud/pveCloud/server/internal/services"
)

/**
 * RegisterAdminRoutes 注册管理端 API 路由。
 *
 * @param group 管理端路由分组
 * @param app API 应用依赖容器
 */
func RegisterAdminRoutes(group *gin.RouterGroup, app *bootstrap.App) {
	systemHandler := admin.NewSystemHandler()
	auditService := services.NewAdminAuditService(app.DB)
	authService := services.NewAdminAuthService(app.DB, app.Redis, app.Config.JWT, auditService)
	authHandler := admin.NewAuthHandler(authService)
	dashboardService := services.NewAdminDashboardService(app.DB)
	dashboardHandler := admin.NewDashboardHandler(dashboardService)
	adminUserService := services.NewAdminUserService(app.DB, auditService)
	adminUserHandler := admin.NewAdminUserHandler(adminUserService)
	adminRoleService := services.NewAdminRoleService(app.DB, auditService)
	adminRoleHandler := admin.NewAdminRoleHandler(adminRoleService)
	systemConfigService := services.NewSystemConfigService(app.DB, auditService)
	systemConfigHandler := admin.NewSystemConfigHandler(systemConfigService)

	group.GET("/ping", systemHandler.Ping)
	group.GET("/auth/captcha", authHandler.Captcha)
	group.POST("/auth/login", authHandler.Login)

	protected := group.Group("")
	protected.Use(middleware.AdminAuth(app.Config.JWT, app.DB))
	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.POST("/auth/refresh", authHandler.Refresh)
	protected.GET("/dashboard", middleware.AdminPermission("dashboard:view"), dashboardHandler.Show)
	protected.GET("/admin-users", middleware.AdminPermission("admin-user:view"), adminUserHandler.List)
	protected.POST("/admin-users", middleware.AdminPermission("admin-user:create"), adminUserHandler.Create)
	protected.GET("/admin-users/:id", middleware.AdminPermission("admin-user:view"), adminUserHandler.Detail)
	protected.PATCH("/admin-users/:id", middleware.AdminPermission("admin-user:update"), adminUserHandler.Update)
	protected.POST("/admin-users/:id/password", middleware.AdminPermission("admin-user:password-reset"), adminUserHandler.ResetPassword)
	protected.GET("/admin-roles", middleware.AdminPermission("admin-role:view"), adminRoleHandler.Roles)
	protected.POST("/admin-roles", middleware.AdminPermission("admin-role:create"), adminRoleHandler.CreateRole)
	protected.GET("/admin-roles/:id", middleware.AdminPermission("admin-role:view"), adminRoleHandler.RoleDetail)
	protected.PATCH("/admin-roles/:id", middleware.AdminPermission("admin-role:update"), adminRoleHandler.UpdateRole)
	protected.GET("/admin-permissions", middleware.AdminPermission("admin-role:view"), adminRoleHandler.Permissions)
	protected.GET("/system-configs", middleware.AdminPermission("system-config:view"), systemConfigHandler.Configs)
	protected.PATCH("/system-configs/:id", middleware.AdminPermission("system-config:update"), systemConfigHandler.Update)
}
