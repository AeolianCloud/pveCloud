<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag, NButton } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listOpLogs } from '@/api/opLog'
import type { OpLog } from '@/api/opLog'

const message = useMessage()

// ── 列表状态 ──────────────────────────────────────────────
const loading = ref(false)
const tableData = ref<OpLog[]>([])
const total = ref(0)

const query = reactive({
  username: '',
  module: '' as string,
  action: '' as string,
  page_num: 1,
  page_size: 20,
})

// 模块选项
const moduleOptions = [
  { label: '全部模块', value: '' },
  { label: '管理员账号', value: 'admin' },
  { label: '角色管理', value: 'role' },
]

// 动作选项
const actionOptions = [
  { label: '全部动作', value: '' },
  { label: '创建', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '状态变更', value: 'set_status' },
  { label: '分配权限', value: 'assign_permissions' },
]

async function loadData() {
  loading.value = true
  try {
    const params = {
      page_num: query.page_num,
      page_size: query.page_size,
      username: query.username || undefined,
      module: query.module || undefined,
      action: query.action || undefined,
    }
    const res = await listOpLogs(params)
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
  query.module = ''
  query.action = ''
  query.page_num = 1
  loadData()
}

// 模块 → 前端路由路径映射（新增模块时在此追加）
const moduleRouteMap: Record<string, string> = {
  admin: '/system/admin-users',
  role:  '/system/roles',
}

// 模块中文名映射
const moduleLabelMap: Record<string, string> = {
  admin: '管理员账号',
  role:  '角色管理',
}

// 动作中文名映射
const actionLabelMap: Record<string, string> = {
  create:             '创建',
  update:             '更新',
  delete:             '删除',
  set_status:         '状态变更',
  assign_permissions: '分配权限',
}

// 动作对应的 Tag 颜色
const actionTagType: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  create:             'success',
  update:             'info',
  delete:             'error',
  set_status:         'warning',
  assign_permissions: 'info',
}

function formatTime(t: string) {
  return new Date(t).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

// ── 表格列 ────────────────────────────────────────────────
const columns: DataTableColumns<OpLog> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '操作人', key: 'username', width: 120 },
  {
    title: '模块',
    key: 'module',
    width: 140,
    render: (row) => {
      const label = moduleLabelMap[row.module] ?? row.module
      const route = moduleRouteMap[row.module]
      // 有对应路由则渲染为可点击链接，点击在新标签页打开
      if (route) {
        return h(
          NButton,
          {
            text: true,
            type: 'primary',
            size: 'small',
            onClick: () => window.open(route, '_blank'),
          },
          { default: () => label },
        )
      }
      return label
    },
  },
  {
    title: '动作',
    key: 'action',
    width: 120,
    render: (row) =>
      h(NTag, {
        type: actionTagType[row.action] ?? 'default',
        size: 'small',
        bordered: false,
      }, { default: () => actionLabelMap[row.action] ?? row.action }),
  },
  {
    title: '目标 ID',
    key: 'target_id',
    width: 90,
    render: (row) => row.target_id || '—',
  },
  { title: 'IP', key: 'ip', width: 140 },
  {
    title: '操作时间',
    key: 'created_at',
    width: 180,
    render: (row) => formatTime(row.created_at),
  },
]

onMounted(loadData)
</script>

<template>
  <div class="op-logs">
    <!-- 搜索栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <n-input
          v-model:value="query.username"
          placeholder="按操作人搜索"
          clearable
          style="width: 180px;"
          @keyup.enter="handleSearch"
        />
        <n-select
          v-model:value="query.module"
          :options="moduleOptions"
          style="width: 130px;"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="query.action"
          :options="actionOptions"
          style="width: 130px;"
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
        :scroll-x="900"
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
.op-logs {
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
