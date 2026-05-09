<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getServerCatalog, type ServerCatalogProduct } from '../../api/product-catalog'

const loading = ref(false)
const products = ref<ServerCatalogProduct[]>([])
const activeProductNo = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const catalog = await getServerCatalog()
    products.value = catalog.products
    activeProductNo.value = catalog.products[0]?.product_no || ''
  } finally {
    loading.value = false
  }
})

const activeProduct = computed(() => products.value.find((product) => product.product_no === activeProductNo.value) || null)

const allRegions = computed(() => {
  const map = new Map<string, string>()
  products.value.forEach((product) => {
    product.plans.forEach((plan) => {
      plan.regions.forEach((region) => map.set(region.region_no, region.name))
    })
  })
  return [...map.values()]
})

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function getMinPrice(product: ServerCatalogProduct) {
  const prices = product.plans
    .map((plan) => plan.prices.find((price) => price.billing_cycle === 'monthly') || plan.prices[0])
    .filter(Boolean)
    .map((price) => price.price_cents)
  return prices.length ? Math.min(...prices) : 0
}
</script>

<template>
  <section class="products-page page-shell">
    <div class="page-hero surface">
      <div>
        <p class="section-label">Products</p>
        <h1 class="page-title">服务器产品目录</h1>
        <p class="page-copy">这里展示公开目录返回的产品、套餐、地域和系统模板，不展示节点、资源池、库存扣减或实例信息。</p>
      </div>
      <div class="hero-meta">
        <span>{{ products.length }} 个产品线</span>
        <span>{{ allRegions.length }} 个销售地域</span>
      </div>
    </div>

    <div v-if="loading" class="loading-panel surface">
      <div class="spinner"></div>
      <span>正在加载产品目录...</span>
    </div>

    <div v-else-if="products.length === 0" class="loading-panel surface">
      暂无可展示的产品目录。
    </div>

    <div v-else class="catalog-layout">
      <aside class="product-list surface">
        <button
          v-for="product in products"
          :key="product.product_no"
          class="product-tab"
          :class="{ active: activeProductNo === product.product_no }"
          type="button"
          @click="activeProductNo = product.product_no"
        >
          <span>{{ product.name }}</span>
          <small>{{ product.plans.length }} 个套餐</small>
        </button>
      </aside>

      <main v-if="activeProduct" class="product-detail">
        <section class="summary-card surface">
          <div>
            <p class="section-label">Selected Product</p>
            <h2>{{ activeProduct.name }}</h2>
            <p>{{ activeProduct.description || activeProduct.summary || '服务器产品展示信息以后端公开目录为准。' }}</p>
          </div>
          <div class="price-badge">
            <span>月付起</span>
            <strong>¥{{ yuan(getMinPrice(activeProduct)) }}</strong>
          </div>
        </section>

        <div class="plan-list">
          <article v-for="plan in activeProduct.plans" :key="plan.plan_no" class="plan-row card">
            <div class="plan-main">
              <div class="plan-title">
                <h3>{{ plan.name }}</h3>
                <span v-if="plan.is_featured" class="tag tag-primary">推荐</span>
                <span v-if="plan.status === 'sold_out'" class="tag tag-warning">售罄</span>
              </div>
              <p>{{ plan.summary || '固定规格服务器套餐' }}</p>
            </div>
            <dl class="spec-strip">
              <div><dt>CPU</dt><dd>{{ plan.cpu_cores }} vCPU</dd></div>
              <div><dt>内存</dt><dd>{{ Math.round(plan.memory_mb / 1024) }} GB</dd></div>
              <div><dt>磁盘</dt><dd>{{ plan.system_disk_gb }} GB</dd></div>
              <div><dt>带宽</dt><dd>{{ plan.bandwidth_mbps }} Mbps</dd></div>
            </dl>
            <div class="plan-footer">
              <span>{{ plan.regions.map((region) => region.name).join(' / ') || '暂无地域' }}</span>
              <RouterLink class="btn btn-outline btn-sm" to="/pricing">查看价格</RouterLink>
            </div>
          </article>
        </div>
      </main>
    </div>
  </section>
</template>

<style scoped>
.products-page {
  display: grid;
  gap: 24px;
}

.page-hero {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  padding: clamp(24px, 4vw, 38px);
}

.page-hero .page-copy {
  max-width: 720px;
  margin-top: 14px;
}

.hero-meta {
  display: grid;
  gap: 10px;
  min-width: 180px;
}

.hero-meta span {
  padding: 12px 14px;
  border-radius: 12px;
  color: var(--c-text-2);
  background: var(--c-surface-dim);
  font-weight: 800;
}

.loading-panel {
  min-height: 220px;
  display: grid;
  place-items: center;
  gap: 12px;
  color: var(--c-text-2);
}

.catalog-layout {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 20px;
  align-items: start;
}

.product-list {
  position: sticky;
  top: 92px;
  display: grid;
  gap: 8px;
  padding: 12px;
}

.product-tab {
  display: grid;
  gap: 4px;
  padding: 14px;
  border-radius: 12px;
  text-align: left;
  cursor: pointer;
  border: 1px solid transparent;
}

.product-tab span {
  font-weight: 800;
}

.product-tab small {
  color: var(--c-text-3);
}

.product-tab.active {
  border-color: var(--c-primary);
  color: var(--c-primary);
  background: var(--c-primary-soft);
}

.product-detail,
.plan-list {
  display: grid;
  gap: 16px;
}

.summary-card {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  padding: 26px;
}

.summary-card h2 {
  margin-top: 8px;
  font-size: clamp(1.7rem, 3vw, 2.4rem);
  letter-spacing: -0.05em;
}

.summary-card p {
  max-width: 720px;
  margin-top: 10px;
  color: var(--c-text-2);
  line-height: 1.7;
}

.price-badge {
  display: grid;
  gap: 4px;
  min-width: 140px;
  padding: 18px;
  border-radius: 14px;
  background: var(--c-primary-soft);
}

.price-badge span {
  color: var(--c-text-2);
  font-weight: 700;
}

.price-badge strong {
  font-size: 2rem;
  letter-spacing: -0.05em;
}

.plan-row {
  display: grid;
  gap: 18px;
  padding: 20px;
}

.plan-title,
.plan-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.plan-title {
  justify-content: flex-start;
}

.plan-title h3 {
  font-size: 1.22rem;
  letter-spacing: -0.03em;
}

.plan-main p,
.plan-footer span {
  color: var(--c-text-2);
  line-height: 1.65;
}

.spec-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.spec-strip div {
  padding: 12px;
  border-radius: 12px;
  background: var(--c-surface-dim);
}

.spec-strip dt {
  color: var(--c-text-3);
  font-size: 0.78rem;
}

.spec-strip dd {
  margin: 4px 0 0;
  font-weight: 800;
}

@media (max-width: 980px) {
  .catalog-layout {
    grid-template-columns: 1fr;
  }

  .product-list {
    position: static;
  }

  .page-hero,
  .summary-card {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 680px) {
  .spec-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
