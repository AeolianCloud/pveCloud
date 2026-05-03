<script setup lang="ts">
import type { SalesRegionItem } from '../../../api/product-catalog'

defineProps<{
  regions: SalesRegionItem[]
  loading: boolean
  statusLabel: (status: string) => string
  statusTagType: (status: string) => string
}>()

defineEmits<{
  create: []
  edit: [item: SalesRegionItem]
}>()
</script>

<template>
  <div class="toolbar">
    <el-button type="primary" @click="$emit('create')">新增地域</el-button>
  </div>
  <el-table :data="regions" v-loading="loading" border stripe>
    <el-table-column prop="name" label="名称" min-width="180" />
    <el-table-column prop="code" label="Code" min-width="140" />
    <el-table-column label="状态" width="120">
      <template #default="{ row }">
        <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="visible" label="展示" width="90">
      <template #default="{ row }">{{ row.visible ? '是' : '否' }}</template>
    </el-table-column>
    <el-table-column label="操作" width="160" fixed="right">
      <template #default="{ row }">
        <el-button link type="primary" @click="$emit('edit', row)">编辑</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
