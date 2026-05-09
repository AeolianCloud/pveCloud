package catalog

import "strings"

const (
	ProductTypeServer = "server"
	StatusActive      = "active"
	StatusSoldOut     = "sold_out"
)

func IsPublicServerProduct(productType string, status string, visible bool) bool {
	return strings.TrimSpace(productType) == ProductTypeServer &&
		strings.TrimSpace(status) == StatusActive &&
		visible
}

func IsPublicServerPlan(status string, visible bool) bool {
	status = strings.TrimSpace(status)
	return visible && (status == StatusActive || status == StatusSoldOut)
}

func HasRenderablePlanParts(priceCount int, regionCount int, templateCount int) bool {
	return priceCount > 0 && regionCount > 0 && templateCount > 0
}
