<template>
  <section class="panel">
    <p class="tag">PAYMENT</p>
    <h2>支付单号 {{ route.params.paymentOrderNo }}</h2>
    <p v-if="loading">正在读取支付状态...</p>
    <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>
    <template v-else-if="paymentOrder">
      <p>当前状态：{{ paymentOrder.pay_status }}</p>
      <p>应付金额：{{ formatAmount(paymentOrder.payable_amount) }}</p>
      <p>关联订单：{{ paymentOrder.order_id }}</p>
      <p v-if="paymentOrder.paid_at">支付时间：{{ paymentOrder.paid_at }}</p>
    </template>
    <button type="button" @click="loadPayment" :disabled="loading">刷新状态</button>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { getPaymentStatus, type PaymentOrder } from '../api/payment'

const route = useRoute()
const paymentOrder = ref<PaymentOrder | null>(null)
const loading = ref(false)
const errorMessage = ref('')

function formatAmount(value: number) {
  return `CNY ${(value / 100).toFixed(2)}`
}

async function loadPayment() {
  loading.value = true
  errorMessage.value = ''

  try {
    paymentOrder.value = await getPaymentStatus(String(route.params.paymentOrderNo))
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '支付状态加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(loadPayment)
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

button {
  margin-top: 14px;
  border: 0;
  border-radius: 16px;
  padding: 12px 16px;
  background: #1d2a33;
  color: white;
}

.error {
  color: #b42318;
}
</style>
