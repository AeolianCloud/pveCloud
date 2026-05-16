<script setup lang="ts">
import { RefreshOutline, SearchOutline } from '@vicons/ionicons5'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NPagination,
  NPopover,
  NSelect,
  NSpace,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import { getAuditLogs, type AuditLogItem, type AuditLogListQuery } from '../../api/audit-log'
import { ADMIN_ROUTE_NAME } from '../../router/constants'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'

const permissionStore = usePermissionStore()
const route = useRoute()
const loading = ref(false)
const refreshing = ref(false)
const errorMessage = ref('')
const logs = ref<AuditLogItem[]>([])
const pagination = reactive({ page: 1, per_page: 15, total: 0 })
const queryForm = reactive({
  admin_id: '',
  action: '',
  object_type: '',
  object_id: '',
  date_range: null as [number, number] | null,
})

const loginActionOptions = [
  { label: '登录成功', value: 'admin.login.success' },
  { label: '登录失败', value: 'admin.login.failed' },
  { label: '登录限流', value: 'admin.login.limited' },
  { label: '验证码限流', value: 'admin.captcha.limited' },
  { label: '退出登录', value: 'admin.logout' },
  { label: '会话刷新', value: 'admin.refresh' },
]

const isLoginPage = computed(() => route.name === ADMIN_ROUTE_NAME.adminSecurityLogs)
const canViewLogs = computed(() =>
  isLoginPage.value
    ? permissionStore.hasPermission(['page.logs.admin-security'])
    : permissionStore.hasPermission(['page.logs.admin-operations', 'page.system-settings.audit-logs']),
)
const canViewSensitive = computed(() => permissionStore.hasPermission('audit-log:sensitive-view'))
const hasLogs = computed(() => logs.value.length > 0)
const pageTitle = computed(() => (isLoginPage.value ? '登录安全' : '操作审计'))
const pageDescription = computed(() =>
  isLoginPage.value ? '查看管理端登录、退出、刷新和限流等安全记录。' : '查看后台资源变更和关键操作审计记录。',
)

function tsToDate(ts: number | null | undefined) {
  if (!ts) return undefined
  const d = new Date(ts)
  return `${d.getFullYear()}-${`${d.getMonth() + 1}`.padStart(2, '0')}-${`${d.getDate()}`.padStart(2, '0')}`
}

function buildQuery(): AuditLogListQuery {
  const r = queryForm.date_range
  return {
    page: pagination.page,
    per_page: pagination.per_page,
    log_type: isLoginPage.value ? 'admin_security' : 'admin_operation',
    admin_id: queryForm.admin_id ? Number(queryForm.admin_id) : undefined,
    action: queryForm.action.trim() || undefined,
    object_type: isLoginPage.value ? undefined : queryForm.object_type.trim() || undefined,
    object_id: queryForm.object_id.trim() || undefined,
    date_from: tsToDate(r?.[0]),
    date_to: tsToDate(r?.[1]),
  }
}

async function loadLogs(options: { refresh?: boolean } = {}) {
  if (!canViewLogs.value) {
    logs.value = []
    errorMessage.value = ''
    return
  }
  loading.value = !options.refresh
  refreshing.value = Boolean(options.refresh)
  errorMessage.value = ''
  try {
    const result = await getAuditLogs(buildQuery())
    logs.value = result.list
    pagination.total = result.total
    pagination.page = result.page
    pagination.per_page = result.per_page
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载失败'
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  void loadLogs()
}

function handleReset() {
  queryForm.admin_id = ''
  queryForm.action = ''
  queryForm.object_type = ''
  queryForm.object_id = ''
  queryForm.date_range = null
  pagination.page = 1
  void loadLogs()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadLogs()
}

function handlePageSizeChange(perPage: number) {
  pagination.per_page = perPage
  pagination.page = 1
  void loadLogs()
}

function formatAction(value: string) {
  const option = loginActionOptions.find((item) => item.value === value)
  return option ? option.label : value
}

function actorName(row: AuditLogItem) {
  if (!row.admin) return '未知管理员'
  return row.admin.display_name || row.admin.username
}

function actorMeta(row: AuditLogItem) {
  if (!row.admin) return '-'
  return row.admin.email || row.admin.username
}

function requestLine(row: AuditLogItem) {
  return `${row.request_method || '-'} ${row.request_path || '-'}`
}

const columns = computed<DataTableColumns<AuditLogItem>>(() => [
  {
    key: 'actor',
    title: '操作者',
    minWidth: 180,
    render: (row) =>
      h('div', { class: 'audit-logs-page__identity' }, [
        h('span', { class: 'audit-logs-page__primary' }, actorName(row)),
        h('span', { class: 'audit-logs-page__secondary' }, actorMeta(row)),
      ]),
  },
  {
    key: 'action',
    title: isLoginPage.value ? '认证动作' : '动作',
    minWidth: 190,
    ellipsis: { tooltip: true },
    render: (row) => (isLoginPage.value ? formatAction(row.action) : row.action),
  },
  {
    key: 'object',
    title: '对象',
    minWidth: 190,
    render: (row) =>
      h('div', { class: 'audit-logs-page__identity' }, [
        h('span', { class: 'audit-logs-page__primary' }, row.object_type),
        h('span', { class: 'audit-logs-page__secondary' }, row.object_id || '-'),
      ]),
  },
  {
    key: 'request',
    title: '请求',
    minWidth: 240,
    ellipsis: { tooltip: true },
    render: (row) => requestLine(row),
  },
  { key: 'ip', title: 'IP', minWidth: 130 },
  { key: 'remark', title: '备注', minWidth: 160, ellipsis: { tooltip: true } },
  {
    key: 'created_at',
    title: '时间',
    minWidth: 180,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    key: 'detail',
    title: '详情',
    width: 110,
    fixed: 'right',
    render: (row) =>
      h(
        NPopover,
        { placement: 'left', trigger: 'click', width: 420 },
        {
          trigger: () => h(NButton, { text: true, type: 'primary' }, { default: () => '查看' }),
          default: () =>
            h('div', { class: 'audit-logs-page__detail' }, [
              h('p', null, [h('strong', null, '请求 ID：'), row.request_id || '-']),
              h('p', null, [h('strong', null, '会话 ID：'), row.session_id || '-']),
              ...(canViewSensitive.value
                ? [
                    h('p', null, [h('strong', null, 'User-Agent：'), row.user_agent || '-']),
                    h('p', null, [h('strong', null, 'Before：')]),
                    h('pre', null, row.before_data || '-'),
                    h('p', null, [h('strong', null, 'After：')]),
                    h('pre', null, row.after_data || '-'),
                  ]
                : [
                    h(
                      NAlert,
                      { type: 'info', showIcon: true, title: '敏感详情需要 audit-log:sensitive-view 权限' },
                      undefined,
                    ),
                  ]),
            ]),
        },
      ),
  },
])

onMounted(() => {
  void loadLogs()
})

watch(isLoginPage, () => {
  handleReset()
})
</script>

<template>
  <div class="audit-logs-page">
    <div class="audit-logs-page__header">
      <div>
        <h2>{{ pageTitle }}</h2>
        <p>{{ pageDescription }}</p>
      </div>
      <NButton :loading="refreshing" @click="loadLogs({ refresh: true })">
        <template #icon>
          <NIcon><RefreshOutline /></NIcon>
        </template>
        刷新
      </NButton>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadLogs">
      <template v-if="!canViewLogs">
        <NCard>
          <EmptyState title="暂无权限" :description="`当前账号没有${pageTitle}查看权限。`" />
        </NCard>
      </template>

      <template v-else>
        <NCard :bordered="false">
          <NForm inline label-placement="left" class="audit-logs-page__filters" @submit.prevent>
            <template v-if="isLoginPage">
              <NFormItem label="管理员 ID">
                <NInput v-model:value="queryForm.admin_id" clearable placeholder="例如 1" @keyup.enter="handleSearch" />
              </NFormItem>
              <NFormItem label="动作类型">
                <NSelect
                  v-model:value="queryForm.action"
                  :options="loginActionOptions"
                  clearable
                  placeholder="全部认证日志"
                  style="min-width: 180px"
                />
              </NFormItem>
              <NFormItem label="时间">
                <NDatePicker v-model:value="queryForm.date_range" type="daterange" clearable />
              </NFormItem>
            </template>

            <template v-else>
              <NFormItem label="管理员 ID">
                <NInput v-model:value="queryForm.admin_id" clearable placeholder="例如 1" @keyup.enter="handleSearch" />
              </NFormItem>
              <NFormItem label="动作">
                <NInput v-model:value="queryForm.action" clearable placeholder="例如 admin.login.success" @keyup.enter="handleSearch" />
              </NFormItem>
              <NFormItem label="对象类型">
                <NInput v-model:value="queryForm.object_type" clearable placeholder="例如 admin_auth" @keyup.enter="handleSearch" />
              </NFormItem>
              <NFormItem label="对象 ID">
                <NInput v-model:value="queryForm.object_id" clearable placeholder="对象标识" @keyup.enter="handleSearch" />
              </NFormItem>
              <NFormItem label="时间">
                <NDatePicker v-model:value="queryForm.date_range" type="daterange" clearable />
              </NFormItem>
            </template>

            <NFormItem :show-label="false">
              <NSpace>
                <NButton type="primary" @click="handleSearch">
                  <template #icon>
                    <NIcon><SearchOutline /></NIcon>
                  </template>
                  查询
                </NButton>
                <NButton @click="handleReset">重置</NButton>
              </NSpace>
            </NFormItem>
          </NForm>
        </NCard>

        <NCard :bordered="false">
          <template v-if="hasLogs">
            <NDataTable
              :columns="columns"
              :data="logs"
              :row-key="(row: AuditLogItem) => row.id"
              striped
              :bordered="false"
            />
            <div class="audit-logs-page__pagination">
              <NPagination
                :page="pagination.page"
                :page-size="pagination.per_page"
                :item-count="pagination.total"
                :page-sizes="[15, 30, 50, 100]"
                show-size-picker
                @update:page="handlePageChange"
                @update:page-size="handlePageSizeChange"
              />
            </div>
          </template>

          <EmptyState
            v-else
            :title="isLoginPage ? '暂无登录安全记录' : '暂无操作审计记录'"
            :description="isLoginPage ? '当前筛选条件下没有可展示的登录安全记录。' : '当前筛选条件下没有可展示的后台操作审计记录。'"
          />
        </NCard>
      </template>
    </QueryState>
  </div>
</template>

<style scoped>
.audit-logs-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.audit-logs-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.audit-logs-page__header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.audit-logs-page__header p {
  margin: 8px 0 0;
  color: rgba(15, 23, 42, 0.55);
}

.audit-logs-page__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0 8px;
}

.audit-logs-page__identity {
  display: grid;
  gap: 4px;
}

.audit-logs-page__primary {
  font-weight: 600;
  color: rgba(15, 23, 42, 0.92);
}

.audit-logs-page__secondary {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}

.audit-logs-page__pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.audit-logs-page__detail {
  display: grid;
  gap: 8px;
}

.audit-logs-page__detail p {
  margin: 0;
  color: rgba(15, 23, 42, 0.78);
  word-break: break-all;
}

.audit-logs-page__detail pre {
  max-height: 160px;
  margin: 0;
  padding: 8px;
  overflow: auto;
  color: rgba(15, 23, 42, 0.92);
  background: rgba(15, 23, 42, 0.04);
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-all;
}

@media (max-width: 768px) {
  .audit-logs-page__header {
    flex-direction: column;
  }
}
</style>
