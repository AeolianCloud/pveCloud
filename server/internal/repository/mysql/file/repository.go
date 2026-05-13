package file

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

type AttachmentListFilters struct {
	Keyword    string
	MimeType   string
	UploaderID uint64
	DateFrom   *time.Time
	DateTo     *time.Time
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Attachments(ctx context.Context, filters AttachmentListFilters, limit int, offset int) ([]FileAttachment, int64, error) {
	query := r.applyAttachmentListFilters(r.db.WithContext(ctx).
		Model(&FileAttachment{}).
		Where("status = ?", "active"), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var attachments []FileAttachment
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&attachments).Error; err != nil {
		return nil, 0, err
	}
	return attachments, total, nil
}

func (r *Repository) CreateAttachment(ctx context.Context, db *gorm.DB, attachment *FileAttachment) error {
	return r.queryDB(db).WithContext(ctx).Create(attachment).Error
}

func (r *Repository) CreateReference(ctx context.Context, db *gorm.DB, reference *FileAttachmentReference) error {
	return r.queryDB(db).WithContext(ctx).Create(reference).Error
}

func (r *Repository) FindAttachmentByID(ctx context.Context, id uint64) (FileAttachment, error) {
	var attachment FileAttachment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&attachment).Error
	return attachment, err
}

func (r *Repository) FindAttachmentByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (FileAttachment, error) {
	var attachment FileAttachment
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&attachment).Error
	return attachment, err
}

func (r *Repository) AttachmentReferences(ctx context.Context, fileID uint64) ([]FileAttachmentReference, error) {
	var rows []FileAttachmentReference
	if err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) CountAttachmentReferences(ctx context.Context, db *gorm.DB, fileID uint64) (int64, error) {
	var count int64
	if err := r.queryDB(db).WithContext(ctx).
		Model(&FileAttachmentReference{}).
		Where("file_id = ?", fileID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) UpdateAttachmentStatus(ctx context.Context, db *gorm.DB, id uint64, status string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&FileAttachment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) applyAttachmentListFilters(db *gorm.DB, filters AttachmentListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		db = db.Where("original_name LIKE ?", keyword)
	}
	if strings.TrimSpace(filters.MimeType) != "" {
		db = db.Where("mime_type = ?", strings.TrimSpace(filters.MimeType))
	}
	if filters.UploaderID > 0 {
		db = db.Where("uploader_id = ?", filters.UploaderID)
	}
	if filters.DateFrom != nil {
		db = db.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		db = db.Where("created_at <= ?", *filters.DateTo)
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
