<script setup lang="ts">
import { EditPen, Plus, Refresh, Search, SwitchButton } from '@element-plus/icons-vue'

import EmptyState from '../../../components/EmptyState.vue'
import type { AdminRoleItem } from '../../../api/admin-role'
import type { PaginationState, RoleQueryFormState } from '../types'

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

function statusTagType(status: string) {
  return status === 'active' ? 'success' : 'info'
}

function formatDateTime(value: string | null) {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date)
}
</script>

<template>
  <el-card v-loading="props.loading" shadow="never" class="admin-settings-tab">
    <div class="admin-settings-tab__toolbar">
      <el-form inline class="admin-settings-tab__filters" @submit.prevent>
        <el-form-item label="关键字">
          <el-input
            v-model="props.queryForm.keyword"
            clearable
            placeholder="搜索编码、名称或说明"
            @keyup.enter="emit('search')"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="props.queryForm.status" clearable placeholder="全部状态">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="emit('search')">查询</el-button>
          <el-button @click="emit('reset')">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="admin-settings-tab__toolbar-actions">
        <el-button :icon="Refresh" :loading="props.refreshing" @click="emit('refresh')">刷新</el-button>
        <el-button v-if="props.canCreateRole" type="primary" :icon="Plus" @click="emit('create')">新建管理组</el-button>
      </div>
    </div>

    <template v-if="!props.canViewRolesResource">
      <EmptyState title="暂无权限" description="当前账号没有管理组权限查看权限。" />
    </template>

    <div v-else-if="props.hasRoles" class="admin-settings-tab__table">
      <el-table :data="props.roles" stripe>
        <el-table-column label="管理组" min-width="220">
          <template #default="{ row }">
            <div class="admin-settings-tab__identity">
              <span class="admin-settings-tab__primary">{{ row.name }}</span>
              <span class="admin-settings-tab__secondary">{{ row.code }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="说明" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ formatStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限码" min-width="220">
          <template #default="{ row }">
            <div class="admin-settings-tab__meta">
              <el-tag size="small" effect="plain">{{ row.permission_codes.length }} 项权限</el-tag>
              <span class="admin-settings-tab__secondary">
                {{ row.permission_codes.slice(0, 3).join(' / ') || '未分配权限' }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="admin-settings-tab__actions">
              <el-button v-if="props.canUpdateRole" link type="primary" :icon="EditPen" @click="emit('edit', row)">编辑</el-button>
              <el-button
                v-if="props.canUpdateRole"
                link
                :type="row.status === 'active' ? 'warning' : 'success'"
                :icon="SwitchButton"
                :loading="props.roleStatusUpdatingId === row.id"
                @click="emit('toggleStatus', row)"
              >
                {{ row.status === 'active' ? '停用' : '启用' }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="admin-settings-tab__pagination">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="props.pagination.page"
          :page-size="props.pagination.per_page"
          :page-sizes="[15, 30, 50, 100]"
          :total="props.pagination.total"
          @current-change="emit('pageChange', $event)"
          @size-change="emit('pageSizeChange', $event)"
        />
      </div>
    </div>

    <EmptyState
      v-else
      title="暂无管理组"
      :description="props.queryForm.keyword || props.queryForm.status ? '未找到符合条件的管理组。' : '当前还没有可展示的管理组。'"
    />
  </el-card>
</template>

<style scoped>
.admin-settings-tab :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.admin-settings-tab__toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
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
  color: var(--el-text-color-primary);
  font-weight: 600;
}

.admin-settings-tab__secondary {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.admin-settings-tab__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 12px;
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

  .admin-settings-tab__toolbar-actions {
    width: 100%;
  }

  .admin-settings-tab__pagination {
    justify-content: flex-start;
    overflow-x: auto;
  }
}
</style>
