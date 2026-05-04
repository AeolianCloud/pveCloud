<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'

const loading = ref(false)
const topPlans = ref<ServerCatalogPlan[]>([])

onMounted(async () => {
  loading.value = true
  try {
    const catalog = await getServerCatalog()
    const allPlans = catalog.products.flatMap(p => p.plans)
    topPlans.value = allPlans.filter(p => p.is_featured).slice(0, 3)
    if (topPlans.value.length === 0) {
      topPlans.value = allPlans.slice(0, 3)
    }
  } finally {
    loading.value = false
  }
})

function yuan(cents: number) {
  return (cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)
}

function monthlyPrice(plan: ServerCatalogPlan) {
  return plan.prices.find((price) => price.billing_cycle === 'monthly') || plan.prices[0]
}

const features = [
  {
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path></svg>',
    title: '极致性能',
    desc: '展示后端公开目录返回的 CPU、内存、磁盘、带宽和价格等固定服务器套餐字段。',
    bg: 'rgba(59, 130, 246, 0.1)', color: '#3b82f6'
  },
  {
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"></path></svg>',
    title: '弹性网络',
    desc: '通过公开目录展示销售地域信息，不展示 PVE 节点、资源池或库存扣减信息。',
    bg: 'rgba(16, 185, 129, 0.1)', color: '#10b981'
  },
  {
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>',
    title: '数据安全',
    desc: '产品、价格、地域和系统模板均通过公开目录展示，前端不承诺后端尚未开放的业务能力。',
    bg: 'rgba(245, 158, 11, 0.1)', color: '#f59e0b'
  },
  {
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>',
    title: '智能运维',
    desc: '当前控制台开放登录态和账号资料维护，订单、支付、实例和工单能力后续再接入。',
    bg: 'rgba(139, 92, 246, 0.1)', color: '#8b5cf6'
  }
]

const regions = [
  { name: '华北-北京', coords: '20%, 65%', active: true },
  { name: '华东-上海', coords: '28%, 70%', active: true },
  { name: '华南-广州', coords: '32%, 75%', active: true },
  { name: '美西-洛杉矶', coords: '35%, 20%', active: false },
  { name: '欧洲-法兰克福', coords: '25%, 45%', active: false }
]

const activeFaq = ref<number | null>(0)
const faqs = [
  { q: 'PVECloud 当前开放什么？', a: '当前开放服务器产品目录、价格展示、用户注册登录、密码找回和账号资料维护。' },
  { q: '现在可以直接购买吗？', a: '当前阶段只开放产品目录、价格展示和用户账号自助能力，购买、支付和实例开通尚未开放。' },
  { q: '服务器是否支持自定义镜像？', a: '是的，除了我们提供的主流 Linux 和 Windows 纯净版镜像外，您可以上传自定义 ISO 安装任何您需要的操作系统。' },
  { q: '产品信息以哪里为准？', a: '页面展示的套餐、价格、地域和系统模板均以后端公开目录返回为准。' }
]
</script>

<template>
  <div>
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="container hero-container">
        <div class="hero-content">
          <div class="hero-badge">
            <span class="pulse-dot"></span>
            新一代高性能云计算平台
          </div>
          <h1 class="hero-title">
            构建未来的 <span class="text-gradient">数字世界</span>
          </h1>
          <p class="hero-desc">
            通过公开目录查看服务器产品、套餐价格、销售地域和系统模板；订单、支付和实例能力后续开放。
          </p>
          <div class="hero-actions">
            <RouterLink to="/products" class="btn btn-primary btn-lg">查看云产品目录</RouterLink>
            <RouterLink to="/pricing" class="btn btn-outline btn-lg">查看价格方案</RouterLink>
          </div>
          
          <div class="hero-stats">
            <div class="stat-item">
              <strong>99.99%</strong>
              <span>服务可用性</span>
            </div>
            <div class="stat-item">
              <strong>100+</strong>
              <span>公开目录字段</span>
            </div>
            <div class="stat-item">
              <strong>24/7</strong>
              <span>用户账号自助</span>
            </div>
          </div>
        </div>
        
        <!-- Hero Visual/Dashboard Mockup -->
        <div class="hero-visual">
          <div class="glass-panel dashboard-mockup">
            <div class="mockup-header">
              <div class="mockup-dots"><span class="dot red"></span><span class="dot yellow"></span><span class="dot green"></span></div>
              <div class="mockup-address">console.pvecloud.com</div>
            </div>
            <div class="mockup-body">
              <div class="mockup-sidebar"></div>
              <div class="mockup-main">
                <div class="mockup-cards">
                  <div class="mockup-card"></div>
                  <div class="mockup-card"></div>
                  <div class="mockup-card"></div>
                </div>
                <div class="mockup-chart">
                  <div class="chart-line"></div>
                </div>
              </div>
            </div>
          </div>
          <!-- Floating elements -->
          <div class="floating-element float-1"></div>
          <div class="floating-element float-2"></div>
        </div>
      </div>
    </section>

    <!-- Partners / Logos -->
    <section class="partners-section">
      <div class="container">
        <p class="partners-title">被超过 10,000+ 开发者与企业团队所信赖</p>
        <div class="partners-grid">
          <div v-for="i in 6" :key="i" class="partner-logo">LOGO {{i}}</div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="features-section py-32">
      <div class="container">
        <div class="section-header">
          <span class="subtitle">Why Choose Us</span>
          <h2>为什么选择 PVECloud？</h2>
          <p>我们不仅仅提供服务器，更提供一整套可靠、易用、高效的基础设施解决方案，让您将全部精力专注于业务代码本身。</p>
        </div>
        
        <div class="grid gap-8 md:grid-cols-2 lg:grid-cols-4">
          <div v-for="f in features" :key="f.title" class="card card-hover feature-card">
            <div class="feature-icon-wrap" :style="{ background: f.bg, color: f.color }">
              <div class="feature-icon" v-html="f.icon"></div>
            </div>
            <h3>{{ f.title }}</h3>
            <p>{{ f.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Featured Products Section -->
    <section class="products-section py-32" style="background: var(--c-surface-dim);">
      <div class="container">
        <div class="section-header">
          <span class="subtitle">Featured Plans</span>
          <h2>爆款推荐，开箱即用</h2>
          <p>精选最受欢迎的计算规格，平衡性能与成本，满足多样化的业务场景。</p>
        </div>

        <div v-if="loading" class="text-center py-20">
          <div class="spinner" style="margin: 0 auto"></div>
          <p style="margin-top: 20px; color: var(--c-text-2);">正在拉取最新方案...</p>
        </div>
        
        <div v-else class="grid gap-8 md:grid-cols-3">
          <div v-for="plan in topPlans" :key="plan.plan_no" class="card plan-card" :class="{ 'plan-featured': plan.is_featured }">
            <div v-if="plan.is_featured" class="plan-popular-badge">最受欢迎</div>
            
            <div class="plan-header">
              <h3 class="plan-name">{{ plan.name }}</h3>
              <p class="plan-summary">{{ plan.summary || '固定规格服务器套餐' }}</p>
            </div>
            
            <div class="plan-price">
              <span class="price-currency">¥</span>
              <span class="price-amount">{{ yuan(monthlyPrice(plan).price_cents) }}</span>
              <span class="price-period">/月</span>
            </div>
            
            <ul class="plan-specs">
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
                <span><strong>{{ plan.cpu_cores }}</strong> vCPU 核处理器</span>
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
                <span><strong>{{ Math.round(plan.memory_mb / 1024) }} GB</strong> 高速内存</span>
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
                <span><strong>{{ plan.system_disk_gb }} GB</strong> 系统盘</span>
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
                <span><strong>{{ plan.bandwidth_mbps }} Mbps</strong> 专线带宽</span>
              </li>
            </ul>
            
            <RouterLink to="/login" class="btn btn-block" :class="plan.is_featured ? 'btn-primary' : 'btn-outline'">
              登录后查看购买入口
            </RouterLink>
          </div>
        </div>
        
        <div class="text-center" style="margin-top: 48px;">
          <RouterLink to="/pricing" class="btn btn-outline">查看所有配置方案</RouterLink>
        </div>
      </div>
    </section>

    <!-- Global Network Map -->
    <section class="network-section py-32">
      <div class="container">
        <div class="grid gap-12 lg:grid-cols-2" style="align-items: center;">
          <div class="network-content">
            <span class="subtitle">Global Network</span>
            <h2 style="font-size: clamp(2rem, 3vw, 2.5rem); font-weight: 800; margin-bottom: 24px;">公开销售地域展示</h2>
            <p style="color: var(--c-text-2); font-size: 1.1rem; line-height: 1.6; margin-bottom: 32px;">
              页面只展示后端公开目录返回的销售地域，不展示 PVE 节点、资源池、库存扣减或实例信息。
            </p>
            <ul class="network-features">
              <li><span class="check-circle"></span> 销售地域以后端公开目录为准</li>
              <li><span class="check-circle"></span> 系统模板按套餐可用关系展示</li>
              <li><span class="check-circle"></span> 不展示未开放的节点和库存信息</li>
            </ul>
          </div>
          
          <div class="network-map glass-panel">
            <div class="map-bg"></div>
            <!-- Dynamic Pins -->
            <div v-for="region in regions" :key="region.name" 
                 class="map-pin" 
                 :class="{ active: region.active }"
                 :style="{ top: region.coords.split(', ')[0], left: region.coords.split(', ')[1] }"
            >
              <div class="pin-dot"></div>
              <div class="pin-pulse"></div>
              <div class="pin-tooltip">{{ region.name }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="faq-section py-32" style="background: var(--c-surface-dim);">
      <div class="container">
        <div class="section-header">
          <h2>常见问题</h2>
          <p>解答您在购买和使用过程中的疑惑</p>
        </div>
        
        <div class="faq-container">
          <div v-for="(item, index) in faqs" :key="index" class="faq-item" :class="{ active: activeFaq === index }">
            <div class="faq-question" @click="activeFaq = activeFaq === index ? null : index">
              <h3>{{ item.q }}</h3>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"></polyline></svg>
            </div>
            <div class="faq-answer" :style="{ maxHeight: activeFaq === index ? '200px' : '0' }">
              <p>{{ item.a }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="cta-section py-32 text-center">
      <div class="container">
        <h2 style="font-size: clamp(2.5rem, 4vw, 3.5rem); font-weight: 800; margin-bottom: 24px;">准备好开始了？</h2>
        <p style="font-size: 1.25rem; color: var(--c-text-2); max-width: 600px; margin: 0 auto 40px;">
          注册后可进入用户中心维护账号资料；购买、支付和实例能力会在后续阶段开放。
        </p>
        <RouterLink to="/register" class="btn btn-primary btn-lg" style="border-radius: 99px;">注册账户</RouterLink>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* Hero Section */
.hero-section {
  position: relative;
  padding: 120px 0 80px;
  overflow: hidden;
}
.hero-container {
  display: grid;
  gap: 48px;
  align-items: center;
}
@media (min-width: 1024px) {
  .hero-container { grid-template-columns: 1fr 1fr; padding-top: 40px; }
}
.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 99px;
  background: var(--c-surface-dim);
  border: 1px solid var(--c-border);
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 32px;
}
.pulse-dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: var(--c-primary);
  box-shadow: 0 0 10px var(--c-primary);
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.4); }
  70% { box-shadow: 0 0 0 6px rgba(59, 130, 246, 0); }
  100% { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0); }
}
.hero-title {
  font-size: clamp(3rem, 5vw, 4.5rem);
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.03em;
  margin-bottom: 24px;
}
.hero-desc {
  font-size: 1.15rem;
  color: var(--c-text-2);
  line-height: 1.6;
  margin-bottom: 40px;
  max-width: 540px;
}
.hero-actions {
  display: flex; gap: 16px; margin-bottom: 48px;
}
.hero-stats {
  display: flex; gap: 32px;
  border-top: 1px solid var(--c-border-light);
  padding-top: 32px;
}
.stat-item { display: flex; flex-direction: column; gap: 4px; }
.stat-item strong { font-size: 1.5rem; font-weight: 800; color: var(--c-text); }
.stat-item span { font-size: 0.875rem; color: var(--c-text-3); font-weight: 600; }

/* Hero Visual Mockup */
.hero-visual { position: relative; perspective: 1000px; }
.dashboard-mockup {
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-lg), var(--shadow-glow);
  transform: rotateY(-5deg) rotateX(5deg);
  transform-style: preserve-3d;
  height: 400px;
  display: flex; flex-direction: column;
}
.mockup-header {
  height: 40px; background: rgba(0,0,0,0.3); border-bottom: 1px solid var(--c-border);
  display: flex; align-items: center; padding: 0 16px; gap: 16px;
}
.mockup-dots { display: flex; gap: 6px; }
.mockup-dots .dot { width: 10px; height: 10px; border-radius: 50%; }
.dot.red { background: #ff5f56; } .dot.yellow { background: #ffbd2e; } .dot.green { background: #27c93f; }
.mockup-address { background: var(--c-surface-dim); padding: 4px 12px; border-radius: 4px; font-size: 0.75rem; color: var(--c-text-3); margin: 0 auto; }
.mockup-body { flex: 1; display: flex; padding: 16px; gap: 16px; }
.mockup-sidebar { width: 60px; border-radius: var(--radius); background: var(--c-surface-dim); border: 1px solid var(--c-border-light); }
.mockup-main { flex: 1; display: flex; flex-direction: column; gap: 16px; }
.mockup-cards { display: flex; gap: 16px; }
.mockup-card { flex: 1; height: 80px; border-radius: var(--radius); background: var(--c-surface-dim); border: 1px solid var(--c-border-light); }
.mockup-chart { flex: 1; border-radius: var(--radius); background: var(--c-surface-dim); border: 1px solid var(--c-border-light); position: relative; overflow: hidden;}
.chart-line {
  position: absolute; bottom: 0; left: 0; right: 0; height: 60%;
  background: linear-gradient(to top, rgba(59,130,246,0.2), transparent);
  border-top: 2px solid var(--c-primary);
  clip-path: polygon(0 100%, 0 50%, 20% 40%, 40% 60%, 60% 30%, 80% 50%, 100% 20%, 100% 100%);
}

.floating-element {
  position: absolute;
  border-radius: var(--radius);
  background: var(--c-card);
  border: 1px solid var(--c-border);
  box-shadow: var(--shadow-lg);
  backdrop-filter: blur(12px);
}
.float-1 { width: 120px; height: 80px; top: 10%; right: -5%; transform: translateZ(50px); animation: float 6s ease-in-out infinite; }
.float-2 { width: 160px; height: 60px; bottom: 20%; left: -10%; transform: translateZ(80px); animation: float 8s ease-in-out infinite reverse; }
@keyframes float {
  0%, 100% { transform: translateY(0) translateZ(50px); }
  50% { transform: translateY(-20px) translateZ(50px); }
}

/* Partners */
.partners-section { border-top: 1px solid var(--c-border-light); border-bottom: 1px solid var(--c-border-light); padding: 40px 0; background: rgba(0,0,0,0.2); }
.partners-title { text-align: center; color: var(--c-text-3); font-size: 0.875rem; font-weight: 600; margin-bottom: 24px; text-transform: uppercase; letter-spacing: 0.05em; }
.partners-grid { display: flex; flex-wrap: wrap; justify-content: center; gap: 40px; opacity: 0.4; filter: grayscale(1); }
.partner-logo { font-size: 1.5rem; font-weight: 900; letter-spacing: -0.05em; }

/* Features */
.feature-card { padding: 32px; }
.feature-icon-wrap { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; margin-bottom: 24px; }
.feature-icon { width: 24px; height: 24px; }
.feature-card h3 { font-size: 1.25rem; font-weight: 700; margin-bottom: 12px; }
.feature-card p { color: var(--c-text-2); font-size: 0.95rem; line-height: 1.6; }

/* Pricing Plans */
.plan-card { padding: 32px; position: relative; display: flex; flex-direction: column; }
.plan-featured { border-color: var(--c-primary); transform: scale(1.02); box-shadow: var(--shadow-lg), var(--shadow-glow); }
.plan-popular-badge {
  position: absolute; top: -14px; left: 50%; transform: translateX(-50%);
  background: linear-gradient(135deg, var(--c-primary), var(--c-accent));
  color: white; padding: 4px 16px; border-radius: 99px; font-size: 0.75rem; font-weight: 700;
}
.plan-header { margin-bottom: 24px; }
.plan-name { font-size: 1.5rem; font-weight: 800; margin-bottom: 8px; }
.plan-summary { color: var(--c-text-2); font-size: 0.9rem; }
.plan-price { display: flex; align-items: baseline; gap: 4px; margin-bottom: 32px; }
.price-currency { font-size: 1.5rem; font-weight: 700; color: var(--c-text); }
.price-amount { font-size: 3rem; font-weight: 800; color: var(--c-text); line-height: 1; letter-spacing: -0.03em; }
.price-period { color: var(--c-text-3); font-size: 1rem; }
.plan-specs { list-style: none; margin-bottom: 32px; flex: 1; display: grid; gap: 16px; }
.plan-specs li { display: flex; align-items: center; gap: 12px; color: var(--c-text-2); font-size: 0.95rem; }
.plan-specs li svg { width: 18px; height: 18px; color: var(--c-primary); flex-shrink: 0; }
.plan-specs strong { color: var(--c-text); font-weight: 600; }

/* Network Map */
.network-features { list-style: none; display: grid; gap: 16px; }
.network-features li { display: flex; align-items: center; gap: 12px; font-size: 1.05rem; font-weight: 500; }
.check-circle { width: 24px; height: 24px; border-radius: 50%; background: var(--c-success-soft); color: var(--c-success); display: flex; align-items: center; justify-content: center; }
.check-circle::after { content: '✓'; font-weight: bold; }
.network-map { height: 400px; position: relative; border-radius: var(--radius-xl); overflow: hidden; }
.map-bg {
  position: absolute; inset: 0;
  background-image: radial-gradient(rgba(255,255,255,0.1) 1px, transparent 1px);
  background-size: 20px 20px;
}
.map-pin { position: absolute; transform: translate(-50%, -50%); cursor: pointer; z-index: 10; }
.pin-dot { width: 12px; height: 12px; border-radius: 50%; background: var(--c-text-3); position: relative; z-index: 2; transition: all 0.3s; }
.map-pin.active .pin-dot { background: var(--c-primary); }
.pin-pulse { position: absolute; inset: -10px; border-radius: 50%; background: rgba(59,130,246,0.3); z-index: 1; animation: mapPulse 2s infinite; opacity: 0; }
.map-pin.active .pin-pulse { opacity: 1; }
@keyframes mapPulse { 0% { transform: scale(0.5); opacity: 1; } 100% { transform: scale(2.5); opacity: 0; } }
.pin-tooltip {
  position: absolute; bottom: 100%; left: 50%; transform: translateX(-50%) translateY(-10px);
  background: var(--c-card); border: 1px solid var(--c-border); padding: 4px 12px; border-radius: 4px;
  font-size: 0.75rem; white-space: nowrap; opacity: 0; visibility: hidden; transition: all 0.2s;
}
.map-pin:hover .pin-tooltip { opacity: 1; visibility: visible; transform: translateX(-50%) translateY(-5px); }

/* FAQ */
.faq-container { max-width: 800px; margin: 0 auto; display: grid; gap: 16px; }
.faq-item { background: var(--c-card); border: 1px solid var(--c-border); border-radius: var(--radius); overflow: hidden; }
.faq-question { padding: 20px 24px; display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none; }
.faq-question h3 { font-size: 1.1rem; font-weight: 600; margin: 0; }
.faq-question svg { transition: transform 0.3s; }
.faq-item.active .faq-question svg { transform: rotate(180deg); color: var(--c-primary); }
.faq-answer { overflow: hidden; transition: max-height 0.3s ease-in-out; }
.faq-answer p { padding: 0 24px 20px; color: var(--c-text-2); line-height: 1.6; }

/* Mobile */
@media (max-width: 768px) {
  .hero-actions { flex-direction: column; }
  .hero-stats { flex-direction: column; gap: 16px; }
}
</style>
