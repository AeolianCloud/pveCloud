package instance

const (
	StatusCreating  = "creating"
	StatusRunning   = "running"
	StatusStopped   = "stopped"
	StatusError     = "error"
	StatusReleasing = "releasing"
	StatusReleased  = "released"

	OperationProvision = "provision"
	OperationStart     = "start"
	OperationStop      = "stop"
	OperationRelease   = "release"
	OperationSync      = "sync"

	OperationStatusRunning   = "running"
	OperationStatusSucceeded = "succeeded"
	OperationStatusFailed    = "failed"

	TaskTypeOperationSync  = "instance_operation_sync"
	TaskTypeExpiryNotice   = "instance_expiry_notice"
	TaskTypeExpiryRelease  = "instance_expiry_release"
	TaskTypeEmailSend      = "notification_email_send"
	TaskTypeSMSPlaceholder = "notification_sms_placeholder"

	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusSucceeded = "succeeded"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"

	NotificationChannelEmail  = "email"
	NotificationChannelSMS    = "sms"
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
	NotificationStatusSkipped = "skipped"
)

func IsKnownStatus(status string) bool {
	switch status {
	case "", StatusCreating, StatusRunning, StatusStopped, StatusError, StatusReleasing, StatusReleased:
		return true
	default:
		return false
	}
}

func CanStart(status string) bool {
	return status == StatusStopped
}

func CanStop(status string) bool {
	return status == StatusRunning
}

func CanRelease(status string) bool {
	return status != StatusReleasing && status != StatusReleased
}

func MapVMStatus(status string) string {
	switch status {
	case StatusRunning:
		return StatusRunning
	case StatusStopped:
		return StatusStopped
	default:
		return StatusError
	}
}

func IsKnownTaskStatus(status string) bool {
	switch status {
	case "", TaskStatusPending, TaskStatusRunning, TaskStatusSucceeded, TaskStatusFailed, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

func IsKnownTaskType(taskType string) bool {
	switch taskType {
	case "", TaskTypeOperationSync, TaskTypeExpiryNotice, TaskTypeExpiryRelease, TaskTypeEmailSend, TaskTypeSMSPlaceholder:
		return true
	default:
		return false
	}
}
