<script setup lang="ts">
import { NButton, NDataTable, NSpace, NTag, type DataTableColumns } from 'naive-ui'
import { computed, h } from 'vue'

import type { ProductItem } from '../../../api/product-catalog'

const props = defineProps<{
  products: ProductItem[]
  loading: boolean
  statusLabel: (status: string) => string
  statusTagType: (status: string) => string
}>()

const emit = defineEmits<{
  create: []
  edit: [item: ProductItem]
  toggleStatus: [item: ProductItem]
  delete: [item: ProductItem]
}>()

type TagType = 'success' | 'warning' | 'error' | 'info' | 'default'

const columns = computed<DataTableColumns<ProductItem>>(() => [
  { key: 'name', title: '名称', minWidth: 180 },
  { key: 'slug', title: 'Slug', minWidth: 160 },
  {
    key: 'status',
    title: '状态',
    width: 120,
    render: (row) =>
      h(NTag, { type: props.statusTagType(row.status) as TagType, size: 'small' }, { default: () => props.statusLabel(row.status) }),
  },
  {
    key: 'visible',
    title: '展示',
    width: 90,
    render: (row) => (row.visible ? '是' : '否'),
  },
  { key: 'sort_order', title: '排序', width: 90 },
  {
    key: 'actions',
    title: '操作',
    width: 280,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => emit('edit', row) }, { default: () => '编辑' }),
          h(NButton, { text: true, type: 'warning', onClick: () => emit('toggleStatus', row) }, { default: () => '切换状态' }),
          h(NButton, { text: true, type: 'error', onClick: () => emit('delete', row) }, { default: () => '删除' }),
        ],
      }),
  },
])
</script>

<template>
  <div class="toolbar">
    <NButton type="primary" @click="$emit('create')">新增产品</NButton>
  </div>
  <NDataTable
    :columns="columns"
    :data="products"
    :loading="loading"
    :row-key="(row: ProductItem) => row.id"
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
