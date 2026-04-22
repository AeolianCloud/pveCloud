<template>
  <div class="shell">
    <header class="masthead">
      <div>
        <p class="eyebrow">PVE CLOUD</p>
        <h1>用户入口</h1>
      </div>
      <nav class="nav">
        <RouterLink to="/login">登录</RouterLink>
        <RouterLink to="/register">注册</RouterLink>
        <RouterLink to="/products">产品</RouterLink>
        <RouterLink to="/orders">订单</RouterLink>
        <RouterLink to="/instances">实例</RouterLink>
        <RouterLink to="/notices">通知</RouterLink>
      </nav>
    </header>

    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<style scoped>
:global(body) {
  margin: 0;
  font-family: "Segoe UI", "PingFang SC", sans-serif;
  background:
    radial-gradient(circle at top left, rgba(255, 186, 73, 0.22), transparent 28%),
    linear-gradient(160deg, #f6f1e8 0%, #f3f7fb 46%, #d7e4ec 100%);
  color: #1d2a33;
}

:global(*) {
  box-sizing: border-box;
}

:global(a) {
  color: inherit;
  text-decoration: none;
}

.shell {
  min-height: 100vh;
  padding: 32px 20px 48px;
}

.masthead {
  max-width: 1160px;
  margin: 0 auto 24px;
  padding: 28px 32px;
  border: 1px solid rgba(29, 42, 51, 0.08);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.76);
  backdrop-filter: blur(16px);
  display: grid;
  gap: 16px;
}

.eyebrow {
  margin: 0 0 6px;
  font-size: 12px;
  letter-spacing: 0.28em;
  color: #81643f;
}

h1 {
  margin: 0;
  font-size: clamp(28px, 4vw, 44px);
}

.nav {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.nav a {
  padding: 10px 14px;
  border-radius: 999px;
  background: #e5edf2;
  font-size: 14px;
}

.nav a.router-link-active {
  background: #1d2a33;
  color: #fff8ef;
}

.content {
  max-width: 1160px;
  margin: 0 auto;
}
</style>
