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
