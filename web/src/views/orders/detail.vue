<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { cancelOrder, getOrderDetail, type OrderDetail } from '../../api/order'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'

const route = useRoute()
const router = useRouter()
const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const errorMessage = ref('')
const order = ref<OrderDetail | null>(null)
const statusText: Record<string, string> = { pending: '待处理', cancelled: '已取消', closed: '已关闭' }
const cycleText: Record<string, string> = { monthly: '月付', quarterly: '季付', semi_yearly: '半年付', yearly: '年付' }
const formatMoney = (cents: number) => `¥${(cents / 100).toFixed(2)}`
const formatMemory = (mb: number) => mb >= 1024 ? `${Math.round(mb / 1024)}GB` : `${mb}MB`

async function loadDetail() {
  loading.value = true
  errorMessage.value = ''
  try {
    order.value = await getOrderDetail(String(route.params.orderNo || ''))
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '订单详情加载失败')
  } finally {
    loading.value = false
  }
}

async function cancel() {
  if (!order.value) return
  const confirmed = await confirmDialog.confirm({
    title: '取消订单',
    message: `确认取消订单 ${order.value.order_no}？取消后无法在当前阶段继续处理。`,
    confirmText: '确认取消',
    cancelText: '保留订单',
    tone: 'danger',
  })
  if (!confirmed) return
  try {
    await cancelOrder(order.value.order_no, '用户取消订单')
    toast.success('订单已取消')
    await loadDetail()
  } catch (err) {
    toast.error(getApiErrorMessage(err, '取消失败'))
  }
}

onMounted(loadDetail)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
      <button type="button" class="mb-6 text-sm font-black text-neutral-600 underline hover:text-neutral-950" @click="router.back()">返回</button>
      <div v-if="loading" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">订单详情加载中...</div>
      <div v-else-if="errorMessage" class="rounded-[1.5rem] border border-red-200 bg-red-50 p-6 text-sm font-bold text-red-700">{{ errorMessage }}</div>
      <article v-else-if="order" class="rounded-[1.5rem] border border-neutral-200 bg-white p-5 shadow-[8px_8px_0_#111] sm:p-6">
        <div class="grid gap-4 border-b border-neutral-200 pb-5 md:grid-cols-[minmax(0,1fr)_10rem] md:items-start">
          <div class="min-w-0"><p class="truncate text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ order.order_no }}</p><h1 class="mt-2 text-2xl font-black text-neutral-950">{{ order.product_name }} · {{ order.plan_name }}</h1><p class="mt-2 text-sm text-neutral-500">订单只表示购买意向，不代表支付或实例交付。</p></div>
          <div class="flex items-center justify-between gap-3 md:block md:text-right"><div class="text-2xl font-black">{{ formatMoney(order.total_amount_cents) }}</div><span class="inline-flex rounded-full border px-3 py-1 text-xs font-black md:mt-2">{{ statusText[order.status] }}</span></div>
        </div>
        <dl class="mt-6 grid gap-3 md:grid-cols-2">
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">计费周期</dt><dd class="mt-1 text-sm font-black">{{ cycleText[order.billing_cycle] || order.billing_cycle }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">销售地域</dt><dd class="mt-1 text-sm font-black">{{ order.region_name }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">系统模板</dt><dd class="mt-1 text-sm font-black">{{ order.template_name }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">网络类型</dt><dd class="mt-1 text-sm font-black">{{ order.network_type_name }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">创建时间</dt><dd class="mt-1 text-sm font-black">{{ order.created_at }}</dd></div>
        </dl>
        <section class="mt-6"><h2 class="text-base font-black">配置快照</h2><div class="mt-3 grid gap-2 text-sm md:grid-cols-4"><div class="rounded-xl border p-3">{{ order.cpu_cores }} 核 CPU</div><div class="rounded-xl border p-3">{{ formatMemory(order.memory_mb) }} 内存</div><div class="rounded-xl border p-3">{{ order.system_disk_gb + order.data_disk_gb }}GB 磁盘</div><div class="rounded-xl border p-3">{{ order.bandwidth_mbps }}M 带宽</div></div></section>
        <section v-if="order.user_note" class="mt-6"><h2 class="text-base font-black">用户备注</h2><p class="mt-2 rounded-xl bg-neutral-50 p-3 text-sm text-neutral-600">{{ order.user_note }}</p></section>
        <div class="mt-6 flex flex-wrap gap-3"><button v-if="order.status === 'pending'" type="button" class="action-pill border border-red-300 px-5 py-2 text-sm font-black text-red-700 hover:bg-red-50" @click="cancel">取消订单</button><RouterLink to="/user/orders" class="action-pill border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">返回订单列表</RouterLink></div>
      </article>
    </div>
  </div>
</template>
