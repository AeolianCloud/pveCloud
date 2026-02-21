// internal/router/router.go
// 路由注册，组装所有 handler。
package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	handleradmin      "pvecloud/backend/internal/handler/admin"
	handlerauth       "pvecloud/backend/internal/handler/auth"
	handlerloginlog   "pvecloud/backend/internal/handler/loginlog"
	handleroplog      "pvecloud/backend/internal/handler/oplog"
	handlerpermission "pvecloud/backend/internal/handler/permission"
	handlerrole       "pvecloud/backend/internal/handler/role"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/session"
	"pvecloud/backend/internal/security"
	svcadmin      "pvecloud/backend/internal/service/admin"
	svcauth       "pvecloud/backend/internal/service/auth"
	svcloginlog   "pvecloud/backend/internal/service/loginlog"
	svcoplog      "pvecloud/backend/internal/service/oplog"
	svcpermission "pvecloud/backend/internal/service/permission"
	svcrole       "pvecloud/backend/internal/service/role"
	"pvecloud/backend/internal/config"
	"pvecloud/backend/pkg/response"
)

// New 构建路由引擎，注入所有依赖。
func New(db *gorm.DB, log *zap.Logger, cfg *config.Config, sessStore session.Store, loginGuard *security.LoginGuard) *gin.Engine {
	if sessStore == nil {
		// 会话存储是登录/鉴权链路的核心依赖：缺失会导致登录后无法鉴权与刷新
		log.Fatal("会话存储未初始化（Redis）")
	}

	r := gin.New()

	// 全局中间件（顺序固定：Recovery → CORS → Logger）
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(log))

	// 404 自定义 JSON 响应，避免返回 HTML
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c)
	})
	r.NoMethod(func(c *gin.Context) {
		// 统一规范：HTTP 状态码只用 200/401/403/404/500
		// 方法不允许也按 404 返回，避免 405 带来前端额外分支处理
		response.NotFound(c)
	})

	// 健康检查（无需鉴权）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── 初始化 service ────────────────────────────────────
	authSvc       := svcauth.New(db, cfg.JWT.Secret, cfg.JWT.ExpireHours, cfg.JWT.RefreshExpireHours, sessStore)
	adminSvc      := svcadmin.New(db)
	roleSvc       := svcrole.New(db)
	permissionSvc := svcpermission.New(db)
	loginlogSvc   := svcloginlog.New(db)
	oplogSvc      := svcoplog.New(db)

	// ── 初始化 handler ────────────────────────────────────
	authHandler       := handlerauth.New(authSvc, loginGuard)
	adminHandler      := handleradmin.New(adminSvc)
	roleHandler       := handlerrole.New(roleSvc)
	permissionHandler := handlerpermission.New(permissionSvc)
	loginlogHandler   := handlerloginlog.New(loginlogSvc)
	oplogHandler      := handleroplog.New(oplogSvc)

	// ── 公开路由（无需 JWT）──────────────────────────────
	open := r.Group("/api/v1/auth")
	{
		open.POST("/login", authHandler.Login)
		open.POST("/refresh", authHandler.Refresh)
	}

	// ── JWT 保护路由 ──────────────────────────────────────
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg.JWT.Secret, sessStore))
	{
		// 退出登录（撤销会话，使 token 立即失效）
		api.POST("/auth/logout", authHandler.Logout)

		api.GET("/profile", authHandler.Profile)

		// 管理员账号管理
		adminUsers := api.Group("/admin-users")
		{
			adminUsers.GET("",              middleware.RequirePermission(db, "admin:list"),   adminHandler.List)
			adminUsers.POST("",             middleware.RequirePermission(db, "admin:create"), middleware.WriteOpLog(db, "admin", "create"), adminHandler.Create)
			adminUsers.PUT("/:id",          middleware.RequirePermission(db, "admin:update"), middleware.WriteOpLog(db, "admin", "update"), adminHandler.Update)
			adminUsers.PATCH("/:id/status", middleware.RequirePermission(db, "admin:status"), middleware.WriteOpLog(db, "admin", "set_status"), adminHandler.SetStatus)
			adminUsers.DELETE("/:id",       middleware.RequirePermission(db, "admin:delete"), middleware.WriteOpLog(db, "admin", "delete"), adminHandler.Delete)
		}

		// 角色管理
		roles := api.Group("/roles")
		{
			roles.GET("",                    middleware.RequirePermission(db, "role:list"),   roleHandler.List)
			roles.GET("/:id",                middleware.RequirePermission(db, "role:list"),   roleHandler.GetByID)
			roles.POST("",                   middleware.RequirePermission(db, "role:create"), middleware.WriteOpLog(db, "role", "create"), roleHandler.Create)
			roles.PUT("/:id",                middleware.RequirePermission(db, "role:update"), middleware.WriteOpLog(db, "role", "update"), roleHandler.Update)
			roles.DELETE("/:id",             middleware.RequirePermission(db, "role:delete"), middleware.WriteOpLog(db, "role", "delete"), roleHandler.Delete)
			roles.PUT("/:id/permissions",    middleware.RequirePermission(db, "role:assign"), middleware.WriteOpLog(db, "role", "assign_permissions"), roleHandler.AssignPermissions)
		}

		// 权限列表（只读，按 group 分组）
		api.GET("/permissions", middleware.RequirePermission(db, "role:list"), permissionHandler.ListGrouped)

		// 登录日志（只读）
		api.GET("/login-logs", middleware.RequirePermission(db, "log:list"), loginlogHandler.List)

		// 操作日志（只读）
		api.GET("/op-logs", middleware.RequirePermission(db, "op:list"), oplogHandler.List)
	}

	return r
}
