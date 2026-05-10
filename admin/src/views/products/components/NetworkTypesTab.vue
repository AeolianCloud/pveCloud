<script setup lang="ts">
import { NButton, NDataTable, NSpace, NTag, type DataTableColumns } from 'naive-ui'
import { computed, h } from 'vue'

import type { NetworkTypeItem } from '../../../api/product-catalog'

const props = defineProps<{
  networkTypes: NetworkTypeItem[]
  loading: boolean
  statusLabel: (status: string) => string
  statusTagType: (status: string) => string
}>()

const emit = defineEmits<{
  create: []
  edit: [item: NetworkTypeItem]
  delete: [item: NetworkTypeItem]
}>()

type TagType = 'success' | 'warning' | 'error' | 'info' | 'default'

const columns = computed<DataTableColumns<NetworkTypeItem>>(() => [
  { key: 'name', title: '名称', minWidth: 180 },
  { key: 'code', title: 'Code', minWidth: 140 },
  { key: 'summary', title: '简介', minWidth: 220 },
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
  {
    key: 'actions',
    title: '操作',
    width: 180,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => emit('edit', row) }, { default: () => '编辑' }),
          h(NButton, { text: true, type: 'error', onClick: () => emit('delete', row) }, { default: () => '删除' }),
        ],
      }),
  },
])
</script>

<template>
  <div class="toolbar">
    <NButton type="primary" @click="$emit('create')">新增网络类型</NButton>
  </div>
  <NDataTable
    :columns="columns"
    :data="networkTypes"
    :loading="loading"
    :row-key="(row: NetworkTypeItem) => row.id"
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
