<script setup lang="ts">
import { Refresh, Search, SwitchButton } from '@element-plus/icons-vue'

import EmptyState from '../../../components/EmptyState.vue'
import type { AdminSessionItem } from '../../../api/admin-session'
import type { AdminSessionQueryFormState, PaginationState } from '../types'

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
  if (status === 'revoked') {
    return '已吊销'
  }
  if (status === 'expired') {
    return '已过期'
  }
  return '活跃'
}

function statusTagType(status: string) {
  if (status === 'revoked') {
    return 'warning'
  }
  if (status === 'expired') {
    return 'info'
  }
  return 'success'
}

function formatRevokeReason(reason: string | null) {
  if (!reason) {
    return ''
  }
  if (reason === 'logout') {
    return '主动退出'
  }
  if (reason === 'refresh') {
    return '刷新轮换'
  }
  if (reason === 'admin_revoke') {
    return '管理员吊销'
  }
  if (reason === 'admin_disabled') {
    return '账号停用'
  }
  if (reason === 'expired') {
    return '会话过期'
  }
  return reason
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
            placeholder="搜索会话 ID、账号、显示名称或 IP"
            @keyup.enter="emit('search')"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="props.queryForm.status" clearable placeholder="全部状态">
            <el-option label="活跃" value="active" />
            <el-option label="已吊销" value="revoked" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="emit('search')">查询</el-button>
          <el-button @click="emit('reset')">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="admin-settings-tab__toolbar-actions">
        <el-button :icon="Refresh" :loading="props.refreshing" @click="emit('refresh')">刷新</el-button>
      </div>
    </div>

    <template v-if="!props.canViewSessionsResource">
      <EmptyState title="暂无权限" description="当前账号没有管理员会话查看权限。" />
    </template>

    <div v-else-if="props.hasSessions" class="admin-settings-tab__table">
      <el-table :data="props.sessions" stripe>
        <el-table-column label="管理员" min-width="180">
          <template #default="{ row }">
            <div class="admin-settings-tab__identity">
              <span class="admin-settings-tab__primary">{{ row.admin_username }}</span>
              <span class="admin-settings-tab__secondary">{{ row.admin_display_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="会话信息" min-width="240">
          <template #default="{ row }">
            <div class="admin-settings-tab__identity">
              <span class="admin-settings-tab__primary">{{ row.session_id }}</span>
              <div class="admin-settings-tab__tags">
                <el-tag v-if="row.is_current" size="small" type="info">当前会话</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="130" align="center">
          <template #default="{ row }">
            <div class="admin-settings-tab__identity admin-settings-tab__identity--center">
              <el-tag :type="statusTagType(row.status)" size="small">
                {{ formatStatusLabel(row.status) }}
              </el-tag>
              <span v-if="row.status !== 'active' && row.revoke_reason" class="admin-settings-tab__secondary">
                {{ formatRevokeReason(row.revoke_reason) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="最近访问" min-width="200">
          <template #default="{ row }">
            <div class="admin-settings-tab__meta">
              <span>{{ formatDateTime(row.last_seen_at) }}</span>
              <span>{{ row.last_seen_ip || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="生命周期" min-width="220">
          <template #default="{ row }">
            <div class="admin-settings-tab__meta">
              <span>签发：{{ formatDateTime(row.issued_at) }}</span>
              <span>过期：{{ formatDateTime(row.expires_at) }}</span>
              <span v-if="row.revoked_at">吊销：{{ formatDateTime(row.revoked_at) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="设备信息" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="admin-settings-tab__secondary">{{ row.user_agent || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="admin-settings-tab__actions">
              <el-tag v-if="row.is_current" size="small" type="info">当前会话不可吊销</el-tag>
              <el-button
                v-else-if="props.canRevokeSession && row.status === 'active'"
                link
                type="danger"
                :icon="SwitchButton"
                :loading="props.sessionRevokingId === row.session_id"
                @click="emit('revoke', row)"
              >
                吊销会话
              </el-button>
              <span v-else class="admin-settings-tab__secondary">-</span>
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
      title="暂无管理员会话"
      :description="props.queryForm.keyword || props.queryForm.status ? '未找到符合条件的管理员会话。' : '当前还没有可展示的管理员会话。'"
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

.admin-settings-tab__identity--center {
  align-items: center;
}

.admin-settings-tab__primary {
  color: var(--el-text-color-primary);
  font-weight: 600;
}

.admin-settings-tab__secondary {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.admin-settings-tab__tags,
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
