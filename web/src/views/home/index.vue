<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'

const loading = ref(false)
const plans = ref<ServerCatalogPlan[]>([])

onMounted(async () => {
  loading.value = true
  try {
    const catalog = await getServerCatalog()
    const allPlans = catalog.products.flatMap((product) => product.plans)
    plans.value = allPlans.filter((plan) => plan.is_featured).slice(0, 3)
    if (plans.value.length === 0) plans.value = allPlans.slice(0, 3)
  } finally {
    loading.value = false
  }
})

const catalogStats = computed(() => {
  const regions = new Set<string>()
  const templates = new Set<string>()
  plans.value.forEach((plan) => {
    plan.regions.forEach((region) => regions.add(region.region_no))
    plan.os_templates.forEach((template) => templates.add(template.template_no))
  })
  return [
    { label: '推荐套餐', value: `${plans.value.length}` },
    { label: '展示地域', value: `${regions.size}` },
    { label: '系统模板', value: `${templates.size}` },
  ]
})

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function monthlyPrice(plan: ServerCatalogPlan) {
  return plan.prices.find((price) => price.billing_cycle === 'monthly') || plan.prices[0]
}
</script>

<template>
  <section class="home-page">
    <div class="container hero-grid">
      <div class="hero-copy">
        <p class="section-label">PVECloud Web</p>
        <h1 class="page-title">清晰展示服务器产品，账号能力先行开放。</h1>
        <p class="page-copy">
          当前用户端聚焦产品目录、价格展示、注册登录、密码找回、账号资料和个人实名。订单、支付和实例交付会在后续阶段开放。
        </p>
        <div class="hero-actions">
          <RouterLink class="btn btn-primary btn-lg" to="/products">查看产品目录</RouterLink>
          <RouterLink class="btn btn-outline btn-lg" to="/pricing">浏览价格</RouterLink>
        </div>
      </div>

      <aside class="surface hero-panel">
        <div class="panel-head">
          <span>PUBLIC CATALOG</span>
          <strong>服务器公开目录</strong>
        </div>
        <div class="stat-grid">
          <div v-for="item in catalogStats" :key="item.label" class="stat-card">
            <strong>{{ item.value }}</strong>
            <span>{{ item.label }}</span>
          </div>
        </div>
        <div class="scope-list">
          <p><b>已开放</b><span>账号自助 / 实名入口 / 产品展示</span></p>
          <p><b>未开放</b><span>下单 / 支付 / 实例 / 工单</span></p>
        </div>
      </aside>
    </div>

    <div class="container">
      <section class="section-block">
        <div class="section-heading">
          <p class="section-label">Featured Plans</p>
          <h2>推荐服务器套餐</h2>
          <RouterLink to="/pricing" class="btn btn-outline btn-sm">完整价格表</RouterLink>
        </div>

        <div v-if="loading" class="loading-card surface">
          <div class="spinner"></div>
          <span>正在读取公开目录...</span>
        </div>
        <div v-else class="plans-grid">
          <article v-for="plan in plans" :key="plan.plan_no" class="plan-card card">
            <div>
              <div class="plan-top">
                <h3>{{ plan.name }}</h3>
                <span v-if="plan.is_featured" class="tag tag-primary">推荐</span>
              </div>
              <p>{{ plan.summary || '固定规格服务器套餐' }}</p>
            </div>
            <div class="price-row">
              <span>¥</span>
              <strong>{{ yuan(monthlyPrice(plan).price_cents) }}</strong>
              <em>/ 月</em>
            </div>
            <dl class="spec-grid">
              <div><dt>CPU</dt><dd>{{ plan.cpu_cores }} vCPU</dd></div>
              <div><dt>内存</dt><dd>{{ Math.round(plan.memory_mb / 1024) }} GB</dd></div>
              <div><dt>系统盘</dt><dd>{{ plan.system_disk_gb }} GB</dd></div>
              <div><dt>带宽</dt><dd>{{ plan.bandwidth_mbps }} Mbps</dd></div>
            </dl>
            <RouterLink class="btn btn-outline btn-block" to="/login">登录后查看购买入口</RouterLink>
          </article>
        </div>
      </section>

      <section class="workflow-strip surface">
        <div>
          <p class="section-label">Current Flow</p>
          <h2>从浏览到账号准备</h2>
        </div>
        <div class="flow-steps">
          <span>浏览产品</span>
          <i></i>
          <span>注册登录</span>
          <i></i>
          <span>维护资料</span>
          <i></i>
          <span>完成实名</span>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.home-page {
  padding: 42px 0 72px;
}

.hero-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) minmax(320px, 0.95fr);
  gap: 28px;
  align-items: stretch;
}

.hero-copy {
  display: grid;
  align-content: center;
  gap: 22px;
  min-height: 440px;
}

.hero-copy .page-copy {
  max-width: 660px;
  font-size: 1.06rem;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero-panel {
  display: grid;
  align-content: space-between;
  gap: 22px;
  padding: 28px;
}

.panel-head {
  display: grid;
  gap: 8px;
}

.panel-head span {
  color: var(--c-text-3);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.14em;
}

.panel-head strong {
  font-size: clamp(1.7rem, 3vw, 2.4rem);
  line-height: 1.05;
  letter-spacing: -0.05em;
}

.stat-grid,
.plans-grid {
  display: grid;
  gap: 16px;
}

.stat-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.stat-card {
  display: grid;
  gap: 4px;
  padding: 18px;
  border: 1px solid var(--c-border);
  border-radius: 14px;
  background: var(--c-surface-dim);
}

.stat-card strong {
  font-size: 1.7rem;
  letter-spacing: -0.04em;
}

.stat-card span,
.scope-list span {
  color: var(--c-text-2);
}

.scope-list {
  display: grid;
  gap: 10px;
}

.scope-list p {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.section-block {
  display: grid;
  gap: 20px;
  padding-top: 56px;
}

.section-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 20px;
}

.section-heading h2,
.workflow-strip h2 {
  font-size: clamp(1.7rem, 3vw, 2.4rem);
  letter-spacing: -0.05em;
}

.plans-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.plan-card {
  display: grid;
  gap: 20px;
  padding: 22px;
}

.plan-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.plan-card h3 {
  font-size: 1.25rem;
  letter-spacing: -0.04em;
}

.plan-card p {
  margin-top: 8px;
  color: var(--c-text-2);
  line-height: 1.65;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.price-row strong {
  font-size: 2.6rem;
  line-height: 1;
  letter-spacing: -0.06em;
}

.price-row em {
  color: var(--c-text-3);
  font-style: normal;
}

.spec-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.spec-grid div {
  padding: 12px;
  border-radius: 12px;
  background: var(--c-surface-dim);
}

.spec-grid dt {
  color: var(--c-text-3);
  font-size: 0.78rem;
}

.spec-grid dd {
  margin: 2px 0 0;
  font-weight: 800;
}

.loading-card {
  min-height: 180px;
  display: grid;
  place-items: center;
  gap: 12px;
  color: var(--c-text-2);
}

.workflow-strip {
  margin-top: 56px;
  padding: 26px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.flow-steps {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  color: var(--c-text-2);
  font-weight: 800;
}

.flow-steps span {
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--c-surface-dim);
}

.flow-steps i {
  width: 20px;
  height: 1px;
  background: var(--c-border-strong);
}

@media (max-width: 980px) {
  .hero-grid,
  .plans-grid {
    grid-template-columns: 1fr;
  }

  .hero-copy {
    min-height: auto;
    padding-top: 18px;
  }

  .workflow-strip {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 620px) {
  .stat-grid,
  .spec-grid {
    grid-template-columns: 1fr;
  }

  .section-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
