<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useWebAuthStore } from '../../store/modules/auth'

const authStore = useWebAuthStore()
const { displayName, user } = storeToRefs(authStore)

const actions = [
  { title: '账号资料', tag: '已开放', desc: '维护邮箱、显示名称和登录密码。', to: '/user/profile', tone: 'success' },
  { title: '实名认证', tag: '购买准备', desc: '提交个人实名，状态以后端核验为准。', to: '/user/real-name', tone: 'primary' },
  { title: '价格方案', tag: '公开展示', desc: '查看当前公开服务器套餐和周期价格。', to: '/pricing', tone: 'primary' },
  { title: '购买与实例', tag: '待开放', desc: '订单、支付、实例和工单后续阶段开放。', to: '', tone: 'warning' },
]
</script>

<template>
  <section class="user-page page-shell">
    <div class="user-hero surface">
      <div>
        <p class="section-label">User Console</p>
        <h1 class="page-title">{{ displayName }}，欢迎回来</h1>
        <p class="page-copy">当前控制台开放账号资料和实名认证，购买与实例交付入口会在业务契约开放后接入。</p>
      </div>
      <div class="profile-box">
        <span>{{ user?.username?.slice(0, 1).toUpperCase() || 'U' }}</span>
        <strong>{{ user?.username }}</strong>
        <small>{{ user?.email }}</small>
      </div>
    </div>

    <div class="summary-grid">
      <div class="summary-card card"><span>会话状态</span><strong>已登录</strong></div>
      <div class="summary-card card"><span>当前阶段</span><strong>账号自助</strong></div>
      <div class="summary-card card"><span>业务边界</span><strong>未开放下单</strong></div>
    </div>

    <div class="action-grid">
      <RouterLink v-for="item in actions" :key="item.title" class="action-card card" :to="item.to || '/user'">
        <span class="tag" :class="item.tone === 'warning' ? 'tag-warning' : item.tone === 'success' ? 'tag-success' : 'tag-primary'">{{ item.tag }}</span>
        <h2>{{ item.title }}</h2>
        <p>{{ item.desc }}</p>
      </RouterLink>
    </div>
  </section>
</template>

<style scoped>
.user-page {
  display: grid;
  gap: 22px;
}

.user-hero {
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 24px;
  padding: clamp(24px, 4vw, 38px);
}

.user-hero .page-copy {
  max-width: 720px;
  margin-top: 14px;
}

.profile-box {
  min-width: 240px;
  display: grid;
  align-content: center;
  gap: 8px;
  padding: 22px;
  border-radius: 16px;
  background: var(--c-surface-dim);
}

.profile-box span {
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  border-radius: 14px;
  color: #fff;
  background: var(--c-primary);
  font-weight: 900;
}

.profile-box small {
  color: var(--c-text-2);
}

.summary-grid,
.action-grid {
  display: grid;
  gap: 16px;
}

.summary-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.summary-card {
  display: grid;
  gap: 6px;
  padding: 18px;
}

.summary-card span {
  color: var(--c-text-3);
  font-weight: 800;
}

.summary-card strong {
  font-size: 1.35rem;
  letter-spacing: -0.04em;
}

.action-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.action-card {
  display: grid;
  gap: 16px;
  align-content: start;
  padding: 22px;
  min-height: 210px;
}

.action-card h2 {
  font-size: 1.3rem;
  letter-spacing: -0.04em;
}

.action-card p {
  color: var(--c-text-2);
  line-height: 1.7;
}

@media (max-width: 980px) {
  .action-grid,
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .user-hero {
    flex-direction: column;
  }
}

@media (max-width: 620px) {
  .action-grid,
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
