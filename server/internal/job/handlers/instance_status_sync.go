package handlers

import "github.com/AeolianCloud/pveCloud/server/internal/job"

func registerInstanceStatusSync(registry job.Registry) {
	registry.Register(job.TaskTypeInstanceStatusSync, unimplemented(job.TaskTypeInstanceStatusSync))
}
