<template>
  <section class="grid">
    <article v-for="product in products" :key="product.id" class="card">
      <p class="meta">{{ product.region }}</p>
      <h2>{{ product.name }}</h2>
      <p>{{ product.spec }}</p>
      <RouterLink :to="`/products/${product.id}`">查看详情</RouterLink>
    </article>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'

const products = [
  { id: 1, region: '华东可售', name: '通用型云主机', spec: '2C4G / 50GB SSD / 峰值 20Mbps' },
  { id: 2, region: '华北可售', name: '高频型云主机', spec: '4C8G / 100GB SSD / 峰值 30Mbps' },
  { id: 3, region: '华南可售', name: '入门型云主机', spec: '1C2G / 40GB SSD / 峰值 10Mbps' },
]
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
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
  color: #9b5d32;
  font-size: 13px;
}

h2 {
  margin: 0 0 10px;
}

a {
  display: inline-block;
  margin-top: 10px;
  color: #1d2a33;
  font-weight: 700;
}
</style>
