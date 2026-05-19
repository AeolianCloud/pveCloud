<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
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

async function loadTasks() {
  loading.value = true
  try {
    const data = await getAsyncTasks(query)
    tasks.value = data.list
    total.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '异步任务加载失败')
  } finally {
    loading.value = false
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

function resetQuery() {
  Object.assign(query, { page: 1, per_page: 15, task_type: '', status: '', object_type: '', object_no: '' })
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
    render: (row) => row.last_error_message || '-',
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
    width: 100,
    fixed: 'right',
    render: (row) =>
      row.status === 'failed' && canRetry.value
        ? h(NButton, { text: true, type: 'primary', onClick: () => retry(row) }, { default: () => '重试' })
        : '-',
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
          <NInput v-model:value="query.task_type" clearable placeholder="task_type" />
        </NFormItem>
        <NFormItem label="对象">
          <NSpace>
            <NInput v-model:value="query.object_type" clearable placeholder="object_type" />
            <NInput v-model:value="query.object_no" clearable placeholder="object_no" />
          </NSpace>
        </NFormItem>
        <NFormItem :show-label="false">
          <NSpace>
            <NButton type="primary" @click="query.page = 1; loadTasks()">查询</NButton>
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
