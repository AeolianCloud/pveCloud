package order

const (
	StatusPending   = "pending"
	StatusCancelled = "cancelled"
	StatusClosed    = "closed"
)

func CanCancel(status string) bool {
	return status == StatusPending
}

func CanClose(status string) bool {
	return status == StatusPending
}

func IsKnownStatus(status string) bool {
	switch status {
	case "", StatusPending, StatusCancelled, StatusClosed:
		return true
	default:
		return false
	}
}
