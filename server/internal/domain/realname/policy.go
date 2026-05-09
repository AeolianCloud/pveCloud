package realname

import "strings"

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"

	ProviderAlipay = "alipay"
	ProviderWechat = "wechat"
	ProviderManual = "manual"
)

func CanManualReview(provider string, status string) bool {
	return strings.TrimSpace(provider) == ProviderManual && strings.TrimSpace(status) == StatusPending
}

func ShouldRejectApprovedWithoutDigest(digest string) bool {
	return strings.TrimSpace(digest) == ""
}

func ProviderUserMessage(resultCode string, resultMessage string) string {
	if strings.TrimSpace(resultMessage) != "" {
		return strings.TrimSpace(resultMessage)
	}
	if strings.TrimSpace(resultCode) != "" {
		return "实名供应商核验失败：" + strings.TrimSpace(resultCode)
	}
	return "实名供应商核验失败"
}

func HasApprovedDigestConflict(duplicateCount int64) bool {
	return duplicateCount > 0
}

func AllowCallbackReplay(replayKey string, inserted bool) bool {
	if strings.TrimSpace(replayKey) == "" {
		return true
	}
	return inserted
}
