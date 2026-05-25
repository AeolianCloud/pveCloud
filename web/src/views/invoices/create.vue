<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { createInvoice, getInvoiceEligibleOrders, type InvoiceEligibleOrderItem, type InvoiceTitleType } from '../../api/invoice'
import { getApiErrorMessage } from '../../api/request'
import { useToast } from '../../composables/useToast'
import { formatDateTime, formatMoney, orderTypeText } from './helpers'

interface InvoiceCreateForm {
  title_type: InvoiceTitleType
  title: string
  tax_no: string
  email: string
  remark: string
}

const route = useRoute()
const router = useRouter()
const toast = useToast()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const eligibleOrders = ref<InvoiceEligibleOrderItem[]>([])
const selectedOrderMap = ref<Record<string, InvoiceEligibleOrderItem>>({})
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, keyword: '' })
const form = reactive<InvoiceCreateForm>({ title_type: 'personal', title: '', tax_no: '', email: '', remark: '' })
const clientToken = ref(newClientToken())
const pendingPreselectOrderNo = ref('')

const selectedOrders = computed(() => Object.values(selectedOrderMap.value))
const selectedAmount = computed(() => selectedOrders.value.reduce((sum, order) => sum + order.amount_cents, 0))
const selectedCurrency = computed(() => selectedOrders.value[0]?.currency || 'CNY')

function newClientToken() {
  if (globalThis.crypto?.randomUUID) return `invoice-${globalThis.crypto.randomUUID()}`
  return `invoice-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

async function loadEligibleOrders() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getInvoiceEligibleOrders(query)
    eligibleOrders.value = data.list
    total.value = data.total
    if (pendingPreselectOrderNo.value) {
      const matched = data.list.find((order) => order.order_no === pendingPreselectOrderNo.value)
      if (matched) {
        selectedOrderMap.value = { ...selectedOrderMap.value, [matched.order_no]: matched }
        pendingPreselectOrderNo.value = ''
      }
    }
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '可开票订单加载失败')
  } finally {
    loading.value = false
  }
}

function isSelected(orderNo: string) {
  return Boolean(selectedOrderMap.value[orderNo])
}

function toggleOrder(order: InvoiceEligibleOrderItem) {
  const next = { ...selectedOrderMap.value }
  if (next[order.order_no]) {
    delete next[order.order_no]
  } else {
    next[order.order_no] = order
  }
  selectedOrderMap.value = next
}

function clearSelected() {
  selectedOrderMap.value = {}
}

async function submit() {
  if (!selectedOrders.value.length) {
    toast.error('请选择要开票的订单')
    return
  }
  if (!form.title.trim()) {
    toast.error('请填写发票抬头')
    return
  }
  if (form.title_type === 'company' && !form.tax_no.trim()) {
    toast.error('企业抬头必须填写税号')
    return
  }

  submitting.value = true
  try {
    // 前端只提交用户选择和抬头资料；订单归属、金额、占用和状态仍以后端事务校验为准。
    const invoice = await createInvoice({
      order_nos: selectedOrders.value.map((order) => order.order_no),
      title_type: form.title_type,
      title: form.title.trim(),
      tax_no: form.title_type === 'company' ? form.tax_no.trim() : null,
      email: form.email.trim() || null,
      remark: form.remark.trim() || null,
      client_token: clientToken.value,
    })
    clientToken.value = newClientToken()
    toast.success('发票申请已提交')
    await router.push(`/user/invoices/${invoice.invoice_no}`)
  } catch (err) {
    toast.error(getApiErrorMessage(err, '提交失败'))
  } finally {
    submitting.value = false
  }
}

function search() {
  query.page = 1
  void loadEligibleOrders()
}

onMounted(() => {
  const orderNo = typeof route.query.order_no === 'string' ? route.query.order_no.trim() : ''
  if (orderNo) {
    query.keyword = orderNo
    pendingPreselectOrderNo.value = orderNo
  }
  void loadEligibleOrders()
})
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-6xl px-4 py-12 sm:px-6 lg:px-8">
      <button type="button" class="mb-6 text-sm font-black text-neutral-600 underline hover:text-neutral-950" @click="router.back()">返回</button>
      <div class="mb-8 border-b border-neutral-200 pb-8">
        <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">New Invoice</p>
        <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">申请发票</h1>
        <p class="mt-3 text-sm text-neutral-500">只展示后端返回的可开票订单，可以选择多笔合并提交一张电子普通发票。</p>
      </div>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem]">
        <section class="min-w-0 rounded-[1.5rem] border border-neutral-200 bg-white p-5 shadow-[8px_8px_0_#111] sm:p-6">
          <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 class="text-lg font-black text-neutral-950">可开票订单</h2>
              <p class="mt-1 text-sm text-neutral-500">订单资格由服务端返回结果决定。</p>
            </div>
            <div class="flex min-w-0 gap-2">
              <input v-model="query.keyword" class="min-w-0 rounded-xl border border-neutral-300 px-4 py-2 text-sm font-bold outline-none focus:border-neutral-950" placeholder="订单编号" @keyup.enter="search" />
              <button type="button" class="action-pill shrink-0 border border-neutral-950 px-4 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white" @click="search">查询</button>
            </div>
          </div>

          <div v-if="loading" class="rounded-2xl border border-neutral-200 bg-neutral-50 p-5 text-sm font-bold text-neutral-600">可开票订单加载中...</div>
          <div v-else-if="errorMessage" class="rounded-2xl border border-red-200 bg-red-50 p-5 text-sm font-bold text-red-700">{{ errorMessage }}</div>
          <div v-else-if="eligibleOrders.length" class="grid gap-3">
            <button
              v-for="order in eligibleOrders"
              :key="order.order_no"
              type="button"
              :class="['grid min-w-0 gap-3 rounded-2xl border p-4 text-left transition sm:grid-cols-[1.5rem_minmax(0,1fr)_8rem] sm:items-center', isSelected(order.order_no) ? 'border-neutral-950 bg-neutral-50 shadow-[4px_4px_0_#111]' : 'border-neutral-200 bg-white hover:border-neutral-950']"
              @click="toggleOrder(order)"
            >
              <span :class="['flex h-6 w-6 items-center justify-center rounded-full border text-xs font-black', isSelected(order.order_no) ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-neutral-300 bg-white text-neutral-400']">{{ isSelected(order.order_no) ? '✓' : '' }}</span>
              <span class="min-w-0">
                <span class="block truncate text-sm font-black text-neutral-950">{{ order.order_no }} · {{ order.product_name }} · {{ order.plan_name }}</span>
                <span class="mt-1 block truncate text-xs font-bold text-neutral-500">
                  {{ orderTypeText[order.order_type] || order.order_type }} · {{ order.payment_status }} · {{ formatDateTime(order.paid_at) }}
                  <span v-if="order.related_instance_no"> · {{ order.related_instance_no }}</span>
                </span>
              </span>
              <span class="text-base font-black text-neutral-950 sm:text-right">{{ formatMoney(order.amount_cents, order.currency) }}</span>
            </button>
          </div>
          <div v-else class="rounded-2xl border border-dashed border-neutral-300 bg-neutral-50 p-8 text-center">
            <p class="text-sm font-black text-neutral-950">暂无可开票订单</p>
            <p class="mt-2 text-sm text-neutral-500">已支付且未被有效发票申请占用的订单会出现在这里。</p>
            <RouterLink to="/user/orders" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">查看订单</RouterLink>
          </div>

          <div v-if="total > query.per_page" class="mt-5 flex justify-center gap-3">
            <button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="query.page <= 1" @click="query.page--; loadEligibleOrders()">上一页</button>
            <button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="eligibleOrders.length < query.per_page" @click="query.page++; loadEligibleOrders()">下一页</button>
          </div>
        </section>

        <aside class="min-w-0 rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5 sm:p-6">
          <h2 class="text-lg font-black text-neutral-950">抬头资料</h2>
          <div class="mt-5 grid grid-cols-2 gap-2 rounded-2xl border border-neutral-200 bg-white p-2">
            <button type="button" :class="['rounded-xl px-3 py-2 text-sm font-black', form.title_type === 'personal' ? 'bg-neutral-950 text-white' : 'text-neutral-600 hover:bg-neutral-100']" @click="form.title_type = 'personal'">个人</button>
            <button type="button" :class="['rounded-xl px-3 py-2 text-sm font-black', form.title_type === 'company' ? 'bg-neutral-950 text-white' : 'text-neutral-600 hover:bg-neutral-100']" @click="form.title_type = 'company'">企业</button>
          </div>

          <form class="mt-5 grid gap-4" @submit.prevent="submit">
            <label class="grid gap-2 text-sm font-black">
              发票抬头
              <input v-model="form.title" maxlength="100" class="w-full rounded-xl border border-neutral-300 px-4 py-3 font-medium outline-none focus:border-neutral-950" />
            </label>
            <label v-if="form.title_type === 'company'" class="grid gap-2 text-sm font-black">
              税号
              <input v-model="form.tax_no" maxlength="64" class="w-full rounded-xl border border-neutral-300 px-4 py-3 font-medium outline-none focus:border-neutral-950" />
            </label>
            <label class="grid gap-2 text-sm font-black">
              接收邮箱
              <input v-model="form.email" maxlength="128" type="email" class="w-full rounded-xl border border-neutral-300 px-4 py-3 font-medium outline-none focus:border-neutral-950" />
            </label>
            <label class="grid gap-2 text-sm font-black">
              备注
              <textarea v-model="form.remark" maxlength="500" class="min-h-28 w-full rounded-xl border border-neutral-300 px-4 py-3 font-medium outline-none focus:border-neutral-950" />
            </label>

            <div class="rounded-2xl border border-neutral-200 bg-white p-4">
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="font-black text-neutral-600">已选订单</span>
                <button v-if="selectedOrders.length" type="button" class="text-xs font-black text-neutral-600 underline" @click="clearSelected">清空</button>
              </div>
              <div class="mt-2 text-2xl font-black text-neutral-950">{{ formatMoney(selectedAmount, selectedCurrency) }}</div>
              <div class="mt-1 text-xs font-bold text-neutral-500">{{ selectedOrders.length }} 个订单</div>
            </div>

            <button type="submit" class="action-pill border border-neutral-950 bg-neutral-950 px-5 py-3 text-sm font-black text-white disabled:opacity-50" :disabled="submitting">
              {{ submitting ? '提交中...' : '提交申请' }}
            </button>
          </form>
        </aside>
      </div>
    </div>
  </div>
</template>
