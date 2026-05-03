<script setup lang="ts">
import { Refresh } from '@element-plus/icons-vue'

import type { WebUserItem } from '../../../api/web-user'
import type { PaginationState, UserQueryState } from '../types'

defineProps<{
  users: WebUserItem[]
  query: UserQueryState
  pagination: PaginationState
  loading: boolean
  canCreate: boolean
  canUpdate: boolean
  canResetPassword: boolean
  statusType: (status: string) => string
}>()

defineEmits<{
  search: []
  refresh: []
  create: []
  edit: [user: WebUserItem]
  resetPassword: [user: WebUserItem]
}>()
</script>

<template>
  <div class="toolbar">
    <el-input v-model="query.keyword" clearable placeholder="用户名 / 邮箱 / 显示名称" @keyup.enter="$emit('search')" />
    <el-select v-model="query.status" clearable placeholder="状态">
      <el-option label="启用" value="active" />
      <el-option label="禁用" value="disabled" />
    </el-select>
    <el-button type="primary" @click="$emit('search')">查询</el-button>
    <el-button :icon="Refresh" :loading="loading" @click="$emit('refresh')">刷新</el-button>
    <el-button v-if="canCreate" type="success" @click="$emit('create')">新建用户</el-button>
  </div>

  <el-table v-loading="loading" :data="users" stripe>
    <el-table-column label="用户名" prop="username" min-width="140" />
    <el-table-column label="邮箱" prop="email" min-width="190" />
    <el-table-column label="显示名称" min-width="140">
      <template #default="{ row }">{{ row.display_name || '-' }}</template>
    </el-table-column>
    <el-table-column label="状态" width="100" align="center">
      <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ row.status }}</el-tag></template>
    </el-table-column>
    <el-table-column label="创建时间" prop="created_at" min-width="180" />
    <el-table-column label="操作" width="190" fixed="right">
      <template #default="{ row }">
        <el-button v-if="canUpdate" size="small" @click="$emit('edit', row)">编辑</el-button>
        <el-button v-if="canResetPassword" size="small" type="warning" @click="$emit('resetPassword', row)">重置密码</el-button>
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
.toolbar .el-input { width: 260px; }
.toolbar .el-select { width: 140px; }
.el-pagination { margin-top: 16px; justify-content: flex-end; }
</style>
