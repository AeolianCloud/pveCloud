package order

const (
	StatusPending      = "pending"
	StatusProvisioning = "provisioning"
	StatusFulfilled    = "fulfilled"
	StatusCancelled    = "cancelled"
	StatusClosed       = "closed"
)

func CanCancel(status string) bool {
	return status == StatusPending
}

func CanClose(status string) bool {
	return status == StatusPending || status == StatusFulfilled
}

func CanProvision(status string) bool {
	return status == StatusPending
}

func IsKnownStatus(status string) bool {
	switch status {
	case "", StatusPending, StatusProvisioning, StatusFulfilled, StatusCancelled, StatusClosed:
		return true
	default:
		return false
	}
}
