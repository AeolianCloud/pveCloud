<script setup lang="ts">
import { NButton, NForm, NFormItem, NInput, NInputNumber, NModal, NSelect, NSwitch } from 'naive-ui'
import { ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'

import type { ProductPayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const props = defineProps<{
  visible: boolean
  mode: DialogMode
  form: ProductPayload
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: []
}>()

const formRef = ref<FormInst | null>(null)

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '上架', value: 'active' },
  { label: '下架', value: 'inactive' },
]

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  slug: [{ required: true, message: '请输入 Slug', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
  sort_order: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
          return new Error('排序必须大于等于 0')
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
    :title="props.mode === 'create' ? '新增产品' : '编辑产品'"
    style="width: 720px"
    :mask-closable="false"
    @update:show="emit('update:visible', $event)"
  >
    <NForm ref="formRef" :model="props.form" :rules="rules" label-placement="left" label-width="110px">
      <NFormItem label="名称" path="name">
        <NInput v-model:value="props.form.name" />
      </NFormItem>
      <NFormItem label="Slug" path="slug">
        <NInput v-model:value="props.form.slug" />
      </NFormItem>
      <NFormItem label="简介">
        <NInput v-model:value="props.form.summary" />
      </NFormItem>
      <NFormItem label="详情">
        <NInput v-model:value="props.form.description" type="textarea" :rows="4" />
      </NFormItem>
      <NFormItem label="状态" path="status">
        <NSelect v-model:value="props.form.status" :options="statusOptions" />
      </NFormItem>
      <NFormItem label="展示">
        <NSwitch v-model:value="props.form.visible" />
      </NFormItem>
      <NFormItem label="排序" path="sort_order">
        <NInputNumber v-model:value="props.form.sort_order" :min="0" />
      </NFormItem>
    </NForm>
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="emit('update:visible', false)">取消</NButton>
        <NButton type="primary" @click="submit">保存</NButton>
      </div>
    </template>
  </NModal>
</template>
