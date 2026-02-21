<script setup lang="ts">
// components/TableCard.vue
// 封装「表格卡片 + 分页」公共结构，减少各页面重复的 n-card / n-data-table / n-pagination 代码。
//
// 使用示例：
//   <TableCard
//     :columns="columns"
//     :data="tableData"
//     :loading="loading"
//     :scroll-x="scrollX"
//     :total="total"
//     v-model:page="query.page_num"
//     v-model:page-size="query.page_size"
//     empty-text="暂无数据"
//     @load="loadData"
//   />
import type { DataTableColumns } from 'naive-ui'
import EmptyState from './EmptyState.vue'

defineProps<{
  // 表格列定义
  columns: DataTableColumns<any>
  // 表格数据
  data: any[]
  // 加载状态
  loading?: boolean
  // 横向滚动最小宽度，undefined 时不启用横向滚动
  scrollX?: number
  // 总记录数
  total: number
  // 当前页码（v-model:page）
  page: number
  // 每页条数（v-model:page-size）
  pageSize: number
  // 空状态提示文字
  emptyText?: string
  // 每页条数可选项
  pageSizes?: number[]
}>()

const emit = defineEmits<{
  // 页码变化
  'update:page': [page: number]
  // 每页条数变化
  'update:pageSize': [size: number]
  // 触发重新加载（页码/页大小变化后由组件内部调用）
  'load': []
}>()

function onPageChange(page: number) {
  emit('update:page', page)
  emit('load')
}

function onPageSizeChange(size: number) {
  emit('update:page', 1)
  emit('update:pageSize', size)
  emit('load')
}
</script>

<template>
  <n-card :bordered="false" class="table-card">
    <n-data-table
      :columns="columns"
      :data="data"
      :loading="loading"
      :pagination="false"
      :scroll-x="scrollX"
      size="small"
      striped
    >
      <template #empty>
        <EmptyState :description="emptyText" />
      </template>
    </n-data-table>

    <div class="pagination">
      <n-pagination
        :page="page"
        :page-size="pageSize"
        :item-count="total"
        :page-sizes="pageSizes ?? [20, 50, 100]"
        show-size-picker
        show-quick-jumper
        @update:page="onPageChange"
        @update:page-size="onPageSizeChange"
      />
    </div>
  </n-card>
</template>

<style scoped>
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
