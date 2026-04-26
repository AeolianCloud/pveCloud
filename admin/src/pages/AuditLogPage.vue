<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { AlertTriangle, ClipboardList, RefreshCw, Search, ShieldAlert } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs, getRiskLogs } from '../api/audit'
import type { AuditLogItem, RiskLogItem } from '../types/audit'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const errorMessage = ref('')
const auditRows = ref<AuditLogItem[]>([])
const riskRows = ref<RiskLogItem[]>([])
const total = ref(0)
const page = ref(Number(route.query.page) || 1)
const perPage = 15

const filters = reactive({
  action: typeof route.query.action === 'string' ? route.query.action : '',
  objectType: typeof route.query.object_type === 'string' ? route.query.object_type : '',
  riskLevel: typeof route.query.risk_level === 'string' ? route.query.risk_level : '',
})

const isRiskPage = computed(() => route.name === 'risk_logs')
const title = computed(() => (isRiskPage.value ? '高危操作日志' : '审计日志'))
const icon = computed(() => (isRiskPage.value ? ShieldAlert : ClipboardList))
const rows = computed(() => (isRiskPage.value ? riskRows.value : auditRows.value))
const lastPage = computed(() => Math.max(1, Math.ceil(total.value / perPage)))
const emptyText = computed(() => (isRiskPage.value ? '暂无高危操作日志' : '暂无审计日志'))

async function loadLogs() {
  loading.value = true
  errorMessage.value = ''
  try {
    const params = {
      page: page.value,
      per_page: perPage,
      action: filters.action || undefined,
      object_type: filters.objectType || undefined,
    }
    if (isRiskPage.value) {
      const result = await getRiskLogs({
        ...params,
        risk_level: filters.riskLevel as 'medium' | 'high' | 'critical' | '',
      })
      riskRows.value = result.list
      total.value = result.total
    } else {
      const result = await getAuditLogs(params)
      auditRows.value = result.list
      total.value = result.total
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '日志加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

async function applyFilters() {
  page.value = 1
  await syncQuery()
  await loadLogs()
}

async function changePage(nextPage: number) {
  page.value = Math.min(Math.max(1, nextPage), lastPage.value)
  await syncQuery()
  await loadLogs()
}

async function syncQuery() {
  await router.replace({
    query: {
      ...route.query,
      page: String(page.value),
      action: filters.action || undefined,
      object_type: filters.objectType || undefined,
      risk_level: isRiskPage.value && filters.riskLevel ? filters.riskLevel : undefined,
    },
  })
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

function riskLabel(level?: string) {
  if (level === 'critical') {
    return '严重'
  }
  if (level === 'high') {
    return '高'
  }
  return '中'
}

function riskClass(level?: string) {
  return `risk-${level || 'medium'}`
}

watch(
  () => route.name,
  () => {
    page.value = 1
    auditRows.value = []
    riskRows.value = []
    loadLogs()
  },
)

onMounted(loadLogs)
</script>

<template>
  <section class="audit-page">
    <header class="audit-toolbar">
      <div class="audit-title">
        <span class="audit-title-icon">
          <component :is="icon" :size="20" aria-hidden="true" />
        </span>
        <div>
          <span>后台安全</span>
          <h1>{{ title }}</h1>
        </div>
      </div>
      <div class="audit-filters">
        <label>
          <Search :size="15" aria-hidden="true" />
          <input v-model="filters.action" type="search" placeholder="操作动作" @keyup.enter="applyFilters" />
        </label>
        <label>
          <ClipboardList :size="15" aria-hidden="true" />
          <input v-model="filters.objectType" type="search" placeholder="对象类型" @keyup.enter="applyFilters" />
        </label>
        <select v-if="isRiskPage" v-model="filters.riskLevel" aria-label="风险等级">
          <option value="">全部风险</option>
          <option value="medium">中风险</option>
          <option value="high">高风险</option>
          <option value="critical">严重风险</option>
        </select>
        <button type="button" @click="applyFilters">
          <Search :size="15" aria-hidden="true" />
          查询
        </button>
        <button type="button" :disabled="loading" title="刷新日志" aria-label="刷新日志" @click="loadLogs">
          <RefreshCw :class="{ spinning: loading }" :size="15" aria-hidden="true" />
          刷新
        </button>
      </div>
    </header>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <article class="audit-table-card">
      <div v-if="loading && rows.length === 0" class="audit-state">日志加载中...</div>
      <div v-else-if="rows.length === 0" class="audit-state">
        <AlertTriangle :size="18" aria-hidden="true" />
        {{ emptyText }}
      </div>
      <div v-else class="audit-table-scroll">
        <table>
          <thead>
            <tr>
              <th>时间</th>
              <th>管理员</th>
              <th v-if="isRiskPage">风险</th>
              <th>动作</th>
              <th>对象</th>
              <th>IP</th>
              <th>{{ isRiskPage ? '风险原因' : '备注' }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>{{ formatDate(row.created_at) }}</td>
              <td>{{ row.admin?.display_name || row.admin?.username || '系统' }}</td>
              <td v-if="isRiskPage">
                <span class="risk-pill" :class="riskClass((row as RiskLogItem).risk_level)">
                  {{ riskLabel((row as RiskLogItem).risk_level) }}
                </span>
              </td>
              <td>{{ row.action }}</td>
              <td>{{ row.object_type }}<small v-if="row.object_id">#{{ row.object_id }}</small></td>
              <td>{{ row.ip || '-' }}</td>
              <td>{{ isRiskPage ? (row as RiskLogItem).risk_reason : row.remark || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="audit-pagination">
        <span>共 {{ total }} 条</span>
        <div>
          <button type="button" :disabled="page <= 1 || loading" @click="changePage(page - 1)">上一页</button>
          <strong>{{ page }} / {{ lastPage }}</strong>
          <button type="button" :disabled="page >= lastPage || loading" @click="changePage(page + 1)">下一页</button>
        </div>
      </footer>
    </article>
  </section>
</template>

<style scoped>
.audit-page {
  display: grid;
  gap: 14px;
}

.audit-toolbar,
.audit-table-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-soft);
}

.audit-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px;
}

.audit-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.audit-title-icon {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: var(--primary);
  background: var(--primary-soft);
}

.audit-title span {
  color: var(--muted);
  font-size: 13px;
  font-weight: 800;
}

.audit-title h1 {
  margin: 3px 0 0;
  font-size: 20px;
  line-height: 1.2;
}

.audit-filters {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.audit-filters label,
.audit-filters button,
.audit-filters select {
  min-height: 34px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 0 11px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  background: var(--panel);
  font-size: 13px;
  font-weight: 750;
}

.audit-filters button,
.audit-filters select {
  cursor: pointer;
}

.audit-filters button:disabled {
  cursor: wait;
  opacity: 0.58;
}

.audit-filters input {
  width: 150px;
  border: 0;
  outline: 0;
  color: var(--text);
  background: transparent;
}

.audit-table-card {
  overflow: hidden;
}

.audit-table-scroll {
  overflow-x: auto;
  padding: 0 10px;
}

table {
  width: 100%;
  min-width: 900px;
  border-collapse: collapse;
  color: var(--muted-strong);
  font-size: 13px;
}

th,
td {
  height: 42px;
  padding: 0 9px;
  border-bottom: 1px solid var(--border);
  text-align: left;
  white-space: nowrap;
}

th {
  height: 38px;
  color: var(--table-head-text);
  background: var(--panel-soft);
  font-weight: 800;
}

td small {
  margin-left: 4px;
  color: var(--muted);
}

.risk-pill {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 8px;
  border-radius: 6px;
  font-weight: 850;
}

.risk-medium {
  color: var(--warning);
  background: var(--warning-soft);
}

.risk-high,
.risk-critical {
  color: var(--danger);
  background: var(--danger-soft);
}

.audit-state {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--muted);
  font-weight: 800;
}

.audit-pagination {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  color: var(--muted);
}

.audit-pagination div {
  display: flex;
  align-items: center;
  gap: 10px;
}

.audit-pagination button {
  min-height: 30px;
  padding: 0 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  background: var(--panel);
  cursor: pointer;
}

.audit-pagination button:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.spinning {
  animation: spin 800ms linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 960px) {
  .audit-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .audit-filters {
    justify-content: flex-start;
  }
}
</style>
