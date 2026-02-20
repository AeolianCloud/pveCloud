<template>
  <section class="panel" style="display: grid; gap: 12px;">
    <h2>工单管理</h2>
    <form style="display: flex; gap: 8px;" @submit.prevent="search">
      <select class="input" v-model="status" style="max-width: 220px;">
        <option value="">全部状态</option>
        <option value="open">open</option>
        <option value="processing">processing</option>
        <option value="closed">closed</option>
      </select>
      <button class="btn" type="submit">筛选</button>
    </form>

    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">ID</th>
          <th align="left">用户</th>
          <th align="left">标题</th>
          <th align="left">优先级</th>
          <th align="left">状态</th>
          <th align="left">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="t in tickets" :key="t.id">
          <td>{{ t.id }}</td>
          <td>{{ t.user_email }}</td>
          <td>{{ t.title }}</td>
          <td>{{ t.priority }}</td>
          <td>{{ t.status }}</td>
          <td style="display: flex; gap: 8px;">
            <button class="btn" @click="reply(t.id)">回复</button>
            <button class="btn" @click="close(t.id)">关闭</button>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../api/http';
import type { AdminTicketItem, ApiResponse } from '../types/api';

const status = ref('');
const tickets = ref<AdminTicketItem[]>([]);

async function load(): Promise<void> {
  const res = await http.get<ApiResponse<AdminTicketItem[]>>('/admin/tickets', { params: { status: status.value } });
  tickets.value = res.data.data ?? [];
}

async function search(): Promise<void> {
  await load();
}

async function reply(id: number): Promise<void> {
  const content = window.prompt('输入回复内容');
  if (!content) return;
  await http.post(`/admin/tickets/${id}/replies`, { content });
  await load();
}

async function close(id: number): Promise<void> {
  await http.post(`/admin/tickets/${id}/close`);
  await load();
}

onMounted(load);
</script>
