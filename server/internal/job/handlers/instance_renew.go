package handlers

import "github.com/AeolianCloud/pveCloud/server/internal/job"

func registerInstanceRenew(registry job.Registry) {
	registry.Register(job.TaskTypeInstanceRenew, unimplemented(job.TaskTypeInstanceRenew))
}
