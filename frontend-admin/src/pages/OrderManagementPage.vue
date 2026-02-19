<template>
  <section class="panel" style="display: grid; gap: 12px;">
    <h2>订单管理</h2>
    <form style="display: flex; gap: 8px;" @submit.prevent="search">
      <select class="input" v-model="status" style="max-width: 220px;">
        <option value="">全部状态</option>
        <option value="pending">pending</option>
        <option value="active">active</option>
        <option value="failed">failed</option>
      </select>
      <button class="btn" type="submit">筛选</button>
    </form>

    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">订单ID</th>
          <th align="left">用户ID</th>
          <th align="left">金额</th>
          <th align="left">状态</th>
          <th align="left">详情</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in orders" :key="o.id">
          <td>{{ o.id }}</td>
          <td>{{ o.user_id }}</td>
          <td>{{ o.amount }}</td>
          <td>{{ o.status }}</td>
          <td><pre style="margin: 0; white-space: pre-wrap;">{{ o.config_snapshot }}</pre></td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../api/http';

const status = ref('');
const orders = ref<any[]>([]);

async function load(): Promise<void> {
  const res = await http.get('/admin/orders', { params: { status: status.value } });
  orders.value = res.data.data ?? [];
}

async function search(): Promise<void> {
  await load();
}

onMounted(load);
</script>
