<script setup lang="ts">
import { NButton, NForm, NFormItem, NInput, NInputNumber, NModal, NSelect, NSwitch } from 'naive-ui'
import { computed, ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'

import type { ProductItem, ProductPlanPayload } from '../../../api/product-catalog'
import type { DialogMode } from '../types'

const props = defineProps<{
  visible: boolean
  mode: DialogMode
  form: ProductPlanPayload
  products: ProductItem[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: []
}>()

const formRef = ref<FormInst | null>(null)

const productOptions = computed(() =>
  props.products.map((p) => ({ label: p.name, value: p.id })),
)

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '上架', value: 'active' },
  { label: '下架', value: 'inactive' },
  { label: '售罄', value: 'sold_out' },
]

const rules: FormRules = {
  product_id: [
    {
      validator: (_rule, value) => {
        if (!value || value < 1) return new Error('请选择所属产品')
        return true
      },
      trigger: 'change',
    },
  ],
  name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编码', trigger: 'blur' }],
  cpu_cores: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 1) {
          return new Error('CPU 核数必须大于等于 1')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  memory_mb: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 128) {
          return new Error('内存必须大于等于 128 MB')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  system_disk_gb: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 1) {
          return new Error('系统盘必须大于等于 1 GB')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  data_disk_gb: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
          return new Error('数据盘必须大于等于 0 GB')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  bandwidth_mbps: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 1) {
          return new Error('带宽必须大于等于 1 Mbps')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  traffic_gb: [
    {
      validator: (_rule, value) => {
        if (value == null) return true
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
          return new Error('流量必须大于等于 0 GB')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  public_ip_count: [
    {
      validator: (_rule, value) => {
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
          return new Error('公网 IP 数必须大于等于 0')
        }
        return true
      },
      trigger: 'change',
    },
  ],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

async function submit() {
  await formRef.value?.validate()
  emit('save')
}
</script>

<template>
  <NModal
    :show="props.visible"
    preset="card"
    :title="props.mode === 'create' ? '新增套餐' : '编辑套餐'"
    style="width: 760px"
    :mask-closable="false"
    @update:show="emit('update:visible', $event)"
  >
    <NForm ref="formRef" :model="props.form" :rules="rules" label-placement="left" label-width="120px">
      <NFormItem label="所属产品" path="product_id">
        <NSelect v-model:value="props.form.product_id" :options="productOptions" placeholder="选择产品" />
      </NFormItem>
      <NFormItem label="套餐名称" path="name">
        <NInput v-model:value="props.form.name" />
      </NFormItem>
      <NFormItem label="编码" path="code">
        <NInput v-model:value="props.form.code" />
      </NFormItem>
      <NFormItem label="简介">
        <NInput v-model:value="props.form.summary" />
      </NFormItem>
      <NFormItem label="CPU 核数" path="cpu_cores">
        <NInputNumber v-model:value="props.form.cpu_cores" :min="1" />
      </NFormItem>
      <NFormItem label="内存 MB" path="memory_mb">
        <NInputNumber v-model:value="props.form.memory_mb" :min="128" />
      </NFormItem>
      <NFormItem label="系统盘 GB" path="system_disk_gb">
        <NInputNumber v-model:value="props.form.system_disk_gb" :min="1" />
      </NFormItem>
      <NFormItem label="数据盘 GB" path="data_disk_gb">
        <NInputNumber v-model:value="props.form.data_disk_gb" :min="0" />
      </NFormItem>
      <NFormItem label="带宽 Mbps" path="bandwidth_mbps">
        <NInputNumber v-model:value="props.form.bandwidth_mbps" :min="1" />
      </NFormItem>
      <NFormItem label="流量 GB" path="traffic_gb">
        <NInputNumber v-model:value="props.form.traffic_gb" :min="0" />
      </NFormItem>
      <NFormItem label="公网 IP 数" path="public_ip_count">
        <NInputNumber v-model:value="props.form.public_ip_count" :min="0" />
      </NFormItem>
      <NFormItem label="状态" path="status">
        <NSelect v-model:value="props.form.status" :options="statusOptions" />
      </NFormItem>
      <NFormItem label="展示">
        <NSwitch v-model:value="props.form.visible" />
      </NFormItem>
      <NFormItem label="推荐">
        <NSwitch v-model:value="props.form.is_featured" />
      </NFormItem>
    </NForm>
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="emit('update:visible', false)">取消</NButton>
        <NButton type="primary" @click="submit">保存</NButton>
      </div>
    </template>
  </NModal>
</template>
