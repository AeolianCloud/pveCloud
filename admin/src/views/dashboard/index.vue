<script setup lang="ts">
import { Checked, Key, Refresh, UserFilled, WarningFilled } from '@element-plus/icons-vue'
import { computed, onMounted, ref, type Component } from 'vue'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import { getAdminDashboard, type DashboardMetric } from '../../api/dashboard'
import { useAuthStore } from '../../store/modules/auth'
import { usePermissionStore } from '../../store/modules/permission'
import MetricCard from './components/MetricCard.vue'

const authStore = useAuthStore()
const permissionStore = usePermissionStore()

const loading = ref(false)
const errorMessage = ref('')
const metrics = ref<DashboardMetric[]>([])
const canViewDashboard = computed(() => permissionStore.hasPermission('dashboard:view'))

const metricMeta: Record<string, { icon: Component }> = {
  active_admins: { icon: UserFilled },
  active_roles: { icon: Checked },
  active_sessions: { icon: Key },
  risk_logs_today: { icon: WarningFilled },
}

const metricCards = computed(() =>
  metrics.value.map((metric) => ({
    key: metric.key,
    title: metric.title,
    value: `${metric.value.toLocaleString()}${metric.unit ? ` ${metric.unit}` : ''}`,
    icon: metricMeta[metric.key]?.icon || UserFilled,
  })),
)

async function loadDashboard() {
  if (!canViewDashboard.value) {
    metrics.value = []
    errorMessage.value = ''
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    const result = await getAdminDashboard()
    authStore.applyDashboardPayload(result)
    metrics.value = result.metrics
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '数据加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadDashboard()
})
</script>

<template>
  <div class="dashboard-page">
    <div class="dashboard-page__header">
      <h2>工作台</h2>
      <el-button :icon="Refresh" :loading="loading" @click="loadDashboard">刷新</el-button>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadDashboard">
      <template v-if="!canViewDashboard">
        <el-card>
          <EmptyState title="暂无权限" description="当前账号没有控制台数据查看权限。" />
        </el-card>
      </template>

      <template v-else-if="metricCards.length === 0">
        <el-card>
          <EmptyState title="暂无数据" description="当前没有可展示的指标。" />
        </el-card>
      </template>

      <template v-else>
        <el-row :gutter="16">
          <el-col v-for="item in metricCards" :key="item.key" :xs="24" :sm="12" :lg="6">
            <MetricCard :title="item.title" :value="item.value" :icon="item.icon" />
          </el-col>
        </el-row>

        <el-card>
          <template #header>登录信息</template>
          <el-descriptions :column="{ xs: 1, sm: 2, lg: 4 }" border>
            <el-descriptions-item label="账号">
              {{ authStore.admin?.username || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="姓名">
              {{ authStore.admin?.display_name || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="authStore.admin?.status === 'active' ? 'success' : 'info'" size="small">
                {{ authStore.admin?.status === 'active' ? '正常' : authStore.admin?.status || '-' }}
              </el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </template>
    </QueryState>
  </div>
</template>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dashboard-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dashboard-page__header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.el-col {
  margin-bottom: 16px;
}
</style>
