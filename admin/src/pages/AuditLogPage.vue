<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { AlertTriangle, ClipboardList, ShieldAlert } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs, getRiskLogs } from '../api/audit'
import AdminEmptyState from '../components/AdminEmptyState.vue'
import AdminPageHeader from '../components/AdminPageHeader.vue'
import AdminTablePanel from '../components/AdminTablePanel.vue'
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
const first = computed(() => (page.value - 1) * perPage)
const emptyText = computed(() => (isRiskPage.value ? '暂无高危操作日志' : '暂无审计日志'))
const riskLevelOptions = [
  { label: '全部风险', value: '' },
  { label: '中风险', value: 'medium' },
  { label: '高风险', value: 'high' },
  { label: '严重风险', value: 'critical' },
]

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

async function changePaginator(event: { first: number; page: number }) {
  page.value = event.page + 1
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

function riskSeverity(level?: string) {
  return level === 'medium' ? 'warn' : 'danger'
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
    <AdminPageHeader eyebrow="后台安全" :title="title" :icon="icon">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText v-model="filters.action" type="search" placeholder="操作动作" @keyup.enter="applyFilters" />
        </IconField>
        <IconField>
          <InputIcon class="pi pi-list" />
          <InputText v-model="filters.objectType" type="search" placeholder="对象类型" @keyup.enter="applyFilters" />
        </IconField>
        <Select
          v-if="isRiskPage"
          v-model="filters.riskLevel"
          :options="riskLevelOptions"
          option-label="label"
          option-value="value"
          aria-label="风险等级"
        />
        <Button type="button" label="查询" icon="pi pi-search" @click="applyFilters" />
        <Button type="button" label="刷新" icon="pi pi-refresh" :loading="loading" severity="secondary" outlined @click="loadLogs" />
    </AdminPageHeader>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

    <AdminTablePanel>
      <AdminEmptyState v-if="loading && rows.length === 0" text="日志加载中..." />
      <AdminEmptyState v-else-if="rows.length === 0" :text="emptyText" :icon="AlertTriangle" />
      <DataTable v-else :value="rows" class="admin-prime-table" data-key="id" striped-rows>
        <Column header="时间">
          <template #body="{ data }">{{ formatDate(data.created_at) }}</template>
        </Column>
        <Column header="管理员">
          <template #body="{ data }">{{ data.admin?.display_name || data.admin?.username || '系统' }}</template>
        </Column>
        <Column v-if="isRiskPage" header="风险">
          <template #body="{ data }">
            <Tag :value="riskLabel((data as RiskLogItem).risk_level)" :severity="riskSeverity((data as RiskLogItem).risk_level)" />
          </template>
        </Column>
        <Column field="action" header="动作" />
        <Column header="对象">
          <template #body="{ data }">{{ data.object_type }}<small v-if="data.object_id">#{{ data.object_id }}</small></template>
        </Column>
        <Column header="IP">
          <template #body="{ data }">{{ data.ip || '-' }}</template>
        </Column>
        <Column :header="isRiskPage ? '风险原因' : '备注'">
          <template #body="{ data }">{{ isRiskPage ? (data as RiskLogItem).risk_reason : data.remark || '-' }}</template>
        </Column>
      </DataTable>
      <template #footer>
        <span>共 {{ total }} 条</span>
        <Paginator :first="first" :rows="perPage" :total-records="total" template="PrevPageLink PageLinks NextPageLink" @page="changePaginator" />
      </template>
    </AdminTablePanel>
  </section>
</template>

<style scoped>
.audit-page {
  display: grid;
  gap: 14px;
}

td small {
  margin-left: 4px;
  color: var(--muted);
}
</style>
