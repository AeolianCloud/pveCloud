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

	TagVisibilityPublic   = "public"
	TagVisibilityInternal = "internal"

	TagStatusActive   = "active"
	TagStatusDisabled = "disabled"

	SLAStatusNormal               = "normal"
	SLAStatusFirstResponseOverdue = "first_response_overdue"
	SLAStatusResolutionOverdue    = "resolution_overdue"
	EventTypeAssign               = "assign"
	EventTypeTransfer             = "transfer"
	EventTypeCollaboratorAdd      = "collaborator_add"
	EventTypeCollaboratorRemove   = "collaborator_remove"
	EventTypeInternalNote         = "internal_note"
	EventTypePriorityUpgrade      = "priority_upgrade"
	EventTypeTagsReplace          = "tags_replace"
	EventTypeAdminReply           = "admin_reply"
	EventTypeAdminClose           = "admin_close"
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

func IsKnownTagVisibility(value string) bool {
	switch value {
	case TagVisibilityPublic, TagVisibilityInternal:
		return true
	default:
		return false
	}
}

func IsKnownTagVisibilityOrEmpty(value string) bool {
	if value == "" {
		return true
	}
	return IsKnownTagVisibility(value)
}

func IsKnownTagStatus(value string) bool {
	switch value {
	case TagStatusActive, TagStatusDisabled:
		return true
	default:
		return false
	}
}

func IsKnownTagStatusOrEmpty(value string) bool {
	if value == "" {
		return true
	}
	return IsKnownTagStatus(value)
}

func IsKnownSLAStatus(value string) bool {
	switch value {
	case "", SLAStatusNormal, SLAStatusFirstResponseOverdue, SLAStatusResolutionOverdue:
		return true
	default:
		return false
	}
}

func PriorityRank(value string) int {
	switch value {
	case PriorityLow:
		return 1
	case PriorityNormal:
		return 2
	case PriorityHigh:
		return 3
	case PriorityUrgent:
		return 4
	default:
		return 0
	}
}

func CanUpgradePriority(current string, next string) bool {
	return PriorityRank(next) > PriorityRank(current)
}

func CanReply(status string) bool {
	return status != StatusClosed
}

func CanClose(status string) bool {
	return status != StatusClosed
}
