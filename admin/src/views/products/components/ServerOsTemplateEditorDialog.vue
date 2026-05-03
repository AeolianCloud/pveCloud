<script setup lang="ts">
import type { ServerOsTemplatePayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const visible = defineModel<boolean>('visible', { required: true })

defineProps<{
  mode: DialogMode
  form: ServerOsTemplatePayload
}>()

defineEmits<{
  save: []
}>()
</script>

<template>
  <el-dialog v-model="visible" :title="mode === 'create' ? '新增模板' : '编辑模板'" width="680px">
    <el-form label-width="110px">
      <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
      <el-form-item label="发行版"><el-input v-model="form.distribution" /></el-form-item>
      <el-form-item label="版本"><el-input v-model="form.version" /></el-form-item>
      <el-form-item label="系统族">
        <el-select v-model="form.os_family">
          <el-option label="Linux" value="linux" />
          <el-option label="Windows" value="windows" />
          <el-option label="BSD" value="bsd" />
        </el-select>
      </el-form-item>
      <el-form-item label="简介"><el-input v-model="form.summary" /></el-form-item>
      <el-form-item label="状态">
        <el-select v-model="form.status">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="inactive" />
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
