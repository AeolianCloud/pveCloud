<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import type { ProductPayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInstance>()

defineProps<{
  mode: DialogMode
  form: ProductPayload
}>()

const emit = defineEmits<{
  save: []
}>()

const rules: FormRules<ProductPayload> = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  slug: [{ required: true, message: '请输入 Slug', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
  sort_order: [{ required: true, type: 'number', min: 0, message: '排序必须大于等于 0', trigger: 'change' }],
}

async function submit() {
  await formRef.value?.validate()
  emit('save')
}
</script>

<template>
  <el-dialog v-model="visible" :title="mode === 'create' ? '新增产品' : '编辑产品'" width="720px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="Slug" prop="slug"><el-input v-model="form.slug" /></el-form-item>
      <el-form-item label="简介"><el-input v-model="form.summary" /></el-form-item>
      <el-form-item label="详情"><el-input v-model="form.description" type="textarea" :rows="4" /></el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="草稿" value="draft" />
          <el-option label="上架" value="active" />
          <el-option label="下架" value="inactive" />
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
