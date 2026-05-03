<script setup lang="ts">
import { Refresh } from '@element-plus/icons-vue'

import type { WebUserSessionItem } from '../../../api/web-user'
import type { PaginationState, SessionQueryState } from '../types'

defineProps<{
  sessions: WebUserSessionItem[]
  query: SessionQueryState
  pagination: PaginationState
  loading: boolean
  canRevoke: boolean
  revokingSessionId: string
  statusType: (status: string) => string
}>()

defineEmits<{
  search: []
  refresh: []
  revoke: [session: WebUserSessionItem]
}>()
</script>

<template>
  <div class="toolbar">
    <el-input-number v-model="query.user_id" :min="1" controls-position="right" placeholder="用户 ID" />
    <el-select v-model="query.status" clearable placeholder="状态">
      <el-option label="活跃" value="active" />
      <el-option label="已吊销" value="revoked" />
      <el-option label="已过期" value="expired" />
    </el-select>
    <el-button type="primary" @click="$emit('search')">查询</el-button>
    <el-button :icon="Refresh" :loading="loading" @click="$emit('refresh')">刷新</el-button>
  </div>

  <el-table v-loading="loading" :data="sessions" stripe>
    <el-table-column label="用户" min-width="180">
      <template #default="{ row }">{{ row.user.username }} / {{ row.user.email }}</template>
    </el-table-column>
    <el-table-column label="状态" width="100" align="center">
      <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ row.status }}</el-tag></template>
    </el-table-column>
    <el-table-column label="签发时间" prop="issued_at" min-width="180" />
    <el-table-column label="过期时间" prop="expires_at" min-width="180" />
    <el-table-column label="最近 IP" min-width="140"><template #default="{ row }">{{ row.last_seen_ip || '-' }}</template></el-table-column>
    <el-table-column label="User-Agent" min-width="220" show-overflow-tooltip><template #default="{ row }">{{ row.user_agent || '-' }}</template></el-table-column>
    <el-table-column label="操作" width="110" fixed="right">
      <template #default="{ row }">
        <el-button v-if="canRevoke && row.status === 'active'" size="small" type="danger" :loading="revokingSessionId === row.session_id" @click="$emit('revoke', row)">吊销</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-pagination
    v-model:current-page="pagination.page"
    v-model:page-size="pagination.per_page"
    background
    layout="total, sizes, prev, pager, next"
    :total="pagination.total"
    :page-sizes="[15, 30, 50, 100]"
    @change="$emit('refresh')"
  />
</template>

<style scoped>
.toolbar { display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 16px; }
.toolbar .el-select { width: 140px; }
.el-pagination { margin-top: 16px; justify-content: flex-end; }
</style>
