<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Edit3, KeyRound, Plus, RefreshCw, Search, ShieldCheck, UserRoundCheck } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getAdminRoles } from '../api/adminRole'
import { createAdminUser, getAdminUsers, resetAdminUserPassword, updateAdminUser } from '../api/adminUser'
import type { AdminRoleItem } from '../types/adminRole'
import type { AdminUserItem, AdminUserStatus } from '../types/adminUser'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const rows = ref<AdminUserItem[]>([])
const roleOptions = ref<AdminRoleItem[]>([])
const total = ref(0)
const page = ref(Number(route.query.page) || 1)
const perPage = 15
const editing = ref<AdminUserItem | null>(null)
const showForm = ref(false)
const showPassword = ref(false)
const passwordTarget = ref<AdminUserItem | null>(null)
const passwordValue = ref('')

const filters = reactive({
  keyword: typeof route.query.keyword === 'string' ? route.query.keyword : '',
  status: typeof route.query.status === 'string' ? (route.query.status as AdminUserStatus) : '',
})

const form = reactive({
  username: '',
  email: '',
  displayName: '',
  password: '',
  status: 'active' as AdminUserStatus,
  roleIDs: [] as number[],
})

const lastPage = computed(() => Math.max(1, Math.ceil(total.value / perPage)))
const formTitle = computed(() => (editing.value ? '编辑管理员' : '创建管理员'))

async function loadUsers() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await getAdminUsers({
      page: page.value,
      per_page: perPage,
      keyword: filters.keyword || undefined,
      status: filters.status ? (filters.status as AdminUserStatus) : undefined,
    })
    rows.value = result.list
    total.value = result.total
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '管理员加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

async function loadRoleOptions() {
  const result = await getAdminRoles({ page: 1, per_page: 100, status: 'active' })
  roleOptions.value = result.list
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
  await loadUsers()
}

async function changePage(nextPage: number) {
  page.value = Math.min(Math.max(1, nextPage), lastPage.value)
  await syncQuery()
  await loadUsers()
}

function openCreate() {
  editing.value = null
  form.username = ''
  form.email = ''
  form.displayName = ''
  form.password = ''
  form.status = 'active'
  form.roleIDs = []
  showForm.value = true
}

function openEdit(row: AdminUserItem) {
  editing.value = row
  form.username = row.username
  form.email = row.email || ''
  form.displayName = row.display_name
  form.password = ''
  form.status = row.status
  form.roleIDs = [...row.role_ids]
  showForm.value = true
}

async function submitForm() {
  submitting.value = true
  errorMessage.value = ''
  try {
    if (editing.value) {
      await updateAdminUser(editing.value.id, {
        email: form.email.trim() || null,
        display_name: form.displayName.trim(),
        status: form.status,
        role_ids: form.roleIDs,
      })
    } else {
      await createAdminUser({
        username: form.username.trim(),
        email: form.email.trim() || null,
        display_name: form.displayName.trim(),
        password: form.password,
        status: form.status,
        role_ids: form.roleIDs,
      })
    }
    showForm.value = false
    await loadUsers()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '提交失败，请稍后重试'
  } finally {
    submitting.value = false
  }
}

async function toggleStatus(row: AdminUserItem) {
  const nextStatus: AdminUserStatus = row.status === 'active' ? 'disabled' : 'active'
  const label = nextStatus === 'active' ? '启用' : '禁用'
  if (!window.confirm(`${label}管理员 ${row.display_name || row.username}？`)) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    await updateAdminUser(row.id, { status: nextStatus })
    await loadUsers()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : `${label}失败，请稍后重试`
  } finally {
    submitting.value = false
  }
}

function openPassword(row: AdminUserItem) {
  passwordTarget.value = row
  passwordValue.value = ''
  showPassword.value = true
}

async function submitPassword() {
  if (!passwordTarget.value) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    await resetAdminUserPassword(passwordTarget.value.id, { password: passwordValue.value })
    showPassword.value = false
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '密码重置失败，请稍后重试'
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

function statusLabel(status: AdminUserStatus) {
  return status === 'active' ? '启用' : '禁用'
}

onMounted(async () => {
  await Promise.all([loadRoleOptions(), loadUsers()])
})
</script>

<template>
  <section class="admin-user-page">
    <header class="admin-user-toolbar">
      <div class="admin-user-title">
        <span class="admin-user-title-icon">
          <UserRoundCheck :size="20" aria-hidden="true" />
        </span>
        <div>
          <span>角色权限</span>
          <h1>管理员账号</h1>
        </div>
      </div>
      <div class="admin-user-filters">
        <label>
          <Search :size="15" aria-hidden="true" />
          <input v-model="filters.keyword" type="search" placeholder="账号 / 邮箱 / 昵称" @keyup.enter="applyFilters" />
        </label>
        <select v-model="filters.status" aria-label="账号状态">
          <option value="">全部状态</option>
          <option value="active">启用</option>
          <option value="disabled">禁用</option>
        </select>
        <button type="button" @click="applyFilters">
          <Search :size="15" aria-hidden="true" />
          查询
        </button>
        <button type="button" :disabled="loading" title="刷新" aria-label="刷新" @click="loadUsers">
          <RefreshCw :class="{ spinning: loading }" :size="15" aria-hidden="true" />
        </button>
        <button class="primary-action" type="button" @click="openCreate">
          <Plus :size="15" aria-hidden="true" />
          创建
        </button>
      </div>
    </header>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <article class="admin-user-table-card">
      <div v-if="loading && rows.length === 0" class="admin-user-state">管理员加载中...</div>
      <div v-else-if="rows.length === 0" class="admin-user-state">
        <ShieldCheck :size="18" aria-hidden="true" />
        暂无管理员账号
      </div>
      <div v-else class="admin-user-table-scroll">
        <table>
          <thead>
            <tr>
              <th>账号</th>
              <th>邮箱</th>
              <th>角色</th>
              <th>状态</th>
              <th>最后登录</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>
                <strong>{{ row.display_name }}</strong>
                <small>{{ row.username }}</small>
              </td>
              <td>{{ row.email || '-' }}</td>
              <td>{{ row.roles.map((role) => role.name).join('，') || '-' }}</td>
              <td>
                <span class="status-pill" :class="`status-${row.status}`">{{ statusLabel(row.status) }}</span>
              </td>
              <td>
                {{ formatDate(row.last_login_at) }}
                <small>{{ row.last_login_ip || '' }}</small>
              </td>
              <td class="row-actions">
                <button type="button" title="编辑" aria-label="编辑" @click="openEdit(row)">
                  <Edit3 :size="15" aria-hidden="true" />
                </button>
                <button type="button" title="重置密码" aria-label="重置密码" @click="openPassword(row)">
                  <KeyRound :size="15" aria-hidden="true" />
                </button>
                <button type="button" :disabled="submitting" @click="toggleStatus(row)">
                  {{ row.status === 'active' ? '禁用' : '启用' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="admin-user-pagination">
        <span>共 {{ total }} 个管理员</span>
        <div>
          <button type="button" :disabled="page <= 1 || loading" @click="changePage(page - 1)">上一页</button>
          <strong>{{ page }} / {{ lastPage }}</strong>
          <button type="button" :disabled="page >= lastPage || loading" @click="changePage(page + 1)">下一页</button>
        </div>
      </footer>
    </article>

    <div v-if="showForm" class="admin-user-modal" role="dialog" aria-modal="true">
      <form class="admin-user-dialog" @submit.prevent="submitForm">
        <header>
          <h2>{{ formTitle }}</h2>
          <button type="button" aria-label="关闭" @click="showForm = false">×</button>
        </header>
        <label>
          <span>账号</span>
          <input v-model="form.username" :disabled="Boolean(editing)" required minlength="3" maxlength="64" />
        </label>
        <label>
          <span>邮箱</span>
          <input v-model="form.email" type="email" maxlength="191" />
        </label>
        <label>
          <span>显示名称</span>
          <input v-model="form.displayName" required maxlength="64" />
        </label>
        <label v-if="!editing">
          <span>初始密码</span>
          <input v-model="form.password" type="password" required minlength="6" maxlength="72" autocomplete="new-password" />
        </label>
        <div class="role-options">
          <span>角色</span>
          <label v-for="role in roleOptions" :key="role.id">
            <input v-model="form.roleIDs" type="checkbox" :value="role.id" />
            <strong>{{ role.name }}</strong>
            <small>{{ role.code }}</small>
          </label>
        </div>
        <label>
          <span>状态</span>
          <select v-model="form.status">
            <option value="active">启用</option>
            <option value="disabled">禁用</option>
          </select>
        </label>
        <footer>
          <button type="button" @click="showForm = false">取消</button>
          <button class="primary-action" type="submit" :disabled="submitting">
            {{ submitting ? '提交中...' : '保存' }}
          </button>
        </footer>
      </form>
    </div>

    <div v-if="showPassword" class="admin-user-modal" role="dialog" aria-modal="true">
      <form class="admin-user-dialog" @submit.prevent="submitPassword">
        <header>
          <h2>重置密码</h2>
          <button type="button" aria-label="关闭" @click="showPassword = false">×</button>
        </header>
        <label>
          <span>新密码</span>
          <input v-model="passwordValue" type="password" required minlength="6" maxlength="72" autocomplete="new-password" />
        </label>
        <footer>
          <button type="button" @click="showPassword = false">取消</button>
          <button class="primary-action" type="submit" :disabled="submitting">
            {{ submitting ? '提交中...' : '确认重置' }}
          </button>
        </footer>
      </form>
    </div>
  </section>
</template>

<style scoped>
.admin-user-page {
  display: grid;
  gap: 14px;
}

.admin-user-toolbar,
.admin-user-table-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-soft);
}

.admin-user-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px;
}

.admin-user-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.admin-user-title-icon {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: var(--primary);
  background: var(--primary-soft);
}

.admin-user-title span {
  color: var(--muted);
  font-size: 13px;
  font-weight: 800;
}

.admin-user-title h1 {
  margin: 3px 0 0;
  font-size: 20px;
  line-height: 1.2;
}

.admin-user-filters {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.admin-user-filters label,
.admin-user-filters button,
.admin-user-filters select,
.admin-user-pagination button,
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

.admin-user-filters input {
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

.admin-user-table-card {
  overflow: hidden;
}

.admin-user-table-scroll {
  overflow-x: auto;
  padding: 0 10px;
}

table {
  width: 100%;
  min-width: 980px;
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

.admin-user-state {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--muted);
  font-weight: 800;
}

.admin-user-pagination {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  color: var(--muted);
}

.admin-user-pagination div {
  display: flex;
  align-items: center;
  gap: 10px;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.admin-user-modal {
  position: fixed;
  inset: 0;
  z-index: 20;
  display: grid;
  place-items: center;
  padding: 20px;
  background: rgb(15 23 42 / 46%);
}

.admin-user-dialog {
  width: min(460px, 100%);
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-strong);
}

.admin-user-dialog header,
.admin-user-dialog footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.admin-user-dialog h2 {
  margin: 0;
  font-size: 18px;
}

.admin-user-dialog label {
  display: grid;
  gap: 6px;
  color: var(--muted-strong);
  font-size: 13px;
  font-weight: 800;
}

.role-options {
  display: grid;
  gap: 8px;
  color: var(--muted-strong);
  font-size: 13px;
  font-weight: 800;
}

.role-options label {
  min-height: 34px;
  display: grid;
  grid-template-columns: 18px minmax(80px, 1fr) minmax(100px, 1fr);
  align-items: center;
  gap: 8px;
  padding: 0 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel-soft);
}

.role-options input[type='checkbox'] {
  min-height: 0;
  width: 14px;
  height: 14px;
  padding: 0;
}

.role-options strong,
.role-options small {
  display: block;
}

.role-options small {
  color: var(--muted);
}

.admin-user-dialog input:not([type='checkbox']),
.admin-user-dialog select {
  min-height: 38px;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0 10px;
  color: var(--text);
  background: var(--panel);
}

.admin-user-dialog header button,
.admin-user-dialog footer button {
  min-height: 34px;
  padding: 0 11px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  background: var(--panel);
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

@media (max-width: 960px) {
  .admin-user-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .admin-user-filters {
    justify-content: flex-start;
  }
}
</style>
