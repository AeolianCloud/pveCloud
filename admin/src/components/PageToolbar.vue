<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  eyebrow?: string
}>()
</script>

<template>
  <div class="page-toolbar">
    <div class="page-toolbar__content">
      <span v-if="eyebrow" class="page-toolbar__eyebrow">{{ eyebrow }}</span>
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-toolbar__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.page-toolbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
}

.page-toolbar__content {
  min-width: 0;
  display: grid;
  gap: 8px;
}

.page-toolbar__eyebrow {
  width: fit-content;
  padding: 6px 12px;
  border-radius: 999px;
  color: var(--pc-accent-strong);
  background: rgba(14, 165, 233, 0.12);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.page-toolbar__content h1 {
  margin: 0;
  color: var(--pc-title-text);
  font-family: var(--pc-display-font);
  font-size: clamp(28px, 3vw, 38px);
  line-height: 1;
}

.page-toolbar__content p {
  max-width: 720px;
  margin: 0;
  color: var(--pc-muted-text);
  line-height: 1.65;
}

.page-toolbar__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

@media (max-width: 767px) {
  .page-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .page-toolbar__actions {
    width: 100%;
  }
}
</style>
