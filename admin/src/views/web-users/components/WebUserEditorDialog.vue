<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import type { EditorMode, UserFormState } from '../types'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInstance>()

defineProps<{
  mode: EditorMode
  title: string
  form: UserFormState
  rules: FormRules<UserFormState>
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
  <el-dialog v-model="visible" :title="title" width="520px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="92px">
      <el-form-item label="用户名" prop="username"><el-input v-model="form.username" :disabled="mode !== 'create'" /></el-form-item>
      <el-form-item label="邮箱" prop="email"><el-input v-model="form.email" /></el-form-item>
      <el-form-item label="显示名称" prop="display_name"><el-input v-model="form.display_name" /></el-form-item>
      <el-form-item v-if="mode === 'create'" label="密码" prop="password"><el-input v-model="form.password" type="password" show-password /></el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="启用" value="active" />
          <el-option label="禁用" value="disabled" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>
