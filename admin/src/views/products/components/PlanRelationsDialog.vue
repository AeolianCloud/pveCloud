<script setup lang="ts">
import type { ProductPlanItem, SalesRegionItem, ServerOsTemplateItem } from '../../../api/product-catalog'

const visible = defineModel<boolean>('visible', { required: true })
const selectedRegionIds = defineModel<number[]>('selectedRegionIds', { required: true })
const selectedTemplateIds = defineModel<number[]>('selectedTemplateIds', { required: true })

defineProps<{
  targetPlan: ProductPlanItem | null
  regions: SalesRegionItem[]
  templates: ServerOsTemplateItem[]
}>()

defineEmits<{
  save: []
}>()
</script>

<template>
  <el-dialog v-model="visible" :title="`关联配置 - ${targetPlan?.name || ''}`" width="760px">
    <el-form label-width="120px">
      <el-form-item label="销售地域">
        <el-select v-model="selectedRegionIds" multiple filterable collapse-tags style="width: 100%">
          <el-option v-for="region in regions" :key="region.id" :label="region.name" :value="region.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="系统模板">
        <el-select v-model="selectedTemplateIds" multiple filterable collapse-tags style="width: 100%">
          <el-option v-for="template in templates" :key="template.id" :label="template.name" :value="template.id" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="$emit('save')">保存</el-button>
    </template>
  </el-dialog>
</template>
