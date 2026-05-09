package systemconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrimitiveAndSecretPolicy(t *testing.T) {
	require.NoError(t, ValidatePrimitiveValue("bool", "true"))
	require.ErrorIs(t, ValidatePrimitiveValue("bool", "yes"), ErrBoolValue)
	require.NoError(t, ValidatePrimitiveValue("int", "1"))
	require.ErrorIs(t, ValidatePrimitiveValue("int", "0"), ErrPositiveInt)
	require.True(t, PreserveSecretWhenBlank(true, " "))
	require.False(t, PreserveSecretWhenBlank(false, " "))
}

func TestRealNameConfigPolicy(t *testing.T) {
	require.Equal(t, []string{"alipay", "wechat"}, SplitProviders(" alipay, wechat "))
	require.NoError(t, ValidateRealNameProviders("alipay,wechat"))
	require.ErrorIs(t, ValidateRealNameProviders("manual"), ErrProviderValue)

	alipay := map[string]string{
		"real_name.alipay.app_id":            "app",
		"real_name.alipay.gateway_url":       "https://gateway",
		"real_name.alipay.return_url":        "https://return",
		"real_name.callback_base_url":        "https://api/callback",
		"real_name.alipay.app_private_key":   "private",
		"real_name.alipay.alipay_public_key": "public",
	}
	require.True(t, AlipayConfigComplete(alipay))
	delete(alipay, "real_name.alipay.app_private_key")
	require.False(t, AlipayConfigComplete(alipay))

	wechat := map[string]string{
		"real_name.wechat.secret_id":    "id",
		"real_name.wechat.secret_key":   "key",
		"real_name.wechat.region":       "ap-guangzhou",
		"real_name.wechat.endpoint":     "faceid.tencentcloudapi.com",
		"real_name.wechat.rule_id":      "rule",
		"real_name.wechat.redirect_url": "https://return",
	}
	require.True(t, WechatConfigComplete(wechat))
	delete(wechat, "real_name.wechat.redirect_url")
	require.False(t, WechatConfigComplete(wechat))
}
