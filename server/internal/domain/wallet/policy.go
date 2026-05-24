// Package wallet contains stable wallet state constants and small policy helpers.
package wallet

const (
	CurrencyCNY = "CNY"

	AccountStatusActive   = "active"
	AccountStatusDisabled = "disabled"

	DirectionCredit = "credit"
	DirectionDebit  = "debit"

	EntryTypeRecharge = "recharge"
	EntryTypePayment  = "payment"
	EntryTypeRefund   = "refund"

	RelatedTypeRecharge = "recharge"
	RelatedTypePayment  = "payment"
	RelatedTypeRefund   = "refund"
	RelatedTypeOrder    = "order"

	RechargeStatusPending = "pending"
	RechargeStatusPaid    = "paid"
	RechargeStatusClosed  = "closed"
	RechargeStatusFailed  = "failed"
)

func IsKnownAccountStatus(status string) bool {
	switch status {
	case AccountStatusActive, AccountStatusDisabled:
		return true
	default:
		return false
	}
}

func IsKnownLedgerDirection(direction string) bool {
	switch direction {
	case DirectionCredit, DirectionDebit:
		return true
	default:
		return false
	}
}

func IsKnownEntryType(entryType string) bool {
	switch entryType {
	case EntryTypeRecharge, EntryTypePayment, EntryTypeRefund:
		return true
	default:
		return false
	}
}

func IsKnownRechargeStatus(status string) bool {
	switch status {
	case RechargeStatusPending, RechargeStatusPaid, RechargeStatusClosed, RechargeStatusFailed:
		return true
	default:
		return false
	}
}
