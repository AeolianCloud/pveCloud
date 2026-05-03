<script setup lang="ts">
import type { SalesRegionPayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const visible = defineModel<boolean>('visible', { required: true })

defineProps<{
  mode: DialogMode
  form: SalesRegionPayload
}>()

defineEmits<{
  save: []
}>()
</script>

<template>
  <el-dialog v-model="visible" :title="mode === 'create' ? '新增地域' : '编辑地域'" width="600px">
    <el-form label-width="100px">
      <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="Code"><el-input v-model="form.code" /></el-form-item>
      <el-form-item label="国家/地区"><el-input v-model="form.country" /></el-form-item>
      <el-form-item label="城市"><el-input v-model="form.city" /></el-form-item>
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
