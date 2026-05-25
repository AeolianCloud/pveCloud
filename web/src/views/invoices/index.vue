<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { cancelInvoice, downloadInvoice, getInvoices, type InvoiceItem } from '../../api/invoice'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'
import { formatDateTime, formatMoney, invoiceStatusClass, invoiceStatusText, saveInvoiceBlob, titleTypeText } from './helpers'

const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const downloading = ref('')
const errorMessage = ref('')
const invoices = ref<InvoiceItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, status: '' })

const statusFilters = [
  { label: '全部', value: '' },
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已开票', value: 'issued' },
  { label: '已驳回', value: 'rejected' },
  { label: '已取消', value: 'cancelled' },
]

async function loadInvoices() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getInvoices(query)
    invoices.value = data.list
    total.value = data.total
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '发票列表加载失败')
  } finally {
    loading.value = false
  }
}

async function cancel(item: InvoiceItem) {
  const confirmed = await confirmDialog.confirm({
    title: '取消发票申请',
    message: `确认取消发票申请 ${item.invoice_no}？取消后会释放关联订单的开票占用。`,
    confirmText: '确认取消',
    cancelText: '保留申请',
    tone: 'danger',
  })
  if (!confirmed) return

  try {
    await cancelInvoice(item.invoice_no, '用户取消发票申请')
    toast.success('发票申请已取消')
    await loadInvoices()
  } catch (err) {
    toast.error(getApiErrorMessage(err, '取消失败'))
  }
}

async function download(item: InvoiceItem) {
  downloading.value = item.invoice_no
  try {
    const blob = await downloadInvoice(item.invoice_no)
    saveInvoiceBlob(blob, item.invoice_no)
  } catch (err) {
    toast.error(getApiErrorMessage(err, '下载失败'))
  } finally {
    downloading.value = ''
  }
}

function selectStatus(value: string) {
  query.status = value
  query.page = 1
  void loadInvoices()
}

onMounted(loadInvoices)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="mb-8 flex flex-col justify-between gap-4 border-b border-neutral-200 pb-8 md:flex-row md:items-end">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Invoices</p>
          <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">发票管理</h1>
          <p class="mt-3 text-sm text-neutral-500">电子普通发票由运营人工处理，已支付订单可合并提交申请。</p>
        </div>
        <RouterLink to="/user/invoices/new" class="action-pill border border-neutral-950 bg-neutral-950 px-5 py-2 text-sm font-black text-white hover:bg-white hover:text-neutral-950">申请发票</RouterLink>
      </div>

      <div class="mb-6 flex flex-wrap gap-3">
        <button
          v-for="item in statusFilters"
          :key="item.value || 'all'"
          type="button"
          :class="['action-pill border px-4 py-2 text-xs font-black', query.status === item.value ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-neutral-300 text-neutral-700 hover:border-neutral-950']"
          @click="selectStatus(item.value)"
        >
          {{ item.label }}
        </button>
      </div>

      <div v-if="loading" class="space-y-3">
        <div v-for="item in 4" :key="item" class="rounded-2xl border border-neutral-200 bg-white p-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_9rem_12rem] lg:items-center">
            <div>
              <div class="skeleton-line h-3 w-36"></div>
              <div class="skeleton-line mt-3 h-5 w-72 max-w-full"></div>
              <div class="skeleton-line mt-3 h-3 w-80 max-w-full"></div>
            </div>
            <div class="skeleton-line h-8 w-28 lg:ml-auto"></div>
            <div class="skeleton-line h-8 w-32 lg:ml-auto"></div>
          </div>
        </div>
      </div>

      <div v-else-if="errorMessage" class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-red-600">发票异常</p>
        <h2 class="mt-3 text-2xl font-black text-neutral-950">发票列表加载失败</h2>
        <p class="mx-auto mt-3 max-w-xl text-sm leading-6 text-neutral-500">{{ errorMessage }}</p>
        <button type="button" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white" @click="loadInvoices">重新加载</button>
      </div>

      <div v-else-if="invoices.length" class="space-y-3">
        <article v-for="item in invoices" :key="item.invoice_no" class="soft-lift rounded-2xl border border-neutral-200 bg-white p-4 sm:p-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_10rem_13rem] lg:items-center">
            <div class="min-w-0">
              <div class="truncate text-[11px] font-black uppercase tracking-[0.14em] text-neutral-500">{{ item.invoice_no }}</div>
              <h2 class="mt-1 truncate text-base font-black text-neutral-950 sm:text-lg">{{ titleTypeText[item.title_type] || item.title_type }} · {{ item.title }}</h2>
              <p class="mt-1 truncate text-xs text-neutral-500 sm:text-sm">
                {{ item.order_count }} 个订单 · {{ formatDateTime(item.created_at) }}
              </p>
              <p v-if="item.invoice_number" class="mt-1 truncate text-xs font-bold text-neutral-500">发票号码：{{ item.invoice_number }}</p>
            </div>
            <div class="flex items-center justify-between gap-3 lg:block lg:text-right">
              <span :class="['inline-flex rounded-full border px-3 py-1 text-xs font-black', invoiceStatusClass[item.status] || 'border-neutral-200 bg-white text-neutral-600']">{{ invoiceStatusText[item.status] || item.status }}</span>
              <div class="text-lg font-black lg:mt-2">{{ formatMoney(item.amount_cents, item.currency) }}</div>
              <div class="text-xs font-bold text-neutral-500">{{ formatDateTime(item.issued_at) }}</div>
            </div>
            <div class="flex flex-wrap gap-2 lg:justify-end">
              <RouterLink :to="`/user/invoices/${item.invoice_no}`" class="action-pill border border-neutral-950 px-3 py-1.5 text-xs font-black hover:bg-neutral-950 hover:text-white">详情</RouterLink>
              <button v-if="item.can_download" type="button" class="action-pill border border-emerald-500 px-3 py-1.5 text-xs font-black text-emerald-700 hover:bg-emerald-50 disabled:opacity-50" :disabled="downloading === item.invoice_no" @click="download(item)">
                {{ downloading === item.invoice_no ? '下载中...' : '下载 PDF' }}
              </button>
              <button v-if="item.can_cancel" type="button" class="action-pill border border-red-300 px-3 py-1.5 text-xs font-black text-red-700 hover:bg-red-50" @click="cancel(item)">取消</button>
            </div>
          </div>
        </article>
      </div>

      <div v-else class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">No Invoices</p>
        <h2 class="mt-3 text-2xl font-black">暂无发票申请</h2>
        <p class="mt-3 text-sm text-neutral-500">可以从可开票订单中选择一笔或多笔合并提交申请。</p>
        <RouterLink to="/user/invoices/new" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">申请发票</RouterLink>
      </div>

      <div v-if="total > query.per_page" class="mt-6 flex justify-center gap-3">
        <button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="query.page <= 1" @click="query.page--; loadInvoices()">上一页</button>
        <button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="invoices.length < query.per_page" @click="query.page++; loadInvoices()">下一页</button>
      </div>
    </div>
  </div>
</template>
