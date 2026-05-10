<script setup lang="ts">
import {
  CheckmarkCircleOutline,
  KeyOutline,
  PersonCircleOutline,
  RefreshOutline,
  WarningOutline,
} from '@vicons/ionicons5'
import { NButton, NCard, NDescriptions, NDescriptionsItem, NGi, NGrid, NIcon, NTag } from 'naive-ui'
import { computed, h, onMounted, ref, type Component } from 'vue'

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
const canViewDashboard = computed(() => permissionStore.hasPermission('page.dashboard'))

const metricMeta: Record<string, { icon: Component }> = {
  active_admins: { icon: PersonCircleOutline },
  active_roles: { icon: CheckmarkCircleOutline },
  active_sessions: { icon: KeyOutline },
  risk_logs_today: { icon: WarningOutline },
}

const metricCards = computed(() =>
  metrics.value.map((metric) => ({
    key: metric.key,
    title: metric.title,
    value: `${metric.value.toLocaleString()}${metric.unit ? ` ${metric.unit}` : ''}`,
    icon: metricMeta[metric.key]?.icon || PersonCircleOutline,
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

const refreshIcon = () => h(NIcon, null, { default: () => h(RefreshOutline) })

onMounted(() => {
  void loadDashboard()
})
</script>

<template>
  <div class="dashboard-page">
    <div class="dashboard-page__header">
      <h2>工作台</h2>
      <NButton :loading="loading" :render-icon="refreshIcon" @click="loadDashboard">刷新</NButton>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadDashboard">
      <template v-if="!canViewDashboard">
        <NCard>
          <EmptyState title="暂无权限" description="当前账号没有控制台数据查看权限。" />
        </NCard>
      </template>

      <template v-else-if="metricCards.length === 0">
        <NCard>
          <EmptyState title="暂无数据" description="当前没有可展示的指标。" />
        </NCard>
      </template>

      <template v-else>
        <NGrid :x-gap="16" :y-gap="16" cols="1 s:2 m:2 l:4" responsive="screen">
          <NGi v-for="item in metricCards" :key="item.key">
            <MetricCard :title="item.title" :value="item.value" :icon="item.icon" />
          </NGi>
        </NGrid>

        <NCard title="登录信息">
          <NDescriptions bordered :column="4" label-placement="left" size="small">
            <NDescriptionsItem label="账号">
              {{ authStore.admin?.username || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem label="姓名">
              {{ authStore.admin?.display_name || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem label="状态">
              <NTag :type="authStore.admin?.status === 'active' ? 'success' : 'default'" size="small">
                {{ authStore.admin?.status === 'active' ? '正常' : authStore.admin?.status || '-' }}
              </NTag>
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>
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
</style>
