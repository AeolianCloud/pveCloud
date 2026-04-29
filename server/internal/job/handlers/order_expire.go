package handlers

import "github.com/AeolianCloud/pveCloud/server/internal/job"

func registerOrderExpire(registry job.Registry) {
	registry.Register(job.TaskTypeOrderExpire, unimplemented(job.TaskTypeOrderExpire))
}
