// Package invoice contains stable invoice workflow constants and state rules.
package invoice

const (
	TypeElectronicNormal = "electronic_normal"

	TitleTypePersonal = "personal"
	TitleTypeCompany  = "company"

	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusIssued     = "issued"
	StatusRejected   = "rejected"
	StatusCancelled  = "cancelled"

	FileRefType = "invoice_application"
)

func IsKnownStatus(status string) bool {
	switch status {
	case "", StatusPending, StatusProcessing, StatusIssued, StatusRejected, StatusCancelled:
		return true
	default:
		return false
	}
}

func IsActiveStatus(status string) bool {
	switch status {
	case StatusPending, StatusProcessing, StatusIssued:
		return true
	default:
		return false
	}
}

func CanCancel(status string) bool {
	return status == StatusPending
}

func CanAccept(status string) bool {
	return status == StatusPending
}

func CanReject(status string) bool {
	return status == StatusPending || status == StatusProcessing
}

func CanIssue(status string) bool {
	return status == StatusProcessing
}

func IsKnownTitleType(titleType string) bool {
	switch titleType {
	case TitleTypePersonal, TitleTypeCompany:
		return true
	default:
		return false
	}
}
