package audit

import (
	"context"
)

// Repository persists audit events to the database.
type Repository interface {
	Record(ctx context.Context, event string, businessType string, businessID uint64, payload []byte) error
}
