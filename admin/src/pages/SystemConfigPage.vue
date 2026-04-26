<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { LockKeyhole, RefreshCw, Save, Settings } from 'lucide-vue-next'

import { getSystemConfigs, updateSystemConfig } from '../api/systemConfig'
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
    <header class="config-toolbar">
      <div class="config-title">
        <span class="config-title-icon"><Settings :size="20" aria-hidden="true" /></span>
        <div>
          <span>运行配置</span>
          <h1>系统设置</h1>
        </div>
      </div>
      <button type="button" :disabled="loading" @click="loadConfigs">
        <RefreshCw :class="{ spinning: loading }" :size="15" aria-hidden="true" />
        刷新
      </button>
    </header>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <div v-if="loading && groups.length === 0" class="config-state">配置加载中...</div>

    <section v-for="group in groups" :key="group.group_name" class="config-group">
      <header>
        <h2>{{ group.group_name }}</h2>
      </header>
      <div class="config-grid">
        <article v-for="item in group.items" :key="item.id" class="config-item">
          <div class="config-item-head">
            <div>
              <strong>{{ item.config_key }}</strong>
              <small>{{ item.description || item.value_type }}</small>
            </div>
            <span v-if="item.is_secret" class="secret-pill"><LockKeyhole :size="13" aria-hidden="true" />敏感</span>
          </div>
          <input
            v-model="editValues[item.id]"
            :type="item.is_secret ? 'password' : 'text'"
            :placeholder="item.is_secret && item.has_value ? '已设置，输入新值覆盖' : '配置值'"
          />
          <footer>
            <span>{{ formatDate(item.updated_at) }}</span>
            <button type="button" :disabled="submittingID === item.id" @click="saveConfig(item)">
              <Save :size="14" aria-hidden="true" />
              {{ submittingID === item.id ? '保存中...' : '保存' }}
            </button>
          </footer>
        </article>
      </div>
    </section>
  </section>
</template>

<style scoped>
.config-page { display: grid; gap: 14px; }
.config-toolbar, .config-group { border: 1px solid var(--border); border-radius: 8px; background: var(--panel); box-shadow: var(--shadow-soft); }
.config-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 14px; padding: 14px; }
.config-title { display: flex; align-items: center; gap: 12px; }
.config-title-icon { width: 38px; height: 38px; display: grid; place-items: center; border-radius: 8px; color: var(--primary); background: var(--primary-soft); }
.config-title span { color: var(--muted); font-size: 13px; font-weight: 800; }
.config-title h1 { margin: 3px 0 0; font-size: 20px; line-height: 1.2; }
.config-toolbar button, .config-item button { min-height: 34px; display: inline-flex; align-items: center; justify-content: center; gap: 7px; padding: 0 11px; border: 1px solid var(--border); border-radius: 8px; color: var(--muted-strong); background: var(--panel); font-size: 13px; font-weight: 750; cursor: pointer; }
.config-group { overflow: hidden; }
.config-group > header { padding: 13px 14px; border-bottom: 1px solid var(--border); background: var(--panel-soft); }
.config-group h2 { margin: 0; font-size: 16px; }
.config-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; padding: 14px; }
.config-item { display: grid; gap: 10px; padding: 12px; border: 1px solid var(--border); border-radius: 8px; background: var(--panel); }
.config-item-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; }
.config-item strong, .config-item small { display: block; }
.config-item strong { color: var(--text); }
.config-item small { margin-top: 4px; color: var(--muted); }
.config-item input { min-height: 38px; border: 1px solid var(--border); border-radius: 8px; padding: 0 10px; color: var(--text); background: var(--panel); }
.config-item footer { display: flex; align-items: center; justify-content: space-between; gap: 10px; color: var(--muted); font-size: 12px; }
.secret-pill { min-height: 22px; display: inline-flex; align-items: center; gap: 4px; padding: 0 8px; border-radius: 6px; color: var(--danger); background: var(--danger-soft); font-weight: 850; }
.config-state { min-height: 220px; display: flex; align-items: center; justify-content: center; color: var(--muted); font-weight: 800; }
button:disabled { cursor: not-allowed; opacity: 0.55; }
.spinning { animation: spin 800ms linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 960px) { .config-toolbar { align-items: flex-start; flex-direction: column; } .config-grid { grid-template-columns: 1fr; } }
</style>
