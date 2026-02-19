<template>
  <section class="panel" style="display: grid; gap: 16px;">
    <h2>仪表盘</h2>
    <div style="display: grid; gap: 12px; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));">
      <article class="panel">
        <p>总用户数</p>
        <strong>{{ metrics.total_users }}</strong>
      </article>
      <article class="panel">
        <p>运行中实例</p>
        <strong>{{ metrics.active_instances }}</strong>
      </article>
      <article class="panel">
        <p>今日收入</p>
        <strong>¥{{ metrics.today_revenue }}</strong>
      </article>
      <article class="panel">
        <p>待处理工单</p>
        <strong>{{ metrics.pending_tickets }}</strong>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../api/http';

const metrics = ref({
  total_users: 0,
  active_instances: 0,
  today_revenue: 0,
  pending_tickets: 0,
});

onMounted(async () => {
  const res = await http.get('/admin/dashboard');
  metrics.value = res.data.data;
});
</script>
