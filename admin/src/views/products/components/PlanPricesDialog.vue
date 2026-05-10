<script setup lang="ts">
import { NButton, NDivider, NForm, NFormItem, NInputNumber, NModal, NSelect } from 'naive-ui'
import { ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'

import type { PlanPricePayload, ProductPlanItem } from '../../../api/product-catalog'

const props = defineProps<{
  visible: boolean
  targetPlan: ProductPlanItem | null
  prices: PlanPricePayload[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: []
}>()

const formRef = ref<FormInst | null>(null)

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
]

const billingCycleLabel: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  semi_yearly: '半年付',
  yearly: '年付',
}

const rules: FormRules = {
  price_cents: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 1) {
          return new Error('价格必须大于等于 1 分')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  original_price_cents: [
    {
      validator: (_rule, value) => {
        if (value == null) return true
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
          return new Error('原价必须大于等于 0 分')
        }
        return true
      },
      trigger: 'change',
    },
  ],
}

async function submit() {
  await formRef.value?.validate()
  emit('save')
}
</script>

<template>
  <NModal
    :show="props.visible"
    preset="card"
    :title="`套餐价格 - ${props.targetPlan?.name || ''}`"
    style="width: 760px"
    :mask-closable="false"
    @update:show="emit('update:visible', $event)"
  >
    <NForm ref="formRef" :model="props.prices" label-placement="left" label-width="130px">
      <div v-for="(price, index) in props.prices" :key="price.billing_cycle" class="price-row">
        <NDivider title-placement="left">{{ billingCycleLabel[price.billing_cycle] || price.billing_cycle }}</NDivider>
        <NFormItem
          label="价格分"
          :path="`${index}.price_cents`"
          :rule="rules.price_cents"
        >
          <NInputNumber v-model:value="price.price_cents" :min="1" />
        </NFormItem>
        <NFormItem
          label="原价分"
          :path="`${index}.original_price_cents`"
          :rule="rules.original_price_cents"
        >
          <NInputNumber v-model:value="price.original_price_cents" :min="0" />
        </NFormItem>
        <NFormItem label="状态">
          <NSelect v-model:value="price.status" :options="statusOptions" />
        </NFormItem>
      </div>
    </NForm>
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="emit('update:visible', false)">取消</NButton>
        <NButton type="primary" @click="submit">保存</NButton>
      </div>
    </template>
  </NModal>
</template>

<style scoped>
.price-row {
  margin-bottom: 8px;
}
</style>
