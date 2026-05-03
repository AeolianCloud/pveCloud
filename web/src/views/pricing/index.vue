<script setup lang="ts">
const tiers = [
  { name: 'Cloud VM-1', tag: '入门', tagColor: 'green', price: '58', cpu: '2 vCPU', ram: '2 GB', disk: '50 GB SSD', bw: '1 TB/月', ip: '1 个' },
  { name: 'Cloud VM-2', tag: '推荐', tagColor: 'primary', price: '128', cpu: '4 vCPU', ram: '8 GB', disk: '100 GB SSD', bw: '3 TB/月', ip: '1 个', featured: true },
  { name: 'Cloud VM-3', tag: '企业', tagColor: 'purple', price: '368', cpu: '8 vCPU', ram: '16 GB', disk: '200 GB SSD', bw: '5 TB/月', ip: '2 个' },
]

const specRows = [
  { label: 'CPU', values: ['2 vCPU', '4 vCPU', '8 vCPU'] },
  { label: '内存', values: ['2 GB', '8 GB', '16 GB'] },
  { label: '系统盘', values: ['50 GB SSD', '100 GB SSD', '200 GB SSD'] },
  { label: '月流量', values: ['1 TB/月', '3 TB/月', '5 TB/月'] },
  { label: '公网 IP', values: ['1 个', '1 个', '2 个'] },
]

const extras = [
  { label: '快照', values: ['最多 3 个', '最多 6 个', '最多 10 个'] },
  { label: '备份', values: ['—', '每周自动', '每日自动'] },
  { label: '安全组', values: ['基础', '标准', '高级'] },
  { label: '技术支持', values: ['工单', '工单 + 在线', '7×24 专属'] },
  { label: 'SLA', values: ['99.9%', '99.95%', '99.99%'] },
]
</script>

<template>
  <section class="page content-page">
    <div class="section-pad">
      <div class="sec-header center" style="margin-bottom: clamp(28px, 4vw, 48px);">
        <p class="label">价格方案</p>
        <h2>选择适合您的云服务器</h2>
        <p>透明定价，按月付费，续费同价。所有方案均包含基础 DDoS 防护和 7×24 系统监控。</p>
      </div>

      <!-- Plan Cards -->
      <div class="pricing-preview-grid" style="margin-bottom: clamp(40px, 6vw, 72px);">
        <div v-for="t in tiers" :key="t.name" class="pricing-preview-card" :class="{ featured: t.featured }">
          <span class="plan-tag" :class="t.tagColor">{{ t.tag }}</span>
          <div class="plan-name">{{ t.name }}</div>
          <div class="plan-price"><strong>¥{{ t.price }}</strong><span>/月</span></div>
          <div class="plan-specs">
            <div class="plan-spec">{{ t.cpu }}</div>
            <div class="plan-spec">{{ t.ram }} 内存</div>
            <div class="plan-spec">{{ t.disk }}</div>
            <div class="plan-spec">{{ t.bw }} 流量</div>
            <div class="plan-spec">{{ t.ip }}</div>
          </div>
          <RouterLink to="/login" class="btn btn-primary btn-sm" style="width:100%">立即选购</RouterLink>
        </div>
      </div>

      <!-- Comparison Table -->
      <div class="sec-header center" style="margin-bottom: clamp(20px, 3vw, 32px);">
        <h2>方案对比</h2>
      </div>
      <div class="pricing-table-wrap">
        <table class="pricing-table">
          <thead>
            <tr>
              <th>规格项</th>
              <th>Cloud VM-1</th>
              <th class="featured-col">Cloud VM-2</th>
              <th>Cloud VM-3</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>月价</td>
              <td class="price-cell">¥58</td>
              <td class="price-cell featured-col">¥128</td>
              <td class="price-cell">¥368</td>
            </tr>
            <tr v-for="row in specRows" :key="row.label">
              <td>{{ row.label }}</td>
              <td>{{ row.values[0] }}</td>
              <td class="featured-col">{{ row.values[1] }}</td>
              <td>{{ row.values[2] }}</td>
            </tr>
            <tr v-for="e in extras" :key="e.label">
              <td>{{ e.label }}</td>
              <td>{{ e.values[0] }}</td>
              <td class="featured-col">{{ e.values[1] }}</td>
              <td>{{ e.values[2] }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- FAQ -->
      <div class="sec-header center" style="margin: clamp(40px, 6vw, 72px) auto 0;">
        <p class="label">常见问题</p>
        <h2>关于云服务器</h2>
      </div>
      <div style="max-width:720px; margin: clamp(20px, 3vw, 32px) auto 0; display:grid; gap:12px;">
        <div style="border:1px solid var(--c-border); border-radius:var(--radius-lg); padding:20px 24px; background:var(--c-surface);">
          <div style="font-weight:680; color:var(--c-text); margin-bottom:6px;">是否支持随时升级配置？</div>
          <div style="color:var(--c-text-2); font-size:.92rem; line-height:1.6;">支持。您可以在控制台随时升级 CPU、内存和磁盘，升级后立即生效，费用按差价折算。</div>
        </div>
        <div style="border:1px solid var(--c-border); border-radius:var(--radius-lg); padding:20px 24px; background:var(--c-surface);">
          <div style="font-weight:680; color:var(--c-text); margin-bottom:6px;">流量超出后如何计费？</div>
          <div style="color:var(--c-text-2); font-size:.92rem; line-height:1.6;">超出套餐流量后按 0.8 元/GB 计费，也可以提前购买流量包降低单价。</div>
        </div>
        <div style="border:1px solid var(--c-border); border-radius:var(--radius-lg); padding:20px 24px; background:var(--c-surface);">
          <div style="font-weight:680; color:var(--c-text); margin-bottom:6px;">支持哪些操作系统？</div>
          <div style="color:var(--c-text-2); font-size:.92rem; line-height:1.6;">支持主流 Linux 发行版（CentOS、Ubuntu、Debian、Rocky Linux）和 Windows Server。</div>
        </div>
        <div style="border:1px solid var(--c-border); border-radius:var(--radius-lg); padding:20px 24px; background:var(--c-surface);">
          <div style="font-weight:680; color:var(--c-text); margin-bottom:6px;">是否提供 API 接口？</div>
          <div style="color:var(--c-text-2); font-size:.92rem; line-height:1.6;">提供完整的 RESTful API，支持实例生命周期管理、监控数据查询和账单查询。</div>
        </div>
      </div>
    </div>
  </section>
</template>
