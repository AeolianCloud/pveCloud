package fileattachment

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
)

/**
 * FileAttachmentHandler 处理文件上传与附件管理接口。
 */
type FileAttachmentHandler struct {
	fileAttachmentService *FileAttachmentService
}

/**
 * NewFileAttachmentHandler 创建文件附件接口处理器。
 *
 * @param fileAttachmentService 文件附件服务
 * @return *FileAttachmentHandler 文件附件接口处理器
 */
func NewFileAttachmentHandler(fileAttachmentService *FileAttachmentService) *FileAttachmentHandler {
	return &FileAttachmentHandler{fileAttachmentService: fileAttachmentService}
}

/**
 * Upload 上传文件。
 *
 * @route POST /admin-api/files/upload
 * @multipart file
 * @response 200 {"code":0,"message":"成功","data":{"id":1,"original_name":"photo.jpg"}}
 * @auth admin jwt, permission file:upload
 */
func (h *FileAttachmentHandler) Upload(c *gin.Context) {
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.fileAttachmentService.maxUploadRequestBytes())
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请选择要上传的文件"))
		return
	}
	defer file.Close()

	result, err := h.fileAttachmentService.Upload(c.Request.Context(), operatorID, file, header)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * List 分页查询文件列表。
 *
 * @route GET /admin-api/files
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission page.file-management
 */
func (h *FileAttachmentHandler) List(c *gin.Context) {
	var query admindto.FileListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.fileAttachmentService.List(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Detail 查看文件详情。
 *
 * @route GET /admin-api/files/{id}
 * @response 200 {"code":0,"message":"成功","data":{"id":1,"original_name":"photo.jpg"}}
 * @auth admin jwt, permission page.file-management
 */
func (h *FileAttachmentHandler) Detail(c *gin.Context) {
	id, ok := support.AdminPathID(c)
	if !ok {
		return
	}

	result, err := h.fileAttachmentService.Detail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Reference 查看文件引用关系。
 *
 * @route GET /admin-api/files/{id}/references
 */
func (h *FileAttachmentHandler) Reference(c *gin.Context) {
	id, ok := support.AdminPathID(c)
	if !ok {
		return
	}

	result, err := h.fileAttachmentService.ReferenceResponse(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Download 下载或预览文件。
 *
 * @route GET /admin-api/files/{id}/download
 */
func (h *FileAttachmentHandler) Download(c *gin.Context) {
	id, ok := support.AdminPathID(c)
	if !ok {
		return
	}

	path, mimeType, filename, err := h.fileAttachmentService.DownloadPath(c.Request.Context(), id)
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

/**
 * Delete 删除文件。
 *
 * @route DELETE /admin-api/files/{id}
 * @response 200 {"code":0,"message":"成功","data":null}
 * @auth admin jwt, permission file:delete
 */
func (h *FileAttachmentHandler) Delete(c *gin.Context) {
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	id, ok := support.AdminPathID(c)
	if !ok {
		return
	}

	if err := h.fileAttachmentService.Delete(c.Request.Context(), operatorID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func isPreviewableMime(value string) bool {
	return strings.HasPrefix(value, "image/") || value == "application/pdf"
}

func urlEncodeFilename(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
