<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'
import { getRealNameStatus, type RealNameStatusResponse } from '../../api/real-name'
import { useWebAppStore } from '../../store/modules/app'
import { useWebAuthStore } from '../../store/modules/auth'

const loading = ref(false)
const plans = ref<ServerCatalogPlan[]>([])
const realNameLoading = ref(false)
const realNameStatus = ref<RealNameStatusResponse | null>(null)
const appStore = useWebAppStore()
const authStore = useWebAuthStore()
const { isLoggedIn } = storeToRefs(authStore)

const billingCycle = ref<'monthly' | 'yearly'>('monthly')

const cycleLabels: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  semi_yearly: '半年付',
  yearly: '年付',
}

onMounted(async () => {
  loading.value = true
  try {
    const [catalog] = await Promise.all([
      getServerCatalog(),
      appStore.loadSiteConfig(),
      loadRealNameStatus(),
    ])
    plans.value = catalog.products.flatMap(p => p.plans)
  } finally {
    loading.value = false
  }
})

watch(isLoggedIn, (loggedIn) => {
  if (loggedIn) {
    void loadRealNameStatus()
    return
  }
  realNameStatus.value = null
})

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function getPrice(plan: ServerCatalogPlan, cycle: 'monthly' | 'yearly') {
  const p = plan.prices.find(pr => pr.billing_cycle === cycle)
  if (p) return p.price_cents
  return plan.prices[0]?.price_cents || 0
}

function getYearlyDiscount(plan: ServerCatalogPlan) {
  const monthly = getPrice(plan, 'monthly') * 12
  const yearly = getPrice(plan, 'yearly')
  if (monthly && yearly && yearly < monthly) {
    const savings = monthly - yearly
    return Math.round((savings / monthly) * 100)
  }
  return 0
}

const displayPlans = computed(() => {
  const sorted = [...plans.value].sort((a, b) => getPrice(a, 'monthly') - getPrice(b, 'monthly'))
  // Just show 4 prominent plans for the pricing comparison
  return sorted.slice(0, 4)
})

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
  if (plan.status === 'sold_out') {
    return { label: '暂时售罄', to: '', disabled: true }
  }
  if (!isLoggedIn.value) {
    return { label: '登录查看购买入口', to: '/login', disabled: false }
  }

  const config = realNameStatus.value?.config || appStore.realNameConfig
  const status = realNameStatus.value?.status
  if (config.required_for_order && config.enabled) {
    if (realNameLoading.value) {
      return { label: '同步实名状态...', to: '', disabled: true }
    }
    if (status !== 'approved') {
      if (status === 'pending') return { label: '查看实名审核进度', to: '/user/real-name', disabled: false }
      if (status === 'rejected') return { label: '重新提交实名', to: '/user/real-name', disabled: false }
      return { label: '查看实名要求', to: '/user/real-name', disabled: false }
    }
    return { label: '已完成实名，购买功能即将开放', to: '', disabled: true }
  }

  return { label: '购买功能即将开放', to: '', disabled: true }
}
</script>

<template>
  <div class="pricing-page">
    <div class="pricing-hero">
      <div class="container text-center">
        <h1 class="hero-title">透明、简单的<span class="text-gradient">产品定价</span></h1>
        <p class="hero-desc">价格来自公开服务器产品目录。当前阶段只展示价格，不开放下单、支付或实例开通。</p>
        
        <!-- Billing Cycle Toggle -->
        <div class="cycle-toggle-wrapper">
          <div class="cycle-toggle">
            <button 
              class="cycle-btn" 
              :class="{ active: billingCycle === 'monthly' }"
              @click="billingCycle = 'monthly'"
            >按月付费</button>
            <button 
              class="cycle-btn" 
              :class="{ active: billingCycle === 'yearly' }"
              @click="billingCycle = 'yearly'"
            >按年付费 <span class="discount-badge">享特惠</span></button>
          </div>
        </div>
      </div>
    </div>

    <section class="pricing-content container">
      <div v-if="loading" class="loading-state text-center py-20">
        <div class="spinner"></div>
        <p style="margin-top:20px; color:var(--c-text-2)">正在同步最新价格清单...</p>
      </div>

      <div v-else-if="displayPlans.length > 0">
        <!-- Modern Pricing Cards -->
        <div class="pricing-cards">
          <div v-for="plan in displayPlans" :key="plan.plan_no" class="pricing-card glass-panel" :class="{ 'is-featured': plan.is_featured }">
            <div v-if="plan.is_featured" class="featured-ribbon">最受欢迎</div>
            
            <div class="pc-header">
              <h3>{{ plan.name }}</h3>
              <p>{{ plan.summary || '通用计算规格，适合大多数场景' }}</p>
            </div>

            <div class="pc-price">
              <span class="currency">¥</span>
              <span class="amount">{{ yuan(getPrice(plan, billingCycle)) }}</span>
              <span class="period">/ {{ cycleLabels[billingCycle] }}</span>
              <div v-if="billingCycle === 'yearly' && getYearlyDiscount(plan) > 0" class="savings">
                省 {{ getYearlyDiscount(plan) }}%
              </div>
            </div>

            <div class="pc-action">
              <RouterLink
                v-if="getAction(plan).to"
                :to="getAction(plan).to"
                class="btn btn-block"
                :class="plan.is_featured ? 'btn-primary' : 'btn-outline'"
              >{{ getAction(plan).label }}</RouterLink>
              <button
                v-else
                class="btn btn-block"
                :class="plan.is_featured ? 'btn-primary' : 'btn-outline'"
                type="button"
                :disabled="getAction(plan).disabled"
              >{{ getAction(plan).label }}</button>
            </div>

            <ul class="pc-specs">
              <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> <strong>{{ plan.cpu_cores }}</strong> 核心处理器</li>
              <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> <strong>{{ Math.round(plan.memory_mb / 1024) }} GB</strong> 内存容量</li>
              <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> <strong>{{ plan.system_disk_gb }} GB</strong> 系统盘</li>
              <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> <strong>{{ plan.bandwidth_mbps }} Mbps</strong> 峰值带宽</li>
              <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> <strong>{{ plan.public_ip_count }}</strong> 个独立公网 IP</li>
            </ul>
          </div>
        </div>

        <!-- Detailed Feature Comparison -->
        <div class="comparison-section py-32">
          <h2 class="text-center" style="font-size: 2rem; font-weight: 800; margin-bottom: 48px;">公开目录字段对比</h2>
          <div class="table-wrap glass-panel">
            <table class="comparison-table">
              <thead>
                <tr>
                  <th class="feature-col">规格特性</th>
                  <th v-for="plan in displayPlans" :key="plan.plan_no" :class="{ 'th-featured': plan.is_featured }">
                    <div class="th-name">{{ plan.name }}</div>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr><td colspan="5" class="tr-group">核心规格</td></tr>
                <tr>
                  <td>处理器规格</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.cpu_cores }} vCPU</td>
                </tr>
                <tr>
                  <td>内存</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ Math.round(plan.memory_mb / 1024) }} GB</td>
                </tr>
                <tr>
                  <td>架构</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.architecture }}</td>
                </tr>

                <tr><td colspan="5" class="tr-group">存储与网络</td></tr>
                <tr>
                  <td>系统盘</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.system_disk_gb }} GB</td>
                </tr>
                <tr>
                  <td>数据盘</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.data_disk_gb }} GB</td>
                </tr>
                <tr>
                  <td>带宽</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.bandwidth_mbps }} Mbps</td>
                </tr>
                <tr>
                  <td>月流量</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.traffic_gb == null ? '未返回' : `${plan.traffic_gb} GB` }}</td>
                </tr>
                <tr>
                  <td>公网 IP</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.public_ip_count }} 个</td>
                </tr>

                <tr><td colspan="5" class="tr-group">可售约束</td></tr>
                <tr>
                  <td>销售地域</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.regions.map((region) => region.name).join(' / ') }}</td>
                </tr>
                <tr>
                  <td>系统模板</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.os_templates.map((template) => template.name).join(' / ') }}</td>
                </tr>
                <tr>
                  <td>状态</td>
                  <td v-for="plan in displayPlans" :key="plan.plan_no">{{ plan.status === 'sold_out' ? '暂时售罄' : '可展示' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <div v-else class="loading-state text-center py-20">
        暂无公开价格，请先在后台配置可展示套餐、价格、地域和系统模板。
      </div>
    </section>
  </div>
</template>

<style scoped>
.pricing-page { padding-bottom: 80px; }

.pricing-hero {
  padding: 100px 0 160px;
  background: 
    radial-gradient(ellipse at 50% 0%, rgba(139, 92, 246, 0.15) 0%, transparent 60%);
  border-bottom: 1px solid var(--c-border-light);
  position: relative;
}
.hero-title { font-size: clamp(2.5rem, 4vw, 3.5rem); font-weight: 800; margin-bottom: 24px; }
.hero-desc { font-size: 1.15rem; color: var(--c-text-2); max-width: 600px; margin: 0 auto 40px; }

/* Cycle Toggle */
.cycle-toggle-wrapper { display: flex; justify-content: center; }
.cycle-toggle {
  display: inline-flex; background: rgba(0,0,0,0.3); border: 1px solid var(--c-border);
  padding: 6px; border-radius: 99px;
}
.cycle-btn {
  padding: 10px 24px; border-radius: 99px; font-size: 0.95rem; font-weight: 600;
  color: var(--c-text-2); cursor: pointer; transition: all 0.3s; position: relative;
}
.cycle-btn.active { background: var(--c-surface); color: var(--c-text); box-shadow: var(--shadow-sm); }
.discount-badge {
  position: absolute; top: -12px; right: -10px;
  background: var(--c-success); color: white; padding: 2px 8px; border-radius: 99px;
  font-size: 0.7rem; font-weight: 800; white-space: nowrap;
}

/* Pricing Cards */
.pricing-cards {
  display: grid; gap: 24px; margin-top: -80px; position: relative; z-index: 10;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}
.pricing-card {
  padding: 40px 32px; border-radius: var(--radius-xl);
  display: flex; flex-direction: column; position: relative; overflow: hidden;
  transition: transform 0.3s, box-shadow 0.3s;
}
.pricing-card:hover { transform: translateY(-8px); box-shadow: var(--shadow-lg), var(--shadow-glow); }
.pricing-card.is-featured { border-color: var(--c-primary); background: rgba(19, 21, 31, 0.95); }
.featured-ribbon {
  position: absolute; top: 0; left: 0; right: 0; background: var(--c-primary); color: white;
  text-align: center; font-size: 0.8rem; font-weight: 700; padding: 6px 0; letter-spacing: 0.05em;
}

.pc-header { margin-bottom: 32px; margin-top: 16px; }
.pc-header h3 { font-size: 1.5rem; font-weight: 800; margin-bottom: 8px; }
.pc-header p { font-size: 0.9rem; color: var(--c-text-2); min-height: 40px; }

.pc-price { display: flex; align-items: baseline; gap: 4px; margin-bottom: 32px; position: relative; }
.pc-price .currency { font-size: 1.5rem; font-weight: 700; }
.pc-price .amount { font-size: 3.5rem; font-weight: 800; line-height: 1; letter-spacing: -0.03em; }
.pc-price .period { font-size: 1rem; color: var(--c-text-3); font-weight: 500; }
.pc-price .savings {
  position: absolute; bottom: -24px; left: 0; color: var(--c-success); font-size: 0.85rem; font-weight: 600;
  background: var(--c-success-soft); padding: 2px 8px; border-radius: 4px;
}

.pc-action { margin-bottom: 32px; }

.pc-specs { list-style: none; display: grid; gap: 16px; flex: 1; }
.pc-specs li { display: flex; align-items: center; gap: 12px; font-size: 0.95rem; color: var(--c-text-2); }
.pc-specs li svg { width: 18px; height: 18px; color: var(--c-primary); }
.pc-specs li strong { color: var(--c-text); }

/* Table */
.table-wrap { border-radius: var(--radius-xl); overflow: hidden; box-shadow: var(--shadow-lg); }
.comparison-table { width: 100%; border-collapse: collapse; text-align: center; }
.comparison-table th, .comparison-table td { padding: 20px; border-bottom: 1px solid var(--c-border-light); }
.comparison-table th { background: rgba(0,0,0,0.4); }
.comparison-table .feature-col { text-align: left; font-size: 1.1rem; font-weight: 700; width: 25%; }
.comparison-table .th-name { font-size: 1.1rem; font-weight: 700; }
.comparison-table .th-featured { background: var(--c-primary-soft); border-top: 3px solid var(--c-primary); }
.tr-group { text-align: left !important; background: var(--c-surface-dim); font-weight: 700; color: var(--c-text); padding: 12px 20px !important; }
.comparison-table td { color: var(--c-text-2); font-size: 0.95rem; }
.comparison-table tr:hover td { background: var(--c-surface-dim); color: var(--c-text); }
.check svg { width: 20px; height: 20px; color: var(--c-success); margin: 0 auto; }
.cross span { color: var(--c-text-3); }

@media (max-width: 992px) {
  .table-wrap { overflow-x: auto; }
  .comparison-table { min-width: 800px; }
}
</style>
