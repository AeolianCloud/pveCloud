<template>
  <section class="panel">
    <p class="tag">产品详情</p>
    <h2>产品 #{{ route.params.id }}</h2>
    <p>这里展示当前 SKU 的规格、带宽、区域可售性和下单入口。</p>
    <ul>
      <li>CPU：4 核</li>
      <li>内存：8 GB</li>
      <li>磁盘：100 GB SSD</li>
      <li>区域：华东 1</li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()
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
</style>
