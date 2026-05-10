<script setup lang="ts">
import { RefreshOutline } from '@vicons/ionicons5'
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NPagination,
  NSelect,
  NSpace,
  NSpin,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h } from 'vue'

import EmptyState from '../../../components/EmptyState.vue'
import type { AdminSessionItem } from '../../../api/admin-session'
import type { AdminSessionQueryFormState, PaginationState } from '../types'
import { formatDateTime } from '../../../utils/datetime'

const props = defineProps<{
  loading: boolean
  refreshing: boolean
  hasSessions: boolean
  canViewSessionsResource: boolean
  canRevokeSession: boolean
  queryForm: AdminSessionQueryFormState
  sessions: AdminSessionItem[]
  pagination: PaginationState
  sessionRevokingId: string | null
}>()

const emit = defineEmits<{
  search: []
  reset: []
  refresh: []
  revoke: [session: AdminSessionItem]
  pageChange: [page: number]
  pageSizeChange: [size: number]
}>()

function formatStatusLabel(status: string) {
  if (status === 'revoked') return '已吊销'
  if (status === 'expired') return '已过期'
  return '活跃'
}

function statusTagType(status: string): 'success' | 'warning' | 'default' {
  if (status === 'revoked') return 'warning'
  if (status === 'expired') return 'default'
  return 'success'
}

function formatRevokeReason(reason: string | null) {
  if (!reason) return ''
  const map: Record<string, string> = {
    logout: '主动退出',
    refresh: '刷新轮换',
    admin_revoke: '管理员吊销',
    admin_disabled: '账号停用',
    expired: '会话过期',
  }
  return map[reason] || reason
}

const statusOptions = [
  { label: '活跃', value: 'active' },
  { label: '已吊销', value: 'revoked' },
  { label: '已过期', value: 'expired' },
]

const columns = computed<DataTableColumns<AdminSessionItem>>(() => [
  {
    key: 'admin',
    title: '管理员',
    minWidth: 180,
    render: (row) =>
      h('div', { class: 'admin-settings-tab__identity' }, [
        h('span', { class: 'admin-settings-tab__primary' }, row.admin_username),
        h('span', { class: 'admin-settings-tab__secondary' }, row.admin_display_name),
      ]),
  },
  {
    key: 'session',
    title: '会话信息',
    minWidth: 240,
    render: (row) => {
      const items: any[] = [h('span', { class: 'admin-settings-tab__primary' }, row.session_id)]
      if (row.is_current) {
        items.push(
          h('div', { class: 'admin-settings-tab__tags' }, [h(NTag, { size: 'small' }, { default: () => '当前会话' })]),
        )
      }
      return h('div', { class: 'admin-settings-tab__identity' }, items)
    },
  },
  {
    key: 'status',
    title: '状态',
    width: 130,
    align: 'center',
    render: (row) => {
      const items: any[] = [
        h(NTag, { type: statusTagType(row.status), size: 'small' }, { default: () => formatStatusLabel(row.status) }),
      ]
      if (row.status !== 'active' && row.revoke_reason) {
        items.push(h('span', { class: 'admin-settings-tab__secondary' }, formatRevokeReason(row.revoke_reason)))
      }
      return h('div', { class: 'admin-settings-tab__identity admin-settings-tab__identity--center' }, items)
    },
  },
  {
    key: 'last_seen',
    title: '最近访问',
    minWidth: 200,
    render: (row) =>
      h('div', { class: 'admin-settings-tab__meta' }, [
        h('span', null, formatDateTime(row.last_seen_at)),
        h('span', null, row.last_seen_ip || '-'),
      ]),
  },
  {
    key: 'lifecycle',
    title: '生命周期',
    minWidth: 220,
    render: (row) => {
      const items: any[] = [
        h('span', null, `签发：${formatDateTime(row.issued_at)}`),
        h('span', null, `过期：${formatDateTime(row.expires_at)}`),
      ]
      if (row.revoked_at) items.push(h('span', null, `吊销：${formatDateTime(row.revoked_at)}`))
      return h('div', { class: 'admin-settings-tab__meta' }, items)
    },
  },
  {
    key: 'user_agent',
    title: '设备信息',
    minWidth: 240,
    ellipsis: { tooltip: true },
    render: (row) => h('span', { class: 'admin-settings-tab__secondary' }, row.user_agent || '-'),
  },
  {
    key: 'actions',
    title: '操作',
    width: 180,
    fixed: 'right',
    render: (row) => {
      if (row.is_current) {
        return h(NTag, { size: 'small' }, { default: () => '当前会话不可吊销' })
      }
      if (props.canRevokeSession && row.status === 'active') {
        return h(
          NButton,
          {
            text: true,
            type: 'error',
            loading: props.sessionRevokingId === row.session_id,
            onClick: () => emit('revoke', row),
          },
          { default: () => '吊销会话' },
        )
      }
      return h('span', { class: 'admin-settings-tab__secondary' }, '-')
    },
  },
])
</script>

<template>
  <NCard :bordered="false" class="admin-settings-tab">
    <NSpin :show="props.loading">
      <div class="admin-settings-tab__body">
        <div class="admin-settings-tab__toolbar">
          <NForm inline label-placement="left" class="admin-settings-tab__filters" @submit.prevent>
            <NFormItem label="关键字">
              <NInput
                v-model:value="props.queryForm.keyword"
                clearable
                placeholder="搜索会话 ID、账号、显示名称或 IP"
                @keyup.enter="emit('search')"
              />
            </NFormItem>
            <NFormItem label="状态">
              <NSelect
                v-model:value="props.queryForm.status"
                :options="statusOptions"
                clearable
                placeholder="全部状态"
                style="min-width: 140px"
              />
            </NFormItem>
            <NFormItem :show-label="false">
              <NSpace>
                <NButton type="primary" @click="emit('search')">查询</NButton>
                <NButton @click="emit('reset')">重置</NButton>
              </NSpace>
            </NFormItem>
          </NForm>

          <div class="admin-settings-tab__toolbar-actions">
            <NButton :loading="props.refreshing" @click="emit('refresh')">
              <template #icon>
                <NIcon><RefreshOutline /></NIcon>
              </template>
              刷新
            </NButton>
          </div>
        </div>

        <template v-if="!props.canViewSessionsResource">
          <EmptyState title="暂无权限" description="当前账号没有管理员会话查看权限。" />
        </template>

        <div v-else-if="props.hasSessions" class="admin-settings-tab__table">
          <NDataTable
            :columns="columns"
            :data="props.sessions"
            :row-key="(row: AdminSessionItem) => row.session_id"
            striped
            :bordered="false"
          />

          <div class="admin-settings-tab__pagination">
            <NPagination
              :page="props.pagination.page"
              :page-size="props.pagination.per_page"
              :item-count="props.pagination.total"
              :page-sizes="[15, 30, 50, 100]"
              show-size-picker
              @update:page="emit('pageChange', $event)"
              @update:page-size="emit('pageSizeChange', $event)"
            />
          </div>
        </div>

        <EmptyState
          v-else
          title="暂无管理员会话"
          :description="props.queryForm.keyword || props.queryForm.status ? '未找到符合条件的管理员会话。' : '当前还没有可展示的管理员会话。'"
        />
      </div>
    </NSpin>
  </NCard>
</template>

<style scoped>
.admin-settings-tab__body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.admin-settings-tab__toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.admin-settings-tab__filters {
  display: flex;
  flex: 1;
  flex-wrap: wrap;
  gap: 8px 0;
}

.admin-settings-tab__toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.admin-settings-tab__table {
  min-height: 240px;
}

.admin-settings-tab__identity,
.admin-settings-tab__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.admin-settings-tab__identity--center {
  align-items: center;
}

.admin-settings-tab__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 8px;
}

.admin-settings-tab__primary {
  color: rgba(15, 23, 42, 0.92);
  font-weight: 600;
}

.admin-settings-tab__secondary {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}

.admin-settings-tab__pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
}
</style>
