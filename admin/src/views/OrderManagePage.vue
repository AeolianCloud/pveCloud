<template>
  <section class="panel">
    <p class="tag">ORDERS</p>
    <h2>订单管理</h2>
    <p v-if="loading">正在加载订单...</p>
    <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="orders.length === 0">当前没有订单。</p>
    <table v-else>
      <thead>
        <tr>
          <th>订单号</th>
          <th>用户</th>
          <th>状态</th>
          <th>金额</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="order in orders" :key="order.id">
          <td>{{ order.order_no }}</td>
          <td>{{ order.user_id }}</td>
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
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 12px 10px;
  text-align: left;
  border-bottom: 1px solid #d9e3dc;
}

.error {
  color: #b42318;
}
</style>
