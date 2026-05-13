package ticket

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	ticketusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/ticket"
)

type Handler struct{ service *ticketusecase.Service }

func NewHandler(service *ticketusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var query webdto.TicketListQuery
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

func (h *Handler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.service.MaxUploadRequestBytes())
	req, headers, ok := bindMultipartCreate(c)
	if !ok {
		return
	}
	defer cleanupMultipartForm(c)
	result, err := h.service.Create(c.Request.Context(), userID, req, headers)
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
	result, err := h.service.Detail(c.Request.Context(), userID, c.Param("ticket_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Reply(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.service.MaxUploadRequestBytes())
	req, headers, ok := bindMultipartMessage(c)
	if !ok {
		return
	}
	defer cleanupMultipartForm(c)
	result, err := h.service.Reply(c.Request.Context(), userID, c.Param("ticket_no"), req, headers)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Close(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req webdto.TicketCloseRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.Close(c.Request.Context(), userID, c.Param("ticket_no"), req)
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
	fileID, ok := pathUint64(c, "file_id")
	if !ok {
		return
	}
	path, mimeType, filename, err := h.service.DownloadPath(c.Request.Context(), userID, c.Param("ticket_no"), fileID)
	if err != nil {
		response.Error(c, err)
		return
	}
	contentDisposition := "inline"
	if !isPreviewableMime(mimeType) {
		contentDisposition = "attachment"
	}
	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename*=UTF-8''%s", contentDisposition, urlEncodeFilename(filename)))
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

func bindMultipartCreate(c *gin.Context) (webdto.TicketCreateRequest, []*multipart.FileHeader, bool) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return webdto.TicketCreateRequest{}, nil, false
	}
	req := webdto.TicketCreateRequest{
		Title:    c.PostForm("title"),
		Category: c.PostForm("category"),
		Priority: c.PostForm("priority"),
		Content:  c.PostForm("content"),
		OrderNo:  c.PostForm("order_no"),
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return webdto.TicketCreateRequest{}, nil, false
	}
	return req, multipartFiles(c), true
}

func bindMultipartMessage(c *gin.Context) (webdto.TicketMessageRequest, []*multipart.FileHeader, bool) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return webdto.TicketMessageRequest{}, nil, false
	}
	req := webdto.TicketMessageRequest{Content: c.PostForm("content")}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return webdto.TicketMessageRequest{}, nil, false
	}
	return req, multipartFiles(c), true
}

func multipartFiles(c *gin.Context) []*multipart.FileHeader {
	if c.Request.MultipartForm == nil || c.Request.MultipartForm.File == nil {
		return nil
	}
	files := c.Request.MultipartForm.File["attachments"]
	if len(files) == 0 {
		files = c.Request.MultipartForm.File["attachments[]"]
	}
	return files
}

func cleanupMultipartForm(c *gin.Context) {
	if c.Request.MultipartForm != nil {
		_ = c.Request.MultipartForm.RemoveAll()
	}
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

func pathUint64(c *gin.Context, name string) (uint64, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		response.Error(c, apperrors.ErrValidation.WithMessage("路径参数格式错误"))
		return 0, false
	}
	return value, true
}

func isPreviewableMime(value string) bool {
	return strings.HasPrefix(value, "image/") || value == "application/pdf"
}

func urlEncodeFilename(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
