<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { cancelOrder, getOrders, type OrderItem } from '../../api/order'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'

const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const errorMessage = ref('')
const orders = ref<OrderItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, status: '' })

const statusText: Record<string, string> = { pending: '待处理', cancelled: '已取消', closed: '已关闭' }
const cycleText: Record<string, string> = { monthly: '月付', quarterly: '季付', semi_yearly: '半年付', yearly: '年付' }
const formatMoney = (cents: number) => `¥${(cents / 100).toFixed(2)}`

async function loadOrders() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getOrders(query)
    orders.value = data.list
    total.value = data.total
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '订单加载失败')
  } finally {
    loading.value = false
  }
}

async function cancel(order: OrderItem) {
  const confirmed = await confirmDialog.confirm({
    title: '取消订单',
    message: `确认取消订单 ${order.order_no}？取消后无法在当前阶段继续处理。`,
    confirmText: '确认取消',
    cancelText: '保留订单',
    tone: 'danger',
  })
  if (!confirmed) return
  try {
    await cancelOrder(order.order_no, '用户取消订单')
    toast.success('订单已取消')
    await loadOrders()
  } catch (err) {
    toast.error(getApiErrorMessage(err, '取消失败'))
  }
}

onMounted(loadOrders)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="mb-8 flex flex-col justify-between gap-4 border-b border-neutral-200 pb-8 md:flex-row md:items-end">
        <div><p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Orders</p><h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">我的订单</h1><p class="mt-3 text-sm text-neutral-500">订单表示购买意向和后台人工处理入口，不代表支付或实例交付。</p></div>
        <RouterLink to="/products" class="rounded-full border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">继续选择套餐</RouterLink>
      </div>
      <div class="mb-6 flex flex-wrap gap-3">
        <button v-for="item in [{ label: '全部', value: '' }, { label: '待处理', value: 'pending' }, { label: '已取消', value: 'cancelled' }, { label: '已关闭', value: 'closed' }]" :key="item.value || 'all'" type="button" :class="['rounded-full border px-4 py-2 text-xs font-black', query.status === item.value ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-neutral-300 text-neutral-700']" @click="query.status = item.value; query.page = 1; loadOrders()">{{ item.label }}</button>
      </div>
      <div v-if="loading" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">订单加载中...</div>
      <div v-else-if="errorMessage" class="rounded-[1.5rem] border border-red-200 bg-red-50 p-6 text-sm font-bold text-red-700">{{ errorMessage }}</div>
      <div v-else-if="orders.length" class="space-y-4">
        <article v-for="order in orders" :key="order.order_no" class="interactive-card rounded-[1.5rem] border border-neutral-200 bg-white p-6">
          <div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
            <div><div class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ order.order_no }}</div><h2 class="mt-2 text-xl font-black text-neutral-950">{{ order.product_name }} · {{ order.plan_name }}</h2><p class="mt-2 text-sm text-neutral-500">{{ cycleText[order.billing_cycle] || order.billing_cycle }} · {{ order.created_at }}</p></div>
            <div class="text-left md:text-right"><div class="text-2xl font-black">{{ formatMoney(order.total_amount_cents) }}</div><span class="mt-2 inline-flex rounded-full border border-neutral-300 px-3 py-1 text-xs font-black">{{ statusText[order.status] }}</span></div>
          </div>
          <div class="mt-5 flex flex-wrap gap-3"><RouterLink :to="`/user/orders/${order.order_no}`" class="rounded-full border border-neutral-950 px-4 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">查看详情</RouterLink><button v-if="order.status === 'pending'" type="button" class="rounded-full border border-red-300 px-4 py-2 text-sm font-black text-red-700 hover:bg-red-50" @click="cancel(order)">取消订单</button></div>
        </article>
      </div>
      <div v-else class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-8 text-center"><h2 class="text-xl font-black">暂无订单</h2><p class="mt-2 text-sm text-neutral-500">从产品中心选择套餐后可以创建订单。</p></div>
      <div v-if="total > query.per_page" class="mt-6 flex justify-center gap-3"><button type="button" class="rounded-full border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="query.page <= 1" @click="query.page--; loadOrders()">上一页</button><button type="button" class="rounded-full border px-4 py-2 text-sm font-black" :disabled="orders.length < query.per_page" @click="query.page++; loadOrders()">下一页</button></div>
    </div>
  </div>
</template>
