<script setup lang="ts">
import { RefreshOutline } from '@vicons/ionicons5'
import {
  NButton,
  NDataTable,
  NIcon,
  NInputNumber,
  NPagination,
  NSelect,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h } from 'vue'

import type { WebUserSessionItem } from '../../../api/web-user'
import type { PaginationState, SessionQueryState } from '../types'

const props = defineProps<{
  sessions: WebUserSessionItem[]
  query: SessionQueryState
  pagination: PaginationState
  loading: boolean
  canRevoke: boolean
  revokingSessionId: string
  statusType: (status: string) => 'success' | 'error' | 'warning' | 'default'
}>()

const emit = defineEmits<{
  search: []
  refresh: []
  revoke: [session: WebUserSessionItem]
}>()

const statusOptions = [
  { label: '活跃', value: 'active' },
  { label: '已吊销', value: 'revoked' },
  { label: '已过期', value: 'expired' },
]

const columns = computed<DataTableColumns<WebUserSessionItem>>(() => [
  {
    key: 'user',
    title: '用户',
    minWidth: 180,
    render: (row) => `${row.user.username} / ${row.user.email}`,
  },
  {
    key: 'status',
    title: '状态',
    width: 100,
    align: 'center',
    render: (row) => h(NTag, { type: props.statusType(row.status), size: 'small' }, { default: () => row.status }),
  },
  { key: 'issued_at', title: '签发时间', minWidth: 180 },
  { key: 'expires_at', title: '过期时间', minWidth: 180 },
  {
    key: 'last_seen_ip',
    title: '最近 IP',
    minWidth: 140,
    render: (row) => row.last_seen_ip || '-',
  },
  {
    key: 'user_agent',
    title: 'User-Agent',
    minWidth: 220,
    ellipsis: { tooltip: true },
    render: (row) => row.user_agent || '-',
  },
  {
    key: 'actions',
    title: '操作',
    width: 110,
    fixed: 'right',
    render: (row) => {
      if (props.canRevoke && row.status === 'active') {
        return h(
          NButton,
          {
            size: 'small',
            type: 'error',
            loading: props.revokingSessionId === row.session_id,
            onClick: () => emit('revoke', row),
          },
          { default: () => '吊销' },
        )
      }
      return '-'
    },
  },
])
</script>

<template>
  <div class="tab-body">
    <div class="toolbar">
      <NInputNumber v-model:value="props.query.user_id" :min="1" placeholder="用户 ID" style="width: 160px" />
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
    </div>

    <NDataTable
      :loading="props.loading"
      :columns="columns"
      :data="props.sessions"
      :row-key="(row: WebUserSessionItem) => row.session_id"
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
