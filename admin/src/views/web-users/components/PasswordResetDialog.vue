<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import type { PasswordFormState } from '../types'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInstance>()

defineProps<{
  form: PasswordFormState
  rules: FormRules<PasswordFormState>
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
  <el-dialog v-model="visible" title="重置密码" width="420px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="新密码" prop="password"><el-input v-model="form.password" type="password" show-password /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>
