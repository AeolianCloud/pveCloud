package payment

import (
	"fmt"
	"net/url"
	"strings"
)

func ValidateProviderConfig(cfg Config, method string) error {
	if strings.TrimSpace(cfg.Provider) == "" {
		return fmt.Errorf("%w: provider missing", ErrIncompleteConfig)
	}
	switch cfg.Provider {
	case ProviderAlipay:
		return validateRequired(cfg, "payment.alipay.app_id", "payment.alipay.gateway_url", "payment.alipay.app_private_key", "payment.alipay.alipay_public_key", "payment.alipay.notify_url", "payment.alipay.return_url")
	case ProviderWechat:
		keys := []string{"payment.wechat.app_id", "payment.wechat.mch_id", "payment.wechat.api_v3_key", "payment.wechat.mch_private_key", "payment.wechat.mch_certificate_serial_no", "payment.wechat.platform_public_key_id", "payment.wechat.platform_public_key", "payment.wechat.notify_url"}
		if method == MethodWechatH5 {
			keys = append(keys, "payment.wechat.h5_scene_info")
		}
		return validateRequired(cfg, keys...)
	default:
		return ErrUnsupportedProvider
	}
}

func ValidateProductionConfig(cfg Config, method string) error {
	if err := ValidateProviderConfig(cfg, method); err != nil {
		return err
	}
	switch cfg.Provider {
	case ProviderAlipay:
		return requireHTTPSURL(cfg.Value("payment.alipay.notify_url"), "payment.alipay.notify_url", false)
	case ProviderWechat:
		if err := requireHTTPSURL(cfg.Value("payment.wechat.notify_url"), "payment.wechat.notify_url", true); err != nil {
			return err
		}
	}
	return nil
}

func validateRequired(cfg Config, keys ...string) error {
	for _, key := range keys {
		if cfg.Value(key) == "" {
			return fmt.Errorf("%w: %s", ErrIncompleteConfig, key)
		}
	}
	return nil
}

func requireHTTPSURL(rawURL, key string, noQuery bool) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("%w: %s must be https url", ErrIncompleteConfig, key)
	}
	if noQuery && parsed.RawQuery != "" {
		return fmt.Errorf("%w: %s must not contain query", ErrIncompleteConfig, key)
	}
	return nil
}

func Summary(fields map[string]any) string {
	pairs := make([]string, 0, len(fields))
	for key, value := range fields {
		if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
			pairs = append(pairs, key+"="+text)
		}
	}
	return strings.Join(pairs, ";")
}
