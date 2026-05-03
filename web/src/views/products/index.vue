<script setup lang="ts">
const specs = [
  {
    name: 'Cloud VM-1', desc: '适合个人测试、轻量应用和开发环境。',
    price: '58', tag: '入门', tagColor: 'green',
    items: [
      { label: 'CPU', value: '2 vCPU' },
      { label: '内存', value: '2 GB' },
      { label: '系统盘', value: '50 GB SSD' },
      { label: '数据盘', value: '可选 10-500 GB' },
      { label: '流量', value: '1 TB/月' },
      { label: '公网 IP', value: '1 个' },
      { label: '带宽', value: '最高 100 Mbps' },
    ],
  },
  {
    name: 'Cloud VM-2', desc: '适合中小型业务、Web 应用和 API 服务。',
    price: '128', tag: '推荐', tagColor: 'primary', featured: true,
    items: [
      { label: 'CPU', value: '4 vCPU' },
      { label: '内存', value: '8 GB' },
      { label: '系统盘', value: '100 GB SSD' },
      { label: '数据盘', value: '可选 10-1000 GB' },
      { label: '流量', value: '3 TB/月' },
      { label: '公网 IP', value: '1 个' },
      { label: '带宽', value: '最高 200 Mbps' },
    ],
  },
  {
    name: 'Cloud VM-3', desc: '适合企业核心系统、数据库和高并发业务。',
    price: '368', tag: '企业', tagColor: 'purple',
    items: [
      { label: 'CPU', value: '8 vCPU' },
      { label: '内存', value: '16 GB' },
      { label: '系统盘', value: '200 GB SSD' },
      { label: '数据盘', value: '可选 10-2000 GB' },
      { label: '流量', value: '5 TB/月' },
      { label: '公网 IP', value: '2 个' },
      { label: '带宽', value: '最高 500 Mbps' },
    ],
  },
  {
    name: 'Cloud VM-4', desc: '适合大规模计算、AI 训练和高性能数据库。',
    price: '688', tag: '旗舰', tagColor: 'purple',
    items: [
      { label: 'CPU', value: '16 vCPU' },
      { label: '内存', value: '32 GB' },
      { label: '系统盘', value: '500 GB SSD' },
      { label: '数据盘', value: '可选 10-4000 GB' },
      { label: '流量', value: '10 TB/月' },
      { label: '公网 IP', value: '2 个' },
      { label: '带宽', value: '最高 1 Gbps' },
    ],
  },
  {
    name: 'Cloud VM-5', desc: '适合核心交易系统和高可用企业架构。',
    price: '1288', tag: '旗舰', tagColor: 'purple',
    items: [
      { label: 'CPU', value: '32 vCPU' },
      { label: '内存', value: '64 GB' },
      { label: '系统盘', value: '1 TB SSD' },
      { label: '数据盘', value: '可选 10-8000 GB' },
      { label: '流量', value: '20 TB/月' },
      { label: '公网 IP', value: '4 个' },
      { label: '带宽', value: '最高 2 Gbps' },
    ],
  },
  {
    name: 'Cloud VM-6', desc: '适合超大规模部署和专属高性能计算集群。',
    price: '2588', tag: '旗舰', tagColor: 'purple',
    items: [
      { label: 'CPU', value: '64 vCPU' },
      { label: '内存', value: '128 GB' },
      { label: '系统盘', value: '2 TB SSD' },
      { label: '数据盘', value: '可选 10-16000 GB' },
      { label: '流量', value: '50 TB/月' },
      { label: '公网 IP', value: '4 个' },
      { label: '带宽', value: '最高 5 Gbps' },
    ],
  },
]
</script>

<template>
  <section class="page content-page">
    <div class="section-pad">
      <div class="sec-header" style="margin-bottom: clamp(28px, 4vw, 48px);">
        <p class="label">产品规格</p>
        <h2>云主机实例</h2>
        <p>从入门到旗舰，覆盖个人测试、中小企业和大规模企业级场景。所有实例均支持按月付费和弹性扩容。</p>
      </div>

      <div class="spec-grid">
        <div v-for="s in specs" :key="s.name" class="spec-card" :class="{ featured: s.featured }">
          <div class="spec-card-head">
            <span class="plan-tag" :class="s.tagColor">{{ s.tag }}</span>
            <h3>{{ s.name }}</h3>
            <p>{{ s.desc }}</p>
          </div>
          <div class="spec-price-row">
            <strong>¥{{ s.price }}</strong><span>/月</span>
          </div>
          <div class="spec-list">
            <div v-for="item in s.items" :key="item.label" class="spec-item">
              <span class="dot"></span>
              <span><strong>{{ item.label }}</strong> {{ item.value }}</span>
            </div>
          </div>
          <RouterLink to="/pricing" class="btn btn-outline btn-sm" style="width:100%">查看价格</RouterLink>
        </div>
      </div>
    </div>
  </section>
</template>
