<script setup lang="ts">
import { NButton, NForm, NFormItem, NInput, NModal } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { ref } from 'vue'

import type { PasswordFormState } from '../types'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInst | null>(null)

defineProps<{
  form: PasswordFormState
  rules: FormRules
  submitting: boolean
}>()

const emit = defineEmits<{
  submit: []
}>()

async function submit() {
  await formRef.value?.validate()
  emit('submit')
}
</script>

<template>
  <NModal :show="visible" preset="card" title="重置密码" style="width: 420px" @update:show="visible = $event">
    <NForm ref="formRef" :model="form" :rules="rules as any" label-placement="top">
      <NFormItem label="新密码" path="password">
        <NInput v-model:value="form.password" type="password" show-password-on="click" />
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
