<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { MonitorCheck, ShieldAlert } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getAdminSessions, revokeAdminSession } from '../api/adminSession'
import AdminEmptyState from '../components/AdminEmptyState.vue'
import AdminPageHeader from '../components/AdminPageHeader.vue'
import AdminTablePanel from '../components/AdminTablePanel.vue'
import type { AdminSessionItem, AdminSessionStatus } from '../types/adminSession'
import { useConfirmAction } from '../utils/confirmAction'

const route = useRoute()
const router = useRouter()
const confirmAction = useConfirmAction()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const rows = ref<AdminSessionItem[]>([])
const total = ref(0)
const page = ref(Number(route.query.page) || 1)
const perPage = 15

const filters = reactive({
  keyword: typeof route.query.keyword === 'string' ? route.query.keyword : '',
  status: typeof route.query.status === 'string' ? (route.query.status as AdminSessionStatus) : '',
})

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '活跃', value: 'active' },
  { label: '已吊销', value: 'revoked' },
  { label: '已过期', value: 'expired' },
]

const first = computed(() => (page.value - 1) * perPage)

async function loadSessions() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await getAdminSessions({
      page: page.value,
      per_page: perPage,
      keyword: filters.keyword || undefined,
      status: filters.status ? (filters.status as AdminSessionStatus) : undefined,
    })
    rows.value = result.list
    total.value = result.total
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '会话加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

async function syncQuery() {
  await router.replace({
    query: {
      ...route.query,
      page: String(page.value),
      keyword: filters.keyword || undefined,
      status: filters.status || undefined,
    },
  })
}

async function applyFilters() {
  page.value = 1
  await syncQuery()
  await loadSessions()
}

async function changePaginator(event: { page: number }) {
  page.value = event.page + 1
  await syncQuery()
  await loadSessions()
}

async function revoke(row: AdminSessionItem) {
  const confirmed = await confirmAction({
    header: '吊销会话',
    message: `确认吊销 ${row.admin?.display_name || row.admin?.username || row.session_id} 的会话吗`,
    acceptLabel: '吊销',
  })
  if (!confirmed) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    await revokeAdminSession(row.id)
    await loadSessions()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '会话吊销失败，请稍后重试'
  } finally {
    submitting.value = false
  }
}

function formatDate(value: string | null) {
  if (!value) {
    return '-'
  }
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

function statusLabel(status: AdminSessionStatus) {
  if (status === 'active') return '活跃'
  if (status === 'revoked') return '已吊销'
  return '已过期'
}

function statusSeverity(status: AdminSessionStatus) {
  return status === 'active' ? 'success' : 'danger'
}

onMounted(loadSessions)
</script>

<template>
  <section class="session-page">
    <AdminPageHeader eyebrow="访问控制" title="登录会话" :icon="MonitorCheck">
      <IconField>
        <InputIcon class="pi pi-search" />
        <InputText v-model="filters.keyword" type="search" placeholder="管理员 / IP / 会话" @keyup.enter="applyFilters" />
      </IconField>
      <Select v-model="filters.status" :options="statusOptions" option-label="label" option-value="value" aria-label="会话状态" />
      <Button type="button" label="查询" icon="pi pi-search" @click="applyFilters" />
      <Button type="button" icon="pi pi-refresh" :loading="loading" severity="secondary" outlined aria-label="刷新" @click="loadSessions" />
    </AdminPageHeader>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

    <AdminTablePanel>
      <AdminEmptyState v-if="loading && rows.length === 0" text="会话加载中..." />
      <AdminEmptyState v-else-if="rows.length === 0" text="暂无登录会话" :icon="ShieldAlert" />
      <DataTable v-else :value="rows" class="admin-prime-table" data-key="id" striped-rows>
        <Column header="管理员">
          <template #body="{ data }">
            <strong>{{ data.admin?.display_name || data.admin?.username || '未知管理员' }}</strong>
            <small>{{ data.session_id }}</small>
          </template>
        </Column>
        <Column header="状态">
          <template #body="{ data }">
            <Tag :value="statusLabel(data.status)" :severity="statusSeverity(data.status)" />
          </template>
        </Column>
        <Column header="最后访问">
          <template #body="{ data }">
            {{ formatDate(data.last_seen_at) }}
            <small>{{ data.last_seen_ip || '-' }}</small>
          </template>
        </Column>
        <Column header="签发 / 过期">
          <template #body="{ data }">
            {{ formatDate(data.issued_at) }}
            <small>{{ formatDate(data.expires_at) }}</small>
          </template>
        </Column>
        <Column header="吊销">
          <template #body="{ data }">
            {{ formatDate(data.revoked_at) }}
            <small>{{ data.revoke_reason || '' }}</small>
          </template>
        </Column>
        <Column header="操作">
          <template #body="{ data }">
            <Button label="吊销" severity="danger" text :disabled="data.status !== 'active' || submitting" @click="revoke(data)" />
          </template>
        </Column>
      </DataTable>
      <template #footer>
        <span>共 {{ total }} 条会话</span>
        <Paginator :first="first" :rows="perPage" :total-records="total" template="PrevPageLink PageLinks NextPageLink" @page="changePaginator" />
      </template>
    </AdminTablePanel>
  </section>
</template>

<style scoped>
.session-page {
  display: grid;
  gap: 14px;
}
</style>
