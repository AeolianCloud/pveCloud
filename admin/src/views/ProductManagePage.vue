<template>
  <section class="panel">
    <p class="tag">PRODUCTS</p>
    <h2>商品与 SKU</h2>
    <p v-if="loading">正在加载商品...</p>
    <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="products.length === 0">当前没有商品。</p>
    <ul v-else class="list">
      <li v-for="product in products" :key="product.product.id">
        <strong>{{ product.product.product_name }}</strong>
        <span>{{ product.product.product_type }} / {{ product.product.status }}</span>
        <span>{{ summarize(product.skus) }}</span>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { listProducts, type SKU, type SaleableProduct } from '../api/catalog'

const products = ref<SaleableProduct[]>([])
const loading = ref(true)
const errorMessage = ref('')

function summarize(skus: SKU[]) {
  if (skus.length === 0) {
    return 'No SKU'
  }

  return skus.map((sku) => sku.sku_name).join(' / ')
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
.panel {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

.list {
  display: grid;
  gap: 14px;
  padding-left: 20px;
}

.list li {
  display: grid;
  gap: 4px;
}

.error {
  color: #b42318;
}
</style>
