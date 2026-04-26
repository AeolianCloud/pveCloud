<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { LockKeyhole, Settings } from 'lucide-vue-next'

import { getSystemConfigs, updateSystemConfig } from '../api/systemConfig'
import AdminEmptyState from '../components/AdminEmptyState.vue'
import AdminPageHeader from '../components/AdminPageHeader.vue'
import type { SystemConfigGroup, SystemConfigItem } from '../types/systemConfig'

const loading = ref(false)
const submittingID = ref<number | null>(null)
const errorMessage = ref('')
const groups = ref<SystemConfigGroup[]>([])
const editValues = ref<Record<number, string>>({})

async function loadConfigs() {
  loading.value = true
  errorMessage.value = ''
  try {
    groups.value = await getSystemConfigs()
    const values: Record<number, string> = {}
    for (const group of groups.value) {
      for (const item of group.items) {
        values[item.id] = item.config_value || ''
      }
    }
    editValues.value = values
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '配置加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

async function saveConfig(item: SystemConfigItem) {
  submittingID.value = item.id
  errorMessage.value = ''
  try {
    await updateSystemConfig(item.id, { config_value: editValues.value[item.id] || '' })
    await loadConfigs()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '配置保存失败，请稍后重试'
  } finally {
    submittingID.value = null
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

onMounted(loadConfigs)
</script>

<template>
  <section class="config-page">
    <AdminPageHeader eyebrow="运行配置" title="系统设置" :icon="Settings">
      <Button type="button" label="刷新" icon="pi pi-refresh" :loading="loading" severity="secondary" outlined @click="loadConfigs" />
    </AdminPageHeader>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
    <AdminEmptyState v-if="loading && groups.length === 0" text="配置加载中..." />

    <Card v-for="group in groups" :key="group.group_name" class="config-group">
      <template #title>{{ group.group_name }}</template>
      <template #content>
      <div class="config-grid">
        <article v-for="item in group.items" :key="item.id" class="config-item">
          <div class="config-item-head">
            <div>
              <strong>{{ item.config_key }}</strong>
              <small>{{ item.description || item.value_type }}</small>
            </div>
            <Tag v-if="item.is_secret" severity="danger">
              <LockKeyhole :size="13" aria-hidden="true" />
              敏感
            </Tag>
          </div>
          <InputText
            v-model="editValues[item.id]"
            :type="item.is_secret ? 'password' : 'text'"
            :placeholder="item.is_secret && item.has_value ? '已设置，输入新值覆盖' : '配置值'"
          />
          <footer>
            <span>{{ formatDate(item.updated_at) }}</span>
            <Button type="button" label="保存" icon="pi pi-save" :loading="submittingID === item.id" size="small" @click="saveConfig(item)" />
          </footer>
        </article>
      </div>
      </template>
    </Card>
  </section>
</template>

<style scoped>
.config-page { display: grid; gap: 14px; }
.config-group { overflow: hidden; border: 1px solid var(--border); border-radius: 8px; background: var(--panel); box-shadow: var(--shadow-soft); }
.config-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; padding: 14px; }
.config-item { display: grid; gap: 10px; padding: 12px; border: 1px solid var(--border); border-radius: 8px; background: var(--panel); }
.config-item-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; }
.config-item strong, .config-item small { display: block; }
.config-item strong { color: var(--text); }
.config-item small { margin-top: 4px; color: var(--muted); }
.config-item footer { display: flex; align-items: center; justify-content: space-between; gap: 10px; color: var(--muted); font-size: 12px; }
@media (max-width: 960px) { .config-grid { grid-template-columns: 1fr; } }
</style>
