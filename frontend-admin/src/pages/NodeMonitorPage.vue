<template>
  <section class="panel" style="display: grid; gap: 12px;">
    <h2>节点监控</h2>
    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">节点</th>
          <th align="left">CPU</th>
          <th align="left">内存</th>
          <th align="left">磁盘</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="n in nodes" :key="n.node">
          <td>{{ n.node }}</td>
          <td>{{ Math.round((n.cpu_usage ?? 0) * 100) }}%</td>
          <td>{{ Math.round((n.memory_usage ?? 0) * 100) }}%</td>
          <td>{{ Math.round((n.disk_usage ?? 0) * 100) }}%</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../api/http';

const nodes = ref<any[]>([]);

onMounted(async () => {
  const res = await http.get('/admin/nodes');
  nodes.value = res.data.data ?? [];
});
</script>
