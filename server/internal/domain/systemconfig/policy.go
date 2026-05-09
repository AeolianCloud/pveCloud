package systemconfig

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrBoolValue     = errors.New("布尔配置只能填写 true 或 false")
	ErrPositiveInt   = errors.New("数字配置必须为正整数")
	ErrProviderValue = errors.New("实名供应商只允许 alipay 或 wechat")
)

func ValidatePrimitiveValue(valueType string, value string) error {
	switch strings.TrimSpace(valueType) {
	case "bool":
		trimmed := strings.TrimSpace(value)
		if trimmed != "true" && trimmed != "false" {
			return ErrBoolValue
		}
	case "int":
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || parsed <= 0 {
			return ErrPositiveInt
		}
	}
	return nil
}

func PreserveSecretWhenBlank(isSecret bool, value string) bool {
	return isSecret && strings.TrimSpace(value) == ""
}

func SplitProviders(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.ToLower(strings.TrimSpace(part))
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func ValidateRealNameProviders(value string) error {
	for _, provider := range SplitProviders(value) {
		if provider != "alipay" && provider != "wechat" {
			return ErrProviderValue
		}
	}
	return nil
}

func AlipayConfigComplete(configs map[string]string) bool {
	return strings.TrimSpace(configs["real_name.alipay.app_id"]) != "" &&
		strings.TrimSpace(configs["real_name.alipay.gateway_url"]) != "" &&
		strings.TrimSpace(configs["real_name.alipay.return_url"]) != "" &&
		(strings.TrimSpace(configs["real_name.alipay.notify_url"]) != "" || strings.TrimSpace(configs["real_name.callback_base_url"]) != "") &&
		strings.TrimSpace(configs["real_name.alipay.app_private_key"]) != "" &&
		strings.TrimSpace(configs["real_name.alipay.alipay_public_key"]) != ""
}

func WechatConfigComplete(configs map[string]string) bool {
	return strings.TrimSpace(configs["real_name.wechat.secret_id"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.secret_key"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.region"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.endpoint"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.rule_id"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.redirect_url"]) != ""
}
