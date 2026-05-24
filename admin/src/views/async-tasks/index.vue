<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NPagination,
  NSelect,
  NSpace,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import { getAsyncTasks, retryAsyncTask, type AsyncTaskItem } from '../../api/async-task'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'
import { confirm, message } from '../../utils/feedback'
import { hasPermissionCode } from '../../utils/permission'

const permissionStore = usePermissionStore()
const loading = ref(false)
const tasks = ref<AsyncTaskItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, task_type: '', status: '', object_type: '', object_no: '' })
const dateRange = ref<[number, number] | null>(null)
const selectedTask = ref<AsyncTaskItem | null>(null)
const detailVisible = ref(false)

const canRetry = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'async-task:retry'))

const statusText: Record<string, string> = {
  pending: '待执行',
  running: '执行中',
  succeeded: '成功',
  failed: '失败',
  cancelled: '已取消',
}

const statusOptions = [
  { label: '待执行', value: 'pending' },
  { label: '执行中', value: 'running' },
  { label: '成功', value: 'succeeded' },
  { label: '失败', value: 'failed' },
  { label: '已取消', value: 'cancelled' },
]

const taskTypeOptions = [
  { label: '实例操作同步', value: 'instance_operation_sync' },
  { label: '实例到期提醒', value: 'instance_expiry_notice' },
  { label: '实例到期释放', value: 'instance_expiry_release' },
  { label: '支付后自动交付', value: 'payment_order_provision' },
  { label: '邮件通知发送', value: 'notification_email_send' },
  { label: '短信通知占位', value: 'notification_sms_placeholder' },
]

function queryParams() {
  const params: Record<string, unknown> = { ...query }
  if (dateRange.value) {
    params.date_from = new Date(dateRange.value[0]).toISOString()
    params.date_to = new Date(dateRange.value[1]).toISOString()
  }
  return params
}

async function loadTasks() {
  loading.value = true
  try {
    const data = await getAsyncTasks(queryParams())
    tasks.value = data.list
    total.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '异步任务加载失败')
  } finally {
    loading.value = false
  }
}

async function copyText(text: string) {
  if (!text) return
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      // Clipboard API may be unavailable on non-secure origins; keep a local fallback for ops consoles.
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    message.success('已复制错误信息')
  } catch {
    message.error('复制失败，请手动选择错误信息')
  }
}

async function retry(row: AsyncTaskItem) {
  try {
    await confirm({ title: '重试异步任务', content: `确认重试任务 ${row.task_no}？`, positiveText: '确认重试', type: 'warning' })
  } catch {
    return
  }
  try {
    await retryAsyncTask(row.task_no)
    message.success('任务已重新进入待执行队列')
    await loadTasks()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '任务重试失败')
  }
}

function openDetail(row: AsyncTaskItem) {
  selectedTask.value = row
  detailVisible.value = true
}

function errorSummary(row: AsyncTaskItem) {
  return [row.last_error_code, row.last_error_message].filter(Boolean).join(' / ')
}

function filterFailedTasks() {
  query.page = 1
  query.status = 'failed'
  void loadTasks()
}

function resetQuery() {
  Object.assign(query, { page: 1, per_page: 15, task_type: '', status: '', object_type: '', object_no: '' })
  dateRange.value = null
  void loadTasks()
}

const columns = computed<DataTableColumns<AsyncTaskItem>>(() => [
  { key: 'task_no', title: '任务号', minWidth: 180 },
  { key: 'task_type', title: '类型', minWidth: 190 },
  {
    key: 'status',
    title: '状态',
    width: 100,
    render: (row) => h(NTag, { size: 'small', type: row.status === 'failed' ? 'error' : row.status === 'succeeded' ? 'success' : 'default' }, { default: () => statusText[row.status] || row.status }),
  },
  {
    key: 'object',
    title: '业务对象',
    minWidth: 160,
    render: (row) => `${row.object_type || '-'} / ${row.object_no || '-'}`,
  },
  {
    key: 'attempts',
    title: '尝试',
    width: 90,
    render: (row) => `${row.attempts}/${row.max_attempts}`,
  },
  {
    key: 'scheduled_at',
    title: '计划时间',
    minWidth: 170,
    render: (row) => formatDateTime(row.scheduled_at),
  },
  {
    key: 'last_error_message',
    title: '最近错误',
    minWidth: 220,
    ellipsis: { tooltip: true },
    render: (row) => {
      const summary = errorSummary(row)
      if (!summary) return '-'
      return h(NSpace, { size: 8, align: 'center', wrap: false }, {
        default: () => [
          h('span', summary),
          h(NButton, { text: true, type: 'primary', onClick: () => copyText(summary) }, { default: () => '复制' }),
        ],
      })
    },
  },
  {
    key: 'locked_by',
    title: 'Worker',
    minWidth: 140,
    render: (row) => row.locked_by || '-',
  },
  {
    key: 'actions',
    title: '操作',
    width: 150,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row) }, { default: () => '详情' }),
          row.status === 'failed' && canRetry.value
            ? h(NButton, { text: true, type: 'primary', onClick: () => retry(row) }, { default: () => '重试' })
            : null,
        ],
      }),
  },
])

onMounted(loadTasks)
</script>

<template>
  <div class="async-tasks-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>异步任务</h2>
          <p class="muted">查看 Worker 后台任务执行状态，并对失败任务执行人工重试。</p>
        </div>
      </template>

      <NForm inline label-placement="left" class="query-form">
        <NFormItem label="状态">
          <NSelect v-model:value="query.status" :options="statusOptions" clearable placeholder="全部" style="width: 140px" />
        </NFormItem>
        <NFormItem label="类型">
          <NSelect v-model:value="query.task_type" :options="taskTypeOptions" clearable filterable placeholder="全部类型" style="width: 190px" />
        </NFormItem>
        <NFormItem label="对象">
          <NSpace>
            <NInput v-model:value="query.object_type" clearable placeholder="object_type" />
            <NInput v-model:value="query.object_no" clearable placeholder="object_no" />
          </NSpace>
        </NFormItem>
        <NFormItem label="创建时间">
          <NDatePicker v-model:value="dateRange" type="datetimerange" clearable style="width: 330px" />
        </NFormItem>
        <NFormItem :show-label="false">
          <NSpace>
            <NButton type="primary" @click="query.page = 1; loadTasks()">查询</NButton>
            <NButton type="error" secondary @click="filterFailedTasks">失败任务</NButton>
            <NButton @click="resetQuery">重置</NButton>
          </NSpace>
        </NFormItem>
      </NForm>

      <NDataTable
        :loading="loading"
        :columns="columns"
        :data="tasks"
        :row-key="(row: AsyncTaskItem) => row.task_no"
        :bordered="false"
      />

      <div class="pagination">
        <NPagination
          v-model:page="query.page"
          v-model:page-size="query.per_page"
          :item-count="total"
          show-size-picker
          :page-sizes="[10, 15, 20, 50]"
          @update:page="loadTasks"
          @update:page-size="loadTasks"
        />
      </div>
    </NCard>

    <NDrawer v-model:show="detailVisible" :width="520" placement="right">
      <NDrawerContent title="任务详情">
        <NDescriptions v-if="selectedTask" :column="1" label-placement="left" bordered size="small">
          <NDescriptionsItem label="任务号">{{ selectedTask.task_no }}</NDescriptionsItem>
          <NDescriptionsItem label="类型">{{ selectedTask.task_type }}</NDescriptionsItem>
          <NDescriptionsItem label="状态">{{ statusText[selectedTask.status] || selectedTask.status }}</NDescriptionsItem>
          <NDescriptionsItem label="业务对象">{{ selectedTask.object_type || '-' }} / {{ selectedTask.object_no || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="计划时间">{{ formatDateTime(selectedTask.scheduled_at) }}</NDescriptionsItem>
          <NDescriptionsItem label="尝试次数">{{ selectedTask.attempts }} / {{ selectedTask.max_attempts }}</NDescriptionsItem>
          <NDescriptionsItem label="锁定 Worker">{{ selectedTask.locked_by || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="锁定到">{{ formatDateTime(selectedTask.locked_until) }}</NDescriptionsItem>
          <NDescriptionsItem label="最近错误码">{{ selectedTask.last_error_code || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="最近错误">
            <NSpace align="center">
              <span>{{ selectedTask.last_error_message || '-' }}</span>
              <NButton v-if="errorSummary(selectedTask)" size="small" text type="primary" @click="copyText(errorSummary(selectedTask))">复制</NButton>
            </NSpace>
          </NDescriptionsItem>
          <NDescriptionsItem label="创建时间">{{ formatDateTime(selectedTask.created_at) }}</NDescriptionsItem>
          <NDescriptionsItem label="完成时间">{{ formatDateTime(selectedTask.completed_at) }}</NDescriptionsItem>
        </NDescriptions>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.page-header h2 {
  margin: 0;
  font-size: 20px;
}
.muted {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}
.query-form {
  margin-bottom: 16px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
