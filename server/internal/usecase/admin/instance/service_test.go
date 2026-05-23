package instance

import (
	"testing"
	"time"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
)

func TestReleaseCompletionUpdatesMarksExpiryOnlyForWorkerRelease(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

	workerUpdates := releaseCompletionUpdates(mysqlinstance.Operation{Action: domaininstance.OperationRelease}, now)
	if workerUpdates["expire_released_at"] != now {
		t.Fatalf("worker release should mark expire_released_at, got %#v", workerUpdates["expire_released_at"])
	}

	adminID := uint64(7)
	adminUpdates := releaseCompletionUpdates(mysqlinstance.Operation{Action: domaininstance.OperationRelease, AdminID: &adminID}, now)
	if _, ok := adminUpdates["expire_released_at"]; ok {
		t.Fatalf("admin release must not mark expire_released_at: %#v", adminUpdates)
	}
}
