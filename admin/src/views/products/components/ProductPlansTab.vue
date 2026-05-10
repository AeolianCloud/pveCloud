<script setup lang="ts">
import { NButton, NDataTable, NSpace, NTag, type DataTableColumns } from 'naive-ui'
import { computed, h } from 'vue'

import type { ProductItem, ProductPlanItem } from '../../../api/product-catalog'

const props = defineProps<{
  plans: ProductPlanItem[]
  productsById: Record<number, ProductItem>
  loading: boolean
  planPublishIssues: (item: ProductPlanItem) => string[]
  statusLabel: (status: string) => string
  statusTagType: (status: string) => string
}>()

const emit = defineEmits<{
  create: []
  edit: [item: ProductPlanItem]
  prices: [item: ProductPlanItem]
  relations: [item: ProductPlanItem]
  toggleStatus: [item: ProductPlanItem]
  delete: [item: ProductPlanItem]
}>()

type TagType = 'success' | 'warning' | 'error' | 'info' | 'default'

const columns = computed<DataTableColumns<ProductPlanItem>>(() => [
  { key: 'name', title: '套餐', minWidth: 180 },
  {
    key: 'product_id',
    title: '产品',
    minWidth: 140,
    render: (row) => props.productsById[row.product_id]?.name || String(row.product_id),
  },
  {
    key: 'spec',
    title: '配置',
    minWidth: 220,
    render: (row) => `${row.cpu_cores}C / ${(row.memory_mb / 1024).toFixed(0)}G / ${row.system_disk_gb}G`,
  },
  {
    key: 'status',
    title: '状态',
    width: 120,
    render: (row) =>
      h(NTag, { type: props.statusTagType(row.status) as TagType, size: 'small' }, { default: () => props.statusLabel(row.status) }),
  },
  {
    key: 'is_featured',
    title: '推荐',
    width: 90,
    render: (row) => (row.is_featured ? '是' : '否'),
  },
  {
    key: 'publish_issues',
    title: '公开检查',
    minWidth: 240,
    render: (row) => {
      const issues = props.planPublishIssues(row)
      if (issues.length === 0) {
        return h(NTag, { type: 'success', size: 'small' }, { default: () => 'Web 可展示' })
      }
      return h(NSpace, { wrap: true, size: 4 }, {
        default: () => issues.map((issue) =>
          h(NTag, { key: issue, type: 'warning', size: 'small' }, { default: () => issue }),
        ),
      })
    },
  },
  {
    key: 'actions',
    title: '操作',
    width: 380,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => emit('edit', row) }, { default: () => '编辑' }),
          h(NButton, { text: true, type: 'success', onClick: () => emit('prices', row) }, { default: () => '价格' }),
          h(NButton, { text: true, type: 'success', onClick: () => emit('relations', row) }, { default: () => '关联' }),
          h(NButton, { text: true, type: 'warning', onClick: () => emit('toggleStatus', row) }, { default: () => '切换状态' }),
          h(NButton, { text: true, type: 'error', onClick: () => emit('delete', row) }, { default: () => '删除' }),
        ],
      }),
  },
])
</script>

<template>
  <div class="toolbar">
    <NButton type="primary" @click="$emit('create')">新增套餐</NButton>
  </div>
  <NDataTable
    :columns="columns"
    :data="plans"
    :loading="loading"
    :row-key="(row: ProductPlanItem) => row.id"
    striped
    bordered
  />
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
