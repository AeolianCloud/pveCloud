package file

import "time"

/**
 * FileAttachment 映射 file_attachments 文件附件表。
 */
type FileAttachment struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	OriginalName   string    `gorm:"column:original_name"`
	StoredName     string    `gorm:"column:stored_name"`
	MimeType       string    `gorm:"column:mime_type"`
	Extension      string    `gorm:"column:extension"`
	Size           uint64    `gorm:"column:size"`
	StoragePath    string    `gorm:"column:storage_path"`
	StorageDriver  string    `gorm:"column:storage_driver"`
	Checksum       string    `gorm:"column:checksum"`
	UploaderID     *uint64   `gorm:"column:uploader_id"`
	UploaderUserID *uint64   `gorm:"column:uploader_user_id"`
	Status         string    `gorm:"column:status"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

/**
 * TableName 返回文件附件表名。
 */
func (FileAttachment) TableName() string {
	return "file_attachments"
}

/**
 * FileAttachmentReference 映射 file_attachment_references 文件引用表。
 */
type FileAttachmentReference struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	FileID    uint64    `gorm:"column:file_id"`
	RefType   string    `gorm:"column:ref_type"`
	RefID     string    `gorm:"column:ref_id"`
	RefName   *string   `gorm:"column:ref_name"`
	RefPath   *string   `gorm:"column:ref_path"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

/**
 * TableName 返回文件引用表名。
 */
func (FileAttachmentReference) TableName() string {
	return "file_attachment_references"
}
