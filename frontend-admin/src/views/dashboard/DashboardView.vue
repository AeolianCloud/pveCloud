<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()

// 欢迎语：按当前小时段
const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

// 格式化最后登录时间
const lastLogin = computed(() => {
  const t = authStore.user?.last_login_at
  if (!t) return '首次登录'
  return new Date(t).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
})
</script>

<template>
  <div class="dashboard">
    <!-- 欢迎横幅 -->
    <div class="welcome-banner">
      <div class="welcome-text">
        <p class="welcome-greeting">{{ greeting }}，{{ authStore.user?.nickname || authStore.user?.username }} 👋</p>
        <p class="welcome-sub">上次登录：{{ lastLogin }}</p>
      </div>
      <div class="welcome-tags">
        <n-tag
          v-for="role in authStore.user?.roles"
          :key="role.id"
          type="info"
          size="small"
          :bordered="false"
        >
          {{ role.label }}
        </n-tag>
      </div>
    </div>

    <!-- 统计卡片区 -->
    <div class="stat-grid">
      <div class="stat-card">
        <div class="stat-icon stat-icon--blue">
          <n-icon size="22"><server-outline /></n-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">—</div>
          <div class="stat-label">虚拟机节点</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon--green">
          <n-icon size="22"><cube-outline /></n-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">—</div>
          <div class="stat-label">运行容器</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon--orange">
          <n-icon size="22"><hardware-chip-outline /></n-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">—</div>
          <div class="stat-label">CPU 使用率</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon--purple">
          <n-icon size="22"><pie-chart-outline /></n-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">—</div>
          <div class="stat-label">内存使用率</div>
        </div>
      </div>
    </div>

    <!-- 提示：功能待接入 -->
    <n-empty
      description="集群数据接入后将在此展示实时监控指标"
      style="margin-top: 48px;"
    />
  </div>
</template>

<script lang="ts">
import {
  ServerOutline,
  CubeOutline,
  HardwareChipOutline,
  PieChartOutline,
} from '@vicons/ionicons5'
export default { components: { ServerOutline, CubeOutline, HardwareChipOutline, PieChartOutline } }
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* ========== 欢迎横幅 ========== */
.welcome-banner {
  background: #fff;
  border-radius: 10px;
  padding: 24px 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.welcome-greeting {
  font-size: 18px;
  font-weight: 600;
  color: #18181c;
  margin-bottom: 6px;
}

.welcome-sub {
  font-size: 13px;
  color: #909399;
}

.welcome-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

/* ========== 统计卡片 ========== */
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  background: #fff;
  border-radius: 10px;
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon--blue   { background: #e8f4fd; color: #4fa8e8; }
.stat-icon--green  { background: #e8f8f0; color: #36b37e; }
.stat-icon--orange { background: #fff3e8; color: #f5a623; }
.stat-icon--purple { background: #f0eeff; color: #7c5cbf; }

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: #18181c;
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

/* ========== 响应式：小屏改为 2 列 ========== */
@media (max-width: 900px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
