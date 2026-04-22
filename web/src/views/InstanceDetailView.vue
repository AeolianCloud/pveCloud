<template>
  <section class="panel">
    <p class="tag">实例详情</p>
    <h2>实例 #{{ route.params.id }}</h2>
    <div class="actions">
      <button type="button">开机</button>
      <button type="button">关机</button>
      <button type="button">重启</button>
      <button type="button">重装</button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()
</script>

<style scoped>
.panel {
  padding: 28px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
}

.tag {
  margin: 0 0 8px;
  color: #9b5d32;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

button {
  border: 0;
  border-radius: 16px;
  padding: 12px 16px;
  background: #1d2a33;
  color: white;
}
</style>
