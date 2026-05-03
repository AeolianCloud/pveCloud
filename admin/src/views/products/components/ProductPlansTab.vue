<script setup lang="ts">
import type { ProductItem, ProductPlanItem } from '../../../api/product-catalog'

defineProps<{
  plans: ProductPlanItem[]
  productsById: Record<number, ProductItem>
  loading: boolean
  planPublishIssues: (item: ProductPlanItem) => string[]
  statusLabel: (status: string) => string
  statusTagType: (status: string) => string
}>()

defineEmits<{
  create: []
  edit: [item: ProductPlanItem]
  prices: [item: ProductPlanItem]
  relations: [item: ProductPlanItem]
  toggleStatus: [item: ProductPlanItem]
}>()
</script>

<template>
  <div class="toolbar">
    <el-button type="primary" @click="$emit('create')">新增套餐</el-button>
  </div>
  <el-table :data="plans" v-loading="loading" border stripe>
    <el-table-column prop="name" label="套餐" min-width="180" />
    <el-table-column label="产品" min-width="140">
      <template #default="{ row }">{{ productsById[row.product_id]?.name || row.product_id }}</template>
    </el-table-column>
    <el-table-column label="配置" min-width="220">
      <template #default="{ row }">
        {{ row.cpu_cores }}C / {{ (row.memory_mb / 1024).toFixed(0) }}G / {{ row.system_disk_gb }}G
      </template>
    </el-table-column>
    <el-table-column label="状态" width="120">
      <template #default="{ row }">
        <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="is_featured" label="推荐" width="90">
      <template #default="{ row }">{{ row.is_featured ? '是' : '否' }}</template>
    </el-table-column>
    <el-table-column label="公开检查" min-width="240">
      <template #default="{ row }">
        <el-tag v-if="planPublishIssues(row).length === 0" type="success">Web 可展示</el-tag>
        <el-space v-else wrap>
          <el-tag v-for="issue in planPublishIssues(row)" :key="issue" type="warning">{{ issue }}</el-tag>
        </el-space>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="320" fixed="right">
      <template #default="{ row }">
        <el-button link type="primary" @click="$emit('edit', row)">编辑</el-button>
        <el-button link type="success" @click="$emit('prices', row)">价格</el-button>
        <el-button link type="success" @click="$emit('relations', row)">关联</el-button>
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
