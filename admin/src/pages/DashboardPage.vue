<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CircleDollarSign, ClipboardCheck, Layers3, Server, TicketCheck, Users } from 'lucide-vue-next'

import { getAdminDashboard } from '../api/dashboard'
import { useAuthStore } from '../stores/auth'
import type { DashboardMetric } from '../types/dashboard'

const auth = useAuthStore()
const loading = ref(false)
const errorMessage = ref('')
const metrics = ref<DashboardMetric[]>([])

const fallbackStats = [
  { key: 'sales', title: '今日销售额', value: '58,960 元', trend: '+23.6%', tone: 'blue', icon: CircleDollarSign },
  { key: 'orders', title: '今日订单数', value: '128', trend: '+18.4%', tone: 'green', icon: ClipboardCheck },
  { key: 'customers', title: '新增客户数', value: '1,356', trend: '+12.7%', tone: 'purple', icon: Users },
  { key: 'instances', title: '运行中实例', value: '3,692', trend: '+8.3%', tone: 'orange', icon: Layers3 },
  { key: 'tickets', title: '待处理工单', value: '32', trend: '-5.9%', tone: 'red', icon: TicketCheck },
]

const productShares = [
  { name: '标准型云服务器', value: '38.7%', color: 'var(--chart-blue)' },
  { name: '计算型云服务器', value: '23.1%', color: 'var(--chart-green)' },
  { name: '内存型云服务器', value: '16.8%', color: 'var(--warning)' },
  { name: 'GPU 云服务器', value: '11.4%', color: 'var(--chart-red)' },
  { name: '其他套餐', value: '10.0%', color: 'var(--chart-muted)' },
]

const resourceItems = [
  { label: 'CPU 使用率', value: '42%', progress: 42, status: '正常', tone: 'success' },
  { label: '内存使用率', value: '55%', progress: 55, status: '正常', tone: 'success' },
  { label: '带宽使用率', value: '68%', progress: 68, status: '关注', tone: 'warn' },
  { label: 'IP 可用数', value: '1,256 个', progress: 82, status: '充足', tone: 'success' },
  { label: '服务 SLA', value: '98.6%', progress: 98.6, status: '稳定', tone: 'info' },
]

const alerts = [
  { level: '严重', text: '节点 1 CPU 使用率超过 90%', time: '05-22 10:15', severity: 'danger' },
  { level: '警告', text: '节点 2 带宽使用率超过 80%', time: '05-22 09:48', severity: 'warn' },
  { level: '提示', text: '公网 IP 可用数量低于预警阈值', time: '05-22 09:20', severity: 'info' },
  { level: '提示', text: '华南节点正在进行计划维护', time: '05-22 08:55', severity: 'secondary' },
]

const orders = [
  { id: 'DD202505220001', customer: '上海智云科技有限公司', product: '4核 8G 5M', node: '上海节点 1', amount: '560.00 元', status: '已支付', createdAt: '05-22 10:23' },
  { id: 'DD202505220002', customer: '杭州数海信息技术有限公司', product: '2核 4G 3M', node: '杭州节点 2', amount: '260.00 元', status: '已支付', createdAt: '05-22 10:18' },
  { id: 'DD202505220003', customer: '深圳云拓网络有限公司', product: '8核 16G 10M', node: '深圳节点 1', amount: '1,360.00 元', status: '待开通', createdAt: '05-22 10:05' },
  { id: 'DD202505220004', customer: '广州星河智能科技', product: 'GPU 1卡 24G', node: '上海节点 1', amount: '2,800.00 元', status: '待支付', createdAt: '05-22 09:58' },
  { id: 'DD202505220005', customer: '北京启航数据服务', product: '2核 4G 5M', node: '北京节点 1', amount: '260.00 元', status: '已支付', createdAt: '05-22 09:47' },
]

const nodes = [
  { name: '上海节点 1', status: '在线', cpu: '42%', memory: '53%', instances: 892 },
  { name: '杭州节点 2', status: '在线', cpu: '38%', memory: '49%', instances: 756 },
  { name: '深圳节点 1', status: '在线', cpu: '45%', memory: '57%', instances: 684 },
  { name: '北京节点 1', status: '在线', cpu: '36%', memory: '43%', instances: 512 },
  { name: '华南节点', status: '维护中', cpu: '41%', memory: '52%', instances: 848 },
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
      value: `${metric.value.toLocaleString()}${metric.unit ? ` ${metric.unit}` : ''}`,
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

function statToneClass(tone: string) {
  return `stat-card--${tone}`
}

function resourceSeverity(tone: string) {
  if (tone === 'warn') {
    return 'warn'
  }
  if (tone === 'info') {
    return 'info'
  }
  return 'success'
}

function orderSeverity(status: string) {
  if (status === '待支付') {
    return 'warn'
  }
  if (status === '待开通') {
    return 'info'
  }
  return 'success'
}

function nodeSeverity(status: string) {
  return status === '维护中' ? 'warn' : 'success'
}

onMounted(loadDashboard)
</script>

<template>
  <section class="dashboard-page">
    <div class="stat-grid">
      <Card v-for="stat in displayStats" :key="stat.key" class="stat-card" :class="statToneClass(stat.tone)">
        <template #content>
          <span class="stat-icon">
            <component :is="stat.icon" :size="22" aria-hidden="true" />
          </span>
          <div class="stat-main">
            <span>{{ stat.title }}</span>
            <strong>{{ stat.value }}</strong>
          </div>
          <div class="stat-footer">
            <span>较昨日</span>
            <strong>{{ stat.trend }}</strong>
          </div>
        </template>
      </Card>
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

    <div class="dashboard-main-grid">
      <Card class="dashboard-card sales-card">
        <template #title>
          <div class="card-header">
            <h2>销售趋势 <span>近 7 日</span></h2>
            <Button label="近 7 天" severity="secondary" text />
          </div>
        </template>
        <template #content>
          <div class="chart-legend"><span></span>销售额</div>
          <div class="line-chart">
            <div class="line-chart-track"></div>
            <div class="line-chart-surface">
              <i class="line-1"></i>
              <i class="line-2"></i>
              <i class="line-3"></i>
              <i class="line-4"></i>
              <i class="line-5"></i>
              <i class="line-6"></i>
              <i class="line-7"></i>
            </div>
          </div>
          <div class="chart-axis">
            <span>05-16</span>
            <span>05-17</span>
            <span>05-18</span>
            <span>05-19</span>
            <span>05-20</span>
            <span>05-21</span>
            <span>05-22</span>
          </div>
        </template>
      </Card>

      <Card class="dashboard-card product-card">
        <template #title>
          <div class="card-header">
            <h2>产品销售占比 <span>本月</span></h2>
          </div>
        </template>
        <template #content>
          <div class="donut-layout">
            <div class="donut-chart">
              <div class="donut-center">
                <strong>2,856,320 元</strong>
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
        </template>
      </Card>
    </div>

    <div class="dashboard-bottom-grid">
      <Card class="dashboard-card resource-card">
        <template #title>
          <div class="card-header">
            <h2>资源使用与告警</h2>
          </div>
        </template>
        <template #content>
          <div class="resource-list">
            <div v-for="item in resourceItems" :key="item.label" class="resource-item">
              <div class="resource-item-head">
                <span>{{ item.label }}</span>
                <Tag :value="item.status" :severity="resourceSeverity(item.tone)" />
              </div>
              <strong>{{ item.value }}</strong>
              <ProgressBar :value="item.progress" :show-value="false" />
            </div>
          </div>
          <div class="alert-list">
            <div v-for="item in alerts" :key="item.text" class="alert-row">
              <Tag :value="item.level" :severity="item.severity" />
              <p>{{ item.text }}</p>
              <time>{{ item.time }}</time>
            </div>
          </div>
        </template>
      </Card>

      <Card class="dashboard-card table-card orders-card">
        <template #title>
          <div class="card-header">
            <h2>最新订单</h2>
          </div>
        </template>
        <template #content>
          <DataTable :value="orders" class="admin-prime-table dashboard-table" striped-rows>
            <Column field="id" header="订单编号" />
            <Column field="customer" header="客户名称" />
            <Column field="product" header="产品规格" />
            <Column field="node" header="节点" />
            <Column field="amount" header="金额" />
            <Column header="状态">
              <template #body="{ data }">
                <Tag :value="data.status" :severity="orderSeverity(data.status)" />
              </template>
            </Column>
            <Column field="createdAt" header="创建时间" />
          </DataTable>
        </template>
      </Card>

      <Card class="dashboard-card table-card node-card">
        <template #title>
          <div class="card-header">
            <h2>节点运行状态</h2>
          </div>
        </template>
        <template #content>
          <DataTable :value="nodes" class="admin-prime-table dashboard-table" striped-rows>
            <Column header="节点">
              <template #body="{ data }">
                <span class="node-title"><Server :size="15" aria-hidden="true" />{{ data.name }}</span>
              </template>
            </Column>
            <Column header="状态">
              <template #body="{ data }">
                <Tag :value="data.status" :severity="nodeSeverity(data.status)" />
              </template>
            </Column>
            <Column field="cpu" header="CPU" />
            <Column field="memory" header="内存" />
            <Column field="instances" header="实例数" />
          </DataTable>
        </template>
      </Card>
    </div>

    <Button class="floating-refresh" icon="pi pi-refresh" rounded :loading="loading" aria-label="刷新控制台" @click="loadDashboard" />
  </section>
</template>

<style scoped>
.dashboard-page {
  position: relative;
  display: grid;
  gap: 14px;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(150px, 1fr));
  gap: 10px;
}

.stat-card {
  overflow: hidden;
}

.stat-card :deep(.p-card-content) {
  position: relative;
  min-height: 124px;
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  grid-template-rows: 58px auto;
  gap: 6px 12px;
  padding: 18px 16px 14px;
}

.stat-icon {
  width: 46px;
  height: 46px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  color: var(--on-primary);
  background: var(--stat-color);
  box-shadow: 0 0 0 5px color-mix(in srgb, var(--stat-color) 18%, transparent);
}

.stat-main {
  min-width: 0;
  display: grid;
  align-content: center;
  gap: 5px;
}

.stat-main span,
.stat-footer span {
  color: var(--muted);
  font-weight: 700;
}

.stat-main strong {
  color: var(--text);
  font-size: 22px;
  line-height: 1;
}

.stat-footer {
  grid-column: 1 / 3;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--success);
  font-size: 12px;
}

.stat-footer strong {
  font-size: 13px;
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

.dashboard-main-grid {
  display: grid;
  grid-template-columns: minmax(420px, 1.35fr) minmax(380px, 1.1fr);
  gap: 14px;
}

.dashboard-bottom-grid {
  display: grid;
  grid-template-columns: minmax(320px, 0.95fr) minmax(520px, 1.35fr) minmax(320px, 0.9fr);
  gap: 14px;
}

.dashboard-card {
  min-width: 0;
  overflow: hidden;
}

.dashboard-card :deep(.p-card-body) {
  gap: 0;
}

.card-header {
  min-height: 46px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card-header h2 {
  margin: 0;
  color: var(--text);
  font-size: 15px;
  line-height: 1.2;
}

.card-header h2 span {
  margin-left: 7px;
  color: var(--muted);
  font-size: 12px;
  font-weight: 500;
}

.sales-card {
  min-height: 266px;
}

.chart-legend {
  display: flex;
  align-items: center;
  gap: 7px;
  padding-bottom: 10px;
  color: var(--muted);
  font-size: 12px;
  font-weight: 700;
}

.chart-legend span {
  width: 18px;
  height: 4px;
  border-radius: 999px;
  background: var(--primary);
}

.line-chart {
  padding: 14px 0 8px;
}

.line-chart-track {
  height: 150px;
  border-radius: 12px;
  background:
    linear-gradient(to bottom, transparent calc(25% - 1px), var(--border) 25%, transparent calc(25% + 1px)),
    linear-gradient(to bottom, transparent calc(50% - 1px), var(--border) 50%, transparent calc(50% + 1px)),
    linear-gradient(to bottom, transparent calc(75% - 1px), var(--border) 75%, transparent calc(75% + 1px));
}

.line-chart-surface {
  position: relative;
  margin-top: -128px;
  height: 116px;
}

.line-chart-surface i {
  position: absolute;
  width: 58px;
  border-top: 3px solid var(--primary);
  transform-origin: left center;
}

.line-1 { top: 64px; left: 0; transform: rotate(4deg); }
.line-2 { top: 66px; left: 58px; transform: rotate(7deg); }
.line-3 { top: 72px; left: 116px; transform: rotate(12deg); }
.line-4 { top: 84px; left: 174px; transform: rotate(-33deg); }
.line-5 { top: 52px; left: 232px; transform: rotate(18deg); }
.line-6 { top: 70px; left: 290px; transform: rotate(23deg); }
.line-7 { top: 92px; left: 348px; transform: rotate(-18deg); }

.chart-axis {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 6px;
  color: var(--muted);
  font-size: 12px;
}

.product-card {
  min-height: 266px;
}

.donut-layout {
  display: grid;
  grid-template-columns: 188px minmax(0, 1fr);
  align-items: center;
  gap: 16px;
  padding-top: 8px;
}

.donut-chart {
  position: relative;
  width: 170px;
  aspect-ratio: 1;
  border-radius: 50%;
  background: conic-gradient(
    var(--chart-blue) 0 38.7%,
    var(--chart-green) 38.7% 61.8%,
    var(--warning) 61.8% 78.6%,
    var(--chart-red) 78.6% 90%,
    var(--chart-muted) 90% 100%
  );
}

.donut-chart::after {
  position: absolute;
  inset: 38px;
  border-radius: 50%;
  background: var(--panel);
  content: "";
}

.donut-center {
  position: absolute;
  inset: 46px;
  z-index: 1;
  display: grid;
  place-content: center;
  text-align: center;
}

.donut-center strong {
  font-size: 17px;
}

.donut-center span {
  margin-top: 6px;
  color: var(--muted);
  font-weight: 700;
}

.share-list {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.share-list li {
  display: grid;
  grid-template-columns: 10px minmax(0, 1fr) auto;
  align-items: center;
  gap: 9px;
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

.resource-card {
  min-height: 298px;
}

.resource-list {
  display: grid;
  gap: 12px;
  padding-top: 8px;
}

.resource-item {
  display: grid;
  gap: 8px;
}

.resource-item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.resource-item span {
  color: var(--muted);
  font-size: 13px;
  font-weight: 700;
}

.resource-item strong {
  color: var(--text);
  font-size: 15px;
}

.alert-list {
  display: grid;
  gap: 10px;
  margin-top: 18px;
}

.alert-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  color: var(--muted-strong);
  font-size: 13px;
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

.dashboard-table :deep(.p-datatable-table) {
  min-width: 620px;
}

.node-card .dashboard-table :deep(.p-datatable-table) {
  min-width: 420px;
}

.node-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--node-title);
  font-weight: 800;
}

.floating-refresh {
  position: fixed;
  right: 18px;
  bottom: 18px;
  box-shadow: var(--shadow);
}

@media (max-width: 1460px) {
  .stat-grid {
    grid-template-columns: repeat(3, minmax(160px, 1fr));
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

  .donut-layout {
    grid-template-columns: 1fr;
    justify-items: center;
  }

  .alert-row {
    grid-template-columns: 1fr;
    justify-items: start;
  }
}
</style>
