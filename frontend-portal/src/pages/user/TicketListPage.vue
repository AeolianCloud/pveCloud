<template>
  <section class="panel grid">
    <h2>工单列表</h2>
    <form style="display: flex; gap: 8px;" @submit.prevent="createTicket">
      <BaseInput v-model="title" placeholder="工单标题" :error="errors.title" />
      <BaseInput v-model="content" placeholder="问题描述" :error="errors.content" />
      <BaseButton type="submit">提交工单</BaseButton>
    </form>

    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">ID</th>
          <th align="left">标题</th>
          <th align="left">状态</th>
          <th align="left">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in tickets" :key="item.id">
          <td>{{ item.id }}</td>
          <td>{{ item.title }}</td>
          <td>{{ item.status }}</td>
          <td><RouterLink :to="`/console/tickets/${item.id}`">查看详情</RouterLink></td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../../api/http';
import BaseInput from '../../components/ui/BaseInput.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import { required, runValidators } from '../../utils/validators';

const title = ref('');
const content = ref('');
const errors = ref({ title: '', content: '' });
const tickets = ref<any[]>([]);

async function load(): Promise<void> {
  const res = await http.get('/user/tickets');
  tickets.value = res.data.data ?? [];
}

async function createTicket(): Promise<void> {
  errors.value.title = runValidators(title.value, [required('标题')]);
  errors.value.content = runValidators(content.value, [required('内容')]);
  if (errors.value.title || errors.value.content) {
    return;
  }

  await http.post('/user/tickets', {
    title: title.value,
    content: content.value,
    priority: 'medium',
  });
  title.value = '';
  content.value = '';
  await load();
}

onMounted(load);
</script>
