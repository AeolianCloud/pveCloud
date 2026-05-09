package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/app/api"
	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
)

/**
 * RegisterAdminRoutes 注册管理端 API 路由。
 *
 * @param group 管理端路由分组
 * @param app API 应用依赖容器
 */
func RegisterAdminRoutes(group *gin.RouterGroup, app *api.App) {
	routes := app.Routes.Admin

	admin := group.Group("")
	admin.Use(middleware.AdminAuditContext())
	admin.GET("/ping", routes.System.Ping)
	admin.GET("/auth/captcha", routes.Auth.Captcha)
	admin.POST("/auth/login", routes.Auth.Login)

	protected := admin.Group("")
	protected.Use(routes.AuthMiddleware)
	protected.GET("/auth/me", routes.Auth.Me)
	protected.POST("/auth/logout", routes.Auth.Logout)
	protected.POST("/auth/refresh", routes.Auth.Refresh)
	protected.GET("/dashboard", middleware.AdminPermission("page.dashboard"), routes.Dashboard.Show)
	protected.GET("/admin-users", middleware.AdminPermission("page.system-settings.admin-users"), routes.AdminUser.List)
	protected.POST("/admin-users", middleware.AdminPermission("admin-user:create"), routes.AdminUser.Create)
	protected.GET("/admin-users/:id", middleware.AdminPermission("page.system-settings.admin-users"), routes.AdminUser.Detail)
	protected.PATCH("/admin-users/:id", middleware.AdminPermission("admin-user:update"), routes.AdminUser.Update)
	protected.POST("/admin-users/:id/password", middleware.AdminPermission("admin-user:password-reset"), routes.AdminUser.ResetPassword)
	protected.GET("/admin-roles", middleware.AdminPermission("page.system-settings.admin-roles"), routes.AdminRole.Roles)
	protected.POST("/admin-roles", middleware.AdminPermission("admin-role:create"), routes.AdminRole.CreateRole)
	protected.GET("/admin-roles/:id", middleware.AdminPermission("page.system-settings.admin-roles"), routes.AdminRole.RoleDetail)
	protected.PATCH("/admin-roles/:id", middleware.AdminPermission("admin-role:update"), routes.AdminRole.UpdateRole)
	protected.GET("/admin-permissions", middleware.AdminPermission("page.system-settings.admin-roles"), routes.AdminRole.Permissions)
	protected.GET("/admin-sessions", middleware.AdminPermission("page.system-settings.admin-sessions"), routes.AdminSession.List)
	protected.PATCH("/admin-sessions/:session_id", middleware.AdminPermission("admin-session:revoke"), routes.AdminSession.Update)
	protected.GET("/audit-logs", middleware.AdminPermission("page.system-settings.audit-logs"), routes.Audit.Logs)
	protected.GET("/system-configs", middleware.AdminPermission("page.system-settings.config"), routes.SystemConfig.Configs)
	protected.PATCH("/system-configs/:id", middleware.AdminPermission("system-config:update"), routes.SystemConfig.Update)
	protected.GET("/users", middleware.AdminPermission("page.web-users"), routes.WebUser.Users)
	protected.POST("/users", middleware.AdminPermission("web-user:create"), routes.WebUser.CreateUser)
	protected.GET("/users/:id", middleware.AdminPermission("page.web-users"), routes.WebUser.UserDetail)
	protected.PATCH("/users/:id", middleware.AdminPermission("web-user:update"), routes.WebUser.UpdateUser)
	protected.POST("/users/:id/password", middleware.AdminPermission("web-user:password-reset"), routes.WebUser.ResetPassword)
	protected.GET("/user-sessions", middleware.AdminPermission("page.web-user-sessions"), routes.WebUser.Sessions)
	protected.PATCH("/user-sessions/:session_id", middleware.AdminPermission("web-user-session:revoke"), routes.WebUser.RevokeSession)
	protected.GET("/real-name-applications", middleware.AdminPermission("page.real-name-management"), routes.RealName.Applications)
	protected.GET("/real-name-applications/:id", middleware.AdminPermission("page.real-name-management"), routes.RealName.Detail)
	protected.POST("/real-name-applications/:id/sync", middleware.AdminPermission("real-name:sync"), routes.RealName.Sync)
	protected.POST("/real-name-applications/:id/review", middleware.AdminPermission("real-name:review"), routes.RealName.Review)
	protected.GET("/products", middleware.AdminPermission("page.products"), routes.ProductCatalog.Products)
	protected.POST("/products", middleware.AdminPermission("product:create"), routes.ProductCatalog.CreateProduct)
	protected.PUT("/products/:id", middleware.AdminPermission("product:update"), routes.ProductCatalog.UpdateProduct)
	protected.PATCH("/products/:id/status", middleware.AdminPermission("product:publish"), routes.ProductCatalog.UpdateProductStatus)
	protected.GET("/product-plans", middleware.AdminPermission("page.products"), routes.ProductCatalog.Plans)
	protected.POST("/product-plans", middleware.AdminPermission("product:create"), routes.ProductCatalog.CreatePlan)
	protected.PUT("/product-plans/:id", middleware.AdminPermission("product:update"), routes.ProductCatalog.UpdatePlan)
	protected.PATCH("/product-plans/:id/status", middleware.AdminPermission("product:publish"), routes.ProductCatalog.UpdatePlanStatus)
	protected.GET("/product-plans/:id/prices", middleware.AdminPermission("page.products"), routes.ProductCatalog.PlanPrices)
	protected.PUT("/product-plans/:id/prices", middleware.AdminPermission("product:update"), routes.ProductCatalog.UpdatePlanPrices)
	protected.GET("/product-plans/:id/regions", middleware.AdminPermission("page.products"), routes.ProductCatalog.PlanRegions)
	protected.PUT("/product-plans/:id/regions", middleware.AdminPermission("product:update"), routes.ProductCatalog.UpdatePlanRegions)
	protected.GET("/product-plans/:id/os-templates", middleware.AdminPermission("page.products"), routes.ProductCatalog.PlanOSTemplates)
	protected.PUT("/product-plans/:id/os-templates", middleware.AdminPermission("product:update"), routes.ProductCatalog.UpdatePlanOSTemplates)
	protected.GET("/sales-regions", middleware.AdminPermission("page.products"), routes.ProductCatalog.SalesRegions)
	protected.POST("/sales-regions", middleware.AdminPermission("product:create"), routes.ProductCatalog.CreateSalesRegion)
	protected.PUT("/sales-regions/:id", middleware.AdminPermission("product:update"), routes.ProductCatalog.UpdateSalesRegion)
	protected.GET("/server-os-templates", middleware.AdminPermission("page.products"), routes.ProductCatalog.ServerOSTemplates)
	protected.POST("/server-os-templates", middleware.AdminPermission("product:create"), routes.ProductCatalog.CreateServerOSTemplate)
	protected.PUT("/server-os-templates/:id", middleware.AdminPermission("product:update"), routes.ProductCatalog.UpdateServerOSTemplate)
	protected.POST("/files/upload", middleware.AdminPermission("file:upload"), routes.FileAttachment.Upload)
	protected.GET("/files", middleware.AdminPermission("page.file-management"), routes.FileAttachment.List)
	protected.GET("/files/:id", middleware.AdminPermission("page.file-management"), routes.FileAttachment.Detail)
	protected.GET("/files/:id/download", middleware.AdminPermission("page.file-management"), routes.FileAttachment.Download)
	protected.GET("/files/:id/references", middleware.AdminPermission("page.file-management"), routes.FileAttachment.Reference)
	protected.DELETE("/files/:id", middleware.AdminPermission("file:delete"), routes.FileAttachment.Delete)
}
