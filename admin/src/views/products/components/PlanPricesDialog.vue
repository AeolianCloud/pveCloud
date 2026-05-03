<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance, FormItemRule } from 'element-plus'

import type { PlanPricePayload, ProductPlanItem } from '../../../api/product-catalog'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInstance>()

defineProps<{
  targetPlan: ProductPlanItem | null
  prices: PlanPricePayload[]
}>()

const emit = defineEmits<{
  save: []
}>()

const priceRules: FormItemRule[] = [
  { required: true, message: '请输入价格', trigger: 'change' },
  {
    validator: (_rule, value, callback) => {
      if (typeof value !== 'number' || !Number.isFinite(value) || value < 1) {
        callback(new Error('价格必须大于等于 1 分'))
        return
      }
      callback()
    },
    trigger: 'change',
  },
]

const originalPriceRules: FormItemRule[] = [
  {
    validator: (_rule, value, callback) => {
      if (value == null) {
        callback()
        return
      }
      if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
        callback(new Error('原价必须大于等于 0 分'))
        return
      }
      callback()
    },
    trigger: 'change',
  },
]

async function submit() {
  await formRef.value?.validate()
  emit('save')
}
</script>

<template>
  <el-dialog v-model="visible" :title="`套餐价格 - ${targetPlan?.name || ''}`" width="760px">
    <el-form ref="formRef" :model="prices" label-width="130px">
      <div v-for="(price, index) in prices" :key="price.billing_cycle" class="price-row">
        <el-divider content-position="left">{{ price.billing_cycle }}</el-divider>
        <el-form-item label="价格分" :prop="`${index}.price_cents`" :rules="priceRules">
          <el-input-number v-model="price.price_cents" :min="1" />
        </el-form-item>
        <el-form-item label="原价分" :prop="`${index}.original_price_cents`" :rules="originalPriceRules">
          <el-input-number v-model="price.original_price_cents" :min="0" />
        </el-form-item>
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
      <el-button type="primary" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.price-row {
  margin-bottom: 8px;
}
</style>
