<script setup lang="ts">
import type { PlanPricePayload, ProductPlanItem } from '../../../api/product-catalog'

const visible = defineModel<boolean>('visible', { required: true })

defineProps<{
  targetPlan: ProductPlanItem | null
  prices: PlanPricePayload[]
}>()

defineEmits<{
  save: []
}>()
</script>

<template>
  <el-dialog v-model="visible" :title="`套餐价格 - ${targetPlan?.name || ''}`" width="760px">
    <el-form label-width="130px">
      <div v-for="price in prices" :key="price.billing_cycle" class="price-row">
        <el-divider content-position="left">{{ price.billing_cycle }}</el-divider>
        <el-form-item label="价格分"><el-input-number v-model="price.price_cents" :min="1" /></el-form-item>
        <el-form-item label="原价分"><el-input-number v-model="price.original_price_cents" :min="0" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="price.status">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
      </div>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="$emit('save')">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.price-row {
  margin-bottom: 8px;
}
</style>
