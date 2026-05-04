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
	fileattachment "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/file_attachment"
	productcatalog "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/product_catalog"
	realname "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/real_name"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/system"
	systemconfig "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/system_config"
	webuser "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/web_user"
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
	webUserService := webuser.NewWebUserService(app.DB, auditService)
	webUserHandler := webuser.NewWebUserHandler(webUserService)
	fileAttachmentService := fileattachment.NewFileAttachmentService(app.DB, auditService, app.Config.Storage)
	fileAttachmentHandler := fileattachment.NewFileAttachmentHandler(fileAttachmentService)
	productCatalogService := productcatalog.NewProductCatalogService(app.DB, auditService)
	productCatalogHandler := productcatalog.NewProductCatalogHandler(productCatalogService)
	realNameService := realname.NewRealNameService(app.DB, auditService)
	realNameHandler := realname.NewRealNameHandler(realNameService)

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
	protected.GET("/users", middleware.AdminPermission("page.web-users"), webUserHandler.Users)
	protected.POST("/users", middleware.AdminPermission("web-user:create"), webUserHandler.CreateUser)
	protected.GET("/users/:id", middleware.AdminPermission("page.web-users"), webUserHandler.UserDetail)
	protected.PATCH("/users/:id", middleware.AdminPermission("web-user:update"), webUserHandler.UpdateUser)
	protected.POST("/users/:id/password", middleware.AdminPermission("web-user:password-reset"), webUserHandler.ResetPassword)
	protected.GET("/user-sessions", middleware.AdminPermission("page.web-user-sessions"), webUserHandler.Sessions)
	protected.PATCH("/user-sessions/:session_id", middleware.AdminPermission("web-user-session:revoke"), webUserHandler.RevokeSession)
	protected.GET("/real-name-applications", middleware.AdminPermission("page.real-name-management"), realNameHandler.Applications)
	protected.GET("/real-name-applications/:id", middleware.AdminPermission("page.real-name-management"), realNameHandler.Detail)
	protected.POST("/real-name-applications/:id/review", middleware.AdminPermission("real-name:review"), realNameHandler.Review)
	protected.GET("/products", middleware.AdminPermission("page.products"), productCatalogHandler.Products)
	protected.POST("/products", middleware.AdminPermission("product:create"), productCatalogHandler.CreateProduct)
	protected.PUT("/products/:id", middleware.AdminPermission("product:update"), productCatalogHandler.UpdateProduct)
	protected.PATCH("/products/:id/status", middleware.AdminPermission("product:publish"), productCatalogHandler.UpdateProductStatus)
	protected.GET("/product-plans", middleware.AdminPermission("page.products"), productCatalogHandler.Plans)
	protected.POST("/product-plans", middleware.AdminPermission("product:create"), productCatalogHandler.CreatePlan)
	protected.PUT("/product-plans/:id", middleware.AdminPermission("product:update"), productCatalogHandler.UpdatePlan)
	protected.PATCH("/product-plans/:id/status", middleware.AdminPermission("product:publish"), productCatalogHandler.UpdatePlanStatus)
	protected.GET("/product-plans/:id/prices", middleware.AdminPermission("page.products"), productCatalogHandler.PlanPrices)
	protected.PUT("/product-plans/:id/prices", middleware.AdminPermission("product:update"), productCatalogHandler.UpdatePlanPrices)
	protected.GET("/product-plans/:id/regions", middleware.AdminPermission("page.products"), productCatalogHandler.PlanRegions)
	protected.PUT("/product-plans/:id/regions", middleware.AdminPermission("product:update"), productCatalogHandler.UpdatePlanRegions)
	protected.GET("/product-plans/:id/os-templates", middleware.AdminPermission("page.products"), productCatalogHandler.PlanOSTemplates)
	protected.PUT("/product-plans/:id/os-templates", middleware.AdminPermission("product:update"), productCatalogHandler.UpdatePlanOSTemplates)
	protected.GET("/sales-regions", middleware.AdminPermission("page.products"), productCatalogHandler.SalesRegions)
	protected.POST("/sales-regions", middleware.AdminPermission("product:create"), productCatalogHandler.CreateSalesRegion)
	protected.PUT("/sales-regions/:id", middleware.AdminPermission("product:update"), productCatalogHandler.UpdateSalesRegion)
	protected.GET("/server-os-templates", middleware.AdminPermission("page.products"), productCatalogHandler.ServerOSTemplates)
	protected.POST("/server-os-templates", middleware.AdminPermission("product:create"), productCatalogHandler.CreateServerOSTemplate)
	protected.PUT("/server-os-templates/:id", middleware.AdminPermission("product:update"), productCatalogHandler.UpdateServerOSTemplate)
	protected.POST("/files/upload", middleware.AdminPermission("file:upload"), fileAttachmentHandler.Upload)
	protected.GET("/files", middleware.AdminPermission("page.file-management"), fileAttachmentHandler.List)
	protected.GET("/files/:id", middleware.AdminPermission("page.file-management"), fileAttachmentHandler.Detail)
	protected.GET("/files/:id/download", middleware.AdminPermission("page.file-management"), fileAttachmentHandler.Download)
	protected.GET("/files/:id/references", middleware.AdminPermission("page.file-management"), fileAttachmentHandler.Reference)
	protected.DELETE("/files/:id", middleware.AdminPermission("file:delete"), fileAttachmentHandler.Delete)
}
