package textutil

import (
	"strconv"
	"strings"
)

func StringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func TrimTo(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func NormalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func Uint64String(id uint64) string {
	return strconv.FormatUint(id, 10)
}
