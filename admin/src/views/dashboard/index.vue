<script setup lang="ts">
import { Checked, Key, UserFilled, WarningFilled } from '@element-plus/icons-vue'
import { computed, onMounted, ref, type Component } from 'vue'

import EmptyState from '../../components/EmptyState.vue'
import PageToolbar from '../../components/PageToolbar.vue'
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

const metricMeta: Record<string, { icon: Component; description: string }> = {
  active_admins: {
    icon: UserFilled,
    description: '当前启用中的后台管理员数量。',
  },
  active_roles: {
    icon: Checked,
    description: '当前可用的后台角色数量。',
  },
  active_sessions: {
    icon: Key,
    description: '当前仍然有效的后台会话数量。',
  },
  risk_logs_today: {
    icon: WarningFilled,
    description: '今天产生的高危操作日志数量。',
  },
}

const metricCards = computed(() =>
  metrics.value.map((metric) => ({
    key: metric.key,
    title: metric.title,
    value: `${metric.value.toLocaleString()}${metric.unit ? ` ${metric.unit}` : ''}`,
    icon: metricMeta[metric.key]?.icon || UserFilled,
    description: metricMeta[metric.key]?.description || '基础后台首页指标。',
  })),
)

const sessionExpiresAt = computed(() => formatDateTime(authStore.session?.expires_at))
const lastIssuedAt = computed(() => formatDateTime(authStore.session?.issued_at))
const hasDashboardPermission = computed(() => permissionStore.hasAllPermissions(['dashboard:view']))

async function loadDashboard() {
  loading.value = true
  errorMessage.value = ''

  try {
    const result = await getAdminDashboard()
    authStore.applyDashboardPayload(result)
    metrics.value = result.metrics
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '控制台数据加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function formatDateTime(value?: string) {
  if (!value) {
    return '未提供'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => {
  void loadDashboard()
})
</script>

<template>
  <section class="dashboard-page">
    <div class="dashboard-page__hero">
      <PageToolbar
        eyebrow="Operations"
        title="控制台"
        description="当前后台前端只保留登录、控制台和 403 页面，控制台聚焦最小可运营信息与权限快照。"
      >
        <template #actions>
          <el-button type="primary" size="large" :loading="loading" @click="loadDashboard">刷新指标</el-button>
        </template>
      </PageToolbar>

      <div class="dashboard-page__signals">
        <div class="dashboard-page__signal">
          <span>角色数量</span>
          <strong>{{ permissionStore.roleIds.length }}</strong>
        </div>
        <div class="dashboard-page__signal">
          <span>权限码数量</span>
          <strong>{{ permissionStore.permissionCodes.length }}</strong>
        </div>
        <div class="dashboard-page__signal">
          <span>首页权限</span>
          <strong>{{ hasDashboardPermission ? '已具备' : '未具备' }}</strong>
        </div>
      </div>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadDashboard">
      <template v-if="metricCards.length === 0">
        <el-card class="dashboard-page__panel">
          <EmptyState title="暂无首页指标" description="当前接口没有返回可展示的基础后台指标。" />
        </el-card>
      </template>

      <template v-else>
        <el-row :gutter="18">
          <el-col v-for="item in metricCards" :key="item.key" :xs="24" :sm="12" :xl="6">
            <MetricCard
              :title="item.title"
              :value="item.value"
              :description="item.description"
              :icon="item.icon"
            />
          </el-col>
        </el-row>

        <el-row :gutter="18" class="dashboard-page__details">
          <el-col :xs="24" :lg="14">
            <el-card class="dashboard-page__panel">
              <template #header>
                <div class="dashboard-page__panel-header">
                  <strong>当前管理员</strong>
                  <span>当前登录态的最小可见摘要</span>
                </div>
              </template>

              <el-descriptions :column="1" border>
                <el-descriptions-item label="显示名称">
                  {{ authStore.admin?.display_name || '未提供' }}
                </el-descriptions-item>
                <el-descriptions-item label="登录账号">
                  {{ authStore.admin?.username || '未提供' }}
                </el-descriptions-item>
                <el-descriptions-item label="账号状态">
                  <el-tag effect="dark" type="success">{{ authStore.admin?.status || '未知' }}</el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="会话签发时间">
                  {{ lastIssuedAt }}
                </el-descriptions-item>
                <el-descriptions-item label="会话过期时间">
                  {{ sessionExpiresAt }}
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>

          <el-col :xs="24" :lg="10">
            <el-card class="dashboard-page__panel">
              <template #header>
                <div class="dashboard-page__panel-header">
                  <strong>权限快照</strong>
                  <span>前端菜单与兼容菜单的最小收口态</span>
                </div>
              </template>

              <el-descriptions :column="1" border>
                <el-descriptions-item label="角色数量">
                  {{ permissionStore.roleIds.length }}
                </el-descriptions-item>
                <el-descriptions-item label="权限码数量">
                  {{ permissionStore.permissionCodes.length }}
                </el-descriptions-item>
                <el-descriptions-item label="前端菜单数量">
                  {{ permissionStore.sidebarMenus.length }}
                </el-descriptions-item>
                <el-descriptions-item label="兼容菜单快照">
                  {{ permissionStore.menuSnapshot.length }}
                </el-descriptions-item>
                <el-descriptions-item label="首页访问权限">
                  <el-tag :type="hasDashboardPermission ? 'success' : 'danger'" effect="dark">
                    {{ hasDashboardPermission ? '已具备' : '未具备' }}
                  </el-tag>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>
        </el-row>
      </template>
    </QueryState>
  </section>
</template>

<style scoped>
.dashboard-page {
  display: grid;
  gap: 22px;
}

.dashboard-page__hero {
  display: grid;
  gap: 22px;
  padding: 26px;
  border-radius: var(--pc-card-radius);
  background:
    radial-gradient(circle at top right, rgba(14, 165, 233, 0.16), transparent 28%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.94), rgba(255, 255, 255, 0.72));
  box-shadow: var(--pc-card-shadow);
  backdrop-filter: blur(16px);
}

.dashboard-page__signals {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.dashboard-page__signal {
  padding: 16px 18px;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.78);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.12);
}

.dashboard-page__signal span {
  display: block;
  margin-bottom: 8px;
  color: var(--pc-subtle-text);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.dashboard-page__signal strong {
  color: var(--pc-title-text);
  font-family: var(--pc-display-font);
  font-size: 26px;
}

.dashboard-page__details {
  margin-top: 0;
}

.dashboard-page__panel {
  border-radius: var(--pc-card-radius);
  background: rgba(255, 255, 255, 0.84);
  box-shadow: var(--pc-card-shadow);
}

.dashboard-page__panel-header {
  display: grid;
  gap: 4px;
}

.dashboard-page__panel-header strong {
  color: var(--pc-title-text);
  font-size: 16px;
}

.dashboard-page__panel-header span {
  color: var(--pc-muted-text);
  font-size: 12px;
}

@media (max-width: 991px) {
  .dashboard-page__hero {
    padding: 20px;
  }

  .dashboard-page__signals {
    grid-template-columns: 1fr;
  }
}
</style>
