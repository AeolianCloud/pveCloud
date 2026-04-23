<template>
  <section class="panel">
    <p class="tag">INSTANCES</p>
    <h2>实例管理</h2>
    <p v-if="loading">正在加载实例...</p>
    <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="instances.length === 0">当前没有实例。</p>
    <ul v-else class="list">
      <li v-for="instance in instances" :key="instance.id">
        {{ instance.instance_no }} / {{ instance.status }} / order {{ instance.order_id }} / user {{ instance.user_id }}
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { listInstances, type Instance } from '../api/instance'

const instances = ref<Instance[]>([])
const loading = ref(true)
const errorMessage = ref('')

onMounted(async () => {
  try {
    instances.value = await listInstances()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '实例加载失败'
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
