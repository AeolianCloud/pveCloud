<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ShieldCheck } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { createAdminRole, getAdminPermissions, getAdminRoles, updateAdminRole } from '../api/adminRole'
import AdminEmptyState from '../components/AdminEmptyState.vue'
import AdminPageHeader from '../components/AdminPageHeader.vue'
import AdminTablePanel from '../components/AdminTablePanel.vue'
import type { AdminPermissionGroup, AdminRoleItem, AdminRoleStatus } from '../types/adminRole'
import { useConfirmAction } from '../utils/confirmAction'

const route = useRoute()
const router = useRouter()
const confirmAction = useConfirmAction()
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

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '启用', value: 'active' },
  { label: '禁用', value: 'disabled' },
]

const first = computed(() => (page.value - 1) * perPage)
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

async function changePaginator(event: { page: number }) {
  page.value = event.page + 1
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
  const confirmed = await confirmAction({
    header: `${label}角色`,
    message: `确认${label}角色 ${row.name} 吗`,
    acceptLabel: label,
  })
  if (!confirmed) {
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

function statusSeverity(status: AdminRoleStatus) {
  return status === 'active' ? 'success' : 'danger'
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
    <AdminPageHeader eyebrow="角色权限" title="角色权限" :icon="ShieldCheck">
      <IconField>
        <InputIcon class="pi pi-search" />
        <InputText v-model="filters.keyword" type="search" placeholder="编码 / 名称 / 说明" @keyup.enter="applyFilters" />
      </IconField>
      <Select v-model="filters.status" :options="statusOptions" option-label="label" option-value="value" aria-label="角色状态" />
      <Button type="button" label="查询" icon="pi pi-search" @click="applyFilters" />
      <Button type="button" icon="pi pi-refresh" :loading="loading" severity="secondary" outlined aria-label="刷新" @click="loadRoles" />
      <Button type="button" label="创建" icon="pi pi-plus" @click="openCreate" />
    </AdminPageHeader>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

    <AdminTablePanel>
      <AdminEmptyState v-if="loading && rows.length === 0" text="角色加载中..." />
      <AdminEmptyState v-else-if="rows.length === 0" text="暂无角色" :icon="ShieldCheck" />
      <DataTable v-else :value="rows" class="admin-prime-table" data-key="id" striped-rows>
        <Column header="角色">
          <template #body="{ data }">
            <strong>{{ data.name }}</strong>
            <small>{{ data.code }}{{ data.description ? ` · ${data.description}` : '' }}</small>
          </template>
        </Column>
        <Column header="权限数">
          <template #body="{ data }">{{ data.permission_codes.length }}</template>
        </Column>
        <Column header="状态">
          <template #body="{ data }">
            <Tag :value="statusLabel(data.status)" :severity="statusSeverity(data.status)" />
          </template>
        </Column>
        <Column header="更新时间">
          <template #body="{ data }">{{ formatDate(data.updated_at) }}</template>
        </Column>
        <Column header="操作">
          <template #body="{ data }">
            <div class="row-actions">
              <Button icon="pi pi-pencil" severity="secondary" text rounded aria-label="编辑" @click="openEdit(data)" />
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
        <span>共 {{ total }} 个角色</span>
        <Paginator :first="first" :rows="perPage" :total-records="total" template="PrevPageLink PageLinks NextPageLink" @page="changePaginator" />
      </template>
    </AdminTablePanel>

    <Dialog v-model:visible="showForm" modal :header="formTitle" class="admin-form-dialog admin-form-dialog--wide">
      <form class="admin-form-grid" @submit.prevent="submitForm">
        <label class="admin-field">
          <span>角色编码</span>
          <InputText v-model="form.code" :disabled="Boolean(editing)" required minlength="2" maxlength="64" />
        </label>
        <label class="admin-field">
          <span>角色名称</span>
          <InputText v-model="form.name" required maxlength="64" />
        </label>
        <label class="admin-field">
          <span>角色说明</span>
          <InputText v-model="form.description" maxlength="255" />
        </label>
        <label class="admin-field">
          <span>状态</span>
          <Select v-model="form.status" :options="statusOptions.slice(1)" option-label="label" option-value="value" />
        </label>
        <div class="permission-panel">
          <section v-for="group in permissionGroups" :key="group.group_name">
            <h3>{{ group.group_name }}</h3>
            <label v-for="permission in group.permissions" :key="permission.code" class="permission-option">
              <Checkbox v-model="form.permissionCodes" :input-id="`permission-${permission.code}`" :value="permission.code" />
              <span>{{ permission.name }}</span>
              <small>{{ permission.code }}</small>
            </label>
          </section>
        </div>
        <footer>
          <Button type="button" label="取消" severity="secondary" outlined @click="showForm = false" />
          <Button type="submit" label="保存" :loading="submitting" />
        </footer>
      </form>
    </Dialog>
  </section>
</template>

<style scoped>
.admin-role-page {
  display: grid;
  gap: 14px;
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 7px;
  min-height: 48px;
}

.permission-panel {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin: 14px 18px 4px;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel-soft);
}

.permission-panel section {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 10px;
  background: var(--panel);
}

.permission-panel h3 {
  margin: 0 0 8px;
  color: var(--text);
  font-size: 13px;
}

.permission-option {
  min-height: 34px;
  display: grid;
  grid-template-columns: 18px minmax(80px, 1fr) minmax(120px, 1.3fr);
  align-items: center;
  gap: 8px;
  color: var(--muted-strong);
  font-size: 12px;
}

.permission-option small {
  color: var(--muted);
}

@media (max-width: 960px) {
  .permission-panel {
    grid-template-columns: 1fr;
  }
}
</style>
