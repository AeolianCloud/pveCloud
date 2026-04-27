<script setup lang="ts">
import { EditPen, Key, Plus, Refresh, Search, SwitchButton } from '@element-plus/icons-vue'

import EmptyState from '../../../components/EmptyState.vue'
import type { AdminRoleItem } from '../../../api/admin-role'
import type { AdminUserItem } from '../../../api/admin-user'
import type { PaginationState, UserQueryFormState } from '../types'

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

function formatRoleOptionLabel(role: AdminRoleItem) {
  return role.status === 'active' ? role.name : `${role.name}（已停用）`
}

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
  <el-card v-loading="loading" shadow="never" class="admin-settings-tab">
    <div class="admin-settings-tab__toolbar">
      <el-form inline class="admin-settings-tab__filters" @submit.prevent>
        <el-form-item label="关键字">
          <el-input
            v-model="props.queryForm.keyword"
            clearable
            placeholder="搜索账号、邮箱或显示名称"
            @keyup.enter="emit('search')"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="props.queryForm.status" clearable placeholder="全部状态">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="props.canViewRolesTab" label="角色">
          <el-select v-model="props.queryForm.role_id" clearable filterable placeholder="全部角色">
            <el-option
              v-for="role in props.roleOptions"
              :key="role.id"
              :label="formatRoleOptionLabel(role)"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="emit('search')">查询</el-button>
          <el-button @click="emit('reset')">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="admin-settings-tab__toolbar-actions">
        <el-button :icon="Refresh" :loading="props.refreshing" @click="emit('refresh')">刷新</el-button>
        <el-button v-if="props.canCreateUser" type="primary" :icon="Plus" @click="emit('create')">新建管理员</el-button>
      </div>
    </div>

    <template v-if="!props.canViewUsersResource">
      <EmptyState title="暂无权限" description="当前账号没有管理员账号查看权限。" />
    </template>

    <div v-else-if="props.hasUsers" class="admin-settings-tab__table">
      <el-table :data="props.users" stripe>
        <el-table-column label="账号" min-width="140">
          <template #default="{ row }">
            <div class="admin-settings-tab__identity">
              <span class="admin-settings-tab__primary">{{ row.username }}</span>
              <span class="admin-settings-tab__secondary">{{ row.display_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="邮箱" prop="email" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.email || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ formatStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="220">
          <template #default="{ row }">
            <div v-if="row.roles.length > 0" class="admin-settings-tab__tags">
              <el-tag v-for="role in row.roles" :key="role.id" size="small" effect="plain">
                {{ role.name }}
              </el-tag>
            </div>
            <span v-else class="admin-settings-tab__secondary">未分配角色</span>
          </template>
        </el-table-column>
        <el-table-column label="最后登录" min-width="220">
          <template #default="{ row }">
            <div class="admin-settings-tab__meta">
              <span>{{ formatDateTime(row.last_login_at) }}</span>
              <span>{{ row.last_login_ip || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <div class="admin-settings-tab__actions">
              <el-button v-if="props.canUpdateUser" link type="primary" :icon="EditPen" @click="emit('edit', row)">编辑</el-button>
              <el-button
                v-if="props.canUpdateUser"
                link
                :type="row.status === 'active' ? 'warning' : 'success'"
                :icon="SwitchButton"
                :loading="props.userStatusUpdatingId === row.id"
                @click="emit('toggleStatus', row)"
              >
                {{ row.status === 'active' ? '停用' : '启用' }}
              </el-button>
              <el-button v-if="props.canResetUserPassword" link type="danger" :icon="Key" @click="emit('resetPassword', row)">重置密码</el-button>
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
      title="暂无管理员"
      :description="props.queryForm.keyword || props.queryForm.status || props.queryForm.role_id ? '未找到符合条件的管理员账号。' : '当前还没有可展示的管理员账号。'"
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

.admin-settings-tab__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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
