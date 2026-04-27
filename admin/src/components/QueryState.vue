<script setup lang="ts">
defineProps<{
  loading?: boolean
  errorMessage?: string
}>()

defineEmits<{
  retry: []
}>()
</script>

<template>
  <div v-if="loading" class="query-state">
    <el-skeleton :rows="4" animated />
  </div>
  <el-result v-else-if="errorMessage" class="query-state" icon="error" title="加载失败" :sub-title="errorMessage">
    <template #extra>
      <el-button type="primary" @click="$emit('retry')">重新加载</el-button>
    </template>
  </el-result>
  <slot v-else />
</template>

<style scoped>
.query-state {
  min-height: 240px;
  display: grid;
  align-items: center;
}
</style>
