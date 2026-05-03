<script setup lang="ts">
import type { ProductItem, ProductPlanPayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const visible = defineModel<boolean>('visible', { required: true })

defineProps<{
  mode: DialogMode
  form: ProductPlanPayload
  products: ProductItem[]
}>()

defineEmits<{
  save: []
}>()
</script>

<template>
  <el-dialog v-model="visible" :title="mode === 'create' ? '新增套餐' : '编辑套餐'" width="760px">
    <el-form label-width="120px">
      <el-form-item label="所属产品">
        <el-select v-model="form.product_id" placeholder="选择产品">
          <el-option v-for="product in products" :key="product.id" :label="product.name" :value="product.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="套餐名称"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
      <el-form-item label="简介"><el-input v-model="form.summary" /></el-form-item>
      <el-form-item label="CPU 核数"><el-input-number v-model="form.cpu_cores" :min="1" /></el-form-item>
      <el-form-item label="内存 MB"><el-input-number v-model="form.memory_mb" :min="128" /></el-form-item>
      <el-form-item label="系统盘 GB"><el-input-number v-model="form.system_disk_gb" :min="1" /></el-form-item>
      <el-form-item label="数据盘 GB"><el-input-number v-model="form.data_disk_gb" :min="0" /></el-form-item>
      <el-form-item label="带宽 Mbps"><el-input-number v-model="form.bandwidth_mbps" :min="1" /></el-form-item>
      <el-form-item label="流量 GB"><el-input-number v-model="form.traffic_gb" :min="0" /></el-form-item>
      <el-form-item label="公网 IP 数"><el-input-number v-model="form.public_ip_count" :min="0" /></el-form-item>
      <el-form-item label="状态">
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
      <el-button type="primary" @click="$emit('save')">保存</el-button>
    </template>
  </el-dialog>
</template>
