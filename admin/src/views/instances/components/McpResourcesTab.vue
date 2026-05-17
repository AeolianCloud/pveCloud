<script setup lang="ts">
import { NButton, NCard, NDataTable, NInput, NSpace, type DataTableColumns } from 'naive-ui'
import { computed } from 'vue'

import type { PveNode, PveStorage, PveVM } from '../../../api/instance'

defineProps<{
  loading: boolean
  selectedNode: string
  nodes: PveNode[]
  storage: PveStorage[]
  vms: PveVM[]
}>()

const emit = defineEmits<{
  'update:selectedNode': [value: string]
  loadNodes: []
  loadStorage: []
  'load-vms': []
}>()

const nodeColumns = computed<DataTableColumns<PveNode>>(() => [
  { key: 'node', title: '节点', render: (row) => String(row.node || row.name || '-') },
  { key: 'status', title: '状态', render: (row) => String(row.status || '-') },
])
const storageColumns = computed<DataTableColumns<PveStorage>>(() => [
  { key: 'storage', title: '存储', render: (row) => String(row.storage || row.name || '-') },
  { key: 'type', title: '类型', render: (row) => String(row.type || '-') },
  { key: 'status', title: '状态', render: (row) => String(row.status || '-') },
])
const vmColumns = computed<DataTableColumns<PveVM>>(() => [
  { key: 'vmid', title: '虚拟机编号', render: (row) => String(row.vmid || '-') },
  { key: 'name', title: '名称', render: (row) => String(row.name || '-') },
  { key: 'status', title: '状态', render: (row) => String(row.status || '-') },
])
</script>

<template>
  <div class="mcp-grid">
    <NCard title="节点" :bordered="false">
      <template #header-extra><NButton size="small" :loading="loading" @click="emit('loadNodes')">刷新</NButton></template>
      <NDataTable :columns="nodeColumns" :data="nodes" :loading="loading" :bordered="false" size="small" />
    </NCard>
    <NCard title="存储" :bordered="false">
      <template #header-extra><NButton size="small" :loading="loading" @click="emit('loadStorage')">刷新</NButton></template>
      <NDataTable :columns="storageColumns" :data="storage" :loading="loading" :bordered="false" size="small" />
    </NCard>
    <NCard title="节点虚拟机" :bordered="false">
      <template #header-extra>
        <NSpace>
          <NInput :value="selectedNode" size="small" placeholder="节点名称" style="width: 160px" @update:value="emit('update:selectedNode', $event)" />
          <NButton size="small" :loading="loading" @click="emit('load-vms')">查询</NButton>
        </NSpace>
      </template>
      <NDataTable :columns="vmColumns" :data="vms" :loading="loading" :bordered="false" size="small" />
    </NCard>
  </div>
</template>
