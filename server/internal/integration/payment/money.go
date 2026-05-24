package payment

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func centsToYuan(cents uint64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func yuanToCents(value string) (uint64, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, err
	}
	if amount < 0 {
		return 0, fmt.Errorf("negative amount")
	}
	return uint64(math.Round(amount * 100)), nil
}
