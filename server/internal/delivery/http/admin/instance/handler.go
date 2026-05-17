package instance

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/httputil"
	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	instanceusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/instance"
)

type Handler struct{ service *instanceusecase.Service }

func NewHandler(service *instanceusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Mappings(c *gin.Context) {
	var query admindto.InstanceMappingListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.ListMappings(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) CreateMapping(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var req admindto.InstanceMappingRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.CreateMapping(c.Request.Context(), operatorID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) UpdateMapping(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.InstanceMappingRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.UpdateMapping(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Nodes(c *gin.Context) {
	result, err := h.service.Nodes(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Node(c *gin.Context) {
	result, err := h.service.Node(c.Request.Context(), c.Param("node"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) NodeVMs(c *gin.Context) {
	result, err := h.service.NodeVMs(c.Request.Context(), c.Param("node"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Storage(c *gin.Context) {
	result, err := h.service.Storage(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ProvisionOrder(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.Provision(c.Request.Context(), operatorID, c.Param("order_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) List(c *gin.Context) {
	var query admindto.InstanceListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Detail(c *gin.Context) {
	result, err := h.service.Detail(c.Request.Context(), c.Param("instance_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Start(c *gin.Context) {
	h.operate(c, h.service.Start)
}

func (h *Handler) Stop(c *gin.Context) {
	h.operate(c, h.service.Stop)
}

func (h *Handler) Release(c *gin.Context) {
	h.operate(c, h.service.Release)
}

func (h *Handler) Sync(c *gin.Context) {
	h.operate(c, h.service.Sync)
}

func (h *Handler) operate(c *gin.Context, fn func(context.Context, uint64, string) (admindto.InstanceDetail, error)) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := fn(c.Request.Context(), operatorID, c.Param("instance_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func currentAdminID(c *gin.Context) (uint64, bool) {
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return operatorID, true
}

func bindQuery(c *gin.Context, target any) bool {
	if err := c.ShouldBindQuery(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return false
	}
	if err := validator.Struct(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return false
	}
	return true
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return false
	}
	if err := validator.Struct(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return false
	}
	return true
}
