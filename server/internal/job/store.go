package job

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

/**
 * GormStore 使用 MariaDB async_tasks 表存储任务状态。
 */
type GormStore struct {
	db       *gorm.DB
	workerID string
	lockTTL  time.Duration
}

/**
 * NewGormStore 创建异步任务数据库存储。
 */
func NewGormStore(db *gorm.DB, workerID string, lockTTL time.Duration) *GormStore {
	return &GormStore{db: db, workerID: strings.TrimSpace(workerID), lockTTL: lockTTL}
}

/**
 * Claim 领取一条到期的 pending 任务，并写入运行锁。
 */
func (s *GormStore) Claim(ctx context.Context, now time.Time) (AsyncTask, error) {
	var claimed AsyncTask
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND run_at <= ?", StatusPending, now).
			Order("priority DESC, run_at ASC, id ASC").
			Limit(1).
			First(&claimed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoTask
		}
		if err != nil {
			return err
		}

		lockedUntil := now.Add(s.lockTTL)
		if err := tx.Model(&AsyncTask{}).
			Where("id = ? AND status = ?", claimed.ID, StatusPending).
			Updates(map[string]any{
				"status":       StatusRunning,
				"locked_by":    s.workerID,
				"locked_until": lockedUntil,
			}).Error; err != nil {
			return err
		}

		claimed.Status = StatusRunning
		claimed.LockedBy = &s.workerID
		claimed.LockedUntil = &lockedUntil
		return nil
	})
	if err != nil {
		return AsyncTask{}, err
	}
	return claimed, nil
}

/**
 * MarkSucceeded 标记任务成功并清空运行锁。
 */
func (s *GormStore) MarkSucceeded(ctx context.Context, task AsyncTask, result any, now time.Time) error {
	resultJSON, err := jsonStringPtr(result)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&AsyncTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"status":       StatusSucceeded,
			"locked_by":    nil,
			"locked_until": nil,
			"last_error":   nil,
			"result":       resultJSON,
			"finished_at":  now,
		}).Error
}

/**
 * MarkRetryableFailure 记录可重试失败并把任务放回 pending。
 */
func (s *GormStore) MarkRetryableFailure(ctx context.Context, task AsyncTask, err error, nextRunAt time.Time, now time.Time) error {
	retryCount := task.RetryCount + 1
	return s.db.WithContext(ctx).Model(&AsyncTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"status":       StatusPending,
			"retry_count":  retryCount,
			"run_at":       nextRunAt,
			"locked_by":    nil,
			"locked_until": nil,
			"last_error":   errorMessage(err),
			"result":       nil,
			"finished_at":  nil,
		}).Error
}

/**
 * MarkFailed 标记任务不可继续重试。
 */
func (s *GormStore) MarkFailed(ctx context.Context, task AsyncTask, err error, now time.Time) error {
	return s.db.WithContext(ctx).Model(&AsyncTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"status":       StatusFailed,
			"locked_by":    nil,
			"locked_until": nil,
			"last_error":   errorMessage(err),
			"result":       nil,
			"finished_at":  now,
		}).Error
}

func jsonStringPtr(value any) (*string, error) {
	if value == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	result := string(encoded)
	return &result, nil
}
