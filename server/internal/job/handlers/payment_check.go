package handlers

import "github.com/AeolianCloud/pveCloud/server/internal/job"

func registerPaymentCheck(registry job.Registry) {
	registry.Register(job.TaskTypePaymentCheck, unimplemented(job.TaskTypePaymentCheck))
}
