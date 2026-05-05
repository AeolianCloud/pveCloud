package realname

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
)

const (
	statusUnverified = "unverified"
	statusPending    = "pending"
	statusApproved   = "approved"
	statusRejected   = "rejected"
	refTypeRealName  = "user_real_name_application"
)

var idCardPattern = regexp.MustCompile(`^[0-9]{17}[0-9Xx]$`)

type RealNameService struct {
	db      *gorm.DB
	storage bootstrap.StorageConfig
}

func NewRealNameService(db *gorm.DB, storage bootstrap.StorageConfig) *RealNameService {
	return &RealNameService{db: db, storage: storage}
}

func (s *RealNameService) Status(ctx context.Context, userID uint64) (webdto.RealNameStatusResponse, error) {
	config, err := s.config(ctx, s.db)
	if err != nil {
		return webdto.RealNameStatusResponse{}, err
	}
	latest, ok, err := s.latest(ctx, s.db, userID)
	if err != nil {
		return webdto.RealNameStatusResponse{}, err
	}
	if !ok {
		return webdto.RealNameStatusResponse{Status: statusUnverified, Config: config}, nil
	}
	summary := applicationSummary(latest)
	return webdto.RealNameStatusResponse{Status: latest.Status, Application: &summary, Config: config}, nil
}

func (s *RealNameService) UploadFile(ctx context.Context, userID uint64, file multipart.File, header *multipart.FileHeader) (webdto.RealNameFileUploadResponse, error) {
	config, err := s.config(ctx, s.db)
	if err != nil {
		return webdto.RealNameFileUploadResponse{}, err
	}
	if !config.Enabled {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrForbidden.WithMessage("实名功能暂未开放")
	}
	originalName := sanitizeOriginalName(header.Filename)
	if originalName == "" {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrValidation.WithMessage("文件名不能为空")
	}
	maxSize := int64(config.ImageMaxSizeMB) * 1024 * 1024
	if header.Size > maxSize {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d MB", config.ImageMaxSizeMB))
	}

	sniff := make([]byte, 512)
	sniffSize, err := io.ReadFull(file, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrInternal.WithMessage("读取文件失败")
	}
	if sniffSize == 0 {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrValidation.WithMessage("文件内容不能为空")
	}
	sniff = sniff[:sniffSize]
	mimeType := http.DetectContentType(sniff)
	if !allowedString(config.AllowedImageTypes, mimeType) || !allowedString(s.storage.AllowedTypes, mimeType) {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrValidation.WithMessage("文件类型不允许")
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(originalName)), ".")
	if !extensionMatchesMime(ext, mimeType) {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrValidation.WithMessage("文件扩展名与内容不匹配")
	}
	storedUUID, err := randomHex(16)
	if err != nil {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrInternal.WithMessage("生成文件名失败")
	}
	now := time.Now()
	storagePath := filepath.Join(fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()), storedUUID+"."+ext)
	absolutePath := filepath.Join(s.storage.LocalPath, storagePath)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0755); err != nil {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrInternal.WithMessage("创建存储目录失败")
	}
	out, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrInternal.WithMessage("保存文件失败")
	}
	hash := sha256.New()
	writer := io.MultiWriter(out, hash)
	writtenBytes, err := writer.Write(sniff)
	written := int64(writtenBytes)
	if err == nil {
		var copied int64
		copied, err = io.Copy(writer, io.LimitReader(file, maxSize-written+1))
		written += copied
	}
	closeErr := out.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(absolutePath)
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrInternal.WithMessage("保存文件失败")
	}
	if written > maxSize {
		_ = os.Remove(absolutePath)
		return webdto.RealNameFileUploadResponse{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d MB", config.ImageMaxSizeMB))
	}
	attachment := models.FileAttachment{
		OriginalName:   originalName,
		StoredName:     storedUUID + "." + ext,
		MimeType:       mimeType,
		Extension:      ext,
		Size:           uint64(written),
		StoragePath:    storagePath,
		StorageDriver:  "local",
		Checksum:       hex.EncodeToString(hash.Sum(nil)),
		UploaderUserID: &userID,
		Status:         "active",
	}
	if err := s.db.WithContext(ctx).Omit("uploader_id").Create(&attachment).Error; err != nil {
		_ = os.Remove(absolutePath)
		return webdto.RealNameFileUploadResponse{}, err
	}
	return webdto.RealNameFileUploadResponse{ID: attachment.ID, OriginalName: attachment.OriginalName, MimeType: attachment.MimeType, Size: attachment.Size, CreatedAt: attachment.CreatedAt}, nil
}

func (s *RealNameService) Submit(ctx context.Context, userID uint64, req webdto.RealNameSubmitRequest) (webdto.RealNameApplicationSummary, error) {
	realName := strings.TrimSpace(req.RealName)
	idNumber := strings.ToUpper(strings.TrimSpace(req.IDNumber))
	if !idCardPattern.MatchString(idNumber) {
		return webdto.RealNameApplicationSummary{}, apperrors.ErrValidation.WithMessage("身份证号码格式错误")
	}
	digest := digestIDNumber(req.IDType, idNumber)
	var created models.UserRealNameApplication
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		config, err := s.config(ctx, tx)
		if err != nil {
			return err
		}
		if !config.Enabled {
			return apperrors.ErrForbidden.WithMessage("实名功能暂未开放")
		}
		latest, ok, err := s.latestForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		attempt := uint(1)
		if ok {
			attempt = latest.SubmitAttempt + 1
			switch latest.Status {
			case statusPending:
				return apperrors.ErrConflict.WithMessage("实名申请审核中，请勿重复提交")
			case statusApproved:
				return apperrors.ErrConflict.WithMessage("实名已通过，不能重复提交")
			case statusRejected:
				if !config.ResubmitEnabled {
					return apperrors.ErrForbidden.WithMessage("实名被拒绝后暂不允许重新提交")
				}
				if int(attempt) > config.MaxSubmitAttempts {
					return apperrors.ErrForbidden.WithMessage("实名提交次数已达上限")
				}
			}
		}
		if err := s.ensureRequiredFiles(ctx, tx, userID, config, req); err != nil {
			return err
		}
		var duplicate int64
		if err := tx.Model(&models.UserRealNameApplication{}).
			Where("id_number_digest = ? AND status = ? AND user_id <> ?", digest, statusApproved, userID).
			Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return apperrors.ErrConflict.WithMessage("该证件号码已完成实名")
		}
		applicationNo, err := applicationNo()
		if err != nil {
			return err
		}
		created = models.UserRealNameApplication{
			ApplicationNo:     applicationNo,
			UserID:            userID,
			RealName:          realName,
			IDType:            req.IDType,
			IDNumberDigest:    digest,
			IDNumberMasked:    maskIDNumber(idNumber),
			IDCardFrontFileID: req.IDCardFrontFileID,
			IDCardBackFileID:  req.IDCardBackFileID,
			HoldCardFileID:    req.HoldCardFileID,
			Status:            statusPending,
			SubmitAttempt:     attempt,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		return s.createFileReferences(ctx, tx, created)
	}); err != nil {
		return webdto.RealNameApplicationSummary{}, err
	}
	summary := applicationSummary(created)
	return summary, nil
}

func (s *RealNameService) RequireApprovedForOrder(ctx context.Context, userID uint64) error {
	config, err := s.config(ctx, s.db)
	if err != nil {
		return err
	}
	if !config.RequiredForOrder {
		return nil
	}
	latest, ok, err := s.latest(ctx, s.db, userID)
	if err != nil {
		return err
	}
	if !ok || latest.Status != statusApproved {
		return apperrors.ErrForbidden.WithMessage("请先完成实名认证后再购买机器")
	}
	return nil
}

func (s *RealNameService) ensureRequiredFiles(ctx context.Context, tx *gorm.DB, userID uint64, config webdto.RealNameConfig, req webdto.RealNameSubmitRequest) error {
	if config.IDCardFrontRequired && req.IDCardFrontFileID == nil {
		return apperrors.ErrValidation.WithMessage("请上传身份证人像面")
	}
	if config.IDCardBackRequired && req.IDCardBackFileID == nil {
		return apperrors.ErrValidation.WithMessage("请上传身份证国徽面")
	}
	if config.HoldCardRequired && req.HoldCardFileID == nil {
		return apperrors.ErrValidation.WithMessage("请上传手持证件照片")
	}
	ids := []uint64{}
	for _, id := range []*uint64{req.IDCardFrontFileID, req.IDCardBackFileID, req.HoldCardFileID} {
		if id != nil && *id > 0 {
			ids = append(ids, *id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := tx.WithContext(ctx).Model(&models.FileAttachment{}).
		Where("id IN ? AND uploader_user_id = ? AND status = ?", ids, userID, "active").
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return apperrors.ErrValidation.WithMessage("实名图片不存在或不属于当前用户")
	}
	return nil
}

func (s *RealNameService) createFileReferences(ctx context.Context, tx *gorm.DB, app models.UserRealNameApplication) error {
	refName := textutil.StringPtr("实名申请 " + app.ApplicationNo)
	refID := fmt.Sprintf("%d", app.ID)
	refs := []models.FileAttachmentReference{}
	for _, id := range []*uint64{app.IDCardFrontFileID, app.IDCardBackFileID, app.HoldCardFileID} {
		if id != nil && *id > 0 {
			refs = append(refs, models.FileAttachmentReference{FileID: *id, RefType: refTypeRealName, RefID: refID, RefName: refName})
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&refs).Error
}

func (s *RealNameService) latest(ctx context.Context, db *gorm.DB, userID uint64) (models.UserRealNameApplication, bool, error) {
	var app models.UserRealNameApplication
	err := db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, false, nil
	}
	return app, err == nil, err
}

func (s *RealNameService) latestForUpdate(ctx context.Context, db *gorm.DB, userID uint64) (models.UserRealNameApplication, bool, error) {
	var app models.UserRealNameApplication
	err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).Order("id DESC").First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, false, nil
	}
	return app, err == nil, err
}

func (s *RealNameService) config(ctx context.Context, db *gorm.DB) (webdto.RealNameConfig, error) {
	config := webdto.RealNameConfig{RequiredForOrder: true, ResubmitEnabled: true, MaxSubmitAttempts: 3, IDCardFrontRequired: true, IDCardBackRequired: true, ImageMaxSizeMB: 5, AllowedImageTypes: []string{"image/jpeg", "image/png", "image/webp"}}
	var rows []models.SystemConfig
	if err := db.WithContext(ctx).Where("config_key LIKE ? AND is_secret = 0", "real_name.%").Find(&rows).Error; err != nil {
		return config, err
	}
	for _, row := range rows {
		value := ""
		if row.ConfigValue != nil {
			value = strings.TrimSpace(*row.ConfigValue)
		}
		switch row.ConfigKey {
		case "real_name.enabled":
			config.Enabled = strings.EqualFold(value, "true")
		case "real_name.required_for_order":
			config.RequiredForOrder = strings.EqualFold(value, "true")
		case "real_name.resubmit_enabled":
			config.ResubmitEnabled = strings.EqualFold(value, "true")
		case "real_name.max_submit_attempts":
			config.MaxSubmitAttempts = positiveInt(value, config.MaxSubmitAttempts)
		case "real_name.id_card_front_required":
			config.IDCardFrontRequired = strings.EqualFold(value, "true")
		case "real_name.id_card_back_required":
			config.IDCardBackRequired = strings.EqualFold(value, "true")
		case "real_name.hold_card_required":
			config.HoldCardRequired = strings.EqualFold(value, "true")
		case "real_name.image_max_size_mb":
			config.ImageMaxSizeMB = positiveInt(value, config.ImageMaxSizeMB)
		case "real_name.allowed_image_types":
			config.AllowedImageTypes = csv(value, config.AllowedImageTypes)
		case "real_name.review_notice":
			config.ReviewNotice = value
		}
	}
	return config, nil
}

func applicationSummary(app models.UserRealNameApplication) webdto.RealNameApplicationSummary {
	return webdto.RealNameApplicationSummary{ApplicationNo: app.ApplicationNo, RealName: app.RealName, IDType: app.IDType, IDNumberMasked: app.IDNumberMasked, Status: app.Status, RejectReason: app.RejectReason, SubmitAttempt: app.SubmitAttempt, CreatedAt: app.CreatedAt, ReviewedAt: app.ReviewedAt}
}

func digestIDNumber(idType, idNumber string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(idType)) + ":" + strings.ToUpper(strings.TrimSpace(idNumber))))
	return hex.EncodeToString(sum[:])
}

func maskIDNumber(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}

func applicationNo() (string, error) {
	random, err := randomHex(4)
	if err != nil {
		return "", err
	}
	return "RN" + time.Now().Format("20060102150405") + strings.ToUpper(random), nil
}

func randomHex(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sanitizeOriginalName(name string) string {
	name = strings.ReplaceAll(filepath.Base(name), "\x00", "")
	return strings.TrimSpace(name)
}

func allowedString(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

func extensionMatchesMime(ext, mimeType string) bool {
	switch mimeType {
	case "image/jpeg":
		return ext == "jpg" || ext == "jpeg"
	case "image/png":
		return ext == "png"
	case "image/webp":
		return ext == "webp"
	default:
		return false
	}
}

func positiveInt(value string, fallback int) int {
	var parsed int
	if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func csv(value string, fallback []string) []string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
