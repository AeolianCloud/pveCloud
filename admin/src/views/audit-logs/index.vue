<script setup lang="ts">
import { Refresh, Search } from '@element-plus/icons-vue'
import { computed, onMounted, reactive, ref } from 'vue'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import { getAuditLogs, type AuditLogItem, type AuditLogListQuery } from '../../api/audit-log'
import { usePermissionStore } from '../../store/modules/permission'

const permissionStore = usePermissionStore()
const loading = ref(false)
const refreshing = ref(false)
const errorMessage = ref('')
const logs = ref<AuditLogItem[]>([])
const pagination = reactive({
  page: 1,
  per_page: 15,
  total: 0,
})
const queryForm = reactive({
  admin_id: '',
  action: '',
  object_type: '',
  object_id: '',
  date_range: [] as string[],
})

const canViewLogs = computed(() => permissionStore.hasPermission('page.system-settings.audit-logs'))
const canViewSensitive = computed(() => permissionStore.hasPermission('audit-log:sensitive-view'))
const hasLogs = computed(() => logs.value.length > 0)

function buildQuery(): AuditLogListQuery {
  const [dateFrom, dateTo] = queryForm.date_range
  return {
    page: pagination.page,
    per_page: pagination.per_page,
    admin_id: queryForm.admin_id ? Number(queryForm.admin_id) : undefined,
    action: queryForm.action.trim() || undefined,
    object_type: queryForm.object_type.trim() || undefined,
    object_id: queryForm.object_id.trim() || undefined,
    date_from: dateFrom || undefined,
    date_to: dateTo || undefined,
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
  queryForm.date_range = []
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

function formatDateTime(value: string | null | undefined) {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date)
}

function actorName(row: AuditLogItem) {
  if (!row.admin) {
    return '未知管理员'
  }
  return row.admin.display_name || row.admin.username
}

function actorMeta(row: AuditLogItem) {
  if (!row.admin) {
    return '-'
  }
  return row.admin.email || row.admin.username
}

function requestLine(row: AuditLogItem) {
  const method = row.request_method || '-'
  const path = row.request_path || '-'
  return `${method} ${path}`
}

onMounted(() => {
  void loadLogs()
})
</script>

<template>
  <div class="audit-logs-page">
    <div class="audit-logs-page__header">
      <div>
        <h2>操作日志</h2>
        <p>查看普通后台操作记录，帮助追踪谁在什么时候操作了后台。</p>
      </div>
      <el-button :icon="Refresh" :loading="refreshing" @click="loadLogs({ refresh: true })">刷新</el-button>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadLogs">
      <template v-if="!canViewLogs">
        <el-card>
          <EmptyState title="暂无权限" description="当前账号没有操作日志查看权限。" />
        </el-card>
      </template>

      <template v-else>
        <el-card shadow="never" class="audit-logs-page__filter-card">
          <el-form inline class="audit-logs-page__filters" @submit.prevent>
            <el-form-item label="管理员 ID">
              <el-input
                v-model="queryForm.admin_id"
                clearable
                placeholder="例如 1"
                @keyup.enter="handleSearch"
              />
            </el-form-item>
            <el-form-item label="动作">
              <el-input
                v-model="queryForm.action"
                clearable
                placeholder="例如 admin.login.success"
                @keyup.enter="handleSearch"
              />
            </el-form-item>
            <el-form-item label="对象类型">
              <el-input
                v-model="queryForm.object_type"
                clearable
                placeholder="例如 admin_auth"
                @keyup.enter="handleSearch"
              />
            </el-form-item>
            <el-form-item label="对象 ID">
              <el-input
                v-model="queryForm.object_id"
                clearable
                placeholder="对象标识"
                @keyup.enter="handleSearch"
              />
            </el-form-item>
            <el-form-item label="时间">
              <el-date-picker
                v-model="queryForm.date_range"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never">
          <template v-if="hasLogs">
            <el-table :data="logs" stripe class="audit-logs-page__table">
              <el-table-column label="操作者" min-width="180">
                <template #default="{ row }">
                  <div class="audit-logs-page__identity">
                    <span class="audit-logs-page__primary">{{ actorName(row) }}</span>
                    <span class="audit-logs-page__secondary">{{ actorMeta(row) }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="动作" prop="action" min-width="190" show-overflow-tooltip />
              <el-table-column label="对象" min-width="190" show-overflow-tooltip>
                <template #default="{ row }">
                  <div class="audit-logs-page__identity">
                    <span class="audit-logs-page__primary">{{ row.object_type }}</span>
                    <span class="audit-logs-page__secondary">{{ row.object_id || '-' }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="请求" min-width="240" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ requestLine(row) }}
                </template>
              </el-table-column>
              <el-table-column label="IP" prop="ip" min-width="130" />
              <el-table-column label="备注" prop="remark" min-width="160" show-overflow-tooltip />
              <el-table-column label="时间" min-width="180">
                <template #default="{ row }">
                  {{ formatDateTime(row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column label="详情" width="110" fixed="right">
                <template #default="{ row }">
                  <el-popover placement="left" width="420" trigger="click">
                    <template #reference>
                      <el-button link type="primary">查看</el-button>
                    </template>
                    <div class="audit-logs-page__detail">
                      <p><strong>请求 ID：</strong>{{ row.request_id || '-' }}</p>
                      <p><strong>会话 ID：</strong>{{ row.session_id || '-' }}</p>
                      <p v-if="canViewSensitive"><strong>User-Agent：</strong>{{ row.user_agent || '-' }}</p>
                      <template v-if="canViewSensitive">
                        <p><strong>Before：</strong></p>
                        <pre>{{ row.before_data || '-' }}</pre>
                        <p><strong>After：</strong></p>
                        <pre>{{ row.after_data || '-' }}</pre>
                      </template>
                      <el-alert
                        v-else
                        type="info"
                        :closable="false"
                        show-icon
                        title="敏感详情需要 audit-log:sensitive-view 权限"
                      />
                    </div>
                  </el-popover>
                </template>
              </el-table-column>
            </el-table>

            <div class="audit-logs-page__pagination">
              <el-pagination
                background
                layout="total, sizes, prev, pager, next"
                :current-page="pagination.page"
                :page-size="pagination.per_page"
                :page-sizes="[15, 30, 50, 100]"
                :total="pagination.total"
                @current-change="handlePageChange"
                @size-change="handlePageSizeChange"
              />
            </div>
          </template>

          <EmptyState
            v-else
            title="暂无操作日志"
            description="当前筛选条件下没有可展示的后台操作记录。"
          />
        </el-card>
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
  color: var(--el-text-color-secondary);
}

.audit-logs-page__filter-card {
  overflow: visible;
}

.audit-logs-page__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0 8px;
}

.audit-logs-page__table {
  width: 100%;
}

.audit-logs-page__identity {
  display: grid;
  gap: 4px;
}

.audit-logs-page__primary {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.audit-logs-page__secondary {
  color: var(--el-text-color-secondary);
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
  color: var(--el-text-color-regular);
  word-break: break-all;
}

.audit-logs-page__detail pre {
  max-height: 160px;
  margin: 0;
  padding: 8px;
  overflow: auto;
  color: var(--el-text-color-primary);
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-all;
}

@media (max-width: 768px) {
  .audit-logs-page__header {
    flex-direction: column;
  }

  .audit-logs-page__pagination {
    justify-content: flex-start;
    overflow-x: auto;
  }
}
</style>
