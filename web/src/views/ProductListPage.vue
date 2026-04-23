<template>
  <section class="stack">
    <header class="header">
      <p class="tag">PRODUCTS</p>
      <h2>真实商品列表</h2>
    </header>

    <p v-if="loading" class="state">正在加载商品...</p>
    <p v-else-if="errorMessage" class="state error">{{ errorMessage }}</p>
    <p v-else-if="products.length === 0" class="state">当前没有可售商品。</p>

    <section v-else class="grid">
      <article v-for="product in products" :key="product.product.id" class="card">
        <p class="meta">{{ product.product.product_type }} / {{ product.product.status }}</p>
        <h2>{{ product.product.product_name }}</h2>
        <p>{{ summarizeSKUs(product.skus) }}</p>
        <RouterLink :to="`/products/${product.product.id}`">查看详情</RouterLink>
      </article>
    </section>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { listProducts, type SKU, type SaleableProduct } from '../api/catalog'

const products = ref<SaleableProduct[]>([])
const loading = ref(true)
const errorMessage = ref('')

function summarizeSKUs(skus: SKU[]) {
  if (skus.length === 0) {
    return '暂无 SKU'
  }

  return skus
    .map((sku) => `${sku.sku_name}: ${sku.cpu_cores}C / ${Math.floor(sku.memory_mb / 1024)}G / ${sku.disk_gb}GB`)
    .join(' | ')
}

onMounted(async () => {
  try {
    products.value = await listProducts()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '商品加载失败'
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

.header h2 {
  margin: 0;
}

.tag {
  margin: 0 0 8px;
  color: #9b5d32;
  font-size: 13px;
  letter-spacing: 0.18em;
}

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
