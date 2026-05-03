<script setup lang="ts">
defineProps<{
  title: string
  description: string
  meta?: string
  metaColor?: string
}>()
</script>

<template>
  <article class="placeholder-card">
    <p v-if="meta" class="card-meta" :data-color="metaColor || 'blue'">{{ meta }}</p>
    <h3>{{ title }}</h3>
    <p>{{ description }}</p>
  </article>
</template>
