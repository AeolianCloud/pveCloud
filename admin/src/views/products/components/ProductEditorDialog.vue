<script setup lang="ts">
import type { ProductPayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const visible = defineModel<boolean>('visible', { required: true })

defineProps<{
  mode: DialogMode
  form: ProductPayload
}>()

defineEmits<{
  save: []
}>()
</script>

<template>
  <el-dialog v-model="visible" :title="mode === 'create' ? '新增产品' : '编辑产品'" width="720px">
    <el-form label-width="110px">
      <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="Slug"><el-input v-model="form.slug" /></el-form-item>
      <el-form-item label="简介"><el-input v-model="form.summary" /></el-form-item>
      <el-form-item label="详情"><el-input v-model="form.description" type="textarea" :rows="4" /></el-form-item>
      <el-form-item label="状态">
        <el-select v-model="form.status">
          <el-option label="草稿" value="draft" />
          <el-option label="上架" value="active" />
          <el-option label="下架" value="inactive" />
        </el-select>
      </el-form-item>
      <el-form-item label="展示"><el-switch v-model="form.visible" /></el-form-item>
      <el-form-item label="排序"><el-input-number v-model="form.sort_order" :min="0" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="$emit('save')">保存</el-button>
    </template>
  </el-dialog>
</template>
