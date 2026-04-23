<template>
  <section class="stack">
    <p v-if="loading" class="state">正在加载商品详情...</p>
    <p v-else-if="errorMessage" class="state error">{{ errorMessage }}</p>

    <template v-else-if="product">
      <section class="panel">
        <p class="tag">PRODUCT DETAIL</p>
        <h2>{{ product.product.product_name }}</h2>
        <p class="summary">{{ product.product.product_type }} / {{ product.product.status }}</p>

        <ul class="sku-list">
          <li v-for="sku in product.skus" :key="sku.id">
            <strong>{{ sku.sku_name }}</strong>
            <span>{{ sku.cpu_cores }}C / {{ Math.floor(sku.memory_mb / 1024) }}G / {{ sku.disk_gb }}GB / {{ sku.bandwidth_mbps }}Mbps</span>
          </li>
        </ul>
      </section>

      <section class="panel">
        <p class="tag">CREATE ORDER</p>
        <h3>用真实接口创建订单</h3>
        <form class="form" @submit.prevent="submitOrder">
          <label>
            <span>SKU</span>
            <select v-model.number="selectedSKUID">
              <option v-for="sku in product.skus" :key="sku.id" :value="sku.id">{{ sku.sku_name }}</option>
            </select>
          </label>
          <label>
            <span>Region ID</span>
            <input v-model.number="regionID" type="number" min="1" />
          </label>
          <label>
            <span>Cycle</span>
            <select v-model="cycle">
              <option value="month">month</option>
              <option value="quarter">quarter</option>
              <option value="year">year</option>
            </select>
          </label>
          <p class="hint">后端当前没有提供 region 列表接口，这里先显式填写真实下单所需的 region id。</p>
          <p v-if="submitError" class="error-text">{{ submitError }}</p>
          <button type="submit" :disabled="submitting">{{ submitting ? '创建中...' : '立即下单' }}</button>
        </form>
      </section>
    </template>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getProduct, type SaleableProduct } from '../api/catalog'
import { createOrder } from '../api/order'

const route = useRoute()
const router = useRouter()
const product = ref<SaleableProduct | null>(null)
const loading = ref(true)
const errorMessage = ref('')
const selectedSKUID = ref(0)
const regionID = ref(1)
const cycle = ref('month')
const submitting = ref(false)
const submitError = ref('')

onMounted(async () => {
  try {
    const item = await getProduct(String(route.params.id))
    product.value = item
    selectedSKUID.value = item.skus[0]?.id ?? 0
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '商品详情加载失败'
  } finally {
    loading.value = false
  }
})

async function submitOrder() {
  submitError.value = ''
  submitting.value = true

  try {
    const payload = await createOrder(selectedSKUID.value, regionID.value, cycle.value)
    await router.push(`/payment/${payload.payment_order.payment_order_no}`)
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : '下单失败'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.stack {
  display: grid;
  gap: 18px;
}

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

.summary {
  color: #4d606e;
}

.sku-list {
  display: grid;
  gap: 12px;
  padding-left: 20px;
}

.sku-list li {
  display: grid;
  gap: 4px;
}

.form {
  display: grid;
  gap: 14px;
}

label {
  display: grid;
  gap: 8px;
}

input,
select,
button {
  border-radius: 16px;
  padding: 12px 14px;
  font: inherit;
}

input,
select {
  border: 1px solid #c8d7e1;
  background: #f9fbfc;
}

button {
  border: 0;
  background: #1d2a33;
  color: white;
}

.hint {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.state {
  margin: 0;
  padding: 18px 20px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.72);
}

.error,
.error-text {
  color: #b42318;
}
</style>
