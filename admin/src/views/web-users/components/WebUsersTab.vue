<script setup lang="ts">
import { RefreshOutline } from '@vicons/ionicons5'
import {
  NButton,
  NDataTable,
  NIcon,
  NInput,
  NPagination,
  NSelect,
  NSpace,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h } from 'vue'

import type { WebUserItem } from '../../../api/web-user'
import type { PaginationState, UserQueryState } from '../types'

const props = defineProps<{
  users: WebUserItem[]
  query: UserQueryState
  pagination: PaginationState
  loading: boolean
  canCreate: boolean
  canUpdate: boolean
  canResetPassword: boolean
  statusType: (status: string) => 'success' | 'error' | 'warning' | 'default'
}>()

const emit = defineEmits<{
  search: []
  refresh: []
  create: []
  edit: [user: WebUserItem]
  resetPassword: [user: WebUserItem]
}>()

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '禁用', value: 'disabled' },
]

const columns = computed<DataTableColumns<WebUserItem>>(() => [
  { key: 'username', title: '用户名', minWidth: 140 },
  { key: 'email', title: '邮箱', minWidth: 190 },
  {
    key: 'display_name',
    title: '显示名称',
    minWidth: 140,
    render: (row) => row.display_name || '-',
  },
  {
    key: 'status',
    title: '状态',
    width: 100,
    align: 'center',
    render: (row) => h(NTag, { type: props.statusType(row.status), size: 'small' }, { default: () => row.status }),
  },
  { key: 'created_at', title: '创建时间', minWidth: 180 },
  {
    key: 'actions',
    title: '操作',
    width: 200,
    fixed: 'right',
    render: (row) => {
      const buttons: any[] = []
      if (props.canUpdate) {
        buttons.push(h(NButton, { size: 'small', onClick: () => emit('edit', row) }, { default: () => '编辑' }))
      }
      if (props.canResetPassword) {
        buttons.push(
          h(
            NButton,
            { size: 'small', type: 'warning', onClick: () => emit('resetPassword', row) },
            { default: () => '重置密码' },
          ),
        )
      }
      return h(NSpace, { size: 8 }, { default: () => buttons })
    },
  },
])
</script>

<template>
  <div class="tab-body">
    <div class="toolbar">
      <NInput
        v-model:value="props.query.keyword"
        clearable
        placeholder="用户名 / 邮箱 / 显示名称"
        style="width: 260px"
        @keyup.enter="emit('search')"
      />
      <NSelect
        v-model:value="props.query.status"
        :options="statusOptions"
        clearable
        placeholder="状态"
        style="width: 140px"
      />
      <NButton type="primary" @click="emit('search')">查询</NButton>
      <NButton :loading="props.loading" @click="emit('refresh')">
        <template #icon>
          <NIcon><RefreshOutline /></NIcon>
        </template>
        刷新
      </NButton>
      <NButton v-if="props.canCreate" type="success" @click="emit('create')">新建用户</NButton>
    </div>

    <NDataTable
      :loading="props.loading"
      :columns="columns"
      :data="props.users"
      :row-key="(row: WebUserItem) => row.id"
      striped
      :bordered="false"
    />

    <div class="pagination">
      <NPagination
        :page="props.pagination.page"
        :page-size="props.pagination.per_page"
        :item-count="props.pagination.total"
        :page-sizes="[15, 30, 50, 100]"
        show-size-picker
        @update:page="(p: number) => { props.pagination.page = p; emit('refresh') }"
        @update:page-size="(s: number) => { props.pagination.per_page = s; props.pagination.page = 1; emit('refresh') }"
      />
    </div>
  </div>
</template>

<style scoped>
.tab-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
}
</style>
