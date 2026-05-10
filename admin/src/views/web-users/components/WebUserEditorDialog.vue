<script setup lang="ts">
import { NButton, NForm, NFormItem, NInput, NModal, NSelect } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { ref } from 'vue'

import type { EditorMode, UserFormState } from '../types'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInst | null>(null)

defineProps<{
  mode: EditorMode
  title: string
  form: UserFormState
  rules: FormRules
  submitting: boolean
}>()

const emit = defineEmits<{
  submit: []
}>()

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '禁用', value: 'disabled' },
]

async function submit() {
  await formRef.value?.validate()
  emit('submit')
}
</script>

<template>
  <NModal :show="visible" preset="card" :title="title" style="width: 520px" @update:show="visible = $event">
    <NForm ref="formRef" :model="form" :rules="rules as any" label-placement="top">
      <NFormItem label="用户名" path="username">
        <NInput v-model:value="form.username" :disabled="mode !== 'create'" />
      </NFormItem>
      <NFormItem label="邮箱" path="email">
        <NInput v-model:value="form.email" />
      </NFormItem>
      <NFormItem label="显示名称" path="display_name">
        <NInput v-model:value="form.display_name" />
      </NFormItem>
      <NFormItem v-if="mode === 'create'" label="密码" path="password">
        <NInput v-model:value="form.password" type="password" show-password-on="click" />
      </NFormItem>
      <NFormItem label="状态" path="status">
        <NSelect v-model:value="form.status" :options="statusOptions" />
      </NFormItem>
    </NForm>
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="visible = false">取消</NButton>
        <NButton type="primary" :loading="submitting" @click="submit">保存</NButton>
      </div>
    </template>
  </NModal>
</template>
