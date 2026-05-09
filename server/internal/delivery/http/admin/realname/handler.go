package realname

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/httputil"
	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	adminrealname "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/realname"
)

type RealNameHandler struct {
	service *adminrealname.RealNameService
}

func NewRealNameHandler(service *adminrealname.RealNameService) *RealNameHandler {
	return &RealNameHandler{service: service}
}

func (h *RealNameHandler) Applications(c *gin.Context) {
	var query admindto.RealNameApplicationListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.service.Applications(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *RealNameHandler) Detail(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	result, err := h.service.Detail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *RealNameHandler) Sync(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	result, err := h.service.Sync(c.Request.Context(), operatorID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *RealNameHandler) Review(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	var req admindto.RealNameReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.service.Review(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
