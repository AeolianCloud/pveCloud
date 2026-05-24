// Package payment contains stable payment and refund state constants.
package payment

const (
	ProviderAlipay = "alipay"
	ProviderWechat = "wechat"
	ProviderWallet = "wallet"

	MethodAlipayPage    = "alipay_page"
	MethodAlipayWap     = "alipay_wap"
	MethodWechatNative  = "wechat_native"
	MethodWechatH5      = "wechat_h5"
	MethodWalletBalance = "wallet_balance"

	StatusPending  = "pending"
	StatusPaid     = "paid"
	StatusClosed   = "closed"
	StatusFailed   = "failed"
	StatusRefunded = "refunded"

	RefundStatusPending   = "pending"
	RefundStatusSucceeded = "succeeded"
	RefundStatusFailed    = "failed"

	EffectStatusActive   = "active"
	EffectStatusReverted = "reverted"

	EffectTypePurchaseInstance = "purchase_instance"
	EffectTypeRenewalExtension = "renewal_extension"
)

func IsKnownProvider(provider string) bool {
	switch provider {
	case ProviderAlipay, ProviderWechat, ProviderWallet:
		return true
	default:
		return false
	}
}

func IsKnownMethod(method string) bool {
	switch method {
	case MethodAlipayPage, MethodAlipayWap, MethodWechatNative, MethodWechatH5, MethodWalletBalance:
		return true
	default:
		return false
	}
}

func ProviderSupportsMethod(provider, method string) bool {
	switch provider {
	case ProviderAlipay:
		return method == MethodAlipayPage || method == MethodAlipayWap
	case ProviderWechat:
		return method == MethodWechatNative || method == MethodWechatH5
	case ProviderWallet:
		return method == MethodWalletBalance
	default:
		return false
	}
}

func IsTerminalStatus(status string) bool {
	switch status {
	case StatusPaid, StatusClosed, StatusFailed, StatusRefunded:
		return true
	default:
		return false
	}
}
