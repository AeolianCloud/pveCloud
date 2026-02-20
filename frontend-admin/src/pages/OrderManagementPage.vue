<template>
  <section class="panel" style="display: grid; gap: 12px;">
    <h2>订单管理</h2>
    <form style="display: flex; gap: 8px; flex-wrap: wrap;" @submit.prevent="search">
      <select class="input" v-model="status" style="max-width: 220px;">
        <option value="">全部状态</option>
        <option value="pending">pending</option>
        <option value="active">active</option>
        <option value="failed">failed</option>
      </select>
      <input class="input" v-model="userId" placeholder="用户ID" style="max-width: 220px;" />
      <input class="input" v-model="dateStart" type="date" style="max-width: 220px;" />
      <input class="input" v-model="dateEnd" type="date" style="max-width: 220px;" />
      <button class="btn" type="submit">筛选</button>
    </form>

    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">订单ID</th>
          <th align="left">用户ID</th>
          <th align="left">金额</th>
          <th align="left">周期</th>
          <th align="left">状态</th>
          <th align="left">创建时间</th>
          <th align="left">详情</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in orders" :key="o.id">
          <td>{{ o.id }}</td>
          <td>{{ o.user_id }}</td>
          <td>{{ o.amount }}</td>
          <td>{{ o.billing_cycle }}</td>
          <td>{{ o.status }}</td>
          <td>{{ o.created_at }}</td>
          <td><pre style="margin: 0; white-space: pre-wrap;">{{ o.config_snapshot }}</pre></td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../api/http';
import type { AdminOrderItem, ApiResponse } from '../types/api';

const status = ref('');
const userId = ref('');
const dateStart = ref('');
const dateEnd = ref('');
const orders = ref<AdminOrderItem[]>([]);

async function load(): Promise<void> {
  const params: Record<string, string> = {};
  if (status.value) {
    params.status = status.value;
  }
  if (userId.value.trim()) {
    params.user_id = userId.value.trim();
  }
  if (dateStart.value || dateEnd.value) {
    params.date_range = `${dateStart.value},${dateEnd.value}`;
  }
  const res = await http.get<ApiResponse<AdminOrderItem[]>>('/admin/orders', { params });
  orders.value = res.data.data ?? [];
}

async function search(): Promise<void> {
  await load();
}

onMounted(load);
</script>
