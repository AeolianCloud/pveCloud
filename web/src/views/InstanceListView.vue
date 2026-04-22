<template>
  <section class="grid">
    <article class="card">
      <p class="meta">运行中</p>
      <h2>实例 I202604220001</h2>
      <p>华东 1 / 4C8G / 到期时间 2026-05-22</p>
      <RouterLink to="/instances/1">查看详情</RouterLink>
    </article>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<style scoped>
.grid {
  display: grid;
  gap: 18px;
}

.card {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
}

.meta {
  margin: 0 0 8px;
  color: #2d8659;
}
</style>
