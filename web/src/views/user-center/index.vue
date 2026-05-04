<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useWebAuthStore } from '../../store/modules/auth'

const authStore = useWebAuthStore()
const { displayName } = storeToRefs(authStore)

const sections = [
  {
    title: '账号资料',
    tag: '已开放',
    tagColor: 'var(--c-green)',
    tagBg: 'var(--c-green-soft)',
    desc: '维护邮箱、显示名称和登录密码。',
    to: '/user/profile',
  },
  {
    title: '实例管理',
    tag: '待开放',
    tagColor: 'var(--c-orange)',
    tagBg: 'var(--c-orange-soft)',
    desc: '查看、启停、重启和销毁云服务器实例。',
    to: '',
  },
  {
    title: '账单中心',
    tag: '待开放',
    tagColor: 'var(--c-orange)',
    tagBg: 'var(--c-orange-soft)',
    desc: '查看消费明细、充值记录和发票管理。',
    to: '',
  },
  {
    title: '安全设置',
    tag: '待开放',
    tagColor: 'var(--c-orange)',
    tagBg: 'var(--c-orange-soft)',
    desc: '管理登录密码、SSH 密钥和访问控制。',
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
