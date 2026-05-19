package asynctask

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	adminaudit "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	adminsupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

const objectType = "async_task"

type AdminAuditService = adminaudit.AdminAuditService
type AdminAuditWriteInput = adminaudit.AdminAuditWriteInput

type Service struct {
	db    *gorm.DB
	tasks *mysqlinstance.Repository
	audit *AdminAuditService
}

func NewService(db *gorm.DB, audit *AdminAuditService) *Service {
	if audit == nil {
		audit = adminaudit.NewAdminAuditService(db)
	}
	return &Service{db: db, tasks: mysqlinstance.NewRepository(db), audit: audit}
}

func (s *Service) List(ctx context.Context, query admindto.AsyncTaskListQuery) (admindto.PageResponse[admindto.AsyncTaskItem], error) {
	if !domaininstance.IsKnownTaskType(query.TaskType) {
		return admindto.PageResponse[admindto.AsyncTaskItem]{}, apperrors.ErrValidation.WithMessage("任务类型不支持")
	}
	if !domaininstance.IsKnownTaskStatus(query.Status) {
		return admindto.PageResponse[admindto.AsyncTaskItem]{}, apperrors.ErrValidation.WithMessage("任务状态不支持")
	}
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.tasks.ListTasks(ctx, mysqlinstance.TaskFilters{TaskType: query.TaskType, Status: query.Status, ObjectType: query.ObjectType, ObjectNo: query.ObjectNo, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.AsyncTaskItem]{}, err
	}
	items := make([]admindto.AsyncTaskItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, taskItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Retry(ctx context.Context, operatorID uint64, taskNo string, req admindto.AsyncTaskRetryRequest) (admindto.AsyncTaskItem, error) {
	var updatedTaskNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		task, err := s.tasks.TaskForUpdate(ctx, tx, strings.TrimSpace(taskNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("异步任务不存在")
		}
		if err != nil {
			return err
		}
		if task.Status != domaininstance.TaskStatusFailed {
			return apperrors.ErrConflict.WithMessage("只有失败任务可以重试")
		}
		updates := map[string]any{"status": domaininstance.TaskStatusPending, "scheduled_at": time.Now(), "locked_by": nil, "locked_until": nil, "last_error_code": nil, "last_error_message": nil, "completed_at": nil}
		if err := s.tasks.UpdateTask(ctx, tx, task.ID, updates); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "async_task.retry", ObjectType: objectType, ObjectID: task.TaskNo, BeforeData: auditSnapshot(task), AfterData: updates, Remark: firstNonEmptyValue(req.Remark, "人工重试异步任务")}); err != nil {
			return err
		}
		updatedTaskNo = task.TaskNo
		return nil
	})
	if err != nil {
		return admindto.AsyncTaskItem{}, err
	}
	updated, err := s.tasks.TaskByNo(ctx, updatedTaskNo)
	if err != nil {
		return admindto.AsyncTaskItem{}, err
	}
	return taskItem(updated), nil
}

func taskItem(task mysqlinstance.Task) admindto.AsyncTaskItem {
	return admindto.AsyncTaskItem{TaskNo: task.TaskNo, TaskType: task.TaskType, Status: task.Status, ObjectType: task.ObjectType, ObjectNo: task.ObjectNo, ScheduledAt: task.ScheduledAt, Attempts: task.Attempts, MaxAttempts: task.MaxAttempts, LastErrorCode: task.LastErrorCode, LastErrorMessage: task.LastErrorMessage, LockedBy: task.LockedBy, LockedUntil: task.LockedUntil, CreatedAt: task.CreatedAt, CompletedAt: task.CompletedAt}
}

func auditSnapshot(task mysqlinstance.Task) map[string]any {
	return map[string]any{"task_no": task.TaskNo, "task_type": task.TaskType, "status": task.Status, "attempts": task.Attempts, "max_attempts": task.MaxAttempts, "object_type": task.ObjectType, "object_no": task.ObjectNo}
}

func firstNonEmptyValue(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	return strings.TrimSpace(*value)
}
