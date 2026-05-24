package order

const (
	StatusPending      = "pending"
	StatusProvisioning = "provisioning"
	StatusFulfilled    = "fulfilled"
	StatusError        = "error"
	StatusCancelled    = "cancelled"
	StatusClosed       = "closed"

	TypePurchase = "purchase"
	TypeRenewal  = "renewal"

	PaymentStatusUnpaid          = "unpaid"
	PaymentStatusPaid            = "paid"
	PaymentStatusManualConfirmed = "manual_confirmed"
	PaymentStatusRefunded        = "refunded"
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

func CanConfirmRenewal(status string, orderType string) bool {
	return status == StatusPending && orderType == TypeRenewal
}

func IsKnownStatus(status string) bool {
	switch status {
	case "", StatusPending, StatusProvisioning, StatusFulfilled, StatusError, StatusCancelled, StatusClosed:
		return true
	default:
		return false
	}
}

func IsKnownType(orderType string) bool {
	switch orderType {
	case "", TypePurchase, TypeRenewal:
		return true
	default:
		return false
	}
}

func BillingCycleMonths(cycle string) (int, bool) {
	switch cycle {
	case "monthly":
		return 1, true
	case "quarterly":
		return 3, true
	case "semi_yearly":
		return 6, true
	case "yearly":
		return 12, true
	default:
		return 0, false
	}
}
