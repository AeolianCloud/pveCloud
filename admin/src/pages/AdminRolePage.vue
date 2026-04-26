<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Edit3, Plus, RefreshCw, Search, ShieldCheck } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { createAdminRole, getAdminPermissions, getAdminRoles, updateAdminRole } from '../api/adminRole'
import type { AdminPermissionGroup, AdminRoleItem, AdminRoleStatus } from '../types/adminRole'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const rows = ref<AdminRoleItem[]>([])
const permissionGroups = ref<AdminPermissionGroup[]>([])
const total = ref(0)
const page = ref(Number(route.query.page) || 1)
const perPage = 15
const editing = ref<AdminRoleItem | null>(null)
const showForm = ref(false)

const filters = reactive({
  keyword: typeof route.query.keyword === 'string' ? route.query.keyword : '',
  status: typeof route.query.status === 'string' ? (route.query.status as AdminRoleStatus) : '',
})

const form = reactive({
  code: '',
  name: '',
  description: '',
  status: 'active' as AdminRoleStatus,
  permissionCodes: [] as string[],
})

const lastPage = computed(() => Math.max(1, Math.ceil(total.value / perPage)))
const formTitle = computed(() => (editing.value ? '编辑角色' : '创建角色'))

async function loadRoles() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await getAdminRoles({
      page: page.value,
      per_page: perPage,
      keyword: filters.keyword || undefined,
      status: filters.status ? (filters.status as AdminRoleStatus) : undefined,
    })
    rows.value = result.list
    total.value = result.total
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '角色加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

async function loadPermissions() {
  permissionGroups.value = await getAdminPermissions()
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
  await loadRoles()
}

async function changePage(nextPage: number) {
  page.value = Math.min(Math.max(1, nextPage), lastPage.value)
  await syncQuery()
  await loadRoles()
}

function openCreate() {
  editing.value = null
  form.code = ''
  form.name = ''
  form.description = ''
  form.status = 'active'
  form.permissionCodes = []
  showForm.value = true
}

function openEdit(row: AdminRoleItem) {
  editing.value = row
  form.code = row.code
  form.name = row.name
  form.description = row.description || ''
  form.status = row.status
  form.permissionCodes = [...row.permission_codes]
  showForm.value = true
}

async function submitForm() {
  submitting.value = true
  errorMessage.value = ''
  try {
    if (editing.value) {
      await updateAdminRole(editing.value.id, {
        name: form.name.trim(),
        description: form.description.trim() || null,
        status: form.status,
        permission_codes: form.permissionCodes,
      })
    } else {
      await createAdminRole({
        code: form.code.trim(),
        name: form.name.trim(),
        description: form.description.trim() || null,
        status: form.status,
        permission_codes: form.permissionCodes,
      })
    }
    showForm.value = false
    await loadRoles()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '提交失败，请稍后重试'
  } finally {
    submitting.value = false
  }
}

async function toggleStatus(row: AdminRoleItem) {
  const nextStatus: AdminRoleStatus = row.status === 'active' ? 'disabled' : 'active'
  const label = nextStatus === 'active' ? '启用' : '禁用'
  if (!window.confirm(`${label}角色 ${row.name}？`)) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    await updateAdminRole(row.id, { status: nextStatus })
    await loadRoles()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : `${label}失败，请稍后重试`
  } finally {
    submitting.value = false
  }
}

function statusLabel(status: AdminRoleStatus) {
  return status === 'active' ? '启用' : '禁用'
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

onMounted(async () => {
  await Promise.all([loadPermissions(), loadRoles()])
})
</script>

<template>
  <section class="admin-role-page">
    <header class="admin-role-toolbar">
      <div class="admin-role-title">
        <span class="admin-role-title-icon">
          <ShieldCheck :size="20" aria-hidden="true" />
        </span>
        <div>
          <span>角色权限</span>
          <h1>角色权限</h1>
        </div>
      </div>
      <div class="admin-role-filters">
        <label>
          <Search :size="15" aria-hidden="true" />
          <input v-model="filters.keyword" type="search" placeholder="编码 / 名称 / 说明" @keyup.enter="applyFilters" />
        </label>
        <select v-model="filters.status" aria-label="角色状态">
          <option value="">全部状态</option>
          <option value="active">启用</option>
          <option value="disabled">禁用</option>
        </select>
        <button type="button" @click="applyFilters">
          <Search :size="15" aria-hidden="true" />
          查询
        </button>
        <button type="button" :disabled="loading" title="刷新" aria-label="刷新" @click="loadRoles">
          <RefreshCw :class="{ spinning: loading }" :size="15" aria-hidden="true" />
        </button>
        <button class="primary-action" type="button" @click="openCreate">
          <Plus :size="15" aria-hidden="true" />
          创建
        </button>
      </div>
    </header>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <article class="admin-role-table-card">
      <div v-if="loading && rows.length === 0" class="admin-role-state">角色加载中...</div>
      <div v-else-if="rows.length === 0" class="admin-role-state">
        <ShieldCheck :size="18" aria-hidden="true" />
        暂无角色
      </div>
      <div v-else class="admin-role-table-scroll">
        <table>
          <thead>
            <tr>
              <th>角色</th>
              <th>权限数</th>
              <th>状态</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>
                <strong>{{ row.name }}</strong>
                <small>{{ row.code }}{{ row.description ? ` · ${row.description}` : '' }}</small>
              </td>
              <td>{{ row.permission_codes.length }}</td>
              <td>
                <span class="status-pill" :class="`status-${row.status}`">{{ statusLabel(row.status) }}</span>
              </td>
              <td>{{ formatDate(row.updated_at) }}</td>
              <td class="row-actions">
                <button type="button" title="编辑" aria-label="编辑" @click="openEdit(row)">
                  <Edit3 :size="15" aria-hidden="true" />
                </button>
                <button type="button" :disabled="submitting" @click="toggleStatus(row)">
                  {{ row.status === 'active' ? '禁用' : '启用' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="admin-role-pagination">
        <span>共 {{ total }} 个角色</span>
        <div>
          <button type="button" :disabled="page <= 1 || loading" @click="changePage(page - 1)">上一页</button>
          <strong>{{ page }} / {{ lastPage }}</strong>
          <button type="button" :disabled="page >= lastPage || loading" @click="changePage(page + 1)">下一页</button>
        </div>
      </footer>
    </article>

    <div v-if="showForm" class="admin-role-modal" role="dialog" aria-modal="true">
      <form class="admin-role-dialog" @submit.prevent="submitForm">
        <header>
          <h2>{{ formTitle }}</h2>
          <button type="button" aria-label="关闭" @click="showForm = false">×</button>
        </header>
        <label>
          <span>角色编码</span>
          <input v-model="form.code" :disabled="Boolean(editing)" required minlength="2" maxlength="64" />
        </label>
        <label>
          <span>角色名称</span>
          <input v-model="form.name" required maxlength="64" />
        </label>
        <label>
          <span>角色说明</span>
          <input v-model="form.description" maxlength="255" />
        </label>
        <label>
          <span>状态</span>
          <select v-model="form.status">
            <option value="active">启用</option>
            <option value="disabled">禁用</option>
          </select>
        </label>
        <div class="permission-panel">
          <section v-for="group in permissionGroups" :key="group.group_name">
            <h3>{{ group.group_name }}</h3>
            <label v-for="permission in group.permissions" :key="permission.code" class="permission-option">
              <input v-model="form.permissionCodes" type="checkbox" :value="permission.code" />
              <span>{{ permission.name }}</span>
              <small>{{ permission.code }}</small>
            </label>
          </section>
        </div>
        <footer>
          <button type="button" @click="showForm = false">取消</button>
          <button class="primary-action" type="submit" :disabled="submitting">
            {{ submitting ? '提交中...' : '保存' }}
          </button>
        </footer>
      </form>
    </div>
  </section>
</template>

<style scoped>
.admin-role-page {
  display: grid;
  gap: 14px;
}

.admin-role-toolbar,
.admin-role-table-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-soft);
}

.admin-role-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px;
}

.admin-role-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.admin-role-title-icon {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: var(--primary);
  background: var(--primary-soft);
}

.admin-role-title span {
  color: var(--muted);
  font-size: 13px;
  font-weight: 800;
}

.admin-role-title h1 {
  margin: 3px 0 0;
  font-size: 20px;
  line-height: 1.2;
}

.admin-role-filters {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.admin-role-filters label,
.admin-role-filters button,
.admin-role-filters select,
.admin-role-pagination button,
.row-actions button {
  min-height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 0 11px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  background: var(--panel);
  font-size: 13px;
  font-weight: 750;
  cursor: pointer;
}

.admin-role-filters input {
  width: 170px;
  border: 0;
  outline: 0;
  color: var(--text);
  background: transparent;
}

.primary-action {
  border-color: transparent !important;
  color: #fff !important;
  background: var(--primary) !important;
}

.admin-role-table-card {
  overflow: hidden;
}

.admin-role-table-scroll {
  overflow-x: auto;
  padding: 0 10px;
}

table {
  width: 100%;
  min-width: 860px;
  border-collapse: collapse;
  color: var(--muted-strong);
  font-size: 13px;
}

th,
td {
  height: 48px;
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

td strong,
td small {
  display: block;
}

td strong {
  color: var(--text);
}

td small {
  margin-top: 3px;
  color: var(--muted);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 8px;
  border-radius: 6px;
  font-weight: 850;
}

.status-active {
  color: var(--success);
  background: var(--success-soft);
}

.status-disabled {
  color: var(--danger);
  background: var(--danger-soft);
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 7px;
  height: 48px;
}

.admin-role-state {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--muted);
  font-weight: 800;
}

.admin-role-pagination {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  color: var(--muted);
}

.admin-role-pagination div {
  display: flex;
  align-items: center;
  gap: 10px;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.admin-role-modal {
  position: fixed;
  inset: 0;
  z-index: 20;
  display: grid;
  place-items: center;
  padding: 20px;
  background: rgb(15 23 42 / 46%);
}

.admin-role-dialog {
  width: min(720px, 100%);
  max-height: min(760px, calc(100vh - 40px));
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-strong);
  overflow: auto;
}

.admin-role-dialog header,
.admin-role-dialog footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.admin-role-dialog h2 {
  margin: 0;
  font-size: 18px;
}

.admin-role-dialog > label {
  display: grid;
  gap: 6px;
  color: var(--muted-strong);
  font-size: 13px;
  font-weight: 800;
}

.admin-role-dialog input:not([type='checkbox']),
.admin-role-dialog select {
  min-height: 38px;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0 10px;
  color: var(--text);
  background: var(--panel);
}

.admin-role-dialog header button,
.admin-role-dialog footer button {
  min-height: 34px;
  padding: 0 11px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  background: var(--panel);
  cursor: pointer;
}

.permission-panel {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.permission-panel section {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 10px;
  background: var(--panel-soft);
}

.permission-panel h3 {
  margin: 0 0 8px;
  color: var(--text);
  font-size: 13px;
}

.permission-option {
  min-height: 30px;
  display: grid;
  grid-template-columns: 18px minmax(80px, 1fr) minmax(120px, 1.3fr);
  align-items: center;
  gap: 8px;
  color: var(--muted-strong);
  font-size: 12px;
}

.permission-option input[type='checkbox'] {
  min-height: 0;
  width: 14px;
  height: 14px;
  padding: 0;
}

.permission-option small {
  color: var(--muted);
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
  .admin-role-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .admin-role-filters {
    justify-content: flex-start;
  }

  .permission-panel {
    grid-template-columns: 1fr;
  }
}
</style>
