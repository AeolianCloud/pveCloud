<template>
  <div style="display: grid; grid-template-columns: 220px 1fr; min-height: 100vh;">
    <aside class="panel" style="border-radius: 0;">
      <h3>用户控制台</h3>
      <nav class="grid">
        <RouterLink to="/console/instances">实例管理</RouterLink>
        <RouterLink to="/console/order-flow">下单流程</RouterLink>
        <RouterLink to="/console/wallet">钱包</RouterLink>
        <RouterLink to="/console/tickets">工单</RouterLink>
      </nav>
    </aside>
    <main style="padding: 16px;">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
// 控制台布局采用侧栏 + 主内容区，确保管理路径清晰可见。
</script>
