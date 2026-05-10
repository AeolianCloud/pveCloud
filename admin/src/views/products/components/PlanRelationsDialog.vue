<script setup lang="ts">
import { NButton, NForm, NFormItem, NModal, NSelect } from 'naive-ui'
import { computed } from 'vue'

import type { ProductPlanItem, SalesRegionItem, ServerOsTemplateItem } from '../../../api/product-catalog'

const props = defineProps<{
  visible: boolean
  targetPlan: ProductPlanItem | null
  regions: SalesRegionItem[]
  templates: ServerOsTemplateItem[]
}>()

const selectedRegionIds = defineModel<number[]>('selectedRegionIds', { required: true })
const selectedTemplateIds = defineModel<number[]>('selectedTemplateIds', { required: true })

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: []
}>()

const regionOptions = computed(() =>
  props.regions.map((r) => ({ label: r.name, value: r.id })),
)

const templateOptions = computed(() =>
  props.templates.map((t) => ({ label: t.name, value: t.id })),
)
</script>

<template>
  <NModal
    :show="props.visible"
    preset="card"
    :title="`关联配置 - ${props.targetPlan?.name || ''}`"
    style="width: 760px"
    :mask-closable="false"
    @update:show="emit('update:visible', $event)"
  >
    <NForm label-placement="left" label-width="120px">
      <NFormItem label="销售地域">
        <NSelect
          v-model:value="selectedRegionIds"
          :options="regionOptions"
          multiple
          filterable
          style="width: 100%"
        />
      </NFormItem>
      <NFormItem label="系统模板">
        <NSelect
          v-model:value="selectedTemplateIds"
          :options="templateOptions"
          multiple
          filterable
          style="width: 100%"
        />
      </NFormItem>
    </NForm>
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="emit('update:visible', false)">取消</NButton>
        <NButton type="primary" @click="$emit('save')">保存</NButton>
      </div>
    </template>
  </NModal>
</template>
