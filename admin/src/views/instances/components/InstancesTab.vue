<script setup lang="ts">
import {
  NButton,
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
import { computed, h } from 'vue'

import type { InstanceItem } from '../../../api/instance'
import { formatDateTime } from '../../../utils/datetime'
import { instanceStatusText } from '../types'

const props = defineProps<{
  loading: boolean
  items: InstanceItem[]
  total: number
  query: { page: number; per_page: number; status: string; instance_no: string; order_no: string; user_keyword: string; date_from: string; date_to: string }
  canOperate: boolean
  canRelease: boolean
  canSync: boolean
}>()

const emit = defineEmits<{
  search: []
  reset: []
  detail: [instanceNo: string]
  start: [item: InstanceItem]
  stop: [item: InstanceItem]
  release: [item: InstanceItem]
  sync: [item: InstanceItem]
}>()

const statusOptions = [
  { label: '创建中', value: 'creating' },
  { label: '运行中', value: 'running' },
  { label: '已停止', value: 'stopped' },
  { label: '异常', value: 'error' },
  { label: '释放中', value: 'releasing' },
  { label: '已释放', value: 'released' },
]

const columns = computed<DataTableColumns<InstanceItem>>(() => [
  { key: 'instance_no', title: '实例编号', minWidth: 170 },
  {
    key: 'user',
    title: '用户',
    minWidth: 150,
    render: (row) => h('div', null, [h('div', { class: 'strong' }, row.user.username), h('div', { class: 'muted' }, row.user.email)]),
  },
  {
    key: 'product',
    title: '实例快照',
    minWidth: 220,
    render: (row) =>
      h('div', null, [
        h('div', { class: 'strong' }, `${row.product_name} · ${row.plan_name}`),
        h('div', { class: 'muted' }, `${row.region_name} · ${row.template_name}`),
      ]),
  },
  {
    key: 'status',
    title: '状态',
    width: 100,
    render: (row) => h(NTag, { type: row.status === 'error' ? 'error' : row.status === 'running' ? 'success' : 'default', size: 'small' }, { default: () => instanceStatusText[row.status] || row.status }),
  },
  {
    key: 'mcp',
    title: '节点 / 编号',
    minWidth: 160,
    render: (row) => `${row.external_node} / ${row.external_vmid}`,
  },
  {
    key: 'created_at',
    title: '创建时间',
    minWidth: 170,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    key: 'actions',
    title: '操作',
    width: 280,
    fixed: 'right',
    render: (row) =>
      h(NSpace, null, {
        default: () => {
          const buttons = [h(NButton, { text: true, type: 'primary', onClick: () => emit('detail', row.instance_no) }, { default: () => '详情' })]
          if (props.canOperate && row.status === 'stopped') buttons.push(h(NButton, { text: true, type: 'success', onClick: () => emit('start', row) }, { default: () => '开机' }))
          if (props.canOperate && row.status === 'running') buttons.push(h(NButton, { text: true, type: 'warning', onClick: () => emit('stop', row) }, { default: () => '关机' }))
          if (props.canSync) buttons.push(h(NButton, { text: true, onClick: () => emit('sync', row) }, { default: () => '同步' }))
          if (props.canRelease && row.status !== 'released' && row.status !== 'releasing') buttons.push(h(NButton, { text: true, type: 'error', onClick: () => emit('release', row) }, { default: () => '释放' }))
          return buttons
        },
      }),
  },
])
</script>

<template>
  <div>
    <NForm inline label-placement="left" class="query-form">
      <NFormItem label="状态">
        <NSelect v-model:value="query.status" :options="statusOptions" clearable placeholder="全部" style="width: 140px" />
      </NFormItem>
      <NFormItem label="实例编号"><NInput v-model:value="query.instance_no" clearable placeholder="INS-" /></NFormItem>
      <NFormItem label="订单编号"><NInput v-model:value="query.order_no" clearable placeholder="ORD-" /></NFormItem>
      <NFormItem label="用户"><NInput v-model:value="query.user_keyword" clearable placeholder="用户名/邮箱" /></NFormItem>
      <NFormItem label="开始"><NInput v-model:value="query.date_from" clearable placeholder="YYYY-MM-DD" /></NFormItem>
      <NFormItem label="结束"><NInput v-model:value="query.date_to" clearable placeholder="YYYY-MM-DD" /></NFormItem>
      <NFormItem :show-label="false">
        <NSpace>
          <NButton type="primary" @click="emit('search')">查询</NButton>
          <NButton @click="emit('reset')">重置</NButton>
        </NSpace>
      </NFormItem>
    </NForm>

    <NDataTable :loading="loading" :columns="columns" :data="items" :row-key="(row: InstanceItem) => row.instance_no" :bordered="false" />

    <div class="pagination">
      <NPagination v-model:page="query.page" v-model:page-size="query.per_page" :item-count="total" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="emit('search')" @update:page-size="emit('search')" />
    </div>
  </div>
</template>
