<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useWebAuthStore } from '../../store/modules/auth'

const authStore = useWebAuthStore()
const { displayName } = storeToRefs(authStore)

const sections = [
  {
    title: '账号资料',
    tag: '已开放',
    tone: 'success',
    index: '01',
    desc: '维护邮箱、显示名称和登录密码。',
    action: '管理资料',
    to: '/user/profile',
  },
  {
    title: '实名认证',
    tag: '购买必需',
    tone: 'primary',
    index: '02',
    desc: '购买机器前需要完成个人实名审核。',
    action: '去实名',
    to: '/user/real-name',
  },
  {
    title: '购买与实例',
    tag: '待开放',
    tone: 'warning',
    index: '03',
    desc: '后续阶段开放购买、支付和实例相关能力。',
    action: '查看计划',
    to: '',
  },
  {
    title: '订单与账单',
    tag: '待开放',
    tone: 'muted',
    index: '04',
    desc: '订单、支付、账单和发票当前不读取真实业务数据。',
    action: '暂未开放',
    to: '',
  },
  {
    title: '安全设置',
    tag: '待开放',
    tone: 'muted',
    index: '05',
    desc: '当前仅在账号资料页开放登录密码修改。',
    action: '稍后开放',
    to: '',
  },
]
</script>

<template>
  <section class="console-page content-page">
    <div class="console-shell">
      <div class="console-hero">
        <div class="hero-copy">
          <p class="eyebrow">USER CONSOLE</p>
          <h1>{{ displayName }}，管理你的云资源入口</h1>
          <p>账号体系、实名认证和产品购买准备已接入。购买机器前请先完成个人实名认证。</p>
          <div class="hero-actions">
            <RouterLink class="btn btn-primary" to="/user/real-name">完成实名认证</RouterLink>
            <RouterLink class="btn btn-outline" to="/pricing">查看服务器价格</RouterLink>
          </div>
        </div>

        <div class="hero-panel">
          <span class="panel-label">购买门禁</span>
          <strong>实名通过后才可购买机器</strong>
          <p>未实名、审核中或被拒绝都会被服务端拦截。</p>
          <div class="panel-steps">
            <span>提交资料</span>
            <i></i>
            <span>后台审核</span>
            <i></i>
            <span>开放购买</span>
          </div>
        </div>
      </div>

      <div class="summary-row">
        <div class="summary-card">
          <span>账号状态</span>
          <strong>已登录</strong>
        </div>
        <div class="summary-card summary-card--hot">
          <span>实名要求</span>
          <strong>购买必需</strong>
        </div>
        <div class="summary-card">
          <span>当前阶段</span>
          <strong>账号自助</strong>
        </div>
      </div>

      <div class="section-title">
        <div>
          <p class="eyebrow">QUICK ACCESS</p>
          <h2>常用入口</h2>
        </div>
        <span>功能会按阶段逐步开放</span>
      </div>

      <div class="user-grid">
        <RouterLink v-for="s in sections" :key="s.title" class="user-card" :data-tone="s.tone" :to="s.to || '/user'">
          <div class="card-top">
            <span class="card-index">{{ s.index }}</span>
            <span class="tag">{{ s.tag }}</span>
          </div>
          <div>
            <h3>{{ s.title }}</h3>
            <p>{{ s.desc }}</p>
          </div>
          <span class="card-action">{{ s.action }}</span>
        </RouterLink>
      </div>
    </div>
  </section>
</template>

<style scoped>
.console-page {
  position: relative;
  min-height: calc(100vh - 96px);
  overflow: hidden;
}

.console-page::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(circle at 12% 8%, rgba(59, 130, 246, 0.22), transparent 30%),
    radial-gradient(circle at 88% 12%, rgba(139, 92, 246, 0.18), transparent 26%),
    linear-gradient(180deg, transparent, rgba(255, 255, 255, 0.02));
}

.console-shell {
  position: relative;
  display: grid;
  gap: clamp(22px, 3vw, 34px);
  width: min(1180px, calc(100% - 40px));
  margin: 0 auto;
  padding: clamp(26px, 5vw, 58px) 0 72px;
}

.console-hero {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(310px, 0.75fr);
  gap: 24px;
  align-items: stretch;
  padding: clamp(28px, 4vw, 46px);
  border: 1px solid var(--c-border);
  border-radius: 34px;
  background:
    linear-gradient(135deg, rgba(59, 130, 246, 0.16), rgba(139, 92, 246, 0.08) 42%, rgba(16, 185, 129, 0.07)),
    rgba(19, 21, 31, 0.72);
  box-shadow: 0 28px 90px rgba(0, 0, 0, 0.32);
  backdrop-filter: blur(22px);
}

[data-theme='light'] .console-hero {
  background:
    linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(139, 92, 246, 0.08) 42%, rgba(16, 185, 129, 0.08)),
    rgba(255, 255, 255, 0.82);
  box-shadow: 0 28px 70px rgba(15, 23, 42, 0.1);
}

.hero-copy {
  display: grid;
  align-content: center;
  gap: 18px;
  max-width: 720px;
}

.hero-copy h1 {
  max-width: 680px;
  font-size: clamp(2.35rem, 5vw, 4.6rem);
  line-height: 0.98;
  letter-spacing: -0.065em;
}

.hero-copy p {
  max-width: 620px;
  color: var(--c-text-2);
  font-size: 1.05rem;
  line-height: 1.8;
}

.eyebrow {
  width: fit-content;
  margin: 0;
  color: var(--c-primary);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.16em;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 4px;
}

.hero-panel {
  position: relative;
  display: grid;
  align-content: space-between;
  gap: 22px;
  min-height: 280px;
  padding: 28px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 26px;
  background:
    radial-gradient(circle at 100% 0%, rgba(16, 185, 129, 0.28), transparent 38%),
    rgba(0, 0, 0, 0.22);
}

[data-theme='light'] .hero-panel {
  border-color: rgba(15, 23, 42, 0.08);
  background:
    radial-gradient(circle at 100% 0%, rgba(16, 185, 129, 0.18), transparent 40%),
    rgba(255, 255, 255, 0.66);
}

.hero-panel::after {
  content: '';
  position: absolute;
  right: -42px;
  bottom: -42px;
  width: 150px;
  height: 150px;
  border-radius: 42px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  transform: rotate(18deg);
}

.panel-label {
  width: fit-content;
  padding: 7px 12px;
  border-radius: 999px;
  color: var(--c-success);
  background: var(--c-success-soft);
  font-size: 0.78rem;
  font-weight: 800;
}

.hero-panel strong {
  max-width: 280px;
  font-size: 1.65rem;
  line-height: 1.18;
  letter-spacing: -0.04em;
}

.hero-panel p {
  max-width: 290px;
  color: var(--c-text-2);
  line-height: 1.7;
}

.panel-steps {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 9px;
  color: var(--c-text-2);
  font-size: 0.82rem;
  font-weight: 700;
}

.panel-steps i {
  width: 22px;
  height: 1px;
  background: var(--c-border);
}

.summary-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.summary-card {
  padding: 20px 22px;
  border: 1px solid var(--c-border);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.04);
}

[data-theme='light'] .summary-card {
  background: rgba(255, 255, 255, 0.78);
}

.summary-card span {
  display: block;
  margin-bottom: 8px;
  color: var(--c-text-3);
  font-size: 0.82rem;
  font-weight: 700;
}

.summary-card strong {
  font-size: 1.18rem;
}

.summary-card--hot {
  border-color: rgba(59, 130, 246, 0.35);
  background: var(--c-primary-soft);
}

.section-title {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 18px;
  margin-top: 8px;
}

.section-title h2 {
  margin-top: 8px;
  font-size: clamp(1.65rem, 3vw, 2.35rem);
  letter-spacing: -0.05em;
}

.section-title > span {
  color: var(--c-text-3);
  font-size: 0.95rem;
}

.user-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 16px;
}

.user-card {
  position: relative;
  display: grid;
  grid-column: span 2;
  align-content: space-between;
  gap: 24px;
  min-height: 230px;
  padding: 24px;
  overflow: hidden;
  border: 1px solid var(--c-border);
  border-radius: 26px;
  background:
    radial-gradient(circle at 100% 0%, rgba(255, 255, 255, 0.08), transparent 42%),
    var(--c-card);
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.18);
  transition: transform 220ms ease, border-color 220ms ease, box-shadow 220ms ease;
}

.user-card:nth-child(1),
.user-card:nth-child(2) {
  grid-column: span 3;
}

.user-card::after {
  content: '';
  position: absolute;
  right: -44px;
  top: -44px;
  width: 132px;
  height: 132px;
  border-radius: 999px;
  background: var(--card-glow, var(--c-primary-soft));
  opacity: 0.7;
  transition: transform 220ms ease, opacity 220ms ease;
}

.user-card[data-tone='primary'] { --card-glow: rgba(59, 130, 246, 0.22); }
.user-card[data-tone='success'] { --card-glow: rgba(16, 185, 129, 0.2); }
.user-card[data-tone='warning'] { --card-glow: rgba(245, 158, 11, 0.2); }
.user-card[data-tone='muted'] { --card-glow: rgba(113, 113, 122, 0.16); }

[data-theme='light'] .user-card {
  background:
    radial-gradient(circle at 100% 0%, rgba(37, 99, 235, 0.08), transparent 44%),
    #fff;
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.08);
}

.user-card:hover {
  transform: translateY(-6px);
  border-color: rgba(59, 130, 246, 0.42);
  box-shadow: var(--shadow-lg), var(--shadow-glow);
}

.user-card:hover::after {
  transform: scale(1.18);
  opacity: 1;
}

.card-top {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card-index {
  color: var(--c-text-3);
  font-size: 0.8rem;
  font-weight: 900;
  letter-spacing: 0.12em;
}

.user-card h3 {
  position: relative;
  z-index: 1;
  margin-bottom: 10px;
  font-size: 1.45rem;
  letter-spacing: -0.03em;
}

.user-card p {
  position: relative;
  z-index: 1;
  color: var(--c-text-2);
  line-height: 1.7;
}

.tag {
  position: relative;
  z-index: 1;
  width: fit-content;
  padding: 6px 11px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 800;
  color: var(--c-text);
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid var(--c-border-light);
}

.card-action {
  position: relative;
  z-index: 1;
  width: fit-content;
  color: var(--c-primary-h);
  font-size: 0.92rem;
  font-weight: 800;
}

.card-action::after {
  content: ' /';
  color: var(--c-text-3);
}

@media (max-width: 980px) {
  .console-hero {
    grid-template-columns: 1fr;
  }

  .user-card,
  .user-card:nth-child(1),
  .user-card:nth-child(2) {
    grid-column: span 3;
  }
}

@media (max-width: 720px) {
  .console-shell {
    width: min(100% - 28px, 1180px);
    padding-top: 24px;
  }

  .console-hero {
    padding: 24px;
    border-radius: 26px;
  }

  .hero-actions,
  .section-title {
    display: grid;
  }

  .summary-row,
  .user-grid {
    grid-template-columns: 1fr;
  }

  .user-card,
  .user-card:nth-child(1),
  .user-card:nth-child(2) {
    grid-column: span 1;
    min-height: 190px;
  }
}
</style>
