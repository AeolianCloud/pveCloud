<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getInvoiceEligibleOrders } from '../../api/invoice'
import { cancelOrder, getOrders, type OrderItem } from '../../api/order'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'

const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const errorMessage = ref('')
const orders = ref<OrderItem[]>([])
const invoiceEligibleNos = ref<Set<string>>(new Set())
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, status: '' })

const statusText: Record<string, string> = { pending: '待处理', provisioning: '交付中', fulfilled: '已交付', error: '交付异常', cancelled: '已取消', closed: '已关闭' }
const orderTypeText: Record<string, string> = { purchase: '新购', renewal: '续费' }
const paymentStatusText: Record<string, string> = { unpaid: '未支付', paid: '已支付', manual_confirmed: '人工确认', refunded: '已退款' }
const cycleText: Record<string, string> = { monthly: '月付', quarterly: '季付', semi_yearly: '半年付', yearly: '年付' }
const formatMoney = (cents: number) => `¥${(cents / 100).toFixed(2)}`

async function loadOrders() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getOrders(query)
    orders.value = data.list
    total.value = data.total
    await loadInvoiceEligibility(data.list)
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '订单加载失败')
  } finally {
    loading.value = false
  }
}

async function loadInvoiceEligibility(currentOrders: OrderItem[]) {
  const next = new Set<string>()
  try {
    // 发票入口显隐只缓存后端可开票接口的返回结果；最终创建仍由发票接口事务校验。
    await Promise.all(currentOrders.map(async (order) => {
      const data = await getInvoiceEligibleOrders({ page: 1, per_page: 5, keyword: order.order_no })
      if (data.list.some((item) => item.order_no === order.order_no)) {
        next.add(order.order_no)
      }
    }))
  } catch {
    invoiceEligibleNos.value = new Set()
    return
  }
  invoiceEligibleNos.value = next
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
        <RouterLink to="/products" class="action-pill border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">继续选择套餐</RouterLink>
      </div>
      <div class="mb-6 flex flex-wrap gap-3">
        <button v-for="item in [{ label: '全部', value: '' }, { label: '待处理', value: 'pending' }, { label: '交付中', value: 'provisioning' }, { label: '已交付', value: 'fulfilled' }, { label: '交付异常', value: 'error' }, { label: '已取消', value: 'cancelled' }, { label: '已关闭', value: 'closed' }]" :key="item.value || 'all'" type="button" :class="['action-pill border px-4 py-2 text-xs font-black', query.status === item.value ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-neutral-300 text-neutral-700 hover:border-neutral-950']" @click="query.status = item.value; query.page = 1; loadOrders()">{{ item.label }}</button>
      </div>
      <div v-if="loading" class="space-y-3">
        <div v-for="item in 4" :key="item" class="rounded-2xl border border-neutral-200 bg-white p-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_9rem_10rem] lg:items-center">
            <div>
              <div class="skeleton-line h-3 w-36"></div>
              <div class="skeleton-line mt-3 h-5 w-64 max-w-full"></div>
              <div class="skeleton-line mt-3 h-3 w-80 max-w-full"></div>
            </div>
            <div class="skeleton-line h-8 w-28 lg:ml-auto"></div>
            <div class="skeleton-line h-8 w-32 lg:ml-auto"></div>
          </div>
        </div>
      </div>
      <div v-else-if="errorMessage" class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-red-600">订单异常</p>
        <h2 class="mt-3 text-2xl font-black text-neutral-950">订单加载失败</h2>
        <p class="mx-auto mt-3 max-w-xl text-sm leading-6 text-neutral-500">{{ errorMessage }}</p>
        <button type="button" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white" @click="loadOrders">重新加载</button>
      </div>
      <div v-else-if="orders.length" class="space-y-3">
        <article v-for="order in orders" :key="order.order_no" class="soft-lift rounded-2xl border border-neutral-200 bg-white p-4 sm:p-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_9rem_10rem] lg:items-center">
            <div class="min-w-0">
              <div class="truncate text-[11px] font-black uppercase tracking-[0.14em] text-neutral-500">{{ order.order_no }}</div>
              <h2 class="mt-1 truncate text-base font-black text-neutral-950 sm:text-lg">{{ order.product_name }} · {{ order.plan_name }}</h2>
              <p class="mt-1 truncate text-xs text-neutral-500 sm:text-sm">{{ orderTypeText[order.order_type] || order.order_type }} · {{ cycleText[order.billing_cycle] || order.billing_cycle }} · {{ order.network_type_name }} · {{ order.created_at }}</p>
              <p v-if="order.related_instance_no" class="mt-1 text-xs font-bold text-neutral-500">关联实例：{{ order.related_instance_no }}</p>
            </div>
            <div class="flex items-center justify-between gap-3 lg:block lg:text-right">
              <span class="inline-flex rounded-full border border-neutral-300 px-3 py-1 text-xs font-black">{{ statusText[order.status] }}</span>
              <div class="mt-1 text-xs font-black text-neutral-500">{{ paymentStatusText[order.payment_status] || order.payment_status }}</div>
              <div class="text-lg font-black lg:mt-2">{{ formatMoney(order.total_amount_cents) }}</div>
            </div>
            <div class="flex flex-wrap gap-2 lg:justify-end">
              <RouterLink :to="`/user/orders/${order.order_no}`" class="action-pill border border-neutral-950 px-3 py-1.5 text-xs font-black hover:bg-neutral-950 hover:text-white">查看详情</RouterLink>
              <RouterLink v-if="invoiceEligibleNos.has(order.order_no)" :to="{ path: '/user/invoices/new', query: { order_no: order.order_no } }" class="action-pill border border-sky-500 px-3 py-1.5 text-xs font-black text-sky-700 hover:bg-sky-50">申请发票</RouterLink>
              <RouterLink v-if="order.status === 'pending' && order.payment_status === 'unpaid'" :to="`/user/orders/${order.order_no}`" class="action-pill border border-emerald-500 px-3 py-1.5 text-xs font-black text-emerald-700 hover:bg-emerald-50">去支付</RouterLink>
              <button v-if="order.status === 'pending'" type="button" class="action-pill border border-red-300 px-3 py-1.5 text-xs font-black text-red-700 hover:bg-red-50" @click="cancel(order)">取消</button>
            </div>
          </div>
        </article>
      </div>
      <div v-else class="state-panel p-8 text-center"><p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">No Orders</p><h2 class="mt-3 text-2xl font-black">暂无订单</h2><p class="mt-3 text-sm text-neutral-500">从产品中心选择套餐后可以创建订单。</p><RouterLink to="/products" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">去选择套餐</RouterLink></div>
      <div v-if="total > query.per_page" class="mt-6 flex justify-center gap-3"><button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="query.page <= 1" @click="query.page--; loadOrders()">上一页</button><button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="orders.length < query.per_page" @click="query.page++; loadOrders()">下一页</button></div>
    </div>
  </div>
</template>
