<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'
import { getRealNameStatus, type RealNameStatusResponse } from '../../api/real-name'
import { useWebAppStore } from '../../store/modules/app'
import { useWebAuthStore } from '../../store/modules/auth'

type BillingCycle = 'monthly' | 'quarterly' | 'semi_yearly' | 'yearly'

const loading = ref(false)
const plans = ref<ServerCatalogPlan[]>([])
const cycle = ref<BillingCycle>('monthly')
const realNameLoading = ref(false)
const realNameStatus = ref<RealNameStatusResponse | null>(null)
const appStore = useWebAppStore()
const authStore = useWebAuthStore()
const { isLoggedIn } = storeToRefs(authStore)

const cycleOptions: Array<{ label: string; value: BillingCycle }> = [
  { label: '月付', value: 'monthly' },
  { label: '季付', value: 'quarterly' },
  { label: '半年付', value: 'semi_yearly' },
  { label: '年付', value: 'yearly' },
]

onMounted(async () => {
  loading.value = true
  try {
    const [catalog] = await Promise.all([
      getServerCatalog(),
      appStore.loadSiteConfig(),
      loadRealNameStatus(),
    ])
    plans.value = catalog.products.flatMap((product) => product.plans)
  } finally {
    loading.value = false
  }
})

watch(isLoggedIn, (value) => {
  if (value) void loadRealNameStatus()
  else realNameStatus.value = null
})

const sortedPlans = computed(() => [...plans.value].sort((a, b) => getPrice(a, 'monthly') - getPrice(b, 'monthly')))

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function getPrice(plan: ServerCatalogPlan, value: BillingCycle) {
  const price = plan.prices.find((item) => item.billing_cycle === value)
  return price?.price_cents ?? plan.prices[0]?.price_cents ?? 0
}

async function loadRealNameStatus() {
  if (!isLoggedIn.value) {
    realNameStatus.value = null
    return
  }
  realNameLoading.value = true
  try {
    realNameStatus.value = await getRealNameStatus()
  } catch {
    realNameStatus.value = null
  } finally {
    realNameLoading.value = false
  }
}

function getAction(plan: ServerCatalogPlan) {
  if (plan.status === 'sold_out') return { label: '暂时售罄', to: '', disabled: true }
  if (!isLoggedIn.value) return { label: '登录查看购买入口', to: '/login', disabled: false }

  const config = realNameStatus.value?.config || appStore.realNameConfig
  const status = realNameStatus.value?.status
  if (config.required_for_order && config.enabled) {
    if (realNameLoading.value) return { label: '同步实名状态...', to: '', disabled: true }
    if (status !== 'approved') return { label: '查看实名要求', to: '/user/real-name', disabled: false }
    return { label: '购买功能即将开放', to: '', disabled: true }
  }
  return { label: '购买功能即将开放', to: '', disabled: true }
}
</script>

<template>
  <section class="pricing-page page-shell">
    <div class="pricing-hero surface">
      <div>
        <p class="section-label">Pricing</p>
        <h1 class="page-title">透明价格表</h1>
        <p class="page-copy">价格、地域、系统模板和售卖状态来自公开服务器产品目录。当前仅展示，不创建订单。</p>
      </div>
      <div class="cycle-switch">
        <button v-for="item in cycleOptions" :key="item.value" type="button" :class="{ active: cycle === item.value }" @click="cycle = item.value">
          {{ item.label }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-panel surface">
      <div class="spinner"></div>
      <span>正在同步价格...</span>
    </div>

    <div v-else-if="sortedPlans.length === 0" class="loading-panel surface">
      暂无公开价格，请先在后台配置可展示套餐。
    </div>

    <div v-else class="price-grid">
      <article v-for="plan in sortedPlans" :key="plan.plan_no" class="price-card card" :class="{ featured: plan.is_featured }">
        <div class="card-head">
          <div>
            <h2>{{ plan.name }}</h2>
            <p>{{ plan.summary || '固定规格服务器套餐' }}</p>
          </div>
          <span v-if="plan.is_featured" class="tag tag-primary">推荐</span>
        </div>

        <div class="price-line">
          <span>¥</span>
          <strong>{{ yuan(getPrice(plan, cycle)) }}</strong>
          <em>/ {{ cycleOptions.find((item) => item.value === cycle)?.label }}</em>
        </div>

        <dl class="spec-list">
          <div><dt>CPU</dt><dd>{{ plan.cpu_cores }} vCPU</dd></div>
          <div><dt>内存</dt><dd>{{ Math.round(plan.memory_mb / 1024) }} GB</dd></div>
          <div><dt>系统盘</dt><dd>{{ plan.system_disk_gb }} GB</dd></div>
          <div><dt>数据盘</dt><dd>{{ plan.data_disk_gb }} GB</dd></div>
          <div><dt>带宽</dt><dd>{{ plan.bandwidth_mbps }} Mbps</dd></div>
          <div><dt>公网 IP</dt><dd>{{ plan.public_ip_count }} 个</dd></div>
        </dl>

        <div class="meta-block">
          <p><b>地域</b>{{ plan.regions.map((region) => region.name).join(' / ') || '未返回' }}</p>
          <p><b>系统</b>{{ plan.os_templates.map((template) => template.name).join(' / ') || '未返回' }}</p>
        </div>

        <RouterLink v-if="getAction(plan).to" class="btn btn-primary btn-block" :to="getAction(plan).to">
          {{ getAction(plan).label }}
        </RouterLink>
        <button v-else class="btn btn-outline btn-block" type="button" :disabled="getAction(plan).disabled">
          {{ getAction(plan).label }}
        </button>
      </article>
    </div>
  </section>
</template>

<style scoped>
.pricing-page {
  display: grid;
  gap: 24px;
}

.pricing-hero {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  padding: clamp(24px, 4vw, 38px);
}

.pricing-hero .page-copy {
  max-width: 680px;
  margin-top: 14px;
}

.cycle-switch {
  display: inline-flex;
  gap: 6px;
  padding: 6px;
  border: 1px solid var(--c-border);
  border-radius: 12px;
  background: var(--c-surface-dim);
}

.cycle-switch button {
  min-height: 36px;
  padding: 0 12px;
  border-radius: 8px;
  color: var(--c-text-2);
  cursor: pointer;
  font-weight: 800;
}

.cycle-switch button.active {
  color: #fff;
  background: var(--c-primary);
}

.loading-panel {
  min-height: 220px;
  display: grid;
  place-items: center;
  gap: 12px;
  color: var(--c-text-2);
}

.price-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.price-card {
  display: grid;
  gap: 20px;
  padding: 22px;
}

.price-card.featured {
  border-color: var(--c-primary);
  box-shadow: var(--shadow);
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.card-head h2 {
  font-size: 1.35rem;
  letter-spacing: -0.04em;
}

.card-head p {
  margin-top: 8px;
  color: var(--c-text-2);
  line-height: 1.6;
}

.price-line {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.price-line strong {
  font-size: 3rem;
  line-height: 1;
  letter-spacing: -0.07em;
}

.price-line em {
  color: var(--c-text-3);
  font-style: normal;
}

.spec-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.spec-list div {
  padding: 12px;
  border-radius: 12px;
  background: var(--c-surface-dim);
}

.spec-list dt {
  color: var(--c-text-3);
  font-size: 0.76rem;
}

.spec-list dd {
  margin: 4px 0 0;
  font-weight: 800;
}

.meta-block {
  display: grid;
  gap: 8px;
  color: var(--c-text-2);
  line-height: 1.55;
}

.meta-block p {
  display: grid;
  gap: 2px;
}

.meta-block b {
  color: var(--c-text);
}

@media (max-width: 1080px) {
  .price-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .pricing-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .cycle-switch {
    flex-wrap: wrap;
  }

  .price-grid {
    grid-template-columns: 1fr;
  }
}
</style>
