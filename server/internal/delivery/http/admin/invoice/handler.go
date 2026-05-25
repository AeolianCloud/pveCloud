package invoice

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	invoiceusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/invoice"
)

type Handler struct{ service *invoiceusecase.Service }

func NewHandler(service *invoiceusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *gin.Context) {
	var query admindto.InvoiceListQuery
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
	result, err := h.service.Detail(c.Request.Context(), c.Param("invoice_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Accept(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.Accept(c.Request.Context(), operatorID, c.Param("invoice_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Reject(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var req admindto.InvoiceRejectRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.Reject(c.Request.Context(), operatorID, c.Param("invoice_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Issue(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var req admindto.InvoiceIssueRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.Issue(c.Request.Context(), operatorID, c.Param("invoice_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) UpdateAdminNote(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var req admindto.InvoiceAdminNoteRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.UpdateAdminNote(c.Request.Context(), operatorID, c.Param("invoice_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Download(c *gin.Context) {
	path, mimeType, filename, err := h.service.DownloadPath(c.Request.Context(), c.Param("invoice_no"))
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

func urlEncodeFilename(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
