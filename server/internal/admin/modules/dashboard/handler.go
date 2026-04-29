package dashboard

import (
	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
)

/**
 * DashboardHandler 处理管理端首页接口。
 */
type DashboardHandler struct {
	dashboardService *AdminDashboardService
}

/**
 * NewDashboardHandler 创建管理端首页接口处理器。
 *
 * @param dashboardService 管理端首页服务
 * @return *DashboardHandler 管理端首页接口处理器
 */
func NewDashboardHandler(dashboardService *AdminDashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

/**
 * Show 处理管理端首页数据查询。
 *
 * @route GET /admin-api/dashboard
 * @response 200 {"code":0,"message":"成功","data":{"admin":{"id":1,"username":"admin","display_name":"超级管理员","status":"active"},"role_ids":[1],"permission_codes":["dashboard:view"],"menus":[],"metrics":[]}}
 * @auth admin jwt, permission dashboard:view
 */
func (h *DashboardHandler) Show(c *gin.Context) {
	adminID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	result, err := h.dashboardService.Get(
		c.Request.Context(),
		adminID,
		middleware.CurrentAdminRoleIDs(c),
		middleware.CurrentAdminPermissionCodes(c),
		currentSessionOrEmpty(c),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

func currentSessionOrEmpty(c *gin.Context) admindto.SessionSummary {
	session, _ := middleware.CurrentAdminSession(c)
	return session
}
