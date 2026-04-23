package task

import (
	"context"
	"fmt"
	"time"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) CreateTask(ctx context.Context, in CreateInput) (Task, error) {
	if in.TaskType == "" || in.BusinessType == "" || in.BusinessID == 0 {
		return Task{}, errorsx.ErrBadRequest
	}

	existing, found, err := s.repo.FindByBusinessKey(ctx, in.TaskType, in.BusinessType, in.BusinessID)
	if err != nil {
		return Task{}, err
	}
	if found {
		return existing, nil
	}

	now := s.now()
	return s.repo.CreateTask(ctx, CreateTaskParams{
		TaskType:      in.TaskType,
		BusinessType:  in.BusinessType,
		BusinessID:    in.BusinessID,
		Status:        "pending",
		Payload:       in.Payload,
		NextRunAt:     now,
		MaxRetryCount: 5,
	})
}

func (s *Service) ListTasks(ctx context.Context, limit int) ([]Task, error) {
	return s.repo.ListTasks(ctx, limit)
}

func NewTaskNo(now time.Time) string {
	return fmt.Sprintf("T%d", now.UnixNano())
}
