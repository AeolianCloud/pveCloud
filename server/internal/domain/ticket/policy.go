package ticket

const (
	StatusWaitingAdmin = "waiting_admin"
	StatusWaitingUser  = "waiting_user"
	StatusClosed       = "closed"

	CategoryAccount   = "account"
	CategoryOrder     = "order"
	CategoryProduct   = "product"
	CategoryTechnical = "technical"
	CategoryBilling   = "billing"
	CategoryOther     = "other"

	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"

	SenderUser  = "user"
	SenderAdmin = "admin"
)

func IsKnownStatus(value string) bool {
	switch value {
	case "", StatusWaitingAdmin, StatusWaitingUser, StatusClosed:
		return true
	default:
		return false
	}
}

func IsKnownCategory(value string) bool {
	switch value {
	case CategoryAccount, CategoryOrder, CategoryProduct, CategoryTechnical, CategoryBilling, CategoryOther:
		return true
	default:
		return false
	}
}

func IsKnownCategoryOrEmpty(value string) bool {
	if value == "" {
		return true
	}
	return IsKnownCategory(value)
}

func IsKnownPriority(value string) bool {
	switch value {
	case "", PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func NormalizePriority(value string) string {
	if value == "" {
		return PriorityNormal
	}
	return value
}

func CanReply(status string) bool {
	return status != StatusClosed
}

func CanClose(status string) bool {
	return status != StatusClosed
}
