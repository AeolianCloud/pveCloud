<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import type { PasswordFormState } from '../types'

const props = defineProps<{
  visible: boolean
  targetLabel: string
  form: PasswordFormState
  rules: FormRules<PasswordFormState>
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
  closed: []
}>()

const formRef = ref<FormInstance>()

watch(
  () => props.visible,
  (value) => {
    if (value) {
      void nextTick(() => {
        formRef.value?.clearValidate()
      })
    }
  },
)

async function handleSubmit() {
  if (!formRef.value) {
    return
  }
  await formRef.value.validate()
  emit('submit')
}
</script>

<template>
  <el-dialog
    :model-value="props.visible"
    title="重置管理员密码"
    width="480px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
    @closed="emit('closed')"
  >
    <el-alert
      type="warning"
      :closable="false"
      class="password-reset-dialog__alert"
      title="密码重置后会立即生效，请通过安全渠道告知管理员。"
    />
    <el-form ref="formRef" :model="props.form" :rules="props.rules" label-width="84px">
      <el-form-item label="管理员">
        <el-input :model-value="props.targetLabel || '-'" disabled />
      </el-form-item>
      <el-form-item label="新密码" prop="password">
        <el-input
          v-model="props.form.password"
          type="password"
          show-password
          placeholder="请输入 6 到 72 位新密码"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="danger" :loading="props.submitting" @click="handleSubmit">确认重置</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.password-reset-dialog__alert {
  margin-bottom: 16px;
}
</style>
