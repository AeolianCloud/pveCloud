package catalog

import (
	"testing"

	mysqlcatalog "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/catalog"
	"github.com/stretchr/testify/require"
)

func TestCatalogProductsRequireRenderablePlanParts(t *testing.T) {
	products := []mysqlcatalog.Product{
		{ID: 1, ProductNo: "P1", Slug: "server-a", Name: "Server A"},
		{ID: 2, ProductNo: "P2", Slug: "server-b", Name: "Server B"},
	}
	plans := []mysqlcatalog.ProductPlan{
		{ID: 10, ProductID: 1, PlanNo: "A1", Code: "a1", Name: "A1"},
		{ID: 20, ProductID: 2, PlanNo: "B1", Code: "b1", Name: "B1"},
	}
	prices := map[uint64][]ServerCatalogPlanPrice{
		10: []ServerCatalogPlanPrice{{BillingCycle: "monthly", PriceCents: 1000, Currency: "CNY"}},
		20: []ServerCatalogPlanPrice{{BillingCycle: "monthly", PriceCents: 1000, Currency: "CNY"}},
	}
	regions := map[uint64][]ServerCatalogRegion{
		10: []ServerCatalogRegion{{RegionNo: "R1", Code: "gz", Name: "Guangzhou"}},
	}
	templates := map[uint64][]ServerCatalogOSTemplate{
		10: []ServerCatalogOSTemplate{{TemplateNo: "T1", Code: "debian", Name: "Debian"}},
		20: []ServerCatalogOSTemplate{{TemplateNo: "T1", Code: "debian", Name: "Debian"}},
	}

	result := catalogProducts(products, plans, prices, regions, templates)

	require.Len(t, result, 1)
	require.Equal(t, "P1", result[0].ProductNo)
	require.Len(t, result[0].Plans, 1)
	require.Equal(t, "A1", result[0].Plans[0].PlanNo)
}
