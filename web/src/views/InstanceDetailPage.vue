<template>
  <section class="panel">
    <p class="tag">INSTANCE DETAIL</p>
    <p v-if="loading">正在加载实例详情...</p>
    <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>
    <template v-else-if="instance">
      <h2>{{ instance.instance_no }}</h2>
      <ul class="details">
        <li>Status: {{ instance.status }}</li>
        <li>Instance Ref: {{ instance.instance_ref }}</li>
        <li>Order ID: {{ instance.order_id }}</li>
        <li>Node ID: {{ instance.node_id }}</li>
        <li>Created At: {{ instance.created_at }}</li>
      </ul>
      <p class="hint">实例操作接口还不在本次真实对接范围内，这里只展示真实实例事实。</p>
    </template>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { getInstance, type Instance } from '../api/instance'

const route = useRoute()
const instance = ref<Instance | null>(null)
const loading = ref(true)
const errorMessage = ref('')

onMounted(async () => {
  try {
    instance.value = await getInstance(String(route.params.id))
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '实例详情加载失败'
  } finally {
    loading.value = false
  }
})
</script>

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

.details {
  display: grid;
  gap: 10px;
  padding-left: 20px;
}

.hint {
  color: #6b7280;
}

.error {
  color: #b42318;
}
</style>
