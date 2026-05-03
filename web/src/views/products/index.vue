<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getServerCatalog, type ServerCatalogPlan, type ServerCatalogProduct } from '../../api/product-catalog'

const loading = ref(false)
const products = ref<ServerCatalogProduct[]>([])

const allPlans = computed(() => products.value.flatMap((product) => product.plans))
const primaryProduct = computed(() => products.value[0])

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function monthlyPrice(plan: ServerCatalogPlan) {
  return plan.prices.find((price) => price.billing_cycle === 'monthly') || plan.prices[0]
}

function memory(plan: ServerCatalogPlan) {
  return `${Math.round(plan.memory_mb / 1024)} GB`
}

onMounted(async () => {
  loading.value = true
  try {
    const catalog = await getServerCatalog()
    products.value = catalog.products
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="page content-page">
    <div class="section-pad">
      <div class="sec-header" style="margin-bottom: clamp(28px, 4vw, 48px);">
        <p class="label">{{ primaryProduct?.slug || '产品规格' }}</p>
        <h2>{{ primaryProduct?.name || '云服务器产品目录' }}</h2>
        <p v-if="primaryProduct?.summary">{{ primaryProduct.summary }}</p>
        <p v-if="primaryProduct?.description" class="product-description">{{ primaryProduct.description }}</p>
        <p v-if="!primaryProduct">展示服务器产品、销售地域和系统模板。当前不包含订单、支付和实例开通。</p>
      </div>

      <div v-if="loading" class="empty-state">正在读取产品目录...</div>
      <template v-else>
        <div v-for="product in products" :key="product.product_no" style="margin-bottom: 48px;">
          <div class="spec-grid">
            <div v-for="plan in product.plans" :key="plan.plan_no" class="spec-card" :class="{ featured: plan.is_featured }">
              <div class="spec-card-head">
                <span class="plan-tag" :class="plan.is_featured ? 'primary' : 'green'">{{ plan.status === 'sold_out' ? '售罄' : plan.is_featured ? '推荐' : '可选' }}</span>
                <h3>{{ plan.name }}</h3>
                <p class="plan-summary">{{ plan.summary || '固定规格云服务器套餐' }}</p>
              </div>
              <div class="spec-price-row"><strong>¥{{ yuan(monthlyPrice(plan).price_cents) }}</strong><span>/月起</span></div>
              <div class="spec-list">
                <div class="spec-item"><span class="dot"></span><span><strong>CPU</strong> {{ plan.cpu_cores }} vCPU</span></div>
                <div class="spec-item"><span class="dot"></span><span><strong>内存</strong> {{ memory(plan) }}</span></div>
                <div class="spec-item"><span class="dot"></span><span><strong>系统盘</strong> {{ plan.system_disk_gb }} GB</span></div>
                <div class="spec-item"><span class="dot"></span><span><strong>流量</strong> {{ plan.traffic_gb == null ? '暂不承诺' : `${plan.traffic_gb} GB/月` }}</span></div>
                <div class="spec-item"><span class="dot"></span><span><strong>地域</strong> {{ plan.regions.map((region) => region.name).join(' / ') }}</span></div>
                <div class="spec-item"><span class="dot"></span><span><strong>模板</strong> {{ plan.os_templates.map((template) => template.name).join(' / ') }}</span></div>
              </div>
              <RouterLink to="/pricing" class="btn btn-outline btn-sm" style="width:100%">查看价格</RouterLink>
            </div>
          </div>
        </div>
        <div v-if="allPlans.length === 0" class="empty-state">暂无公开套餐，请稍后查看。</div>
      </template>
    </div>
  </section>
</template>

<style scoped>
.empty-state { padding: 40px 0; text-align: center; color: var(--c-text-2); }
.product-description,
.plan-summary { color: var(--c-text-2); line-height: 1.7; white-space: pre-line; }
</style>
