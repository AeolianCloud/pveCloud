<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ShieldCheck, UserRoundCheck } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getAdminRoles } from '../api/adminRole'
import { createAdminUser, getAdminUsers, resetAdminUserPassword, updateAdminUser } from '../api/adminUser'
import AdminEmptyState from '../components/AdminEmptyState.vue'
import AdminPageHeader from '../components/AdminPageHeader.vue'
import AdminTablePanel from '../components/AdminTablePanel.vue'
import type { AdminRoleItem } from '../types/adminRole'
import type { AdminUserItem, AdminUserStatus } from '../types/adminUser'
import { useConfirmAction } from '../utils/confirmAction'

const route = useRoute()
const router = useRouter()
const confirmAction = useConfirmAction()
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

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '启用', value: 'active' },
  { label: '禁用', value: 'disabled' },
]

const first = computed(() => (page.value - 1) * perPage)
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

async function changePaginator(event: { page: number }) {
  page.value = event.page + 1
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
  const confirmed = await confirmAction({
    header: `${label}管理员`,
    message: `确认${label}管理员 ${row.display_name || row.username} 吗`,
    acceptLabel: label,
  })
  if (!confirmed) {
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

function statusSeverity(status: AdminUserStatus) {
  return status === 'active' ? 'success' : 'danger'
}

onMounted(async () => {
  await Promise.all([loadRoleOptions(), loadUsers()])
})
</script>

<template>
  <section class="admin-user-page">
    <AdminPageHeader eyebrow="角色权限" title="管理员账号" :icon="UserRoundCheck">
      <IconField>
        <InputIcon class="pi pi-search" />
        <InputText v-model="filters.keyword" type="search" placeholder="账号 / 邮箱 / 昵称" @keyup.enter="applyFilters" />
      </IconField>
      <Select v-model="filters.status" :options="statusOptions" option-label="label" option-value="value" aria-label="账号状态" />
      <Button type="button" label="查询" icon="pi pi-search" @click="applyFilters" />
      <Button type="button" icon="pi pi-refresh" :loading="loading" severity="secondary" outlined aria-label="刷新" @click="loadUsers" />
      <Button type="button" label="创建" icon="pi pi-plus" @click="openCreate" />
    </AdminPageHeader>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

    <AdminTablePanel>
      <AdminEmptyState v-if="loading && rows.length === 0" text="管理员加载中..." />
      <AdminEmptyState v-else-if="rows.length === 0" text="暂无管理员账号" :icon="ShieldCheck" />
      <DataTable v-else :value="rows" class="admin-prime-table" data-key="id" striped-rows>
        <Column header="账号">
          <template #body="{ data }">
            <strong>{{ data.display_name }}</strong>
            <small>{{ data.username }}</small>
          </template>
        </Column>
        <Column field="email" header="邮箱">
          <template #body="{ data }">{{ data.email || '-' }}</template>
        </Column>
        <Column header="角色">
          <template #body="{ data }">{{ data.roles.map((role: AdminRoleItem) => role.name).join('，') || '-' }}</template>
        </Column>
        <Column header="状态">
          <template #body="{ data }">
            <Tag :value="statusLabel(data.status)" :severity="statusSeverity(data.status)" />
          </template>
        </Column>
        <Column header="最后登录">
          <template #body="{ data }">
            {{ formatDate(data.last_login_at) }}
            <small>{{ data.last_login_ip || '' }}</small>
          </template>
        </Column>
        <Column header="操作">
          <template #body="{ data }">
            <div class="row-actions">
              <Button icon="pi pi-pencil" severity="secondary" text rounded aria-label="编辑" @click="openEdit(data)" />
              <Button icon="pi pi-key" severity="secondary" text rounded aria-label="重置密码" @click="openPassword(data)" />
              <Button
                :label="data.status === 'active' ? '禁用' : '启用'"
                :severity="data.status === 'active' ? 'danger' : 'success'"
                text
                :disabled="submitting"
                @click="toggleStatus(data)"
              />
            </div>
          </template>
        </Column>
      </DataTable>
      <template #footer>
        <span>共 {{ total }} 个管理员</span>
        <Paginator :first="first" :rows="perPage" :total-records="total" template="PrevPageLink PageLinks NextPageLink" @page="changePaginator" />
      </template>
    </AdminTablePanel>

    <Dialog v-model:visible="showForm" modal :header="formTitle" class="admin-form-dialog">
      <form class="admin-form-grid" @submit.prevent="submitForm">
        <label class="admin-field">
          <span>账号</span>
          <InputText v-model="form.username" :disabled="Boolean(editing)" required minlength="3" maxlength="64" />
        </label>
        <label class="admin-field">
          <span>邮箱</span>
          <InputText v-model="form.email" type="email" maxlength="191" />
        </label>
        <label class="admin-field">
          <span>显示名称</span>
          <InputText v-model="form.displayName" required maxlength="64" />
        </label>
        <label v-if="!editing" class="admin-field">
          <span>初始密码</span>
          <Password v-model="form.password" required minlength="6" maxlength="72" autocomplete="new-password" toggle-mask :feedback="false" />
        </label>
        <div class="role-options">
          <span class="role-options-title">角色</span>
          <label v-for="role in roleOptions" :key="role.id">
            <Checkbox v-model="form.roleIDs" :input-id="`admin-role-${role.id}`" :value="role.id" />
            <strong>{{ role.name }}</strong>
            <small>{{ role.code }}</small>
          </label>
        </div>
        <label class="admin-field">
          <span>状态</span>
          <Select v-model="form.status" :options="statusOptions.slice(1)" option-label="label" option-value="value" />
        </label>
        <footer>
          <Button type="button" label="取消" severity="secondary" outlined @click="showForm = false" />
          <Button type="submit" label="保存" :loading="submitting" />
        </footer>
      </form>
    </Dialog>

    <Dialog v-model:visible="showPassword" modal header="重置密码" class="admin-form-dialog admin-form-dialog--compact">
      <form class="admin-form-grid admin-form-grid--compact" @submit.prevent="submitPassword">
        <label class="admin-field">
          <span>新密码</span>
          <Password v-model="passwordValue" required minlength="6" maxlength="72" autocomplete="new-password" toggle-mask :feedback="false" />
        </label>
        <footer>
          <Button type="button" label="取消" severity="secondary" outlined @click="showPassword = false" />
          <Button type="submit" label="确认重置" :loading="submitting" />
        </footer>
      </form>
    </Dialog>
  </section>
</template>

<style scoped>
.admin-user-page {
  display: grid;
  gap: 14px;
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 7px;
  min-height: 48px;
}

.role-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 10px;
  margin: 14px 18px 4px;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  background: var(--panel-soft);
  font-size: 13px;
  font-weight: 800;
}

.role-options-title {
  grid-column: 1 / -1;
  color: var(--text);
}

.role-options label {
  min-height: 42px;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  align-items: center;
  gap: 8px 9px;
  padding: 8px 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
}

.role-options strong,
.role-options small {
  display: block;
}

.role-options strong {
  min-width: 0;
}

.role-options small {
  grid-column: 2;
  margin-top: -5px;
  color: var(--muted);
}

@media (max-width: 960px) {
  .role-options {
    grid-template-columns: 1fr;
  }
}
</style>
