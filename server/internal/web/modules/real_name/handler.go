package realname

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	realnameintegration "github.com/AeolianCloud/pveCloud/server/internal/platform/integrations/realname"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/web/middleware"
)

const providerCallbackMaxBodyBytes = 1 << 20

type RealNameHandler struct {
	service *RealNameService
}

func NewRealNameHandler(service *RealNameService) *RealNameHandler {
	return &RealNameHandler{service: service}
}

func (h *RealNameHandler) Status(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	result, err := h.service.Status(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *RealNameHandler) Submit(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	var req webdto.RealNameSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.service.Submit(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *RealNameHandler) Sync(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	var req webdto.RealNameSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.service.Sync(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *RealNameHandler) ProviderCallback(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, providerCallbackMaxBodyBytes)
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
	if err := c.Request.ParseForm(); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	request := realnameintegration.CallbackRequest{
		Method:      c.Request.Method,
		Headers:     c.Request.Header.Clone(),
		Query:       c.Request.URL.Query(),
		Form:        c.Request.PostForm,
		RawBody:     rawBody,
		ContentType: c.ContentType(),
	}
	if err := h.service.ProviderCallback(c.Request.Context(), c.Param("provider"), request); err != nil {
		response.Error(c, err)
		return
	}
	c.String(200, "success")
}
