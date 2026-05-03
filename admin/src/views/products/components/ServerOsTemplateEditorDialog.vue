<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import type { ServerOsTemplatePayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInstance>()

defineProps<{
  mode: DialogMode
  form: ServerOsTemplatePayload
}>()

const emit = defineEmits<{
  save: []
}>()

const rules: FormRules<ServerOsTemplatePayload> = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编码', trigger: 'blur' }],
  distribution: [{ required: true, message: '请输入发行版', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本', trigger: 'blur' }],
  os_family: [{ required: true, message: '请选择系统族', trigger: 'change' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
  sort_order: [{ required: true, type: 'number', min: 0, message: '排序必须大于等于 0', trigger: 'change' }],
}

async function submit() {
  await formRef.value?.validate()
  emit('save')
}
</script>

<template>
  <el-dialog v-model="visible" :title="mode === 'create' ? '新增模板' : '编辑模板'" width="680px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="编码" prop="code"><el-input v-model="form.code" /></el-form-item>
      <el-form-item label="发行版" prop="distribution"><el-input v-model="form.distribution" /></el-form-item>
      <el-form-item label="版本" prop="version"><el-input v-model="form.version" /></el-form-item>
      <el-form-item label="系统族" prop="os_family">
        <el-select v-model="form.os_family">
          <el-option label="Linux" value="linux" />
          <el-option label="Windows" value="windows" />
          <el-option label="BSD" value="bsd" />
        </el-select>
      </el-form-item>
      <el-form-item label="简介"><el-input v-model="form.summary" /></el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="inactive" />
        </el-select>
      </el-form-item>
      <el-form-item label="展示"><el-switch v-model="form.visible" /></el-form-item>
      <el-form-item label="排序" prop="sort_order"><el-input-number v-model="form.sort_order" :min="0" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>
