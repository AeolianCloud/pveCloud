<script setup lang="ts">
import Toolbar from 'primevue/toolbar'
import type { Component } from 'vue'

defineProps<{
  eyebrow: string
  title: string
  icon: Component
}>()
</script>

<template>
  <Toolbar class="admin-page-header">
    <template #start>
      <div class="admin-page-heading">
        <span class="admin-page-heading-icon">
          <component :is="icon" :size="20" aria-hidden="true" />
        </span>
        <div>
          <span>{{ eyebrow }}</span>
          <h1>{{ title }}</h1>
        </div>
      </div>
    </template>
    <template #end>
      <div class="admin-page-actions">
        <slot />
      </div>
    </template>
  </Toolbar>
</template>

<style scoped>
.admin-page-header {
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-soft);
}

.admin-page-heading {
  display: flex;
  align-items: center;
  gap: 12px;
}

.admin-page-heading-icon {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: var(--primary);
  background: var(--primary-soft);
}

.admin-page-heading span {
  color: var(--muted);
  font-size: 13px;
  font-weight: 800;
}

.admin-page-heading h1 {
  margin: 3px 0 0;
  color: var(--text);
  font-size: 20px;
  line-height: 1.2;
}

.admin-page-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 960px) {
  .admin-page-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .admin-page-actions {
    justify-content: flex-start;
  }
}
</style>
