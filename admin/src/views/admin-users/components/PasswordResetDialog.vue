<script setup lang="ts">
import { NAlert, NButton, NForm, NFormItem, NInput, NModal } from 'naive-ui'
import { nextTick, ref, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'

import type { PasswordFormState } from '../types'

const props = defineProps<{
  visible: boolean
  targetLabel: string
  form: PasswordFormState
  rules: FormRules
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
  closed: []
}>()

const formRef = ref<FormInst | null>(null)

watch(
  () => props.visible,
  (value) => {
    if (value) {
      void nextTick(() => formRef.value?.restoreValidation())
    }
  },
)

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  emit('submit')
}

function handleAfterLeave() {
  emit('closed')
}
</script>

<template>
  <NModal
    :show="props.visible"
    preset="card"
    title="重置管理员密码"
    style="width: 480px"
    :mask-closable="false"
    :on-after-leave="handleAfterLeave"
    @update:show="emit('update:visible', $event)"
  >
    <NAlert type="warning" :show-icon="true" style="margin-bottom: 16px">
      密码重置后会立即生效，请通过安全渠道告知管理员。
    </NAlert>
    <NForm ref="formRef" :model="props.form" :rules="props.rules as any" label-placement="top">
      <NFormItem label="管理员">
        <NInput :value="props.targetLabel || '-'" disabled />
      </NFormItem>
      <NFormItem label="新密码" path="password">
        <NInput
          v-model:value="props.form.password"
          type="password"
          show-password-on="click"
          placeholder="请输入 6 到 72 位新密码"
        />
      </NFormItem>
    </NForm>

    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="emit('update:visible', false)">取消</NButton>
        <NButton type="error" :loading="props.submitting" @click="handleSubmit">确认重置</NButton>
      </div>
    </template>
  </NModal>
</template>
