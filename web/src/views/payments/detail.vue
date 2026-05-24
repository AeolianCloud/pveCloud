<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { getPayment, type PaymentStatus } from '../../api/payment'
import { getApiErrorMessage } from '../../api/request'

const route = useRoute()
const loading = ref(false)
const errorMessage = ref('')
const payment = ref<PaymentStatus | null>(null)
let timer: number | undefined

const statusText: Record<string, string> = {
  pending: '待支付',
  paid: '已支付',
  closed: '已关闭',
  failed: '支付失败',
  refunded: '已退款',
}

const methodText: Record<string, string> = {
  alipay_page: '支付宝电脑网页',
  alipay_wap: '支付宝手机网页',
  wechat_native: '微信扫码',
  wechat_h5: '微信 H5',
}

const terminal = computed(() => {
  const status = payment.value?.status
  return status === 'paid' || status === 'closed' || status === 'failed' || status === 'refunded'
})

const formatMoney = (cents: number) => `¥${(cents / 100).toFixed(2)}`

async function loadPayment() {
  loading.value = true
  errorMessage.value = ''
  try {
    payment.value = await getPayment(String(route.params.paymentNo || ''))
    if (terminal.value) stopPolling()
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '支付状态加载失败')
    stopPolling()
  } finally {
    loading.value = false
  }
}

function startPolling() {
  stopPolling()
  timer = window.setInterval(() => {
    if (!terminal.value) void loadPayment()
  }, 3000)
}

function stopPolling() {
  if (timer) {
    window.clearInterval(timer)
    timer = undefined
  }
}

onMounted(async () => {
  await loadPayment()
  if (!terminal.value) startPolling()
})

onBeforeUnmount(stopPolling)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
      <RouterLink to="/user/orders" class="mb-6 inline-flex text-sm font-black text-neutral-600 underline hover:text-neutral-950">返回订单</RouterLink>
      <div v-if="loading && !payment" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">支付状态加载中...</div>
      <div v-else-if="errorMessage" class="rounded-[1.5rem] border border-red-200 bg-red-50 p-6 text-sm font-bold text-red-700">{{ errorMessage }}</div>
      <article v-else-if="payment" class="rounded-[1.5rem] border border-neutral-200 bg-white p-6 shadow-[8px_8px_0_#111]">
        <div class="flex flex-col gap-4 border-b border-neutral-200 pb-5 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ payment.payment_no }}</p>
            <h1 class="mt-2 text-2xl font-black text-neutral-950">{{ methodText[payment.method] || payment.method }}</h1>
            <p class="mt-2 text-sm text-neutral-500">订单 {{ payment.order_no }}</p>
          </div>
          <div class="sm:text-right">
            <div class="text-2xl font-black">{{ formatMoney(payment.amount_cents) }}</div>
            <span class="mt-2 inline-flex rounded-full border px-3 py-1 text-xs font-black">{{ statusText[payment.status] || payment.status }}</span>
          </div>
        </div>

        <section class="mt-6 grid gap-3 md:grid-cols-2">
          <div class="rounded-xl bg-neutral-50 p-3"><div class="text-xs font-black text-neutral-500">过期时间</div><div class="mt-1 text-sm font-black">{{ payment.expires_at }}</div></div>
          <div class="rounded-xl bg-neutral-50 p-3"><div class="text-xs font-black text-neutral-500">订单状态</div><div class="mt-1 text-sm font-black">{{ payment.order_status }} / {{ payment.order_payment_status }}</div></div>
        </section>

        <section v-if="payment.status === 'pending'" class="mt-6">
          <div v-if="payment.qr_code_url" class="rounded-xl border border-neutral-200 p-4">
            <div class="text-sm font-black text-neutral-700">微信扫码支付</div>
            <div class="mt-3 break-all rounded-lg bg-neutral-950 p-4 text-sm font-bold text-white">{{ payment.qr_code_url }}</div>
          </div>
          <a v-if="payment.redirect_url" :href="payment.redirect_url" class="action-pill mt-4 inline-flex border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">打开支付页面</a>
          <p class="mt-4 text-sm text-neutral-500">页面会自动刷新支付状态。</p>
        </section>

        <section v-else class="mt-6 flex flex-wrap gap-3">
          <RouterLink :to="`/user/orders/${payment.order_no}`" class="action-pill border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">查看订单</RouterLink>
          <RouterLink v-if="payment.related_instance_no" :to="`/user/instances/${payment.related_instance_no}`" class="action-pill border border-neutral-300 px-5 py-2 text-sm font-black hover:bg-neutral-100">查看实例</RouterLink>
        </section>
      </article>
    </div>
  </div>
</template>
