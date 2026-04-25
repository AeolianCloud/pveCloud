<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  BadgeCheck,
  Box,
  CircleDollarSign,
  ClipboardCheck,
  Layers3,
  RefreshCw,
  Server,
  ShoppingCart,
  TicketCheck,
  Users,
} from 'lucide-vue-next'

import { getAdminDashboard } from '../api/dashboard'
import { useAuthStore } from '../stores/auth'
import type { DashboardMetric } from '../types/dashboard'

const auth = useAuthStore()
const loading = ref(false)
const errorMessage = ref('')
const metrics = ref<DashboardMetric[]>([])

const fallbackStats = [
  { key: 'sales', title: '今日销售额', value: '￥258,960', trend: '+23.6%', tone: 'blue', icon: CircleDollarSign },
  { key: 'orders', title: '今日订单数', value: '128', trend: '+18.4%', tone: 'green', icon: ClipboardCheck },
  { key: 'customers', title: '新增客户数', value: '1,356', trend: '+12.7%', tone: 'purple', icon: Users },
  { key: 'instances', title: '运行云服务器', value: '3,692', trend: '+8.3%', tone: 'orange', icon: Layers3 },
  { key: 'tickets', title: '待处理工单', value: '32', trend: '-5.9%', tone: 'red', icon: TicketCheck },
]

const chartPoints = '0,62 72,64 144,68 216,72 288,24 360,52 432,76 504,58'
const sparkline = '0,48 18,43 36,45 54,30 72,36 90,28 108,33 126,14'

const productShares = [
  { name: '标准型云服务器', value: '38.7%', color: '#2f7cf6' },
  { name: '计算型云服务器', value: '23.1%', color: '#2fc28b' },
  { name: '内存型云服务器', value: '16.8%', color: '#f59e0b' },
  { name: 'GPU云服务器', value: '11.4%', color: '#ff4f5c' },
  { name: '其他套餐', value: '10.0%', color: '#7c8aa4' },
]

const resourceItems = [
  { label: 'CPU使用率', value: '42%', status: '正常', tone: 'green' },
  { label: '内存使用率', value: '55%', status: '正常', tone: 'green' },
  { label: '带宽使用率', value: '68%', status: '关注', tone: 'orange' },
  { label: 'IP可用数', value: '1,256个', status: '充足', tone: 'green' },
  { label: '服务SLA达成率', value: '98.6%', status: '稳定', tone: 'green' },
]

const alerts = [
  { level: '严重', text: '节点1 CPU使用率超过 90%', time: '05-22 10:15', tone: 'red' },
  { level: '警告', text: '节点2 带宽使用率超过 80%', time: '05-22 09:48', tone: 'orange' },
  { level: '提示', text: '公网 IP 可用数量低于预警阈值', time: '05-22 09:20', tone: 'yellow' },
  { level: '提示', text: '华南节点正在进行计划维护', time: '05-22 08:55', tone: 'blue' },
]

const orders = [
  ['DD202505220001', '上海智云科技有限公司', '4核8G 5M', '上海节点1', '￥1,560.00', '已支付', '05-22 10:23'],
  ['DD202505220002', '杭州数海信息技术有限公司', '2核4G 3M', '杭州节点2', '￥960.00', '已支付', '05-22 10:18'],
  ['DD202505220003', '深圳云拓网络有限公司', '8核16G 10M', '深圳节点1', '￥3,360.00', '待开通', '05-22 10:05'],
  ['DD202505220004', '广州星河智能科技', 'GPU 1卡24G', '上海节点1', '￥6,800.00', '待支付', '05-22 09:58'],
  ['DD202505220005', '北京启航数据服务', '2核4G 5M', '北京节点1', '￥1,260.00', '已支付', '05-22 09:47'],
]

const nodes = [
  ['上海节点1', '在线', '42%', '53%', '892'],
  ['杭州节点2', '在线', '38%', '49%', '756'],
  ['深圳节点1', '在线', '45%', '57%', '684'],
  ['北京节点1', '在线', '36%', '43%', '512'],
  ['华南节点', '维护中', '41%', '52%', '848'],
]

const displayStats = computed(() => {
  if (metrics.value.length === 0) {
    return fallbackStats
  }

  return fallbackStats.map((item, index) => {
    const metric = metrics.value[index]
    if (!metric) {
      return item
    }
    return {
      ...item,
      key: metric.key,
      title: metric.title,
      value: `${metric.value.toLocaleString()}${metric.unit || ''}`,
    }
  })
})

async function loadDashboard() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await getAdminDashboard()
    auth.applyDashboard(result)
    metrics.value = result.metrics
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '控制台数据加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboard)
</script>

<template>
  <section class="dashboard-page">
    <div class="stat-grid">
      <article v-for="stat in displayStats" :key="stat.key" class="stat-card" :class="`stat-card--${stat.tone}`">
        <span class="stat-icon">
          <component :is="stat.icon" :size="28" aria-hidden="true" />
        </span>
        <div class="stat-main">
          <span>{{ stat.title }}</span>
          <strong>{{ stat.value }}</strong>
        </div>
        <div class="stat-footer">
          <span>较昨日</span>
          <strong>{{ stat.trend }}</strong>
        </div>
        <svg class="stat-sparkline" viewBox="0 0 126 58" preserveAspectRatio="none" aria-hidden="true">
          <polyline :points="sparkline" fill="none" stroke="currentColor" stroke-width="3" />
        </svg>
      </article>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <div class="dashboard-main-grid">
      <article class="dashboard-card sales-card">
        <header class="card-header">
          <h2>销售趋势<span>近 7 日</span></h2>
          <button class="ghost-select" type="button">近7天</button>
        </header>
        <div class="chart-legend"><span></span>销售额（元）</div>
        <div class="line-chart">
          <svg viewBox="0 0 504 104" preserveAspectRatio="none" aria-hidden="true">
            <defs>
              <linearGradient id="salesArea" x1="0" x2="0" y1="0" y2="1">
                <stop offset="0%" stop-color="currentColor" stop-opacity="0.22" />
                <stop offset="100%" stop-color="currentColor" stop-opacity="0.02" />
              </linearGradient>
            </defs>
            <polygon :points="`${chartPoints} 504,104 0,104`" fill="url(#salesArea)" />
            <polyline :points="chartPoints" fill="none" stroke="currentColor" stroke-width="3" />
            <circle cx="288" cy="24" r="5" fill="currentColor" />
          </svg>
          <div class="chart-grid-lines" aria-hidden="true"></div>
        </div>
        <div class="chart-axis">
          <span>05-16</span><span>05-17</span><span>05-18</span><span>05-19</span><span>05-20</span><span>05-21</span
          ><span>05-22</span>
        </div>
      </article>

      <article class="dashboard-card product-card">
        <header class="card-header">
          <h2>产品销售占比<span>本月</span></h2>
        </header>
        <div class="donut-layout">
          <div class="donut-chart">
            <div class="donut-center">
              <strong>￥1,856,320</strong>
              <span>月销售额</span>
            </div>
          </div>
          <ul class="share-list">
            <li v-for="item in productShares" :key="item.name">
              <span :style="{ background: item.color }"></span>
              <em>{{ item.name }}</em>
              <strong>{{ item.value }}</strong>
            </li>
          </ul>
        </div>
      </article>

      <article class="dashboard-card quick-card">
        <header class="card-header">
          <h2>快捷操作</h2>
        </header>
        <div class="quick-actions">
          <button type="button"><ShoppingCart :size="20" aria-hidden="true" />创建订单</button>
          <button type="button"><Box :size="20" aria-hidden="true" />新增套餐</button>
          <button type="button"><Users :size="20" aria-hidden="true" />客户管理</button>
          <button type="button"><TicketCheck :size="20" aria-hidden="true" />工单处理</button>
          <button type="button"><BadgeCheck :size="20" aria-hidden="true" />权限配置</button>
        </div>
      </article>
    </div>

    <div class="dashboard-bottom-grid">
      <article class="dashboard-card resource-card">
        <header class="card-header">
          <h2>资源使用与告警</h2>
          <a href="/audit">查看监控</a>
        </header>
        <div class="resource-list">
          <div v-for="item in resourceItems" :key="item.label" class="resource-item">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <div class="progress-track"><i :class="`tone-${item.tone}`"></i></div>
            <em :class="`tone-${item.tone}`">{{ item.status }}</em>
          </div>
        </div>
        <div class="alert-list">
          <div v-for="item in alerts" :key="item.text" class="alert-row">
            <span :class="`dot-${item.tone}`"></span>
            <strong>{{ item.level }}</strong>
            <p>{{ item.text }}</p>
            <time>{{ item.time }}</time>
          </div>
        </div>
      </article>

      <article class="dashboard-card table-card orders-card">
        <header class="card-header">
          <h2>最新订单</h2>
          <a href="/orders">查看订单</a>
        </header>
        <div class="table-scroll">
          <table>
            <thead>
              <tr>
                <th>订单编号</th>
                <th>客户名称</th>
                <th>产品规格</th>
                <th>节点</th>
                <th>金额</th>
                <th>状态</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="order in orders" :key="order[0]">
                <td><a href="/orders">{{ order[0] }}</a></td>
                <td>{{ order[1] }}</td>
                <td>{{ order[2] }}</td>
                <td>{{ order[3] }}</td>
                <td>{{ order[4] }}</td>
                <td><span class="status-pill">{{ order[5] }}</span></td>
                <td>{{ order[6] }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="dashboard-card table-card node-card">
        <header class="card-header">
          <h2>节点运行状态</h2>
          <a href="/instances">查看节点</a>
        </header>
        <div class="table-scroll">
          <table>
            <thead>
              <tr>
                <th>节点</th>
                <th>状态</th>
                <th>CPU</th>
                <th>内存</th>
                <th>实例数</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="node in nodes" :key="node[0]">
                <td><Server :size="17" aria-hidden="true" />{{ node[0] }}</td>
                <td><span class="node-status">{{ node[1] }}</span></td>
                <td>{{ node[2] }}</td>
                <td>{{ node[3] }}</td>
                <td>{{ node[4] }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </div>

    <button class="floating-refresh" type="button" title="刷新控制台" aria-label="刷新控制台" @click="loadDashboard">
      <RefreshCw :class="{ spinning: loading }" :size="18" aria-hidden="true" />
    </button>
  </section>
</template>

<style scoped>
.dashboard-page {
  position: relative;
  display: grid;
  gap: 18px;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(170px, 1fr));
  gap: 14px;
}

.stat-card {
  position: relative;
  min-height: 162px;
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr);
  grid-template-rows: 76px auto;
  gap: 8px 18px;
  overflow: hidden;
  padding: 28px 22px 18px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow);
}

.stat-icon {
  width: 62px;
  height: 62px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  color: #ffffff;
  background: var(--stat-color);
  box-shadow: 0 0 0 7px color-mix(in srgb, var(--stat-color) 18%, transparent);
}

.stat-main {
  min-width: 0;
  display: grid;
  align-content: center;
  gap: 8px;
}

.stat-main span,
.stat-footer span {
  color: var(--muted);
  font-weight: 700;
}

.stat-main strong {
  color: var(--text);
  font-size: 28px;
  line-height: 1;
}

.stat-footer {
  grid-column: 1 / 3;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--success);
}

.stat-footer strong {
  font-size: 15px;
}

.stat-sparkline {
  position: absolute;
  right: 18px;
  bottom: 18px;
  width: 92px;
  height: 42px;
  color: var(--stat-color);
}

.stat-card--blue {
  --stat-color: var(--primary);
}

.stat-card--green {
  --stat-color: var(--success);
}

.stat-card--purple {
  --stat-color: var(--purple);
}

.stat-card--orange {
  --stat-color: var(--orange);
}

.stat-card--red {
  --stat-color: var(--danger);
}

.stat-card--red .stat-footer {
  color: var(--success);
}

.dashboard-main-grid {
  display: grid;
  grid-template-columns: minmax(420px, 1.35fr) minmax(380px, 1.1fr) minmax(220px, 0.58fr);
  gap: 18px;
}

.dashboard-bottom-grid {
  display: grid;
  grid-template-columns: minmax(320px, 0.95fr) minmax(520px, 1.35fr) minmax(320px, 0.9fr);
  gap: 18px;
}

.dashboard-card {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-soft);
}

.card-header {
  min-height: 54px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 18px;
}

.card-header h2 {
  margin: 0;
  color: var(--text);
  font-size: 18px;
  line-height: 1.2;
}

.card-header h2 span {
  margin-left: 7px;
  color: var(--muted);
  font-size: 14px;
  font-weight: 500;
}

.card-header a,
.ghost-select {
  color: var(--primary);
  background: transparent;
  text-decoration: none;
  font-size: 13px;
  font-weight: 700;
}

.ghost-select {
  min-height: 36px;
  padding: 0 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  cursor: pointer;
}

.sales-card {
  min-height: 326px;
  padding-bottom: 20px;
}

.chart-legend {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 22px 8px;
  color: var(--muted);
  font-size: 14px;
  font-weight: 700;
}

.chart-legend span {
  width: 18px;
  height: 4px;
  border-radius: 999px;
  background: var(--primary);
}

.line-chart {
  position: relative;
  height: 178px;
  margin: 0 28px 0 22px;
  color: var(--primary);
}

.line-chart svg,
.chart-grid-lines {
  position: absolute;
  inset: 0;
}

.chart-grid-lines {
  background:
    linear-gradient(to bottom, transparent calc(25% - 1px), var(--border) 25%, transparent calc(25% + 1px)),
    linear-gradient(to bottom, transparent calc(50% - 1px), var(--border) 50%, transparent calc(50% + 1px)),
    linear-gradient(to bottom, transparent calc(75% - 1px), var(--border) 75%, transparent calc(75% + 1px));
  opacity: 0.78;
}

.chart-axis {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 8px;
  padding: 0 24px;
  color: var(--muted);
  font-size: 14px;
}

.product-card {
  min-height: 326px;
}

.donut-layout {
  display: grid;
  grid-template-columns: 238px minmax(0, 1fr);
  align-items: center;
  gap: 20px;
  padding: 20px 24px 30px;
}

.donut-chart {
  position: relative;
  width: 218px;
  aspect-ratio: 1;
  border-radius: 50%;
  background: conic-gradient(
    #2f7cf6 0 38.7%,
    #2fc28b 38.7% 61.8%,
    #f59e0b 61.8% 78.6%,
    #ff4f5c 78.6% 90%,
    #7c8aa4 90% 100%
  );
}

.donut-chart::after {
  position: absolute;
  inset: 48px;
  border-radius: 50%;
  background: var(--panel);
  content: "";
}

.donut-center {
  position: absolute;
  inset: 58px;
  z-index: 1;
  display: grid;
  place-content: center;
  text-align: center;
}

.donut-center strong {
  font-size: 22px;
}

.donut-center span {
  margin-top: 8px;
  color: var(--muted);
  font-weight: 700;
}

.share-list {
  display: grid;
  gap: 18px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.share-list li {
  display: grid;
  grid-template-columns: 10px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  color: var(--muted-strong);
}

.share-list li span {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.share-list em {
  min-width: 0;
  overflow: hidden;
  font-style: normal;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quick-card {
  min-height: 326px;
}

.quick-actions {
  display: grid;
  gap: 12px;
  padding: 10px 18px 18px;
}

.quick-actions button {
  min-height: 42px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 18px;
  border: 1px solid color-mix(in srgb, currentColor 18%, var(--border));
  border-radius: 6px;
  color: var(--primary);
  background: color-mix(in srgb, currentColor 8%, var(--panel));
  cursor: pointer;
  font-weight: 800;
}

.quick-actions button:nth-child(2) {
  color: var(--success);
}

.quick-actions button:nth-child(3) {
  color: var(--purple);
}

.quick-actions button:nth-child(4) {
  color: var(--orange);
}

.quick-actions button:nth-child(5) {
  color: var(--primary);
}

.resource-card {
  min-height: 354px;
}

.resource-list {
  display: grid;
  grid-template-columns: repeat(5, minmax(74px, 1fr));
  border-top: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
}

.resource-item {
  min-width: 0;
  display: grid;
  gap: 8px;
  padding: 18px 12px;
  border-right: 1px solid var(--border);
}

.resource-item:last-child {
  border-right: 0;
}

.resource-item span {
  color: var(--muted);
  font-size: 13px;
  font-weight: 700;
}

.resource-item strong {
  font-size: 18px;
}

.progress-track {
  height: 7px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--panel-strong);
}

.progress-track i {
  display: block;
  width: 72%;
  height: 100%;
  border-radius: inherit;
  background: currentColor;
}

.resource-item em {
  font-size: 13px;
  font-style: normal;
  font-weight: 800;
}

.tone-green {
  color: var(--success);
}

.tone-orange {
  color: var(--orange);
}

.alert-list {
  display: grid;
  padding: 12px 18px 18px;
}

.alert-row {
  min-height: 40px;
  display: grid;
  grid-template-columns: 9px 42px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  color: var(--muted-strong);
  font-size: 13px;
}

.alert-row span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-red {
  background: var(--danger);
}

.dot-orange {
  background: var(--orange);
}

.dot-yellow {
  background: #f7c948;
}

.dot-blue {
  background: var(--primary);
}

.alert-row strong {
  font-weight: 800;
}

.alert-row p {
  min-width: 0;
  overflow: hidden;
  margin: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.alert-row time {
  color: var(--muted);
}

.table-card {
  min-height: 354px;
}

.table-scroll {
  overflow-x: auto;
  padding: 0 14px 16px;
}

table {
  width: 100%;
  min-width: 620px;
  border-collapse: collapse;
  color: var(--muted-strong);
  font-size: 13px;
}

.node-card table {
  min-width: 420px;
}

th {
  height: 44px;
  color: #53627a;
  background: var(--panel-soft);
  font-weight: 800;
  text-align: left;
}

td {
  height: 50px;
  border-bottom: 1px solid var(--border);
}

th,
td {
  padding: 0 10px;
  white-space: nowrap;
}

td a {
  color: var(--primary);
  font-weight: 800;
  text-decoration: none;
}

.status-pill,
.node-status {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: 6px;
  color: var(--success);
  background: var(--success-soft);
  font-weight: 800;
}

.node-status::before {
  width: 8px;
  height: 8px;
  margin-right: 7px;
  border-radius: 50%;
  background: currentColor;
  content: "";
}

.node-card td:first-child {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #2e4262;
  font-weight: 800;
}

.floating-refresh {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 44px;
  height: 44px;
  display: grid;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: 50%;
  color: var(--primary);
  background: var(--panel);
  box-shadow: var(--shadow);
  cursor: pointer;
}

.spinning {
  animation: spin 800ms linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1460px) {
  .stat-grid {
    grid-template-columns: repeat(3, minmax(180px, 1fr));
  }

  .dashboard-main-grid,
  .dashboard-bottom-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .stat-grid {
    grid-template-columns: 1fr;
  }

  .dashboard-main-grid,
  .dashboard-bottom-grid {
    gap: 14px;
  }

  .donut-layout {
    grid-template-columns: 1fr;
    justify-items: center;
  }

  .resource-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .resource-item {
    border-bottom: 1px solid var(--border);
  }

  .alert-row {
    grid-template-columns: 9px 42px minmax(0, 1fr);
  }

  .alert-row time {
    display: none;
  }
}
</style>
