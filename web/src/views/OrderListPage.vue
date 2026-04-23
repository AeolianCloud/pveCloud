<template>
  <section class="panel">
    <p class="tag">ORDERS</p>
    <h2>我的真实订单</h2>

    <p v-if="loading" class="state">正在加载订单...</p>
    <p v-else-if="errorMessage" class="state error">{{ errorMessage }}</p>
    <p v-else-if="orders.length === 0" class="state">当前还没有订单。</p>

    <table v-else>
      <thead>
        <tr>
          <th>订单号</th>
          <th>SKU ID</th>
          <th>状态</th>
          <th>应付金额</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="order in orders" :key="order.id">
          <td>{{ order.order_no }}</td>
          <td>{{ order.sku_id }}</td>
          <td>{{ order.status }}</td>
          <td>{{ formatAmount(order.payable_amount) }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { listOrders, type Order } from '../api/order'

const orders = ref<Order[]>([])
const loading = ref(true)
const errorMessage = ref('')

function formatAmount(value: number) {
  return `CNY ${(value / 100).toFixed(2)}`
}

onMounted(async () => {
  try {
    orders.value = await listOrders()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '订单加载失败'
  } finally {
    loading.value = false
  }
})
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

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 14px 10px;
  text-align: left;
  border-bottom: 1px solid #e1eaef;
}

.state {
  margin: 0 0 14px;
  color: #4d606e;
}

.error {
  color: #b42318;
}
</style>
