package invoice

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	invoiceusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/invoice"
)

type Handler struct{ service *invoiceusecase.Service }

func NewHandler(service *invoiceusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) EligibleOrders(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var query webdto.InvoiceEligibleOrderQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.EligibleOrders(c.Request.Context(), userID, query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req webdto.InvoiceCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var query webdto.InvoiceListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.List(c.Request.Context(), userID, query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Detail(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Detail(c.Request.Context(), userID, c.Param("invoice_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Cancel(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req webdto.InvoiceCancelRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.Cancel(c.Request.Context(), userID, c.Param("invoice_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Download(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	path, mimeType, filename, err := h.service.DownloadPath(c.Request.Context(), userID, c.Param("invoice_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", urlEncodeFilename(filename)))
	c.Header("Cache-Control", "no-store, private")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.File(path)
}

func currentUserID(c *gin.Context) (uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return userID, true
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

func urlEncodeFilename(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
