<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useWebAuthStore } from '../../store/modules/auth'

const authStore = useWebAuthStore()
const { displayName } = storeToRefs(authStore)

const sections = [
  {
    title: '账号资料',
    tag: '已开放',
    tagColor: 'var(--c-success)',
    tagBg: 'var(--c-success-soft)',
    desc: '维护邮箱、显示名称和登录密码。',
    to: '/user/profile',
  },
  {
    title: '购买与实例',
    tag: '待开放',
    tagColor: 'var(--c-warning)',
    tagBg: 'var(--c-warning-soft)',
    desc: '后续阶段开放购买、支付和实例相关能力。',
    to: '',
  },
  {
    title: '订单与账单',
    tag: '待开放',
    tagColor: 'var(--c-warning)',
    tagBg: 'var(--c-warning-soft)',
    desc: '订单、支付、账单和发票当前不读取真实业务数据。',
    to: '',
  },
  {
    title: '安全设置',
    tag: '待开放',
    tagColor: 'var(--c-warning)',
    tagBg: 'var(--c-warning-soft)',
    desc: '当前仅在账号资料页开放登录密码修改。',
    to: '',
  },
]
</script>

<template>
  <section class="page content-page">
    <div class="section-pad">
      <div class="sec-header" style="margin-bottom: clamp(28px, 4vw, 48px);">
        <p class="label">用户中心</p>
        <h2>{{ displayName }}，欢迎进入控制台</h2>
        <p>当前已接入用户登录态。以下入口仍为占位，不读取真实业务数据。</p>
      </div>

      <div class="user-grid">
        <RouterLink v-for="s in sections" :key="s.title" class="user-card" :to="s.to || '/user'">
          <span class="tag" :style="{ color: s.tagColor, background: s.tagBg }">{{ s.tag }}</span>
          <h3>{{ s.title }}</h3>
          <p>{{ s.desc }}</p>
        </RouterLink>
      </div>
    </div>
  </section>
</template>
