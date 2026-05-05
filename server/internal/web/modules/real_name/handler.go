package realname

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/web/middleware"
)

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

func (h *RealNameHandler) UploadFile(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	maxBytes, err := h.service.MaxUploadRequestBytes(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请选择上传文件"))
		return
	}
	defer file.Close()
	result, err := h.service.UploadFile(c.Request.Context(), userID, file, header)
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
