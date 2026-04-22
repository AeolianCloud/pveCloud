<template>
  <div class="shell">
    <aside class="sidebar">
      <p class="eyebrow">PVE CLOUD</p>
      <h1>管理后台</h1>
      <nav>
        <RouterLink to="/">仪表盘</RouterLink>
        <RouterLink to="/users">用户</RouterLink>
        <RouterLink to="/products">产品</RouterLink>
        <RouterLink to="/orders">订单</RouterLink>
        <RouterLink to="/instances">实例</RouterLink>
        <RouterLink to="/tasks">任务</RouterLink>
      </nav>
    </aside>

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
    linear-gradient(180deg, rgba(222, 235, 225, 0.72), rgba(240, 238, 232, 0.72)),
    linear-gradient(120deg, #102021, #22383c 55%, #4b6655);
  color: #f5f1e8;
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
  display: grid;
  grid-template-columns: 260px 1fr;
}

.sidebar {
  padding: 28px 22px;
  background: rgba(10, 23, 25, 0.72);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
}

.eyebrow {
  margin: 0 0 8px;
  font-size: 12px;
  letter-spacing: 0.28em;
  color: #bbd3ae;
}

h1 {
  margin: 0 0 24px;
  font-size: 32px;
}

nav {
  display: grid;
  gap: 10px;
}

nav a {
  padding: 12px 14px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.06);
}

nav a.router-link-active {
  background: #d7e7c7;
  color: #132224;
  font-weight: 700;
}

.content {
  padding: 28px;
}

@media (max-width: 900px) {
  .shell {
    grid-template-columns: 1fr;
  }
}
</style>
