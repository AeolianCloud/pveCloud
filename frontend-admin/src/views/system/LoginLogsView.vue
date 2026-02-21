<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listLoginLogs } from '@/api/loginLog'
import type { LoginLog } from '@/api/loginLog'
import { useTableScroll } from '@/composables/useTableScroll'
import { formatTime } from '@/utils/format'
import TableCard from '@/components/TableCard.vue'

const message = useMessage()

const { scrollX } = useTableScroll(1000)

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
  <div class="page-container">
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
    <TableCard
      :columns="columns"
      :data="tableData"
      :loading="loading"
      :scroll-x="scrollX"
      :total="total"
      :page="query.page_num"
      :page-size="query.page_size"
      empty-text="暂无登录记录"
      @update:page="(p) => (query.page_num = p)"
      @update:page-size="(s) => (query.page_size = s)"
      @load="loadData"
    />
  </div>
</template>

<style scoped>
.page-container {
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
</style>
