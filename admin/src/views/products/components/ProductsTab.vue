<script setup lang="ts">
import type { ProductItem } from '../../../api/product-catalog'

defineProps<{
  products: ProductItem[]
  loading: boolean
  statusLabel: (status: string) => string
  statusTagType: (status: string) => string
}>()

defineEmits<{
  create: []
  edit: [item: ProductItem]
  toggleStatus: [item: ProductItem]
}>()
</script>

<template>
  <div class="toolbar">
    <el-button type="primary" @click="$emit('create')">新增产品</el-button>
  </div>
  <el-table :data="products" v-loading="loading" border stripe>
    <el-table-column prop="name" label="名称" min-width="180" />
    <el-table-column prop="slug" label="Slug" min-width="160" />
    <el-table-column label="状态" width="120">
      <template #default="{ row }">
        <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="visible" label="展示" width="90">
      <template #default="{ row }">{{ row.visible ? '是' : '否' }}</template>
    </el-table-column>
    <el-table-column prop="sort_order" label="排序" width="90" />
    <el-table-column label="操作" width="220" fixed="right">
      <template #default="{ row }">
        <el-button link type="primary" @click="$emit('edit', row)">编辑</el-button>
        <el-button link type="warning" @click="$emit('toggleStatus', row)">切换状态</el-button>
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
