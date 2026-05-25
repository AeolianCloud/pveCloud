<script setup lang="ts">
import {
  CalendarClearOutline,
  CardOutline,
  ChatbubblesOutline,
  CheckmarkCircleOutline,
  DocumentTextOutline,
  GridOutline,
  KeyOutline,
  PersonCircleOutline,
  ReceiptOutline,
  RefreshOutline,
  ServerOutline,
  ShieldCheckmarkOutline,
  TimeOutline,
  WarningOutline,
} from '@vicons/ionicons5'
import { NButton, NCard, NDescriptions, NDescriptionsItem, NGi, NGrid, NIcon, NTag } from 'naive-ui'
import { computed, h, onMounted, ref, type Component } from 'vue'
import { RouterLink } from 'vue-router'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import { getAdminDashboard, type DashboardBusinessMetric, type DashboardMetric } from '../../api/dashboard'
import { useAuthStore } from '../../store/modules/auth'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'
import type { SidebarMenuItem } from '../../utils/permission'
import MetricCard from './components/MetricCard.vue'

const authStore = useAuthStore()
const permissionStore = usePermissionStore()

const loading = ref(false)
const errorMessage = ref('')
const metrics = ref<DashboardMetric[]>([])
const businessMetrics = ref<DashboardBusinessMetric[]>([])
const lastLoadedAt = ref('')
const canViewDashboard = computed(() => permissionStore.hasPermission('page.dashboard'))

const metricMeta: Record<string, { icon: Component }> = {
  active_admins: { icon: PersonCircleOutline },
  active_roles: { icon: CheckmarkCircleOutline },
  active_sessions: { icon: KeyOutline },
  audit_logs_today: { icon: WarningOutline },
}

const businessMetricMeta: Record<string, { icon: Component }> = {
  pending_orders: { icon: ReceiptOutline },
  order_errors: { icon: WarningOutline },
  instance_errors: { icon: ServerOutline },
  failed_async_tasks: { icon: TimeOutline },
  pending_tickets: { icon: ChatbubblesOutline },
  invoice_todo: { icon: DocumentTextOutline },
  payment_exceptions: { icon: CardOutline },
}

const metricCards = computed(() =>
  metrics.value.map((metric) => ({
    key: metric.key,
    title: metric.title,
    value: `${metric.value.toLocaleString()}${metric.unit ? ` ${metric.unit}` : ''}`,
    icon: metricMeta[metric.key]?.icon || PersonCircleOutline,
  })),
)

const businessMetricCards = computed(() =>
  businessMetrics.value.map((metric) => ({
    key: metric.key,
    title: metric.title,
    value: `${metric.value.toLocaleString()}${metric.unit ? ` ${metric.unit}` : ''}`,
    description: metric.description,
    targetPath: metric.target_path,
    canOpen: Boolean(metric.target_path && (!metric.target_permission || permissionStore.hasPermission(metric.target_permission))),
    severity: metric.severity,
    tagType: metric.severity === 'error' ? 'error' as const : metric.severity === 'warning' ? 'warning' as const : 'info' as const,
    icon: businessMetricMeta[metric.key]?.icon || WarningOutline,
  })),
)

const quickEntries = computed(() => flattenMenus(permissionStore.sidebarMenus).filter((item) => item.path !== '/dashboard'))

const overviewCards = computed(() => [
  {
    key: 'roles',
    title: '当前角色',
    value: permissionStore.roleIds.length.toLocaleString(),
    description: '已授予角色数量',
    icon: ShieldCheckmarkOutline,
  },
  {
    key: 'permissions',
    title: '权限节点',
    value: permissionStore.permissionCodes.length.toLocaleString(),
    description: '当前账号可用权限',
    icon: CheckmarkCircleOutline,
  },
  {
    key: 'menus',
    title: '可访问页面',
    value: flattenMenus(permissionStore.sidebarMenus).length.toLocaleString(),
    description: '侧栏菜单可见入口',
    icon: GridOutline,
  },
])

const sessionStatus = computed(() => {
  const expiresAt = authStore.session?.expires_at
  if (!expiresAt) return { label: '未知', type: 'default' as const }

  const expiresTime = new Date(expiresAt).getTime()
  if (Number.isNaN(expiresTime)) return { label: '未知', type: 'default' as const }

  return expiresTime > Date.now()
    ? { label: '有效', type: 'success' as const }
    : { label: '已过期', type: 'warning' as const }
})

function flattenMenus(items: SidebarMenuItem[]): SidebarMenuItem[] {
  return items.flatMap((item) => (item.children?.length ? flattenMenus(item.children) : [item]))
}

async function loadDashboard() {
  if (!canViewDashboard.value) {
    metrics.value = []
    businessMetrics.value = []
    errorMessage.value = ''
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    const result = await getAdminDashboard()
    authStore.applyDashboardPayload(result)
    metrics.value = result.metrics
    businessMetrics.value = result.business_metrics || []
    lastLoadedAt.value = formatDateTime(new Date().toISOString())
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
    <div class="dashboard-hero">
      <div>
        <div class="dashboard-hero__eyebrow">PVE Cloud Admin</div>
        <h2>工作台</h2>
        <p>集中查看基础后台运行状态、业务待办异常、当前会话和可访问管理入口。</p>
      </div>
      <div class="dashboard-hero__actions">
        <div class="dashboard-hero__time">最近更新：{{ lastLoadedAt || '-' }}</div>
        <NButton type="primary" :loading="loading" :render-icon="refreshIcon" @click="loadDashboard">刷新数据</NButton>
      </div>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadDashboard">
      <template v-if="!canViewDashboard">
        <NCard>
          <EmptyState title="暂无权限" description="当前账号没有控制台数据查看权限。" />
        </NCard>
      </template>

      <template v-else-if="metricCards.length === 0 && businessMetricCards.length === 0">
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

        <div class="dashboard-page__content">
          <div class="dashboard-page__main">
            <NCard title="业务待办与异常">
              <div v-if="businessMetricCards.length > 0" class="business-metric-grid">
                <RouterLink v-for="item in businessMetricCards.filter((metric) => metric.canOpen)" :key="item.key" :to="item.targetPath || '/dashboard'" class="business-metric-card business-metric-card--link">
                  <div class="business-metric-card__icon" :class="`business-metric-card__icon--${item.severity}`">
                    <NIcon :size="20"><component :is="item.icon" /></NIcon>
                  </div>
                  <div class="business-metric-card__body">
                    <div class="business-metric-card__top">
                      <span>{{ item.title }}</span>
                      <NTag :type="item.tagType" size="small" round>{{ item.severity === 'error' ? '异常' : item.severity === 'warning' ? '待办' : '信息' }}</NTag>
                    </div>
                    <div class="business-metric-card__value">{{ item.value }}</div>
                    <div class="business-metric-card__description">{{ item.description }}</div>
                  </div>
                </RouterLink>
                <div v-for="item in businessMetricCards.filter((metric) => !metric.canOpen)" :key="item.key" class="business-metric-card business-metric-card--disabled">
                  <div class="business-metric-card__icon" :class="`business-metric-card__icon--${item.severity}`">
                    <NIcon :size="20"><component :is="item.icon" /></NIcon>
                  </div>
                  <div class="business-metric-card__body">
                    <div class="business-metric-card__top">
                      <span>{{ item.title }}</span>
                      <NTag :type="item.tagType" size="small" round>{{ item.severity === 'error' ? '异常' : item.severity === 'warning' ? '待办' : '信息' }}</NTag>
                    </div>
                    <div class="business-metric-card__value">{{ item.value }}</div>
                    <div class="business-metric-card__description">{{ item.description }}</div>
                  </div>
                </div>
              </div>
              <EmptyState v-else title="暂无业务指标" description="当前没有可展示的业务待办或异常。" />
            </NCard>

            <NCard title="账号与会话">
              <NDescriptions bordered :column="2" label-placement="left" size="small">
                <NDescriptionsItem label="账号">
                  {{ authStore.admin?.username || '-' }}
                </NDescriptionsItem>
                <NDescriptionsItem label="姓名">
                  {{ authStore.admin?.display_name || '-' }}
                </NDescriptionsItem>
                <NDescriptionsItem label="账号状态">
                  <NTag :type="authStore.admin?.status === 'active' ? 'success' : 'default'" size="small">
                    {{ authStore.admin?.status === 'active' ? '正常' : authStore.admin?.status || '-' }}
                  </NTag>
                </NDescriptionsItem>
                <NDescriptionsItem label="会话状态">
                  <NTag :type="sessionStatus.type" size="small">{{ sessionStatus.label }}</NTag>
                </NDescriptionsItem>
                <NDescriptionsItem label="签发时间">
                  {{ formatDateTime(authStore.session?.issued_at) }}
                </NDescriptionsItem>
                <NDescriptionsItem label="过期时间">
                  {{ formatDateTime(authStore.session?.expires_at) }}
                </NDescriptionsItem>
              </NDescriptions>
            </NCard>

            <NCard title="快捷入口">
              <div v-if="quickEntries.length > 0" class="quick-entry-grid">
                <RouterLink v-for="entry in quickEntries" :key="entry.path" :to="entry.path" class="quick-entry">
                  <div class="quick-entry__icon">
                    <NIcon :size="18"><GridOutline /></NIcon>
                  </div>
                  <div>
                    <div class="quick-entry__title">{{ entry.title }}</div>
                    <div class="quick-entry__path">{{ entry.path }}</div>
                  </div>
                </RouterLink>
              </div>
              <EmptyState v-else title="暂无快捷入口" description="当前账号除工作台外暂无其它可访问菜单。" />
            </NCard>
          </div>

          <div class="dashboard-page__side">
            <NCard title="权限概览">
              <div class="overview-list">
                <div v-for="item in overviewCards" :key="item.key" class="overview-item">
                  <div class="overview-item__icon">
                    <NIcon :size="18"><component :is="item.icon" /></NIcon>
                  </div>
                  <div class="overview-item__body">
                    <div class="overview-item__top">
                      <span>{{ item.title }}</span>
                      <strong>{{ item.value }}</strong>
                    </div>
                    <div class="overview-item__description">{{ item.description }}</div>
                  </div>
                </div>
              </div>
            </NCard>

            <NCard title="运行提示">
              <div class="hint-list">
                <div class="hint-item">
                  <NIcon :size="18"><CalendarClearOutline /></NIcon>
                  <span>首页指标只读汇总当前已开放的基础后台和业务运营数据。</span>
                </div>
                <div class="hint-item">
                  <NIcon :size="18"><ShieldCheckmarkOutline /></NIcon>
                  <span>前端菜单只做可见性控制，最终授权以服务端 RBAC 为准。</span>
                </div>
                <div class="hint-item">
                  <NIcon :size="18"><TimeOutline /></NIcon>
                  <span>长时间停留后建议刷新数据，确保权限和会话信息同步。</span>
                </div>
                <div class="hint-item">
                  <NIcon :size="18"><WarningOutline /></NIcon>
                  <span>业务卡片只提供入口跳转，处理动作仍在对应页面按权限和状态执行。</span>
                </div>
              </div>
            </NCard>
          </div>
        </div>
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

.dashboard-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 24px;
  border-radius: 18px;
  color: #fff;
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.42), transparent 32%),
    linear-gradient(135deg, #0f172a 0%, #1d4ed8 100%);
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.18);
}

.dashboard-hero h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  line-height: 1.25;
}

.dashboard-hero p {
  margin: 8px 0 0;
  color: rgba(255, 255, 255, 0.78);
}

.dashboard-hero__eyebrow {
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.58);
}

.dashboard-hero__actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
  flex-shrink: 0;
}

.dashboard-hero__time {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.68);
}

.dashboard-page__content {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 16px;
}

.dashboard-page__main,
.dashboard-page__side {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.quick-entry-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.business-metric-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.business-metric-card {
  display: flex;
  gap: 12px;
  min-width: 0;
  padding: 14px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 14px;
  color: inherit;
  text-decoration: none;
  background: rgba(248, 250, 252, 0.78);
}

.business-metric-card--link {
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.business-metric-card--link:hover {
  border-color: rgba(37, 99, 235, 0.32);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.08);
  transform: translateY(-1px);
}

.business-metric-card--disabled {
  opacity: 0.78;
}

.business-metric-card__icon {
  width: 42px;
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 12px;
}

.business-metric-card__icon--info {
  color: #2563eb;
  background: rgba(37, 99, 235, 0.1);
}

.business-metric-card__icon--warning {
  color: #d97706;
  background: rgba(217, 119, 6, 0.12);
}

.business-metric-card__icon--error {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.1);
}

.business-metric-card__body {
  min-width: 0;
  flex: 1;
}

.business-metric-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  font-size: 13px;
  font-weight: 600;
  color: rgba(15, 23, 42, 0.72);
}

.business-metric-card__value {
  margin-top: 6px;
  font-size: 24px;
  font-weight: 750;
  line-height: 1.1;
  color: #0f172a;
}

.business-metric-card__description {
  margin-top: 5px;
  color: rgba(15, 23, 42, 0.54);
  font-size: 12px;
  line-height: 1.5;
}

.quick-entry {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding: 14px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 14px;
  color: inherit;
  text-decoration: none;
  background: rgba(248, 250, 252, 0.72);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.quick-entry:hover {
  border-color: rgba(37, 99, 235, 0.32);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.08);
  transform: translateY(-1px);
}

.quick-entry__icon,
.overview-item__icon {
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 12px;
  color: #2563eb;
  background: rgba(37, 99, 235, 0.1);
}

.quick-entry__title {
  overflow: hidden;
  font-weight: 600;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.quick-entry__path {
  margin-top: 2px;
  overflow: hidden;
  font-size: 12px;
  color: rgba(15, 23, 42, 0.46);
  white-space: nowrap;
  text-overflow: ellipsis;
}

.overview-list,
.hint-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.overview-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.86);
}

.overview-item__body {
  min-width: 0;
  flex: 1;
}

.overview-item__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}

.overview-item__top strong {
  font-size: 20px;
  color: #0f172a;
}

.overview-item__description {
  margin-top: 2px;
  font-size: 12px;
  color: rgba(15, 23, 42, 0.52);
}

.hint-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  color: rgba(15, 23, 42, 0.68);
  line-height: 1.6;
}

.hint-item .n-icon {
  margin-top: 3px;
  color: #2563eb;
  flex-shrink: 0;
}

@media (max-width: 1080px) {
  .dashboard-page__content {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .dashboard-hero {
    flex-direction: column;
    padding: 20px;
  }

  .dashboard-hero h2 {
    font-size: 24px;
  }

  .dashboard-hero__actions {
    width: 100%;
    align-items: stretch;
  }

  .dashboard-hero__time {
    text-align: left;
  }

  .quick-entry-grid {
    grid-template-columns: 1fr;
  }

  .business-metric-grid {
    grid-template-columns: 1fr;
  }
}
</style>
