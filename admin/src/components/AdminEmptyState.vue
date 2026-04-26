<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  text: string
  icon?: Component
}>()
</script>

<template>
  <Message class="admin-empty-state" severity="secondary" :closable="false">
    <span class="admin-empty-state-content">
      <component :is="icon" v-if="icon" :size="18" aria-hidden="true" />
      {{ text }}
    </span>
  </Message>
</template>

<style scoped>
.admin-empty-state {
  min-height: 220px;
  display: grid;
  place-items: center;
}

.admin-empty-state :deep(.p-message-content) {
  justify-content: center;
}

.admin-empty-state-content {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--muted);
  font-weight: 800;
}
</style>
