package handlers

import "github.com/AeolianCloud/pveCloud/server/internal/job"

func registerInstanceCreate(registry job.Registry) {
	registry.Register(job.TaskTypeInstanceCreate, unimplemented(job.TaskTypeInstanceCreate))
}
