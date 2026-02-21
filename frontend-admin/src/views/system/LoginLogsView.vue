<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listLoginLogs } from '@/api/loginLog'
import type { LoginLog } from '@/api/loginLog'

const message = useMessage()

// ── 列表状态 ──────────────────────────────────────────────
const loading = ref(false)
const tableData = ref<LoginLog[]>([])
const total = ref(0)

const query = reactive({
  username: '',
  status: '' as number | '',
  page_num: 1,
  page_size: 20,
})

const statusOptions = [
  { label: '全部', value: '' },
  { label: '成功', value: 1 },
  { label: '失败', value: 0 },
]

async function loadData() {
  loading.value = true
  try {
    // status 为空字符串时不传该参数
    const params = {
      page_num: query.page_num,
      page_size: query.page_size,
      username: query.username || undefined,
      status: query.status === '' ? undefined : query.status,
    }
    const res = await listLoginLogs(params)
    const { list, total: t } = res.data.data
    tableData.value = list
    total.value = t
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() { query.page_num = 1; loadData() }
function handleReset() {
  query.username = ''
  query.status = ''
  query.page_num = 1
  loadData()
}

// 格式化时间
function formatTime(t: string) {
  return new Date(t).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

// 简化 User-Agent，只取浏览器名+版本
function shortUA(ua: string) {
  if (!ua) return '—'
  const match = ua.match(/(Chrome|Firefox|Safari|Edge|Edg)\/[\d.]+/)
  return match ? match[0].replace('Edg/', 'Edge/') : ua.slice(0, 40)
}

// ── 表格列 ────────────────────────────────────────────────
const columns: DataTableColumns<LoginLog> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户名', key: 'username', width: 140 },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(NTag, {
        type: row.status === 1 ? 'success' : 'error',
        size: 'small',
        bordered: false,
      }, { default: () => (row.status === 1 ? '成功' : '失败') }),
  },
  { title: 'IP', key: 'ip', width: 140 },
  {
    title: '浏览器',
    key: 'user_agent',
    render: (row) => shortUA(row.user_agent),
  },
  {
    title: '备注',
    key: 'remark',
    width: 160,
    render: (row) => row.remark || '—',
  },
  {
    title: '登录时间',
    key: 'created_at',
    width: 180,
    render: (row) => formatTime(row.created_at),
  },
]

onMounted(loadData)
</script>

<template>
  <div class="login-logs">
    <!-- 搜索栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <n-input
          v-model:value="query.username"
          placeholder="按用户名搜索"
          clearable
          style="width: 200px;"
          @keyup.enter="handleSearch"
        />
        <n-select
          v-model:value="query.status"
          :options="statusOptions"
          style="width: 120px;"
          @update:value="handleSearch"
        />
        <n-button @click="handleSearch">搜索</n-button>
        <n-button @click="handleReset">重置</n-button>
      </div>
      <span class="total-tip">共 {{ total }} 条记录</span>
    </div>

    <!-- 数据表格 -->
    <n-card :bordered="false" class="table-card">
      <n-data-table
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :pagination="false"
        :scroll-x="1000"
        size="small"
        striped
      />

      <div class="pagination">
        <n-pagination
          v-model:page="query.page_num"
          v-model:page-size="query.page_size"
          :item-count="total"
          :page-sizes="[20, 50, 100]"
          show-size-picker
          show-quick-jumper
          @update:page="loadData"
          @update:page-size="() => { query.page_num = 1; loadData() }"
        />
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.login-logs {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  padding: 16px 20px;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.total-tip {
  font-size: 13px;
  color: #909399;
}

.table-card {
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
