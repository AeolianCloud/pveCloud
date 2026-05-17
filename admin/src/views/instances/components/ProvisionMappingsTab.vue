<script setup lang="ts">
import { NButton, NDataTable, NForm, NFormItem, NInput, NPagination, NSelect, NSpace, NTag, type DataTableColumns } from 'naive-ui'
import { computed, h } from 'vue'

import type { InstanceMappingItem } from '../../../api/instance'
import { formatDateTime } from '../../../utils/datetime'
import { mappingStatusText } from '../types'

const props = defineProps<{
  loading: boolean
  items: InstanceMappingItem[]
  total: number
  query: { page: number; per_page: number; status: string; plan_no: string; region_no: string; template_no: string; network_type_no: string }
  canProvision: boolean
}>()

const emit = defineEmits<{
  search: []
  reset: []
  create: []
  edit: [item: InstanceMappingItem]
}>()

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
]

const columns = computed<DataTableColumns<InstanceMappingItem>>(() => [
  { key: 'mapping_no', title: '映射编号', minWidth: 170 },
  {
    key: 'scope',
    title: '匹配范围',
    minWidth: 240,
    render: (row) =>
      h('div', null, [
        h('div', { class: 'strong' }, row.plan_no),
        h('div', { class: 'muted' }, `${row.region_no} · ${row.template_no} · ${row.network_type_no || '不限网络'}`),
      ]),
  },
  {
    key: 'target',
    title: '目标资源',
    minWidth: 220,
    render: (row) =>
      h('div', null, [
        h('div', { class: 'strong' }, `${row.node} / ${row.storage}`),
        h('div', { class: 'muted' }, row.disk_source),
      ]),
  },
  {
    key: 'vmid',
    title: '编号范围',
    minWidth: 150,
    render: (row) => `${row.next_vmid} / ${row.vmid_start}-${row.vmid_end}`,
  },
  {
    key: 'status',
    title: '状态',
    width: 90,
    render: (row) => h(NTag, { size: 'small', type: row.status === 'active' ? 'success' : 'default' }, { default: () => mappingStatusText[row.status] }),
  },
  { key: 'updated_at', title: '更新时间', minWidth: 170, render: (row) => formatDateTime(row.updated_at) },
  {
    key: 'actions',
    title: '操作',
    width: 100,
    fixed: 'right',
    render: (row) => props.canProvision ? h(NButton, { text: true, type: 'primary', onClick: () => emit('edit', row) }, { default: () => '编辑' }) : '-',
  },
])
</script>

<template>
  <div>
    <div class="toolbar">
      <NForm inline label-placement="left" class="query-form">
        <NFormItem label="状态"><NSelect v-model:value="query.status" :options="statusOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
        <NFormItem label="套餐"><NInput v-model:value="query.plan_no" clearable placeholder="套餐编号" /></NFormItem>
        <NFormItem label="地域"><NInput v-model:value="query.region_no" clearable placeholder="地域编号" /></NFormItem>
        <NFormItem label="模板"><NInput v-model:value="query.template_no" clearable placeholder="模板编号" /></NFormItem>
        <NFormItem :show-label="false">
          <NSpace><NButton type="primary" @click="emit('search')">查询</NButton><NButton @click="emit('reset')">重置</NButton></NSpace>
        </NFormItem>
      </NForm>
      <NButton v-if="canProvision" type="primary" @click="emit('create')">新增映射</NButton>
    </div>

    <NDataTable :loading="loading" :columns="columns" :data="items" :row-key="(row: InstanceMappingItem) => row.id" :bordered="false" />

    <div class="pagination">
      <NPagination v-model:page="query.page" v-model:page-size="query.per_page" :item-count="total" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="emit('search')" @update:page-size="emit('search')" />
    </div>
  </div>
</template>
