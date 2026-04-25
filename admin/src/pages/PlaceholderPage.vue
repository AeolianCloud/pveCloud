<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() => String(route.meta.title || '页面'))
</script>

<template>
  <section class="state-panel">
    <h2>{{ title }}</h2>
    <p>基础布局已就绪。</p>
  </section>
</template>
