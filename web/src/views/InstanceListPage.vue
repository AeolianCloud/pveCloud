<template>
  <section class="stack">
    <header>
      <p class="tag">INSTANCES</p>
      <h2>我的实例</h2>
    </header>

    <p v-if="loading" class="state">正在加载实例...</p>
    <p v-else-if="errorMessage" class="state error">{{ errorMessage }}</p>
    <p v-else-if="instances.length === 0" class="state">当前没有实例。</p>

    <section v-else class="grid">
      <article v-for="instance in instances" :key="instance.id" class="card">
        <p class="meta">{{ instance.status }}</p>
        <h2>{{ instance.instance_no }}</h2>
        <p>Order {{ instance.order_id }} / Node {{ instance.node_id }}</p>
        <RouterLink :to="`/instances/${instance.id}`">查看详情</RouterLink>
      </article>
    </section>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

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
.stack {
  display: grid;
  gap: 18px;
}

.tag {
  margin: 0 0 8px;
  color: #9b5d32;
  font-size: 13px;
  letter-spacing: 0.18em;
}

.grid {
  display: grid;
  gap: 18px;
}

.card {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
}

.meta {
  margin: 0 0 8px;
  color: #2d8659;
}

.state {
  margin: 0;
  padding: 18px 20px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.72);
}

.error {
  color: #b42318;
}
</style>
