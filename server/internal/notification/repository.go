package notification

import "context"

type Repository interface {
	Create(ctx context.Context, n Notification) (Notification, error)
	ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error)
	MarkRead(ctx context.Context, id uint64, userID uint64) error
	CountUnread(ctx context.Context, userID uint64) (int, error)
}
