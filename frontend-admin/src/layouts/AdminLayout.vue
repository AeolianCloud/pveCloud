<template>
  <div style="display: grid; grid-template-columns: 220px 1fr; min-height: 100vh;">
    <aside class="panel" style="border-radius: 0;">
      <h3>运营后台</h3>
      <nav style="display: grid; gap: 10px;">
        <RouterLink to="/dashboard">仪表盘</RouterLink>
        <RouterLink to="/users">用户管理</RouterLink>
        <RouterLink to="/products">商品管理</RouterLink>
        <RouterLink to="/orders">订单管理</RouterLink>
        <RouterLink to="/tickets">工单管理</RouterLink>
        <RouterLink to="/nodes">节点监控</RouterLink>
      </nav>
    </aside>

    <main style="padding: 16px;">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
// 后台统一布局，保持管理路径可预测。
</script>
