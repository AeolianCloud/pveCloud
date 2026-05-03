<script setup lang="ts">
const plans = [
  {
    tag: '入门', tagColor: 'green', name: 'Cloud VM-1',
    price: '58', cpu: '2 vCPU', ram: '2 GB', disk: '50 GB SSD', bw: '1 TB/月', ip: '1 个公网 IP',
  },
  {
    tag: '推荐', tagColor: 'primary', name: 'Cloud VM-2', featured: true,
    price: '128', cpu: '4 vCPU', ram: '8 GB', disk: '100 GB SSD', bw: '3 TB/月', ip: '1 个公网 IP',
  },
  {
    tag: '企业', tagColor: 'purple', name: 'Cloud VM-3',
    price: '368', cpu: '8 vCPU', ram: '16 GB', disk: '200 GB SSD', bw: '5 TB/月', ip: '2 个公网 IP',
  },
]

const features = [
  { icon: 'cpu', color: 'blue', title: '高性能计算', desc: '基于 KVM 虚拟化，Intel/AMD x86 架构，保障计算性能稳定可靠。' },
  { icon: 'lock', color: 'green', title: '安全隔离', desc: '每台实例独立运行，支持安全组、防火墙和网络隔离策略。' },
  { icon: 'net', color: 'orange', title: '弹性网络', desc: '支持弹性公网 IP、内网互通和自定义网络拓扑。' },
  { icon: 'disk', color: 'purple', title: '快照与备份', desc: '支持系统盘和数据盘快照，按需回滚和备份恢复。' },
  { icon: 'chart', color: 'blue', title: '实时监控', desc: 'CPU、内存、磁盘和网络流量可视化监控，异常自动告警。' },
  { icon: 'gear', color: 'green', title: '灵活管理', desc: 'Web 控制台 + API 双入口，支持实例启停、重装和密码重置。' },
]

const trustItems = [
  { value: '99.95%', label: '可用性 SLA' },
  { value: '5', label: '数据中心' },
  { value: '< 10ms', label: '内网延迟' },
  { value: '7x24', label: '技术支持' },
]
</script>

<template>
  <section class="page home-page">
    <!-- Hero -->
    <div class="hero">
      <div class="hero-bg-blob blob-1"></div>
      <div class="hero-bg-blob blob-2"></div>
      
      <div style="position: relative; z-index: 2;">
        <div class="hero-badge">Cloud Server</div>
        <h1>按需弹性的<em>云主机</em>服务</h1>
        <p class="hero-desc">
          基于虚拟化技术提供高性能云服务器，按需付费、弹性扩容。
          从个人测试到企业生产，一站式满足您的计算需求。
        </p>
        <div class="hero-actions">
          <RouterLink to="/pricing" class="btn btn-primary">查看价格方案</RouterLink>
          <RouterLink to="/products" class="btn btn-outline">了解产品规格</RouterLink>
        </div>
      </div>

      <div class="hero-pricing" style="position: relative; z-index: 2;">
        <div class="hero-plan-name">入门方案</div>
        <div class="hero-plan-title">Cloud VM-1</div>
        <div class="hero-plan-price"><strong>58</strong><span>元/月起</span></div>
        <div class="hero-specs">
          <div class="hero-spec-row"><span>CPU</span><strong>2 vCPU</strong></div>
          <div class="hero-spec-row"><span>内存</span><strong>2 GB</strong></div>
          <div class="hero-spec-row"><span>系统盘</span><strong>50 GB SSD</strong></div>
          <div class="hero-spec-row"><span>流量</span><strong>1 TB/月</strong></div>
          <div class="hero-spec-row"><span>公网 IP</span><strong>1 个</strong></div>
        </div>
        <RouterLink to="/pricing" class="btn btn-primary" style="width:100%">立即选购</RouterLink>
      </div>
    </div>

    <!-- Features -->
    <div class="section-pad">
      <div class="sec-header center" style="margin-bottom: clamp(32px, 5vw, 56px);">
        <p class="label">核心能力</p>
        <h2>为开发者和企业打造的云基础设施</h2>
        <p>从实例创建到运维管理，全流程自助化，按量计费，用多少付多少。</p>
      </div>
      <div class="features-grid">
        <div v-for="f in features" :key="f.title" class="feature-card">
          <div class="feature-icon" :class="[f.color, f.icon]"></div>
          <h3>{{ f.title }}</h3>
          <p>{{ f.desc }}</p>
        </div>
      </div>
    </div>

    <!-- Pricing Preview -->
    <div class="section-pad" style="background: var(--c-surface-dim); border-top: 1px solid var(--c-border-light); border-bottom: 1px solid var(--c-border-light);">
      <div class="sec-header center" style="margin-bottom: clamp(32px, 5vw, 56px);">
        <p class="label">价格方案</p>
        <h2>透明定价，按需选择</h2>
        <p>所有方案均支持按月付费，续费同价，无隐藏费用。</p>
      </div>
      <div class="pricing-preview-grid">
        <div v-for="p in plans" :key="p.name" class="pricing-preview-card" :class="{ featured: p.featured }">
          <span class="plan-tag" :class="p.tagColor">{{ p.tag }}</span>
          <div class="plan-name">{{ p.name }}</div>
          <div class="plan-price"><strong>{{ p.price }}</strong><span>元/月</span></div>
          <div class="plan-specs">
            <div class="plan-spec">{{ p.cpu }}</div>
            <div class="plan-spec">{{ p.ram }} 内存</div>
            <div class="plan-spec">{{ p.disk }}</div>
            <div class="plan-spec">{{ p.bw }} 流量</div>
            <div class="plan-spec">{{ p.ip }}</div>
          </div>
          <div class="plan-cta">
            <RouterLink to="/pricing" class="btn btn-outline btn-sm" style="width:100%">查看方案详情</RouterLink>
          </div>
        </div>
      </div>
    </div>

    <!-- Trust -->
    <div class="section-pad">
      <div class="sec-header center" style="margin-bottom: clamp(32px, 5vw, 56px);">
        <p class="label">为什么选择 pveCloud</p>
        <h2>稳定、安全、高性价比</h2>
      </div>
      <div class="trust-bar">
        <div v-for="t in trustItems" :key="t.label" class="trust-item">
          <strong>{{ t.value }}</strong>
          <span>{{ t.label }}</span>
        </div>
      </div>
    </div>

    <!-- CTA -->
    <div class="cta-band">
      <h2>准备好开始了吗</h2>
      <p>几分钟内即可创建您的第一台云服务器，按需付费，随时扩容。</p>
      <div class="cta-actions">
        <RouterLink to="/pricing" class="btn btn-white">立即选购</RouterLink>
        <RouterLink to="/products" class="btn btn-ghost">了解产品</RouterLink>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero-bg-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  z-index: 1;
  animation: float 10s infinite alternate ease-in-out;
  opacity: 0.6;
}
.blob-1 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.4) 0%, transparent 70%);
  top: -100px;
  left: -100px;
}
.blob-2 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.3) 0%, transparent 70%);
  bottom: -200px;
  right: -100px;
  animation-delay: -5s;
}

@keyframes float {
  0% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(20px, -20px) scale(1.05); }
  100% { transform: translate(-20px, 20px) scale(0.95); }
}
</style>
