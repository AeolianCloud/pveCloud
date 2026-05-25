<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { cancelInvoice, downloadInvoice, getInvoiceDetail, type InvoiceDetail } from '../../api/invoice'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'
import { formatDateTime, formatMoney, invoiceStatusClass, invoiceStatusText, orderTypeText, saveInvoiceBlob, titleTypeText } from './helpers'

const route = useRoute()
const router = useRouter()
const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const downloading = ref(false)
const errorMessage = ref('')
const invoice = ref<InvoiceDetail | null>(null)

async function loadDetail() {
  loading.value = true
  errorMessage.value = ''
  try {
    invoice.value = await getInvoiceDetail(String(route.params.invoiceNo || ''))
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '发票详情加载失败')
  } finally {
    loading.value = false
  }
}

async function cancel() {
  if (!invoice.value) return
  const confirmed = await confirmDialog.confirm({
    title: '取消发票申请',
    message: `确认取消发票申请 ${invoice.value.invoice_no}？`,
    confirmText: '确认取消',
    cancelText: '保留申请',
    tone: 'danger',
  })
  if (!confirmed) return

  try {
    await cancelInvoice(invoice.value.invoice_no, '用户取消发票申请')
    toast.success('发票申请已取消')
    await loadDetail()
  } catch (err) {
    toast.error(getApiErrorMessage(err, '取消失败'))
  }
}

async function download() {
  if (!invoice.value) return
  downloading.value = true
  try {
    const blob = await downloadInvoice(invoice.value.invoice_no)
    saveInvoiceBlob(blob, invoice.value.invoice_no)
  } catch (err) {
    toast.error(getApiErrorMessage(err, '下载失败'))
  } finally {
    downloading.value = false
  }
}

onMounted(loadDetail)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
      <button type="button" class="mb-6 text-sm font-black text-neutral-600 underline hover:text-neutral-950" @click="router.back()">返回</button>

      <div v-if="loading" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">发票详情加载中...</div>
      <div v-else-if="errorMessage" class="rounded-[1.5rem] border border-red-200 bg-red-50 p-6 text-sm font-bold text-red-700">{{ errorMessage }}</div>

      <article v-else-if="invoice" class="rounded-[1.5rem] border border-neutral-200 bg-white p-5 shadow-[8px_8px_0_#111] sm:p-6">
        <div class="grid gap-4 border-b border-neutral-200 pb-5 md:grid-cols-[minmax(0,1fr)_12rem] md:items-start">
          <div class="min-w-0">
            <p class="truncate text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ invoice.invoice_no }}</p>
            <h1 class="mt-2 text-2xl font-black text-neutral-950">{{ titleTypeText[invoice.title_type] || invoice.title_type }} · {{ invoice.title }}</h1>
            <p class="mt-2 text-sm text-neutral-500">电子普通发票申请详情。</p>
          </div>
          <div class="flex items-center justify-between gap-3 md:block md:text-right">
            <div class="text-2xl font-black">{{ formatMoney(invoice.amount_cents, invoice.currency) }}</div>
            <span :class="['inline-flex rounded-full border px-3 py-1 text-xs font-black md:mt-2', invoiceStatusClass[invoice.status] || 'border-neutral-200 bg-white text-neutral-600']">{{ invoiceStatusText[invoice.status] || invoice.status }}</span>
          </div>
        </div>

        <dl class="mt-6 grid gap-3 md:grid-cols-2">
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">发票类型</dt><dd class="mt-1 text-sm font-black">电子普通发票</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">订单数量</dt><dd class="mt-1 text-sm font-black">{{ invoice.order_count }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">税号</dt><dd class="mt-1 break-all text-sm font-black">{{ invoice.tax_no || '-' }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">接收邮箱</dt><dd class="mt-1 break-all text-sm font-black">{{ invoice.email || '-' }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">申请时间</dt><dd class="mt-1 text-sm font-black">{{ formatDateTime(invoice.created_at) }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">受理时间</dt><dd class="mt-1 text-sm font-black">{{ formatDateTime(invoice.accepted_at) }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">开票时间</dt><dd class="mt-1 text-sm font-black">{{ formatDateTime(invoice.issued_at) }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">发票号码</dt><dd class="mt-1 break-all text-sm font-black">{{ invoice.invoice_number || '-' }}</dd></div>
        </dl>

        <section v-if="invoice.remark || invoice.reject_reason || invoice.cancel_reason" class="mt-6 grid gap-3">
          <div v-if="invoice.remark" class="rounded-xl border border-neutral-200 bg-white p-4">
            <h2 class="text-sm font-black text-neutral-950">申请备注</h2>
            <p class="mt-2 whitespace-pre-wrap text-sm text-neutral-600">{{ invoice.remark }}</p>
          </div>
          <div v-if="invoice.reject_reason" class="rounded-xl border border-red-200 bg-red-50 p-4">
            <h2 class="text-sm font-black text-red-800">驳回原因</h2>
            <p class="mt-2 whitespace-pre-wrap text-sm text-red-700">{{ invoice.reject_reason }}</p>
          </div>
          <div v-if="invoice.cancel_reason" class="rounded-xl border border-neutral-200 bg-neutral-50 p-4">
            <h2 class="text-sm font-black text-neutral-950">取消原因</h2>
            <p class="mt-2 whitespace-pre-wrap text-sm text-neutral-600">{{ invoice.cancel_reason }}</p>
          </div>
        </section>

        <section class="mt-6">
          <h2 class="text-base font-black text-neutral-950">订单明细</h2>
          <div class="mt-3 overflow-hidden rounded-2xl border border-neutral-200">
            <div v-for="order in invoice.orders" :key="order.order_no" class="grid gap-2 border-b border-neutral-100 p-4 text-sm last:border-b-0 md:grid-cols-[minmax(0,1fr)_7rem_9rem] md:items-center">
              <div class="min-w-0">
                <RouterLink :to="`/user/orders/${order.order_no}`" class="truncate font-black text-neutral-950 underline">{{ order.order_no }}</RouterLink>
                <div class="mt-1 truncate text-xs font-bold text-neutral-500">{{ orderTypeText[order.order_type] || order.order_type }} · {{ order.product_name || '-' }} · {{ order.plan_name || '-' }}</div>
              </div>
              <div class="text-xs font-bold text-neutral-500">{{ formatDateTime(order.paid_at) }}</div>
              <div class="font-black md:text-right">{{ formatMoney(order.order_amount_cents, order.currency) }}</div>
            </div>
          </div>
        </section>

        <div class="mt-6 flex flex-wrap gap-3">
          <button v-if="invoice.can_download" type="button" class="action-pill border border-emerald-500 px-5 py-2 text-sm font-black text-emerald-700 hover:bg-emerald-50 disabled:opacity-50" :disabled="downloading" @click="download">
            {{ downloading ? '下载中...' : '下载 PDF' }}
          </button>
          <button v-if="invoice.can_cancel" type="button" class="action-pill border border-red-300 px-5 py-2 text-sm font-black text-red-700 hover:bg-red-50" @click="cancel">取消申请</button>
          <RouterLink to="/user/invoices" class="action-pill border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">返回发票列表</RouterLink>
        </div>
      </article>
    </div>
  </div>
</template>
