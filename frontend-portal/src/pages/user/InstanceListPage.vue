<template>
  <section class="panel">
    <h2>实例列表</h2>
    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">实例</th>
          <th align="left">状态</th>
          <th align="left">IP</th>
          <th align="left">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="item.id">
          <td><RouterLink :to="`/console/instances/${item.id}`">{{ item.name }}</RouterLink></td>
          <td><span class="panel" style="padding: 2px 8px;">{{ item.status }}</span></td>
          <td>{{ item.ip }}</td>
          <td style="display: flex; gap: 8px;">
            <button class="btn" @click="operate(item.id, 'start')">开机</button>
            <button class="btn" @click="operate(item.id, 'stop')">关机</button>
            <button class="btn" @click="operate(item.id, 'reboot')">重启</button>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../../api/http';

const items = ref<any[]>([]);

async function load(): Promise<void> {
  const res = await http.get('/user/instances');
  items.value = res.data.data ?? [];
}

async function operate(id: number, action: 'start' | 'stop' | 'reboot'): Promise<void> {
  await http.post(`/user/instances/${id}/${action}`);
  await load();
}

onMounted(load);
</script>
