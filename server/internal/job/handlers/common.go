package handlers

import (
	"context"
	"fmt"

	"github.com/AeolianCloud/pveCloud/server/internal/job"
)

func unimplemented(taskType string) job.Handler {
	return func(ctx context.Context, task job.AsyncTask) job.HandlerResult {
		return job.PermanentFailure(fmt.Errorf("%s handler is not implemented yet", taskType))
	}
}
