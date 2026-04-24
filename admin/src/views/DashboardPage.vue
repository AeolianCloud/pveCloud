<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDashboardStats, type DashboardStats } from '../api/dashboard'

const stats = ref<DashboardStats | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    stats.value = await getDashboardStats()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="panel">
    <p class="tag">DASHBOARD</p>
    <h2>管理后台</h2>

    <p v-if="loading">加载中...</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <div v-else class="stats-grid">
      <div class="stat-card">
        <span class="stat-value">{{ stats?.total_orders ?? 0 }}</span>
        <span class="stat-label">总订单</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.pending_orders ?? 0 }}</span>
        <span class="stat-label">待支付</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.total_instances ?? 0 }}</span>
        <span class="stat-label">总实例</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.running_instances ?? 0 }}</span>
        <span class="stat-label">运行中</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.total_users ?? 0 }}</span>
        <span class="stat-label">总用户</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.pending_tasks ?? 0 }}</span>
        <span class="stat-label">待处理任务</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

h2 { margin: 0 0 16px; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 16px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  padding: 16px;
  border-radius: 12px;
  background: rgba(85, 114, 87, 0.08);
}

.stat-value {
  font-size: 1.8em;
  font-weight: 700;
}

.stat-label {
  font-size: 0.85em;
  color: #666;
}

.error { color: #c00; }
</style>
