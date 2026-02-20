<template>
  <section class="panel" style="display: grid; gap: 12px;">
    <h2>用户管理</h2>
    <form style="display: flex; gap: 8px;" @submit.prevent="search">
      <BaseInput v-model="keyword" placeholder="按邮箱搜索" />
      <BaseButton type="submit">搜索</BaseButton>
    </form>

    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">ID</th>
          <th align="left">邮箱</th>
          <th align="left">角色</th>
          <th align="left">状态</th>
          <th align="left">余额</th>
          <th align="left">实例数</th>
          <th align="left">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in users" :key="u.id">
          <td>{{ u.id }}</td>
          <td>{{ u.email }}</td>
          <td>{{ u.role }}</td>
          <td>{{ u.status }}</td>
          <td>{{ u.balance }}</td>
          <td>{{ u.instance_count }}</td>
          <td style="display: flex; gap: 8px;">
            <BaseButton @click="toggle(u.id)">禁用/启用</BaseButton>
            <BaseButton @click="forceLogout(u.id)">强制下线</BaseButton>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../api/http';
import BaseInput from '../components/ui/BaseInput.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import type { AdminUserItem, ApiResponse } from '../types/api';

const keyword = ref('');
const users = ref<AdminUserItem[]>([]);

async function load(): Promise<void> {
  const res = await http.get<ApiResponse<AdminUserItem[]>>('/admin/users', { params: { keyword: keyword.value } });
  users.value = res.data.data ?? [];
}

async function search(): Promise<void> {
  await load();
}

async function toggle(id: number): Promise<void> {
  await http.post(`/admin/users/${id}/toggle-status`);
  await load();
}

async function forceLogout(id: number): Promise<void> {
  await http.post(`/admin/users/${id}/force-logout`);
}

onMounted(load);
</script>
