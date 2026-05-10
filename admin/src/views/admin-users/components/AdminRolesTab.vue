<script setup lang="ts">
import { AddOutline, RefreshOutline } from '@vicons/ionicons5'
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
import type { PaginationState, RoleQueryFormState } from '../types'
import { formatDateTime } from '../../../utils/datetime'

const props = defineProps<{
  loading: boolean
  refreshing: boolean
  hasRoles: boolean
  canViewRolesResource: boolean
  canCreateRole: boolean
  canUpdateRole: boolean
  queryForm: RoleQueryFormState
  roles: AdminRoleItem[]
  pagination: PaginationState
  roleStatusUpdatingId: number | null
}>()

const emit = defineEmits<{
  search: []
  reset: []
  refresh: []
  create: []
  edit: [role: AdminRoleItem]
  toggleStatus: [role: AdminRoleItem]
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

const columns = computed<DataTableColumns<AdminRoleItem>>(() => [
  {
    key: 'name',
    title: '管理组',
    minWidth: 220,
    render: (row) =>
      h('div', { class: 'admin-settings-tab__identity' }, [
        h('span', { class: 'admin-settings-tab__primary' }, row.name),
        h('span', { class: 'admin-settings-tab__secondary' }, row.code),
      ]),
  },
  {
    key: 'description',
    title: '说明',
    minWidth: 240,
    ellipsis: { tooltip: true },
    render: (row) => row.description || '-',
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
    key: 'permission_codes',
    title: '权限码',
    minWidth: 220,
    render: (row) =>
      h(NTag, { size: 'small' }, { default: () => `${row.permission_codes.length} 项权限` }),
  },
  {
    key: 'updated_at',
    title: '更新时间',
    minWidth: 180,
    render: (row) => formatDateTime(row.updated_at),
  },
  {
    key: 'actions',
    title: '操作',
    width: 200,
    fixed: 'right',
    render: (row) => {
      const buttons: any[] = []
      if (props.canUpdateRole) {
        buttons.push(
          h(NButton, { text: true, type: 'primary', onClick: () => emit('edit', row) }, { default: () => '编辑' }),
          h(
            NButton,
            {
              text: true,
              type: row.status === 'active' ? 'warning' : 'success',
              loading: props.roleStatusUpdatingId === row.id,
              onClick: () => emit('toggleStatus', row),
            },
            { default: () => (row.status === 'active' ? '停用' : '启用') },
          ),
        )
      }
      return h(NSpace, { size: 8 }, { default: () => buttons })
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
                placeholder="搜索编码、名称或说明"
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
            <NButton v-if="props.canCreateRole" type="primary" @click="emit('create')">
              <template #icon>
                <NIcon><AddOutline /></NIcon>
              </template>
              新建管理组
            </NButton>
          </div>
        </div>

        <template v-if="!props.canViewRolesResource">
          <EmptyState title="暂无权限" description="当前账号没有管理组权限查看权限。" />
        </template>

        <div v-else-if="props.hasRoles" class="admin-settings-tab__table">
          <NDataTable
            :columns="columns"
            :data="props.roles"
            :row-key="(row: AdminRoleItem) => row.id"
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
          title="暂无管理组"
          :description="props.queryForm.keyword || props.queryForm.status ? '未找到符合条件的管理组。' : '当前还没有可展示的管理组。'"
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

.admin-settings-tab__identity {
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

.admin-settings-tab__pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
}
</style>
