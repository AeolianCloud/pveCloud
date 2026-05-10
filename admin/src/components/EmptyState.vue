<script setup lang="ts">
import { NEmpty } from 'naive-ui'

defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <NEmpty>
    <template #default>
      <div class="empty-state">
        <strong>{{ title }}</strong>
        <p v-if="description">{{ description }}</p>
      </div>
    </template>
  </NEmpty>
</template>

<style scoped>
.empty-state {
  display: grid;
  gap: 8px;
  text-align: center;
}

.empty-state strong {
  font-size: 16px;
}

.empty-state p {
  max-width: 360px;
  margin: 0 auto;
  color: rgba(15, 23, 42, 0.55);
  line-height: 1.6;
}
</style>
