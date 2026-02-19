<template>
  <section class="grid">
    <article class="panel">
      <h2>工单详情 #{{ ticketId }}</h2>
      <ul>
        <li v-for="row in replies" :key="row.id">
          <strong>{{ row.is_admin ? '管理员' : '用户' }}:</strong> {{ row.content }}
        </li>
      </ul>
    </article>

    <article class="panel">
      <h3>追加回复</h3>
      <form style="display: flex; gap: 8px;" @submit.prevent="submitReply">
        <BaseInput v-model="content" placeholder="输入回复内容" :error="error" />
        <BaseButton type="submit">发送</BaseButton>
      </form>
    </article>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import http from '../../api/http';
import BaseInput from '../../components/ui/BaseInput.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import { required, runValidators } from '../../utils/validators';

const route = useRoute();
const ticketId = Number(route.params.id);
const content = ref('');
const error = ref('');
const replies = ref<any[]>([]);

async function load(): Promise<void> {
  const res = await http.get(`/user/tickets/${ticketId}/replies`);
  replies.value = res.data.data ?? [];
}

async function submitReply(): Promise<void> {
  error.value = runValidators(content.value, [required('回复内容')]);
  if (error.value) {
    return;
  }
  await http.post(`/user/tickets/${ticketId}/replies`, { content: content.value });
  content.value = '';
  await load();
}

onMounted(load);
</script>
