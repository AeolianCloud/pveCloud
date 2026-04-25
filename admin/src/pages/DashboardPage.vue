<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RefreshCw, ShieldCheck } from 'lucide-vue-next'

import { getAdminDashboard } from '../api/dashboard'
import { useAuthStore } from '../stores/auth'
import type { DashboardMetric } from '../types/dashboard'

const auth = useAuthStore()
const loading = ref(false)
const errorMessage = ref('')
const metrics = ref<DashboardMetric[]>([])

const permissionCount = computed(() => auth.permissionCodes.length)

async function loadDashboard() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await getAdminDashboard()
    auth.applyDashboard(result)
    metrics.value = result.metrics
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '首页数据加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboard)
</script>

<template>
  <section class="dashboard-overview">
    <div class="summary-panel">
      <ShieldCheck :size="28" aria-hidden="true" />
      <div>
        <h2>{{ auth.admin?.display_name || auth.admin?.username }}</h2>
        <p>{{ auth.admin?.username }} 已登录</p>
      </div>
      <button class="icon-only-button panel-action" type="button" title="刷新" @click="loadDashboard">
        <RefreshCw :size="18" aria-hidden="true" />
      </button>
    </div>

    <div class="metric-grid">
      <div v-for="metric in metrics" :key="metric.key" class="metric-card">
        <span>{{ metric.title }}</span>
        <strong>{{ metric.value }}{{ metric.unit || '' }}</strong>
      </div>
      <div v-if="loading" class="metric-card">
        <span>加载中</span>
        <strong>...</strong>
      </div>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <div class="permission-panel">
      <h2>权限码 {{ permissionCount }}</h2>
      <div class="permission-list">
        <span v-for="code in auth.permissionCodes" :key="code">{{ code }}</span>
        <span v-if="auth.permissionCodes.length === 0">暂无权限</span>
      </div>
    </div>
  </section>
</template>
