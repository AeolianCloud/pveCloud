<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'

const loading = ref(false)
const plans = ref<ServerCatalogPlan[]>([])

const visiblePlans = computed(() => plans.value.slice(0, 4))

const cycleLabels: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  semi_yearly: '半年付',
  yearly: '年付',
}

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function monthlyPrice(plan: ServerCatalogPlan) {
  return plan.prices.find((price) => price.billing_cycle === 'monthly') || plan.prices[0]
}

function memory(plan: ServerCatalogPlan) {
  return `${Math.round(plan.memory_mb / 1024)} GB`
}

function ctaText(plan: ServerCatalogPlan) {
  return plan.status === 'sold_out' ? '暂时售罄' : '购买功能即将开放'
}

onMounted(async () => {
  loading.value = true
  try {
    const catalog = await getServerCatalog()
    plans.value = catalog.products.flatMap((product) => product.plans)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="page content-page">
    <div class="section-pad">
      <div class="sec-header center" style="margin-bottom: clamp(28px, 4vw, 48px);">
        <p class="label">价格方案</p>
        <h2>服务器套餐价格</h2>
        <p>价格、地域和系统模板均来自后台产品目录。当前阶段不开放下单和支付。</p>
      </div>

      <div v-if="loading" class="empty-state">正在读取价格目录...</div>
      <div v-else-if="visiblePlans.length === 0" class="empty-state">暂无公开价格，请先在后台配置套餐、价格、地域和系统模板。</div>

      <div v-else class="pricing-preview-grid" style="margin-bottom: clamp(40px, 6vw, 72px);">
        <div v-for="plan in visiblePlans" :key="plan.plan_no" class="pricing-preview-card" :class="{ featured: plan.is_featured }">
          <span class="plan-tag" :class="plan.is_featured ? 'primary' : 'green'">{{ plan.status === 'sold_out' ? '售罄' : plan.is_featured ? '推荐' : '可选' }}</span>
          <div class="plan-name">{{ plan.name }}</div>
          <p v-if="plan.summary" class="plan-summary">{{ plan.summary }}</p>
          <div class="plan-price"><strong>¥{{ yuan(monthlyPrice(plan).price_cents) }}</strong><span>/月起</span></div>
          <div class="plan-specs">
            <div class="plan-spec">{{ plan.cpu_cores }} vCPU</div>
            <div class="plan-spec">{{ memory(plan) }} 内存</div>
            <div class="plan-spec">{{ plan.system_disk_gb }} GB 系统盘</div>
            <div class="plan-spec">{{ plan.bandwidth_mbps }} Mbps 带宽</div>
            <div class="plan-spec">{{ plan.public_ip_count }} 个公网 IP</div>
          </div>
          <div class="price-cycles">
            <div v-for="price in plan.prices" :key="price.billing_cycle" class="price-cycle">
              <span>{{ cycleLabels[price.billing_cycle] }}</span><strong>¥{{ yuan(price.price_cents) }}</strong>
            </div>
          </div>
          <RouterLink to="/login" class="btn btn-primary btn-sm" style="width:100%">{{ ctaText(plan) }}</RouterLink>
        </div>
      </div>

      <div v-if="visiblePlans.length > 0" class="pricing-table-wrap">
        <table class="pricing-table">
          <thead>
            <tr>
              <th>规格项</th>
              <th v-for="plan in visiblePlans" :key="plan.plan_no" :class="{ 'featured-col': plan.is_featured }">{{ plan.name }}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>月价</td>
              <td v-for="plan in visiblePlans" :key="plan.plan_no" class="price-cell" :class="{ 'featured-col': plan.is_featured }">¥{{ yuan(monthlyPrice(plan).price_cents) }}</td>
            </tr>
            <tr>
              <td>简介</td>
              <td v-for="plan in visiblePlans" :key="plan.plan_no" :class="{ 'featured-col': plan.is_featured }">{{ plan.summary || '固定规格云服务器套餐' }}</td>
            </tr>
            <tr>
              <td>CPU</td>
              <td v-for="plan in visiblePlans" :key="plan.plan_no" :class="{ 'featured-col': plan.is_featured }">{{ plan.cpu_cores }} vCPU</td>
            </tr>
            <tr>
              <td>内存</td>
              <td v-for="plan in visiblePlans" :key="plan.plan_no" :class="{ 'featured-col': plan.is_featured }">{{ memory(plan) }}</td>
            </tr>
            <tr>
              <td>系统盘</td>
              <td v-for="plan in visiblePlans" :key="plan.plan_no" :class="{ 'featured-col': plan.is_featured }">{{ plan.system_disk_gb }} GB</td>
            </tr>
            <tr>
              <td>销售地域</td>
              <td v-for="plan in visiblePlans" :key="plan.plan_no" :class="{ 'featured-col': plan.is_featured }">{{ plan.regions.map((region) => region.name).join(' / ') }}</td>
            </tr>
            <tr>
              <td>系统模板</td>
              <td v-for="plan in visiblePlans" :key="plan.plan_no" :class="{ 'featured-col': plan.is_featured }">{{ plan.os_templates.map((template) => template.name).join(' / ') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

<style scoped>
.empty-state { padding: 40px 0; text-align: center; color: var(--c-text-2); }
.price-cycles { display: grid; gap: 8px; margin: 16px 0; }
.price-cycle { display: flex; justify-content: space-between; font-size: .9rem; color: var(--c-text-2); }
.price-cycle strong { color: var(--c-text); }
.plan-summary { min-height: 44px; margin: 10px 0 0; color: var(--c-text-2); font-size: .9rem; line-height: 1.55; }
</style>
