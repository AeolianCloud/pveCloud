<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listNotices, markNoticeRead, type Notice } from '../api/notice'

const notices = ref<Notice[]>([])
const loading = ref(true)
const error = ref('')

async function handleMarkRead(id: number) {
  try {
    await markNoticeRead(id)
    const n = notices.value.find(n => n.id === id)
    if (n) n.is_read = true
  } catch {
    // ignore
  }
}

onMounted(async () => {
  try {
    notices.value = await listNotices()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载通知失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="panel">
    <p class="tag">NOTICES</p>
    <h2>通知中心</h2>

    <p v-if="loading">加载中...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-else-if="notices.length === 0">暂无通知</p>

    <ul v-else class="notice-list">
      <li v-for="n in notices" :key="n.id" :class="{ unread: !n.is_read }">
        <div class="notice-header">
          <strong>{{ n.title }}</strong>
          <span class="time">{{ n.created_at }}</span>
        </div>
        <p>{{ n.body }}</p>
        <button v-if="!n.is_read" @click="handleMarkRead(n.id)">标记已读</button>
      </li>
    </ul>
  </section>
</template>

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

.notice-list {
  list-style: none;
  padding: 0;
}

.notice-list li {
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.notice-list li.unread {
  background: rgba(155, 93, 50, 0.06);
  padding-left: 8px;
  border-radius: 6px;
}

.notice-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.time {
  color: #999;
  font-size: 0.85em;
}

button {
  margin-top: 6px;
  padding: 4px 12px;
  border: 1px solid #9b5d32;
  border-radius: 6px;
  background: transparent;
  color: #9b5d32;
  cursor: pointer;
}

.error { color: #c00; }
</style>
