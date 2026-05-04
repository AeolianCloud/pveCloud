<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getServerCatalog, type ServerCatalogProduct } from '../../api/product-catalog'

const loading = ref(false)
const products = ref<ServerCatalogProduct[]>([])
const activeProductNo = ref<string | null>(null)

onMounted(async () => {
  loading.value = true
  try {
    const catalog = await getServerCatalog()
    products.value = catalog.products
    if (products.value.length > 0) {
      activeProductNo.value = products.value[0].product_no
    }
  } finally {
    loading.value = false
  }
})

const activeProduct = computed<ServerCatalogProduct | null>(() => {
  if (!activeProductNo.value) return null
  return products.value.find(p => p.product_no === activeProductNo.value) || null
})

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function getMinPrice(product: ServerCatalogProduct) {
  if (!product.plans.length) return 0
  const prices = product.plans.map(p => {
    const monthly = p.prices.find(pr => pr.billing_cycle === 'monthly')
    return monthly ? monthly.price_cents : p.prices[0]?.price_cents || 0
  })
  return Math.min(...prices)
}
</script>

<template>
  <div class="products-page">
    <!-- Hero Section -->
    <section class="page-hero">
      <div class="container hero-container text-center">
        <h1 class="hero-title">为每一个业务场景<br/><span class="text-gradient">提供最佳算力支撑</span></h1>
        <p class="hero-desc">无论您是个人开发者还是大型企业团队，我们提供丰富多样的实例规格，从经济型通用计算到极速内存优化型，总有一款适合您的业务需求。</p>
        
        <div class="hero-features">
          <div class="hf-item"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> 灵活扩展，秒级部署</div>
          <div class="hf-item"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> 99.99% SLA 保障</div>
          <div class="hf-item"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg> 全系配备 NVMe 固态硬盘</div>
        </div>
      </div>
    </section>

    <!-- Interactive Products Explorer -->
    <section class="products-explorer py-20">
      <div class="container">
        
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>正在为您加载产品矩阵数据...</p>
        </div>
        
        <div v-else-if="products.length === 0" class="empty-state">
          暂无可用的产品目录，请稍后查看。
        </div>

        <div v-else class="catalog-grid">
          <!-- Sidebar Categories -->
          <aside class="catalog-sidebar">
            <h3 class="sidebar-title">产品线分类</h3>
            <div class="category-list">
              <button 
                v-for="prod in products" 
                :key="prod.product_no"
                class="category-btn"
                :class="{ active: activeProductNo === prod.product_no }"
                @click="activeProductNo = prod.product_no"
              >
                <span class="cat-name">{{ prod.name }}</span>
                <span class="cat-count">{{ prod.plans.length }} 款系列</span>
              </button>
            </div>
            
            <div class="help-box glass-panel">
              <h4>选型建议</h4>
              <p>不知道该如何选择？</p>
              <RouterLink to="/contact" class="btn btn-outline btn-sm btn-block" style="margin-top:12px;">联系架构师</RouterLink>
            </div>
          </aside>

          <!-- Product Lines -->
          <main class="catalog-main" v-if="activeProduct">
            <div class="category-header">
              <h2>{{ activeProduct.name }}</h2>
              <p>{{ activeProduct.summary || '满足您多维度业务需求的弹性计算实例。' }}</p>
            </div>

            <div class="products-list">
              <div class="product-card card card-hover">
                <div class="product-header">
                  <div class="product-icon">
                    <svg v-if="activeProduct.name.includes('通用')" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect><rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect><line x1="6" y1="6" x2="6.01" y2="6"></line><line x1="6" y1="18" x2="6.01" y2="18"></line></svg>
                    <svg v-else-if="activeProduct.name.includes('内存')" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 4H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2z"></path><line x1="12" y1="16" x2="12" y2="20"></line><line x1="8" y1="16" x2="8" y2="20"></line><line x1="16" y1="16" x2="16" y2="20"></line><line x1="8" y1="4" x2="8" y2="8"></line><line x1="16" y1="4" x2="16" y2="8"></line><line x1="12" y1="4" x2="12" y2="8"></line></svg>
                    <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
                  </div>
                  <div class="product-meta">
                    <h3>{{ activeProduct.name }}</h3>
                    <p>{{ activeProduct.description }}</p>
                  </div>
                  <div class="product-price-min">
                    <span class="label">起售价</span>
                    <div class="price"><strong>¥{{ yuan(getMinPrice(activeProduct)) }}</strong> / 月</div>
                  </div>
                </div>

                <div class="plans-grid">
                  <div v-for="plan in activeProduct.plans.slice(0, 4)" :key="plan.plan_no" class="plan-mini-card">
                    <div class="plan-mini-name">{{ plan.name }}</div>
                    <div class="plan-mini-specs">
                      <span>{{ plan.cpu_cores }}vCPU</span>
                      <span class="dot"></span>
                      <span>{{ Math.round(plan.memory_mb / 1024) }}GB</span>
                      <span class="dot"></span>
                      <span>{{ plan.system_disk_gb }}GB</span>
                    </div>
                  </div>
                  <div v-if="activeProduct.plans.length > 4" class="plan-mini-more">
                    +{{ activeProduct.plans.length - 4 }} 更多规格
                  </div>
                </div>

                <div class="product-actions">
                  <RouterLink to="/pricing" class="btn btn-outline">查看详细定价</RouterLink>
                  <RouterLink to="/login" class="btn btn-primary">立即配置并购买</RouterLink>
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </section>

    <!-- Technical Infrastructure -->
    <section class="infrastructure py-32">
      <div class="container text-center">
        <h2 style="font-size: 2.5rem; font-weight: 800; margin-bottom: 24px;">硬核底层基础设施保障</h2>
        <p style="color:var(--c-text-2); font-size: 1.1rem; max-width: 700px; margin: 0 auto 64px;">采用目前业界顶级的数据中心架构与硬件，每一个比特的处理都追求极速与稳定。</p>
        
        <div class="infra-grid">
          <div class="infra-item glass-panel">
            <h4>计算架构</h4>
            <p>全系搭载全新第三代 AMD EPYC 或 Intel Xeon Scalable 铂金级可扩展处理器，提供强悍算力。</p>
          </div>
          <div class="infra-item glass-panel">
            <h4>存储系统</h4>
            <p>基于分布式 Ceph 架构的纯 NVMe 闪存池，支持极速快照、无损容灾及高达百万级 IOPS 并发吞吐。</p>
          </div>
          <div class="infra-item glass-panel">
            <h4>网络接入</h4>
            <p>多线 BGP 智能路由切换，自带内网万兆隔离环境，外网接入 T 级 DDoS 流量清洗与硬防。</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.products-page {
  background: var(--c-bg);
}

/* Hero */
.page-hero {
  position: relative;
  padding: 100px 0 60px;
  background: 
    radial-gradient(ellipse at 50% -20%, rgba(59, 130, 246, 0.2) 0%, transparent 60%);
  border-bottom: 1px solid var(--c-border-light);
}
.hero-title {
  font-size: clamp(2.5rem, 4vw, 4rem);
  font-weight: 800;
  line-height: 1.2;
  margin-bottom: 24px;
}
.hero-desc {
  font-size: 1.15rem;
  color: var(--c-text-2);
  max-width: 680px;
  margin: 0 auto 32px;
  line-height: 1.6;
}
.hero-features {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 24px;
  color: var(--c-text);
  font-weight: 500;
}
.hf-item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--c-surface-dim);
  padding: 8px 16px;
  border-radius: 99px;
  border: 1px solid var(--c-border-light);
}
.hf-item svg { width: 18px; height: 18px; color: var(--c-primary); }

/* Explorer */
.loading-state, .empty-state { text-align: center; padding: 100px 20px; color: var(--c-text-2); }
.spinner { margin: 0 auto 16px; }

.catalog-grid {
  display: grid;
  gap: 40px;
  align-items: flex-start;
}
@media (min-width: 1024px) {
  .catalog-grid { grid-template-columns: 280px 1fr; }
}

.catalog-sidebar {
  position: sticky;
  top: 100px;
}
.sidebar-title {
  font-size: 0.9rem;
  text-transform: uppercase;
  color: var(--c-text-3);
  letter-spacing: 0.05em;
  font-weight: 700;
  margin-bottom: 16px;
  padding: 0 12px;
}
.category-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 32px;
}
.category-btn {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-radius: var(--radius-sm);
  color: var(--c-text-2);
  font-size: 1.05rem;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}
.category-btn:hover {
  background: var(--c-surface-dim);
  color: var(--c-text);
}
.category-btn.active {
  background: var(--c-primary-soft);
  color: var(--c-primary);
  border-color: rgba(59, 130, 246, 0.2);
}
.cat-count { font-size: 0.8rem; background: var(--c-surface-dim); padding: 2px 8px; border-radius: 99px; }
.category-btn.active .cat-count { background: var(--c-primary); color: #fff; }

.help-box {
  padding: 24px;
  border-radius: var(--radius);
  text-align: center;
}
.help-box h4 { font-weight: 700; margin-bottom: 8px; }
.help-box p { font-size: 0.9rem; color: var(--c-text-2); }

.category-header { margin-bottom: 32px; border-bottom: 1px solid var(--c-border-light); padding-bottom: 24px; }
.category-header h2 { font-size: 2rem; font-weight: 800; margin-bottom: 8px; }
.category-header p { color: var(--c-text-2); font-size: 1.05rem; }

.products-list { display: grid; gap: 24px; }
.product-card { padding: 32px; display: flex; flex-direction: column; gap: 24px; }
.product-header { display: flex; gap: 20px; align-items: flex-start; flex-wrap: wrap; }
.product-icon {
  width: 56px; height: 56px; border-radius: 16px;
  background: var(--c-primary-soft); color: var(--c-primary);
  display: flex; align-items: center; justify-content: center;
}
.product-icon svg { width: 28px; height: 28px; }
.product-meta { flex: 1; min-width: 200px; }
.product-meta h3 { font-size: 1.4rem; font-weight: 800; margin-bottom: 8px; }
.product-meta p { color: var(--c-text-2); font-size: 0.95rem; line-height: 1.5; }
.product-price-min { text-align: right; }
.product-price-min .label { font-size: 0.85rem; color: var(--c-text-3); text-transform: uppercase; font-weight: 600; display: block; margin-bottom: 4px; }
.product-price-min .price { font-size: 0.95rem; color: var(--c-text-2); }
.product-price-min .price strong { font-size: 1.5rem; color: var(--c-text); font-weight: 800; }

.plans-grid { display: flex; flex-wrap: wrap; gap: 12px; }
.plan-mini-card {
  background: var(--c-surface-dim);
  border: 1px solid var(--c-border-light);
  border-radius: var(--radius-sm);
  padding: 12px 16px;
  flex: 1; min-width: 180px;
}
.plan-mini-name { font-weight: 600; margin-bottom: 4px; font-size: 0.95rem; }
.plan-mini-specs { display: flex; align-items: center; gap: 6px; font-size: 0.85rem; color: var(--c-text-2); }
.dot { width: 4px; height: 4px; border-radius: 50%; background: var(--c-border); }
.plan-mini-more {
  display: flex; align-items: center; justify-content: center;
  padding: 12px 16px; border-radius: var(--radius-sm);
  background: rgba(0,0,0,0.2); border: 1px dashed var(--c-border);
  color: var(--c-text-2); font-size: 0.9rem; font-weight: 600;
  cursor: pointer;
}

.product-actions {
  display: flex; justify-content: flex-end; gap: 16px;
  padding-top: 24px; border-top: 1px solid var(--c-border-light);
}

@media (max-width: 768px) {
  .product-header { flex-direction: column; }
  .product-price-min { text-align: left; }
  .product-actions { flex-direction: column; }
}

/* Infrastructure */
.infra-grid { display: grid; gap: 24px; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); }
.infra-item { padding: 40px; text-align: left; border-radius: var(--radius-xl); }
.infra-item h4 { font-size: 1.25rem; font-weight: 800; margin-bottom: 12px; color: var(--c-text); }
.infra-item p { color: var(--c-text-2); line-height: 1.6; }
</style>
