package fileattachment

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	fileAttachmentObjectType = "file_attachment"
	fileUploadAction         = "file.upload"
	fileDeleteAction         = "file.delete"
	multipartOverheadBytes   = int64(1 << 20)
)

/**
 * FileAttachmentService 处理文件上传与附件管理。
 */
type FileAttachmentService struct {
	db           *gorm.DB
	auditService *AdminAuditService
	config       bootstrap.StorageConfig
}

/**
 * NewFileAttachmentService 创建文件附件服务。
 *
 * @param db 数据库连接
 * @param auditService 后台审计服务
 * @param config 存储配置
 * @return *FileAttachmentService 文件附件服务
 */
func NewFileAttachmentService(db *gorm.DB, auditService *AdminAuditService, config bootstrap.StorageConfig) *FileAttachmentService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &FileAttachmentService{
		db:           db,
		auditService: auditService,
		config:       config,
	}
}

/**
 * Upload 上传文件。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param file 上传的文件
 * @param header 文件头信息
 * @return admindto.FileUploadResponse 上传结果
 * @return error 上传失败原因
 */
func (s *FileAttachmentService) Upload(ctx context.Context, operatorID uint64, file multipart.File, header *multipart.FileHeader) (admindto.FileUploadResponse, error) {
	originalName := sanitizeOriginalName(header.Filename)
	if originalName == "" {
		return admindto.FileUploadResponse{}, apperrors.ErrValidation.WithMessage("文件名不能为空")
	}

	// 校验文件大小
	if header.Size > s.config.MaxSize {
		return admindto.FileUploadResponse{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", s.config.MaxSize))
	}
	sniff := make([]byte, 512)
	sniffSize, err := io.ReadFull(file, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return admindto.FileUploadResponse{}, apperrors.ErrInternal.WithMessage("读取文件失败")
	}
	if sniffSize == 0 {
		return admindto.FileUploadResponse{}, apperrors.ErrValidation.WithMessage("文件内容不能为空")
	}
	sniff = sniff[:sniffSize]
	if int64(sniffSize) > s.config.MaxSize {
		return admindto.FileUploadResponse{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", s.config.MaxSize))
	}

	// 安全校验：文件名、扩展名、声明 MIME 和真实文件头必须一致。
	mimeType := detectMimeType(sniff)
	if err := s.validateFile(originalName, header, mimeType); err != nil {
		return admindto.FileUploadResponse{}, err
	}

	// 生成 UUID 作为存储文件名
	uuid, err := generateUUID()
	if err != nil {
		return admindto.FileUploadResponse{}, apperrors.ErrInternal.WithMessage("生成文件名失败")
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext != "" {
		ext = ext[1:] // 移除点号
	}

	// 构建存储路径
	now := time.Now()
	storagePath := filepath.Join(
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		uuid+"."+ext,
	)
	absolutePath := filepath.Join(s.config.LocalPath, storagePath)

	// 确保目录存在
	dir := filepath.Dir(absolutePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return admindto.FileUploadResponse{}, apperrors.ErrInternal.WithMessage("创建存储目录失败")
	}

	// 流式写入文件并同步计算 SHA-256，避免按文件大小整体占用内存。
	out, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return admindto.FileUploadResponse{}, apperrors.ErrInternal.WithMessage("保存文件失败")
	}
	hash := sha256.New()
	writer := io.MultiWriter(out, hash)
	writtenBytes, err := writer.Write(sniff)
	written := int64(writtenBytes)
	if err == nil {
		var copied int64
		copied, err = io.Copy(writer, io.LimitReader(file, s.config.MaxSize-written+1))
		written += copied
	}
	closeErr := out.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(absolutePath)
		return admindto.FileUploadResponse{}, apperrors.ErrInternal.WithMessage("保存文件失败")
	}
	if written > s.config.MaxSize {
		_ = os.Remove(absolutePath)
		return admindto.FileUploadResponse{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", s.config.MaxSize))
	}
	checksumHex := hex.EncodeToString(hash.Sum(nil))

	// 创建数据库记录
	attachment := models.FileAttachment{
		OriginalName:  originalName,
		StoredName:    uuid + "." + ext,
		MimeType:      mimeType,
		Extension:     ext,
		Size:          uint64(written),
		StoragePath:   storagePath,
		StorageDriver: "local",
		Checksum:      checksumHex,
		UploaderID:    operatorID,
		Status:        "active",
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&attachment).Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     fileUploadAction,
			ObjectType: fileAttachmentObjectType,
			ObjectID:   textutil.Uint64String(attachment.ID),
			AfterData: map[string]any{
				"id":            attachment.ID,
				"original_name": attachment.OriginalName,
				"mime_type":     attachment.MimeType,
				"size":          attachment.Size,
				"checksum":      attachment.Checksum,
			},
			Remark: "上传文件",
		})
	}); err != nil {
		// 数据库或审计失败时清理已写入的文件。
		_ = os.Remove(absolutePath)
		return admindto.FileUploadResponse{}, err
	}

	return admindto.FileUploadResponse{
		ID:           attachment.ID,
		OriginalName: attachment.OriginalName,
		MimeType:     attachment.MimeType,
		Size:         attachment.Size,
		URL:          fmt.Sprintf("/admin-api/files/%d", attachment.ID),
		CreatedAt:    attachment.CreatedAt,
	}, nil
}

/**
 * List 分页查询文件列表。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @return admindto.PageResponse[admindto.FileItem] 分页结果
 * @return error 查询失败原因
 */
func (s *FileAttachmentService) List(ctx context.Context, query admindto.FileListQuery) (admindto.PageResponse[admindto.FileItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Model(&models.FileAttachment{}).Where("status = ?", "active")

	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where("original_name LIKE ?", keyword)
	}
	if query.MimeType != "" {
		db = db.Where("mime_type = ?", strings.TrimSpace(query.MimeType))
	}
	if query.UploaderID > 0 {
		db = db.Where("uploader_id = ?", query.UploaderID)
	}
	if query.DateFrom != "" {
		from, err := parseTime(query.DateFrom)
		if err != nil {
			return admindto.PageResponse[admindto.FileItem]{}, apperrors.ErrValidation.WithMessage("开始时间格式错误")
		}
		db = db.Where("created_at >= ?", from)
	}
	if query.DateTo != "" {
		to, err := parseTime(query.DateTo)
		if err != nil {
			return admindto.PageResponse[admindto.FileItem]{}, apperrors.ErrValidation.WithMessage("结束时间格式错误")
		}
		db = db.Where("created_at <= ?", to)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.FileItem]{}, err
	}

	var attachments []models.FileAttachment
	if err := db.Order("id DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&attachments).Error; err != nil {
		return admindto.PageResponse[admindto.FileItem]{}, err
	}

	// 批量查询上传者信息
	uploaderIDs := make([]uint64, 0, len(attachments))
	for _, a := range attachments {
		uploaderIDs = append(uploaderIDs, a.UploaderID)
	}
	uploaderMap, err := s.uploaderSummaryByIDs(ctx, uploaderIDs)
	if err != nil {
		return admindto.PageResponse[admindto.FileItem]{}, err
	}

	items := make([]admindto.FileItem, 0, len(attachments))
	for _, a := range attachments {
		item := admindto.FileItem{
			ID:           a.ID,
			OriginalName: a.OriginalName,
			MimeType:     a.MimeType,
			Extension:    a.Extension,
			Size:         a.Size,
			URL:          fmt.Sprintf("/admin-api/files/%d", a.ID),
			Uploader:     uploaderMap[a.UploaderID],
			CreatedAt:    a.CreatedAt,
		}
		items = append(items, item)
	}
	return support.PageResponse(items, total, page, perPage), nil
}

/**
 * Detail 查询文件详情。
 *
 * @param ctx 请求上下文
 * @param id 文件 ID
 * @return admindto.FileItem 文件详情
 * @return error 查询失败原因
 */
func (s *FileAttachmentService) Detail(ctx context.Context, id uint64) (admindto.FileDetailResponse, error) {
	return s.detailResponse(ctx, id)
}

func (s *FileAttachmentService) detailResponse(ctx context.Context, id uint64) (admindto.FileDetailResponse, error) {
	attachment, err := s.findAttachment(ctx, id)
	if err != nil {
		return admindto.FileDetailResponse{}, err
	}
	if attachment.Status != "active" {
		return admindto.FileDetailResponse{}, apperrors.ErrNotFound.WithMessage("文件不存在")
	}

	uploaderMap, err := s.uploaderSummaryByIDs(ctx, []uint64{attachment.UploaderID})
	if err != nil {
		return admindto.FileDetailResponse{}, err
	}
	reference, err := s.referenceResponse(ctx, id)
	if err != nil {
		return admindto.FileDetailResponse{}, err
	}
	canDelete := reference.ReferenceCount == 0

	return admindto.FileDetailResponse{
		FileItem: admindto.FileItem{
			ID:           attachment.ID,
			OriginalName: attachment.OriginalName,
			MimeType:     attachment.MimeType,
			Extension:    attachment.Extension,
			Size:         attachment.Size,
			URL:          fmt.Sprintf("/admin-api/files/%d", attachment.ID),
			Uploader:     uploaderMap[attachment.UploaderID],
			CreatedAt:    attachment.CreatedAt,
		},
		StorageDriver:  attachment.StorageDriver,
		Checksum:       attachment.Checksum,
		ReferenceCount: reference.ReferenceCount,
		References:     reference.References,
		DownloadURL:    fmt.Sprintf("/admin-api/files/%d/download", attachment.ID),
		CanDelete:      canDelete,
	}, nil
}

func (s *FileAttachmentService) DownloadPath(ctx context.Context, id uint64) (string, string, string, error) {
	attachment, err := s.findAttachment(ctx, id)
	if err != nil {
		return "", "", "", err
	}
	if attachment.Status != "active" {
		return "", "", "", apperrors.ErrNotFound.WithMessage("文件不存在")
	}
	reference, err := s.referenceResponse(ctx, id)
	if err != nil {
		return "", "", "", err
	}
	for _, item := range reference.References {
		if item.RefType == "user_real_name_application" {
			return "", "", "", apperrors.ErrForbidden.WithMessage("历史实名附件不提供下载或预览")
		}
	}
	absolutePath, err := s.safeStoragePath(attachment.StoragePath)
	if err != nil {
		return "", "", "", err
	}
	if _, err := os.Stat(absolutePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", "", apperrors.ErrNotFound.WithMessage("文件不存在")
		}
		return "", "", "", err
	}
	return absolutePath, attachment.MimeType, attachment.OriginalName, nil
}

func (s *FileAttachmentService) ReferenceResponse(ctx context.Context, id uint64) (admindto.FileReferenceResponse, error) {
	attachment, err := s.findAttachment(ctx, id)
	if err != nil {
		return admindto.FileReferenceResponse{}, err
	}
	if attachment.Status != "active" {
		return admindto.FileReferenceResponse{}, apperrors.ErrNotFound.WithMessage("文件不存在")
	}
	return s.referenceResponse(ctx, id)
}

func (s *FileAttachmentService) referenceResponse(ctx context.Context, id uint64) (admindto.FileReferenceResponse, error) {
	var rows []models.FileAttachmentReference
	if err := s.db.WithContext(ctx).Where("file_id = ?", id).Order("id ASC").Find(&rows).Error; err != nil {
		return admindto.FileReferenceResponse{}, err
	}
	references := make([]admindto.FileReferenceItem, 0, len(rows))
	for _, row := range rows {
		references = append(references, admindto.FileReferenceItem{
			ID:        row.ID,
			FileID:    row.FileID,
			RefType:   row.RefType,
			RefID:     row.RefID,
			RefName:   row.RefName,
			RefPath:   row.RefPath,
			CreatedAt: row.CreatedAt,
		})
	}
	return admindto.FileReferenceResponse{
		FileID:         id,
		ReferenceCount: int64(len(references)),
		References:     references,
	}, nil
}

/**
 * Delete 删除文件（软删除）。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param id 文件 ID
 * @return error 删除失败原因
 */
func (s *FileAttachmentService) Delete(ctx context.Context, operatorID uint64, id uint64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var attachment models.FileAttachment
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&attachment).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("文件不存在")
		}
		if err != nil {
			return err
		}
		if attachment.Status != "active" {
			return apperrors.ErrNotFound.WithMessage("文件不存在")
		}
		var referenceCount int64
		if err := tx.Model(&models.FileAttachmentReference{}).Where("file_id = ?", id).Count(&referenceCount).Error; err != nil {
			return err
		}
		if referenceCount > 0 {
			return apperrors.ErrConflict.WithMessage("文件仍被业务记录引用，禁止删除")
		}
		// 软删除
		if err := tx.Model(&attachment).Update("status", "deleted").Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     fileDeleteAction,
			ObjectType: fileAttachmentObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: map[string]any{
				"id":            attachment.ID,
				"original_name": attachment.OriginalName,
				"mime_type":     attachment.MimeType,
				"size":          attachment.Size,
			},
			Remark: "删除文件",
		})
	}); err != nil {
		return err
	}

	return nil
}

func (s *FileAttachmentService) maxUploadRequestBytes() int64 {
	if s.config.MaxSize <= 0 {
		return multipartOverheadBytes
	}
	return s.config.MaxSize + multipartOverheadBytes
}

func (s *FileAttachmentService) safeStoragePath(storagePath string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(storagePath))
	if cleanPath == "." || filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", apperrors.ErrNotFound.WithMessage("文件不存在")
	}
	root, err := filepath.Abs(s.config.LocalPath)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(root, cleanPath))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", apperrors.ErrNotFound.WithMessage("文件不存在")
	}
	return target, nil
}

func (s *FileAttachmentService) findAttachment(ctx context.Context, id uint64) (models.FileAttachment, error) {
	var attachment models.FileAttachment
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&attachment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.FileAttachment{}, apperrors.ErrNotFound.WithMessage("文件不存在")
	}
	if err != nil {
		return models.FileAttachment{}, err
	}
	return attachment, nil
}

func (s *FileAttachmentService) validateFile(originalName string, header *multipart.FileHeader, detectedMimeType string) error {
	// 检查扩展名
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext != "" {
		ext = ext[1:] // 移除点号
	}
	if ext == "" {
		return apperrors.ErrValidation.WithMessage("文件扩展名不能为空")
	}

	// 危险文件类型黑名单
	dangerousExts := map[string]bool{
		"php": true, "php3": true, "php4": true, "php5": true, "phtml": true,
		"exe": true, "msi": true, "bat": true, "cmd": true, "com": true,
		"sh": true, "bash": true, "zsh": true,
		"js": true, "vbs": true, "vbe": true, "wsf": true,
		"html": true, "htm": true, "svg": true,
		"jar": true, "war": true, "ear": true,
		"py": true, "pl": true, "rb": true,
		"asp": true, "aspx": true, "jsp": true,
	}
	if dangerousExts[ext] {
		return apperrors.ErrValidation.WithMessage("不允许上传该类型的文件")
	}
	allowedExtMimeTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"pdf":  "application/pdf",
	}
	expectedMimeType, ok := allowedExtMimeTypes[ext]
	if !ok {
		return apperrors.ErrValidation.WithMessage("不允许上传该扩展名的文件")
	}

	// MIME 类型白名单
	allowedMimeTypes := make(map[string]bool)
	for _, t := range s.config.AllowedTypes {
		allowedMimeTypes[t] = true
	}
	if !allowedMimeTypes[expectedMimeType] {
		return apperrors.ErrValidation.WithMessage("当前配置不允许上传该类型文件")
	}
	contentType := header.Header.Get("Content-Type")
	if contentType != "" && !allowedMimeTypes[contentType] {
		return apperrors.ErrValidation.WithMessage("不允许上传该MIME类型的文件")
	}
	if contentType != "" && contentType != expectedMimeType {
		return apperrors.ErrValidation.WithMessage("文件扩展名与声明类型不一致")
	}
	if detectedMimeType != expectedMimeType {
		return apperrors.ErrValidation.WithMessage("文件内容与扩展名不一致")
	}

	return nil
}

func sanitizeOriginalName(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	name = strings.ReplaceAll(name, "\x00", "")
	return name
}

func (s *FileAttachmentService) uploaderSummaryByIDs(ctx context.Context, ids []uint64) (map[uint64]*admindto.FileUploaderSummary, error) {
	result := make(map[uint64]*admindto.FileUploaderSummary)
	if len(ids) == 0 {
		return result, nil
	}

	var users []models.AdminUser
	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	for _, u := range users {
		uid := u.ID
		result[uid] = &admindto.FileUploaderSummary{
			ID:          uid,
			Username:    u.Username,
			DisplayName: u.DisplayName,
		}
	}
	return result, nil
}

func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func detectMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// JPEG
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	// PNG
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	// GIF
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}
	// WebP
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	// PDF
	if data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46 {
		return "application/pdf"
	}

	return "application/octet-stream"
}

func parseTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", value)
}
