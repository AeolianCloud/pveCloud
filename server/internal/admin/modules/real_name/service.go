package realname

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	realNameObjectType   = "user_real_name"
	realNameReviewAction = "real_name.review"
)

type RealNameService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

func NewRealNameService(db *gorm.DB, auditService *AdminAuditService) *RealNameService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &RealNameService{db: db, auditService: auditService}
}

func (s *RealNameService) Applications(ctx context.Context, query admindto.RealNameApplicationListQuery) (admindto.PageResponse[admindto.RealNameApplicationItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.applicationDB(ctx)
	db, err := applyFilters(db, query)
	if err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	var rows []applicationRow
	if err := db.Select(applicationSelect()).Order("applications.id DESC").Limit(perPage).Offset((page - 1) * perPage).Scan(&rows).Error; err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	items := make([]admindto.RealNameApplicationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.item(false))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

func (s *RealNameService) Detail(ctx context.Context, id uint64) (admindto.RealNameApplicationItem, error) {
	var row applicationRow
	err := s.applicationDB(ctx).Select(applicationSelect()).Where("applications.id = ?", id).Scan(&row).Error
	if err != nil {
		return admindto.RealNameApplicationItem{}, err
	}
	if row.ID == 0 {
		return admindto.RealNameApplicationItem{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
	}
	return row.item(true), nil
}

func (s *RealNameService) Review(ctx context.Context, operatorID uint64, id uint64, req admindto.RealNameReviewRequest) (admindto.RealNameApplicationItem, error) {
	status := strings.TrimSpace(req.Status)
	reason := textutil.NormalizeOptionalString(req.RejectReason)
	if status == "rejected" && (reason == nil || strings.TrimSpace(*reason) == "") {
		return admindto.RealNameApplicationItem{}, apperrors.ErrValidation.WithMessage("拒绝原因不能为空")
	}
	if status == "approved" {
		reason = nil
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.UserRealNameApplication
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("实名申请不存在")
		}
		if err != nil {
			return err
		}
		if current.Status != "pending" {
			return apperrors.ErrConflict.WithMessage("只有待审核申请可以审核")
		}
		if status == "approved" {
			var duplicate int64
			if err := tx.Model(&models.UserRealNameApplication{}).Where("id_number_digest = ? AND status = ? AND user_id <> ?", current.IDNumberDigest, "approved", current.UserID).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return apperrors.ErrConflict.WithMessage("该证件号码已被其它用户实名通过")
			}
		}
		now := time.Now()
		updates := map[string]any{"status": status, "review_admin_id": operatorID, "reviewed_at": now, "reject_reason": reason}
		if err := tx.Model(&models.UserRealNameApplication{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: realNameReviewAction, ObjectType: realNameObjectType, ObjectID: textutil.Uint64String(id), BeforeData: auditSnapshot(current), AfterData: map[string]any{"id": id, "status": status, "reject_reason": reason}, Remark: "审核实名申请"})
	}); err != nil {
		return admindto.RealNameApplicationItem{}, err
	}
	return s.Detail(ctx, id)
}

func applyFilters(db *gorm.DB, query admindto.RealNameApplicationListQuery) (*gorm.DB, error) {
	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR applications.real_name LIKE ? OR applications.application_no LIKE ?", keyword, keyword, keyword, keyword)
	}
	if query.Status != "" {
		db = db.Where("applications.status = ?", strings.TrimSpace(query.Status))
	}
	if query.IDType != "" {
		db = db.Where("applications.id_type = ?", strings.TrimSpace(query.IDType))
	}
	if query.DateFrom != "" {
		from, err := time.ParseInLocation("2006-01-02", query.DateFrom, time.Local)
		if err != nil {
			return nil, apperrors.ErrValidation.WithMessage("开始时间格式错误")
		}
		db = db.Where("applications.created_at >= ?", from)
	}
	if query.DateTo != "" {
		to, err := time.ParseInLocation("2006-01-02", query.DateTo, time.Local)
		if err != nil {
			return nil, apperrors.ErrValidation.WithMessage("结束时间格式错误")
		}
		db = db.Where("applications.created_at < ?", to.Add(24*time.Hour))
	}
	return db, nil
}

func (s *RealNameService) applicationDB(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).
		Table("user_real_name_applications AS applications").
		Joins("JOIN users ON users.id = applications.user_id").
		Joins("LEFT JOIN file_attachments AS front ON front.id = applications.id_card_front_file_id").
		Joins("LEFT JOIN file_attachments AS back ON back.id = applications.id_card_back_file_id").
		Joins("LEFT JOIN file_attachments AS hold ON hold.id = applications.hold_card_file_id").
		Joins("LEFT JOIN admin_users AS admins ON admins.id = applications.review_admin_id")
}

func auditSnapshot(app models.UserRealNameApplication) map[string]any {
	return map[string]any{"id": app.ID, "application_no": app.ApplicationNo, "user_id": app.UserID, "status": app.Status, "id_number_masked": app.IDNumberMasked}
}

func applicationSelect() string {
	return `applications.id, applications.application_no, applications.real_name, applications.id_type, applications.id_number_masked, applications.status, applications.submit_attempt, applications.reviewed_at, applications.reject_reason, applications.created_at, applications.updated_at,
		users.id AS user_id, users.username, users.email, users.display_name, users.status AS user_status,
		front.id AS front_file_id, front.original_name AS front_original_name, front.mime_type AS front_mime_type, front.size AS front_size, front.created_at AS front_created_at,
		back.id AS back_file_id, back.original_name AS back_original_name, back.mime_type AS back_mime_type, back.size AS back_size, back.created_at AS back_created_at,
		hold.id AS hold_file_id, hold.original_name AS hold_original_name, hold.mime_type AS hold_mime_type, hold.size AS hold_size, hold.created_at AS hold_created_at,
		admins.id AS review_admin_id, admins.username AS review_admin_username, admins.email AS review_admin_email, admins.display_name AS review_admin_display_name, admins.status AS review_admin_status
	`
}

type applicationRow struct {
	ID                     uint64
	ApplicationNo          string
	RealName               string
	IDType                 string
	IDNumberMasked         string
	Status                 string
	SubmitAttempt          uint
	ReviewedAt             *time.Time
	RejectReason           *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	UserID                 uint64
	Username               string
	Email                  string
	DisplayName            *string
	UserStatus             string
	FrontFileID            *uint64
	FrontOriginalName      *string
	FrontMimeType          *string
	FrontSize              *uint64
	FrontCreatedAt         *time.Time
	BackFileID             *uint64
	BackOriginalName       *string
	BackMimeType           *string
	BackSize               *uint64
	BackCreatedAt          *time.Time
	HoldFileID             *uint64
	HoldOriginalName       *string
	HoldMimeType           *string
	HoldSize               *uint64
	HoldCreatedAt          *time.Time
	ReviewAdminID          *uint64
	ReviewAdminUsername    *string
	ReviewAdminEmail       *string
	ReviewAdminDisplayName *string
	ReviewAdminStatus      *string
}

func (r applicationRow) item(includeFiles bool) admindto.RealNameApplicationItem {
	item := admindto.RealNameApplicationItem{ID: r.ID, ApplicationNo: r.ApplicationNo, User: admindto.RealNameUserSummary{ID: r.UserID, Username: r.Username, Email: r.Email, DisplayName: r.DisplayName, Status: r.UserStatus}, RealName: r.RealName, IDType: r.IDType, IDNumberMasked: r.IDNumberMasked, Status: r.Status, SubmitAttempt: r.SubmitAttempt, ReviewedAt: r.ReviewedAt, RejectReason: r.RejectReason, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	if r.ReviewAdminID != nil {
		item.ReviewAdmin = &admindto.RealNameUserSummary{ID: *r.ReviewAdminID, Username: stringValue(r.ReviewAdminUsername), Email: stringValue(r.ReviewAdminEmail), DisplayName: r.ReviewAdminDisplayName, Status: stringValue(r.ReviewAdminStatus)}
	}
	if includeFiles {
		item.IDCardFrontFile = fileSummary(r.FrontFileID, r.FrontOriginalName, r.FrontMimeType, r.FrontSize, r.FrontCreatedAt)
		item.IDCardBackFile = fileSummary(r.BackFileID, r.BackOriginalName, r.BackMimeType, r.BackSize, r.BackCreatedAt)
		item.HoldCardFile = fileSummary(r.HoldFileID, r.HoldOriginalName, r.HoldMimeType, r.HoldSize, r.HoldCreatedAt)
	}
	return item
}

func fileSummary(id *uint64, name *string, mimeType *string, size *uint64, createdAt *time.Time) *admindto.RealNameFileSummary {
	if id == nil || name == nil || mimeType == nil || size == nil || createdAt == nil {
		return nil
	}
	return &admindto.RealNameFileSummary{ID: *id, OriginalName: *name, MimeType: *mimeType, Size: *size, CreatedAt: *createdAt}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
