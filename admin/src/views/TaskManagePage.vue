<template>
  <section class="panel">
    <p class="tag">TASKS</p>
    <h2>异步任务中心</h2>
    <p v-if="loading">正在加载任务...</p>
    <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="tasks.length === 0">当前没有任务。</p>
    <ul v-else class="list">
      <li v-for="task in tasks" :key="task.id">
        {{ task.task_no }} / {{ task.task_type }} / {{ task.status }} / {{ task.retry_count }} of {{ task.max_retry_count }}
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { listTasks, type Task } from '../api/task'

const tasks = ref<Task[]>([])
const loading = ref(true)
const errorMessage = ref('')

onMounted(async () => {
  try {
    tasks.value = await listTasks()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '任务加载失败'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.panel {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

.list {
  display: grid;
  gap: 10px;
  padding-left: 20px;
}

.error {
  color: #b42318;
}
</style>
