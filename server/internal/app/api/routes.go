package api

import (
	"github.com/gin-gonic/gin"

	adminrolehttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/adminrole"
	adminsessionhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/adminsession"
	adminuserhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/adminuser"
	audithttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/audit"
	adminauthhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/auth"
	dashboardhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/dashboard"
	fileattachmenthttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/fileattachment"
	adminmiddleware "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	adminorderhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/order"
	productcataloghttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/productcatalog"
	adminrealnamehttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/realname"
	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/system"
	systemconfighttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/systemconfig"
	webuserhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/webuser"
	webauthhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/auth"
	cataloghttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/catalog"
	webmiddleware "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
	weborderhttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/order"
	webrealnamehttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/realname"
	siteconfighttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/siteconfig"
	userprofilehttp "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/userprofile"
	mysqlcatalog "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/catalog"
	mysqlfile "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/file"
	mysqlsystemconfig "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/systemconfig"
	adminroleusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/adminrole"
	adminsessionusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/adminsession"
	adminuserusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/adminuser"
	auditusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"
	adminauthusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/auth"
	dashboardusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dashboard"
	fileattachmentusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/fileattachment"
	adminorderusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/order"
	productcatalogusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/productcatalog"
	adminrealnameusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/realname"
	systemconfigusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/systemconfig"
	webuserusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/webuser"
	webauthusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/auth"
	catalogusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/catalog"
	weborderusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/order"
	webrealnameusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/realname"
	siteconfigusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/siteconfig"
	userprofileusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/userprofile"
)

type AdminRouteSet struct {
	System         *system.SystemHandler
	Auth           *adminauthhttp.AuthHandler
	Dashboard      *dashboardhttp.DashboardHandler
	AdminUser      *adminuserhttp.AdminUserHandler
	AdminRole      *adminrolehttp.AdminRoleHandler
	AdminSession   *adminsessionhttp.AdminSessionHandler
	SystemConfig   *systemconfighttp.SystemConfigHandler
	WebUser        *webuserhttp.WebUserHandler
	FileAttachment *fileattachmenthttp.FileAttachmentHandler
	ProductCatalog *productcataloghttp.ProductCatalogHandler
	RealName       *adminrealnamehttp.RealNameHandler
	Order          *adminorderhttp.Handler
	Audit          *audithttp.AdminAuditHandler
	AuthMiddleware gin.HandlerFunc
}

type WebRouteSet struct {
	SiteConfig     *siteconfighttp.Handler
	Auth           *webauthhttp.UserAuthHandler
	UserProfile    *userprofilehttp.UserProfileHandler
	ProductCatalog *cataloghttp.Handler
	RealName       *webrealnamehttp.RealNameHandler
	Order          *weborderhttp.Handler
	AuthMiddleware gin.HandlerFunc
}

type RouteSets struct {
	Admin AdminRouteSet
	Web   WebRouteSet
}

func NewRouteSets(app *App) RouteSets {
	auditService := auditusecase.NewAdminAuditService(app.DB)
	adminAuthService := adminauthusecase.NewAdminAuthService(app.DB, app.Redis, app.Config.JWT, auditService)
	webAuthService := webauthusecase.NewUserAuthService(app.DB, app.Redis, app.Config.JWT, app.Config.Mail)

	siteConfigRepository := mysqlsystemconfig.NewRepository(app.DB)
	fileRepository := mysqlfile.NewRepository(app.DB)
	productCatalogRepository := mysqlcatalog.NewRepository(app.DB)
	webRealNameService := webrealnameusecase.NewRealNameService(app.DB, app.Redis)

	return RouteSets{
		Admin: AdminRouteSet{
			System:         system.NewSystemHandler(),
			Auth:           adminauthhttp.NewAuthHandler(adminAuthService),
			Dashboard:      dashboardhttp.NewDashboardHandler(dashboardusecase.NewAdminDashboardService(app.DB)),
			AdminUser:      adminuserhttp.NewAdminUserHandler(adminuserusecase.NewAdminUserService(app.DB, auditService)),
			AdminRole:      adminrolehttp.NewAdminRoleHandler(adminroleusecase.NewAdminRoleService(app.DB, auditService)),
			AdminSession:   adminsessionhttp.NewAdminSessionHandler(adminsessionusecase.NewAdminSessionService(app.DB, auditService)),
			SystemConfig:   systemconfighttp.NewSystemConfigHandler(systemconfigusecase.NewSystemConfigService(app.DB, auditService)),
			WebUser:        webuserhttp.NewWebUserHandler(webuserusecase.NewWebUserService(app.DB, auditService)),
			FileAttachment: fileattachmenthttp.NewFileAttachmentHandler(fileattachmentusecase.NewFileAttachmentService(app.DB, auditService, app.Config.Storage)),
			ProductCatalog: productcataloghttp.NewProductCatalogHandler(productcatalogusecase.NewProductCatalogService(app.DB, auditService)),
			RealName:       adminrealnamehttp.NewRealNameHandler(adminrealnameusecase.NewRealNameService(app.DB, app.Redis, auditService)),
			Order:          adminorderhttp.NewHandler(adminorderusecase.NewService(app.DB, auditService)),
			Audit:          audithttp.NewAdminAuditHandler(auditService, adminmiddleware.CurrentAdminPermissionCodes),
			AuthMiddleware: adminmiddleware.AdminAuth(adminAuthService),
		},
		Web: WebRouteSet{
			SiteConfig:     siteconfighttp.NewHandler(siteconfigusecase.NewSiteConfigService(siteConfigRepository, fileRepository, app.Config.Storage)),
			Auth:           webauthhttp.NewUserAuthHandler(webAuthService),
			UserProfile:    userprofilehttp.NewUserProfileHandler(userprofileusecase.NewUserProfileService(app.DB)),
			ProductCatalog: cataloghttp.NewHandler(catalogusecase.NewServerCatalogService(productCatalogRepository)),
			RealName:       webrealnamehttp.NewRealNameHandler(webRealNameService),
			Order:          weborderhttp.NewHandler(weborderusecase.NewService(app.DB, webRealNameService)),
			AuthMiddleware: webmiddleware.UserAuth(webAuthService),
		},
	}
}
