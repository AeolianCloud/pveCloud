<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import type { ProductItem, ProductPlanPayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const visible = defineModel<boolean>('visible', { required: true })
const formRef = ref<FormInstance>()

defineProps<{
  mode: DialogMode
  form: ProductPlanPayload
  products: ProductItem[]
}>()

const emit = defineEmits<{
  save: []
}>()

const rules: FormRules<ProductPlanPayload> = {
  product_id: [{ required: true, type: 'number', min: 1, message: '请选择所属产品', trigger: 'change' }],
  name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编码', trigger: 'blur' }],
  cpu_cores: [{ required: true, type: 'number', min: 1, message: 'CPU 核数必须大于等于 1', trigger: 'change' }],
  memory_mb: [{ required: true, type: 'number', min: 128, message: '内存必须大于等于 128 MB', trigger: 'change' }],
  system_disk_gb: [{ required: true, type: 'number', min: 1, message: '系统盘必须大于等于 1 GB', trigger: 'change' }],
  data_disk_gb: [{ required: true, type: 'number', min: 0, message: '数据盘必须大于等于 0 GB', trigger: 'change' }],
  bandwidth_mbps: [{ required: true, type: 'number', min: 1, message: '带宽必须大于等于 1 Mbps', trigger: 'change' }],
  traffic_gb: [
    {
      validator: (_rule, value, callback) => {
        if (value == null) {
          callback()
          return
        }
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
          callback(new Error('流量必须大于等于 0 GB'))
          return
        }
        callback()
      },
      trigger: 'change',
    },
  ],
  public_ip_count: [{ required: true, type: 'number', min: 0, message: '公网 IP 数必须大于等于 0', trigger: 'change' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

async function submit() {
  await formRef.value?.validate()
  emit('save')
}
</script>

<template>
  <el-dialog v-model="visible" :title="mode === 'create' ? '新增套餐' : '编辑套餐'" width="760px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item label="所属产品" prop="product_id">
        <el-select v-model="form.product_id" placeholder="选择产品">
          <el-option v-for="product in products" :key="product.id" :label="product.name" :value="product.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="套餐名称" prop="name"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="编码" prop="code"><el-input v-model="form.code" /></el-form-item>
      <el-form-item label="简介"><el-input v-model="form.summary" /></el-form-item>
      <el-form-item label="CPU 核数" prop="cpu_cores"><el-input-number v-model="form.cpu_cores" :min="1" /></el-form-item>
      <el-form-item label="内存 MB" prop="memory_mb"><el-input-number v-model="form.memory_mb" :min="128" /></el-form-item>
      <el-form-item label="系统盘 GB" prop="system_disk_gb"><el-input-number v-model="form.system_disk_gb" :min="1" /></el-form-item>
      <el-form-item label="数据盘 GB" prop="data_disk_gb"><el-input-number v-model="form.data_disk_gb" :min="0" /></el-form-item>
      <el-form-item label="带宽 Mbps" prop="bandwidth_mbps"><el-input-number v-model="form.bandwidth_mbps" :min="1" /></el-form-item>
      <el-form-item label="流量 GB" prop="traffic_gb"><el-input-number v-model="form.traffic_gb" :min="0" /></el-form-item>
      <el-form-item label="公网 IP 数" prop="public_ip_count"><el-input-number v-model="form.public_ip_count" :min="0" /></el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="草稿" value="draft" />
          <el-option label="上架" value="active" />
          <el-option label="下架" value="inactive" />
          <el-option label="售罄" value="sold_out" />
        </el-select>
      </el-form-item>
      <el-form-item label="展示"><el-switch v-model="form.visible" /></el-form-item>
      <el-form-item label="推荐"><el-switch v-model="form.is_featured" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>
