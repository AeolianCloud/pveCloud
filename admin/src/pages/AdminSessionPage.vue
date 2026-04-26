<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { MonitorCheck, RefreshCw, Search, ShieldAlert } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getAdminSessions, revokeAdminSession } from '../api/adminSession'
import type { AdminSessionItem, AdminSessionStatus } from '../types/adminSession'

const route = useRoute()
const router = useRouter()
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

const lastPage = computed(() => Math.max(1, Math.ceil(total.value / perPage)))

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

async function changePage(nextPage: number) {
  page.value = Math.min(Math.max(1, nextPage), lastPage.value)
  await syncQuery()
  await loadSessions()
}

async function revoke(row: AdminSessionItem) {
  if (!window.confirm(`吊销 ${row.admin?.display_name || row.admin?.username || row.session_id} 的会话？`)) {
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

onMounted(loadSessions)
</script>

<template>
  <section class="session-page">
    <header class="session-toolbar">
      <div class="session-title">
        <span class="session-title-icon"><MonitorCheck :size="20" aria-hidden="true" /></span>
        <div>
          <span>访问控制</span>
          <h1>登录会话</h1>
        </div>
      </div>
      <div class="session-filters">
        <label>
          <Search :size="15" aria-hidden="true" />
          <input v-model="filters.keyword" type="search" placeholder="管理员 / IP / 会话" @keyup.enter="applyFilters" />
        </label>
        <select v-model="filters.status" aria-label="会话状态">
          <option value="">全部状态</option>
          <option value="active">活跃</option>
          <option value="revoked">已吊销</option>
          <option value="expired">已过期</option>
        </select>
        <button type="button" @click="applyFilters"><Search :size="15" aria-hidden="true" />查询</button>
        <button type="button" :disabled="loading" aria-label="刷新" @click="loadSessions">
          <RefreshCw :class="{ spinning: loading }" :size="15" aria-hidden="true" />
        </button>
      </div>
    </header>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <article class="session-table-card">
      <div v-if="loading && rows.length === 0" class="session-state">会话加载中...</div>
      <div v-else-if="rows.length === 0" class="session-state">
        <ShieldAlert :size="18" aria-hidden="true" />
        暂无登录会话
      </div>
      <div v-else class="session-table-scroll">
        <table>
          <thead>
            <tr>
              <th>管理员</th>
              <th>状态</th>
              <th>最后访问</th>
              <th>签发 / 过期</th>
              <th>吊销</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>
                <strong>{{ row.admin?.display_name || row.admin?.username || '未知管理员' }}</strong>
                <small>{{ row.session_id }}</small>
              </td>
              <td><span class="status-pill" :class="`status-${row.status}`">{{ statusLabel(row.status) }}</span></td>
              <td>
                {{ formatDate(row.last_seen_at) }}
                <small>{{ row.last_seen_ip || '-' }}</small>
              </td>
              <td>
                {{ formatDate(row.issued_at) }}
                <small>{{ formatDate(row.expires_at) }}</small>
              </td>
              <td>
                {{ formatDate(row.revoked_at) }}
                <small>{{ row.revoke_reason || '' }}</small>
              </td>
              <td>
                <button type="button" :disabled="row.status !== 'active' || submitting" @click="revoke(row)">吊销</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="session-pagination">
        <span>共 {{ total }} 条会话</span>
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
.session-page { display: grid; gap: 14px; }
.session-toolbar, .session-table-card { border: 1px solid var(--border); border-radius: 8px; background: var(--panel); box-shadow: var(--shadow-soft); }
.session-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 14px; padding: 14px; }
.session-title { display: flex; align-items: center; gap: 12px; }
.session-title-icon { width: 38px; height: 38px; display: grid; place-items: center; border-radius: 8px; color: var(--primary); background: var(--primary-soft); }
.session-title span { color: var(--muted); font-size: 13px; font-weight: 800; }
.session-title h1 { margin: 3px 0 0; font-size: 20px; line-height: 1.2; }
.session-filters { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
.session-filters label, .session-filters button, .session-filters select, .session-pagination button, td button { min-height: 34px; display: inline-flex; align-items: center; justify-content: center; gap: 7px; padding: 0 11px; border: 1px solid var(--border); border-radius: 8px; color: var(--muted-strong); background: var(--panel); font-size: 13px; font-weight: 750; cursor: pointer; }
.session-filters input { width: 170px; border: 0; outline: 0; color: var(--text); background: transparent; }
.session-table-card { overflow: hidden; }
.session-table-scroll { overflow-x: auto; padding: 0 10px; }
table { width: 100%; min-width: 980px; border-collapse: collapse; color: var(--muted-strong); font-size: 13px; }
th, td { height: 48px; padding: 0 9px; border-bottom: 1px solid var(--border); text-align: left; white-space: nowrap; }
th { height: 38px; color: var(--table-head-text); background: var(--panel-soft); font-weight: 800; }
td strong, td small { display: block; }
td strong { color: var(--text); }
td small { margin-top: 3px; color: var(--muted); }
.status-pill { display: inline-flex; align-items: center; min-height: 22px; padding: 0 8px; border-radius: 6px; font-weight: 850; }
.status-active { color: var(--success); background: var(--success-soft); }
.status-revoked, .status-expired { color: var(--danger); background: var(--danger-soft); }
.session-state { min-height: 220px; display: flex; align-items: center; justify-content: center; gap: 8px; color: var(--muted); font-weight: 800; }
.session-pagination { min-height: 52px; display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 0 14px; color: var(--muted); }
.session-pagination div { display: flex; align-items: center; gap: 10px; }
button:disabled { cursor: not-allowed; opacity: 0.55; }
.spinning { animation: spin 800ms linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 960px) { .session-toolbar { align-items: flex-start; flex-direction: column; } .session-filters { justify-content: flex-start; } }
</style>
