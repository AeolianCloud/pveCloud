<script setup lang="ts">
import {
  AddOutline,
  RefreshOutline,
  SearchOutline,
} from '@vicons/ionicons5'
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
import type { AdminRoleItem } from '../../../api/admin-role'
import type { AdminUserItem } from '../../../api/admin-user'
import type { PaginationState, UserQueryFormState } from '../types'
import { formatDateTime } from '../../../utils/datetime'

const props = defineProps<{
  loading: boolean
  refreshing: boolean
  hasUsers: boolean
  canViewUsersResource: boolean
  canViewRolesTab: boolean
  canCreateUser: boolean
  canUpdateUser: boolean
  canResetUserPassword: boolean
  queryForm: UserQueryFormState
  roleOptions: AdminRoleItem[]
  users: AdminUserItem[]
  pagination: PaginationState
  userStatusUpdatingId: number | null
}>()

const emit = defineEmits<{
  search: []
  reset: []
  refresh: []
  create: []
  edit: [user: AdminUserItem]
  toggleStatus: [user: AdminUserItem]
  resetPassword: [user: AdminUserItem]
  pageChange: [page: number]
  pageSizeChange: [size: number]
}>()

function formatStatusLabel(status: string) {
  return status === 'active' ? '启用' : '停用'
}

function statusTagType(status: string): 'success' | 'default' {
  return status === 'active' ? 'success' : 'default'
}

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'disabled' },
]

const roleSelectOptions = computed(() =>
  props.roleOptions.map((role) => ({
    label: role.status === 'active' ? role.name : `${role.name}（已停用）`,
    value: role.id,
  })),
)

const columns = computed<DataTableColumns<AdminUserItem>>(() => [
  {
    key: 'username',
    title: '账号',
    minWidth: 140,
    render: (row) =>
      h('div', { class: 'admin-settings-tab__identity' }, [
        h('span', { class: 'admin-settings-tab__primary' }, row.username),
        h('span', { class: 'admin-settings-tab__secondary' }, row.display_name),
      ]),
  },
  {
    key: 'email',
    title: '邮箱',
    minWidth: 220,
    ellipsis: { tooltip: true },
    render: (row) => row.email || '-',
  },
  {
    key: 'status',
    title: '状态',
    width: 100,
    align: 'center',
    render: (row) =>
      h(NTag, { type: statusTagType(row.status), size: 'small' }, { default: () => formatStatusLabel(row.status) }),
  },
  {
    key: 'roles',
    title: '角色',
    minWidth: 220,
    render: (row) => {
      if (row.roles.length === 0) {
        return h('span', { class: 'admin-settings-tab__secondary' }, '未分配角色')
      }
      return h(
        'div',
        { class: 'admin-settings-tab__tags' },
        row.roles.map((r) => h(NTag, { size: 'small' }, { default: () => r.name })),
      )
    },
  },
  {
    key: 'last_login',
    title: '最后登录',
    minWidth: 220,
    render: (row) =>
      h('div', { class: 'admin-settings-tab__meta' }, [
        h('span', null, formatDateTime(row.last_login_at)),
        h('span', null, row.last_login_ip || '-'),
      ]),
  },
  {
    key: 'created_at',
    title: '创建时间',
    minWidth: 180,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    key: 'actions',
    title: '操作',
    width: 280,
    fixed: 'right',
    render: (row) => {
      const buttons: any[] = []
      if (props.canUpdateUser) {
        buttons.push(
          h(NButton, { text: true, type: 'primary', onClick: () => emit('edit', row) }, { default: () => '编辑' }),
          h(
            NButton,
            {
              text: true,
              type: row.status === 'active' ? 'warning' : 'success',
              loading: props.userStatusUpdatingId === row.id,
              onClick: () => emit('toggleStatus', row),
            },
            { default: () => (row.status === 'active' ? '停用' : '启用') },
          ),
        )
      }
      if (props.canResetUserPassword) {
        buttons.push(
          h(NButton, { text: true, type: 'error', onClick: () => emit('resetPassword', row) }, { default: () => '重置密码' }),
        )
      }
      return h(NSpace, { size: 8 }, { default: () => buttons })
    },
  },
])

void SearchOutline
</script>

<template>
  <NCard :bordered="false" class="admin-settings-tab">
    <NSpin :show="loading">
      <div class="admin-settings-tab__body">
        <div class="admin-settings-tab__toolbar">
          <NForm inline label-placement="left" class="admin-settings-tab__filters" @submit.prevent>
            <NFormItem label="关键字">
              <NInput
                v-model:value="props.queryForm.keyword"
                clearable
                placeholder="搜索账号、邮箱或显示名称"
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
            <NFormItem v-if="props.canViewRolesTab" label="角色">
              <NSelect
                v-model:value="props.queryForm.role_id"
                :options="roleSelectOptions"
                clearable
                filterable
                placeholder="全部角色"
                style="min-width: 180px"
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
            <NButton v-if="props.canCreateUser" type="primary" @click="emit('create')">
              <template #icon>
                <NIcon><AddOutline /></NIcon>
              </template>
              新建管理员
            </NButton>
          </div>
        </div>

        <template v-if="!props.canViewUsersResource">
          <EmptyState title="暂无权限" description="当前账号没有管理员账号查看权限。" />
        </template>

        <div v-else-if="props.hasUsers" class="admin-settings-tab__table">
          <NDataTable
            :columns="columns"
            :data="props.users"
            :row-key="(row: AdminUserItem) => row.id"
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
          title="暂无管理员"
          :description="props.queryForm.keyword || props.queryForm.status || props.queryForm.role_id ? '未找到符合条件的管理员账号。' : '当前还没有可展示的管理员账号。'"
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

.admin-settings-tab__primary {
  color: rgba(15, 23, 42, 0.92);
  font-weight: 600;
}

.admin-settings-tab__secondary {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}

.admin-settings-tab__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.admin-settings-tab__pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
}

@media (max-width: 960px) {
  .admin-settings-tab__toolbar {
    flex-direction: column;
  }
}
</style>
