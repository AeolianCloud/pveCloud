<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'

const loading = ref(false)
const plans = ref<ServerCatalogPlan[]>([])

const featuredPlans = computed(() => {
  const featured = plans.value.filter((plan) => plan.is_featured)
  return (featured.length > 0 ? featured : plans.value).slice(0, 3)
})

const heroPlan = computed(() => featuredPlans.value[0])

function monthlyPrice(plan: ServerCatalogPlan) {
  return plan.prices.find((price) => price.billing_cycle === 'monthly') || plan.prices[0]
}

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function memory(plan: ServerCatalogPlan) {
  return `${Math.round(plan.memory_mb / 1024)} GB`
}

function traffic(plan: ServerCatalogPlan) {
  return plan.traffic_gb == null ? '暂不承诺' : `${plan.traffic_gb} GB/月`
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

const features = [
  { icon: 'cpu', color: 'blue', title: '固定套餐', desc: '先开放固定规格服务器套餐，价格和配置由后台统一维护。' },
  { icon: 'lock', color: 'green', title: '销售地域', desc: '地域是销售展示约束，不提前绑定 PVE 节点或资源池。' },
  { icon: 'disk', color: 'purple', title: '系统模板', desc: '服务器系统模板独立管理，避免与图片附件概念混淆。' },
]
</script>

<template>
  <section class="page home-page">
    <div class="hero">
      <div class="hero-bg-blob blob-1"></div>
      <div class="hero-bg-blob blob-2"></div>
      <div style="position: relative; z-index: 2;">
        <div class="hero-badge">Server Catalog</div>
        <h1>按固定套餐展示的<em>云服务器</em></h1>
        <p class="hero-desc">
          当前阶段开放服务器产品目录，支持查看套餐、价格、销售地域和系统模板。
          订单、支付和实例开通能力仍在后续阶段开放。
        </p>
        <div class="hero-actions">
          <RouterLink to="/pricing" class="btn btn-primary">查看套餐价格</RouterLink>
          <RouterLink to="/products" class="btn btn-outline">了解产品能力</RouterLink>
        </div>
      </div>

      <div v-if="heroPlan" class="hero-pricing" style="position: relative; z-index: 2;">
        <div class="hero-plan-name">推荐套餐</div>
        <div class="hero-plan-title">{{ heroPlan.name }}</div>
        <p v-if="heroPlan.summary" class="hero-plan-summary">{{ heroPlan.summary }}</p>
        <div class="hero-plan-price"><strong>{{ yuan(monthlyPrice(heroPlan).price_cents) }}</strong><span>元/月起</span></div>
        <div class="hero-specs">
          <div class="hero-spec-row"><span>CPU</span><strong>{{ heroPlan.cpu_cores }} vCPU</strong></div>
          <div class="hero-spec-row"><span>内存</span><strong>{{ memory(heroPlan) }}</strong></div>
          <div class="hero-spec-row"><span>系统盘</span><strong>{{ heroPlan.system_disk_gb }} GB</strong></div>
          <div class="hero-spec-row"><span>流量</span><strong>{{ traffic(heroPlan) }}</strong></div>
          <div class="hero-spec-row"><span>公网 IP</span><strong>{{ heroPlan.public_ip_count }} 个</strong></div>
        </div>
        <RouterLink to="/pricing" class="btn btn-primary" style="width:100%">查看详情</RouterLink>
      </div>
    </div>

    <div class="section-pad">
      <div class="sec-header center" style="margin-bottom: clamp(32px, 5vw, 56px);">
        <p class="label">目录能力</p>
        <h2>先把售卖服务器的基础事实做准</h2>
        <p>产品目录只负责展示和可售约束，不承载交易、支付或实例交付。</p>
      </div>
      <div class="features-grid">
        <div v-for="f in features" :key="f.title" class="feature-card">
          <div class="feature-icon" :class="[f.color, f.icon]"></div>
          <h3>{{ f.title }}</h3>
          <p>{{ f.desc }}</p>
        </div>
      </div>
    </div>

    <div class="section-pad" style="background: var(--c-surface-dim); border-top: 1px solid var(--c-border-light); border-bottom: 1px solid var(--c-border-light);">
      <div class="sec-header center" style="margin-bottom: clamp(32px, 5vw, 56px);">
        <p class="label">推荐套餐</p>
        <h2>后台维护，前台展示</h2>
        <p v-if="loading">正在读取产品目录...</p>
        <p v-else-if="featuredPlans.length === 0">暂无公开套餐，请先在后台配置套餐、价格、地域和系统模板。</p>
      </div>
      <div v-if="featuredPlans.length > 0" class="pricing-preview-grid">
        <div v-for="p in featuredPlans" :key="p.plan_no" class="pricing-preview-card" :class="{ featured: p.is_featured }">
          <span class="plan-tag" :class="p.is_featured ? 'primary' : 'green'">{{ p.status === 'sold_out' ? '售罄' : p.is_featured ? '推荐' : '可选' }}</span>
          <div class="plan-name">{{ p.name }}</div>
          <p v-if="p.summary" class="plan-summary">{{ p.summary }}</p>
          <div class="plan-price"><strong>{{ yuan(monthlyPrice(p).price_cents) }}</strong><span>元/月起</span></div>
          <div class="plan-specs">
            <div class="plan-spec">{{ p.cpu_cores }} vCPU</div>
            <div class="plan-spec">{{ memory(p) }} 内存</div>
            <div class="plan-spec">{{ p.system_disk_gb }} GB 系统盘</div>
            <div class="plan-spec">{{ traffic(p) }} 流量</div>
            <div class="plan-spec">{{ p.regions.map((region) => region.name).join(' / ') }}</div>
          </div>
          <div class="plan-cta">
            <RouterLink to="/pricing" class="btn btn-outline btn-sm" style="width:100%">查看方案详情</RouterLink>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero-bg-blob { position: absolute; border-radius: 50%; filter: blur(80px); z-index: 1; animation: float 10s infinite alternate ease-in-out; opacity: 0.6; }
.blob-1 { width: 400px; height: 400px; background: radial-gradient(circle, rgba(99, 102, 241, 0.4) 0%, transparent 70%); top: -100px; left: -100px; }
.blob-2 { width: 500px; height: 500px; background: radial-gradient(circle, rgba(59, 130, 246, 0.3) 0%, transparent 70%); bottom: -200px; right: -100px; animation-delay: -5s; }
.hero-plan-summary,
.plan-summary { color: var(--c-text-2); font-size: .9rem; line-height: 1.6; margin: 8px 0 0; }
@keyframes float { 0% { transform: translate(0, 0) scale(1); } 50% { transform: translate(20px, -20px) scale(1.05); } 100% { transform: translate(-20px, 20px) scale(0.95); } }
</style>
