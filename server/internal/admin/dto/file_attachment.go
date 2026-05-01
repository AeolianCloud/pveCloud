package dto

import "time"

/**
 * FileListQuery 表示文件列表查询参数。
 */
type FileListQuery struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	PerPage    int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword    string `form:"keyword" validate:"omitempty,max=255"`
	MimeType   string `form:"mime_type" validate:"omitempty,max=128"`
	UploaderID uint64 `form:"uploader_id" validate:"omitempty"`
	DateFrom   string `form:"date_from" validate:"omitempty"`
	DateTo     string `form:"date_to" validate:"omitempty"`
}

/**
 * FileItem 表示文件列表项。
 */
type FileItem struct {
	ID           uint64               `json:"id"`
	OriginalName string               `json:"original_name"`
	MimeType     string               `json:"mime_type"`
	Extension    string               `json:"extension"`
	Size         uint64               `json:"size"`
	URL          string               `json:"url"`
	Uploader     *FileUploaderSummary `json:"uploader"`
	CreatedAt    time.Time            `json:"created_at"`
}

/**
 * FileReferenceItem 表示文件引用记录。
 */
type FileReferenceItem struct {
	ID        uint64    `json:"id"`
	FileID    uint64    `json:"file_id"`
	RefType   string    `json:"ref_type"`
	RefID     string    `json:"ref_id"`
	RefName   *string   `json:"ref_name"`
	RefPath   *string   `json:"ref_path"`
	CreatedAt time.Time `json:"created_at"`
}

/**
 * FileReferenceResponse 表示文件引用查询结果。
 */
type FileReferenceResponse struct {
	FileID         uint64              `json:"file_id"`
	ReferenceCount int64               `json:"reference_count"`
	References     []FileReferenceItem `json:"references"`
}

/**
 * FileDetailResponse 表示文件详情。
 */
type FileDetailResponse struct {
	FileItem
	StorageDriver  string              `json:"storage_driver"`
	Checksum       string              `json:"checksum"`
	ReferenceCount int64               `json:"reference_count"`
	References     []FileReferenceItem `json:"references"`
	DownloadURL    string              `json:"download_url"`
	CanDelete      bool                `json:"can_delete"`
}

/**
 * FileUploaderSummary 表示文件上传者摘要。
 */
type FileUploaderSummary struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

/**
 * FileUploadResponse 表示文件上传成功响应。
 */
type FileUploadResponse struct {
	ID           uint64    `json:"id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	Size         uint64    `json:"size"`
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"created_at"`
}
