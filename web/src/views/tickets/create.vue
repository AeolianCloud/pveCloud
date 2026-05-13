<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getOrders, type OrderItem } from '../../api/order'
import { getApiErrorMessage } from '../../api/request'
import { createTicket } from '../../api/ticket'
import AppSelect, { type AppSelectOption } from '../../components/AppSelect.vue'
import { useToast } from '../../composables/useToast'

const router = useRouter()
const toast = useToast()
const submitting = ref(false)
const ordersLoading = ref(false)
const attachmentInput = ref<HTMLInputElement | null>(null)
const orders = ref<OrderItem[]>([])
const form = reactive({ title: '', category: 'other', priority: 'normal', order_no: '', content: '' })
const pendingAttachments = ref<PendingAttachment[]>([])
let attachmentSeq = 0

interface PendingAttachment {
  id: number
  file: File
  previewUrl: string | null
}

const categories: AppSelectOption[] = [
  { label: '账号问题', value: 'account', description: '登录、资料、认证等账号相关问题' },
  { label: '订单问题', value: 'order', description: '订单创建、状态或人工处理咨询' },
  { label: '产品咨询', value: 'product', description: '套餐、地域、配置和售前咨询' },
  { label: '技术支持', value: 'technical', description: '使用过程中遇到的技术问题' },
  { label: '账务问题', value: 'billing', description: '费用、发票或账务记录咨询' },
  { label: '其它问题', value: 'other', description: '不属于以上分类的通用问题' },
]
const priorities: AppSelectOption[] = [
  { label: '低', value: 'low', description: '一般咨询或非紧急问题' },
  { label: '普通', value: 'normal', description: '默认优先级，适合大多数工单' },
  { label: '高', value: 'high', description: '影响正常使用，需要优先处理' },
  { label: '紧急', value: 'urgent', description: '严重影响业务连续性的问题' },
]

const selectedOrder = computed(() => orders.value.find((order) => order.order_no === form.order_no) || null)
const files = computed(() => pendingAttachments.value.map((item) => item.file))
const statusText: Record<string, string> = { pending: '待处理', cancelled: '已取消', closed: '已关闭' }
const statusClass: Record<string, string> = {
  pending: 'border-amber-200 bg-amber-50 text-amber-700',
  cancelled: 'border-red-200 bg-red-50 text-red-700',
  closed: 'border-neutral-200 bg-neutral-100 text-neutral-700',
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const data = await getOrders({ page: 1, per_page: 100 })
    orders.value = data.list
  } catch (err) {
    toast.error(getApiErrorMessage(err, '订单加载失败'))
  } finally {
    ordersLoading.value = false
  }
}

function onFilesChange(event: Event) {
  const input = event.target as HTMLInputElement
  const selected = Array.from(input.files || [])
  const remaining = 5 - pendingAttachments.value.length
  if (selected.length > remaining) {
    toast.error('单条消息最多 5 个附件')
  }
  const accepted = selected.slice(0, Math.max(remaining, 0))
  pendingAttachments.value.push(...accepted.map(toPendingAttachment))
  input.value = ''
}

function toPendingAttachment(file: File): PendingAttachment {
  return {
    id: ++attachmentSeq,
    file,
    previewUrl: file.type.startsWith('image/') ? URL.createObjectURL(file) : null,
  }
}

function removeAttachment(id: number) {
  const item = pendingAttachments.value.find((attachment) => attachment.id === id)
  if (item?.previewUrl) URL.revokeObjectURL(item.previewUrl)
  pendingAttachments.value = pendingAttachments.value.filter((attachment) => attachment.id !== id)
  if (attachmentInput.value) attachmentInput.value.value = ''
}

function clearAttachments() {
  pendingAttachments.value.forEach((attachment) => {
    if (attachment.previewUrl) URL.revokeObjectURL(attachment.previewUrl)
  })
  pendingAttachments.value = []
}

function attachmentSize(size: number) {
  if (size >= 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  if (size >= 1024) return `${Math.round(size / 1024)} KB`
  return `${size} B`
}

function selectOrder(orderNo: string) {
  form.order_no = orderNo
}

async function submit() {
  if (!form.title.trim() || !form.content.trim()) {
    toast.error('请填写标题和问题描述')
    return
  }
  submitting.value = true
  try {
    const ticket = await createTicket({ ...form, order_no: form.order_no.trim() || undefined, files: files.value })
    clearAttachments()
    toast.success('工单已提交')
    await router.push(`/user/tickets/${ticket.ticket_no}`)
  } catch (err) {
    toast.error(getApiErrorMessage(err, '提交失败'))
  } finally {
    submitting.value = false
  }
}

onMounted(loadOrders)
onBeforeUnmount(clearAttachments)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
      <button type="button" class="mb-6 text-sm font-black text-neutral-600 underline hover:text-neutral-950" @click="router.back()">返回</button>
      <div class="mb-8 border-b border-neutral-200 pb-8">
        <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">New Ticket</p>
        <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">提交工单</h1>
      </div>
      <form class="min-w-0 overflow-hidden rounded-[1.5rem] border border-neutral-200 bg-white p-5 shadow-[8px_8px_0_#111] sm:p-6" @submit.prevent="submit">
        <div class="grid min-w-0 gap-4 md:grid-cols-2">
          <label class="grid min-w-0 gap-2 text-sm font-black">标题<input v-model="form.title" class="w-full min-w-0 rounded-xl border border-neutral-300 px-4 py-3 font-medium outline-none focus:border-neutral-950" maxlength="160" /></label>
          <div class="grid min-w-0 gap-2 text-sm font-black md:row-span-3">
            <div class="flex items-center justify-between gap-3">
              <span>关联订单（可选）</span>
              <span v-if="ordersLoading" class="text-xs font-bold text-neutral-500">加载中...</span>
            </div>
            <div class="min-w-0 rounded-2xl border border-neutral-200 bg-neutral-50 p-2">
              <button type="button" :class="['mb-2 flex w-full min-w-0 items-center justify-between gap-3 rounded-xl border px-3 py-2 text-left text-sm', !form.order_no ? 'border-neutral-950 bg-white shadow-sm' : 'border-transparent bg-transparent hover:bg-white']" @click="selectOrder('')">
                <span class="min-w-0 truncate font-black text-neutral-950">不关联订单</span>
                <span class="shrink-0 rounded-full border border-neutral-200 bg-white px-2 py-0.5 text-[11px] font-black text-neutral-500">默认</span>
              </button>
              <div v-if="orders.length" class="grid max-h-56 gap-2 overflow-y-auto pr-1">
                <button v-for="order in orders" :key="order.order_no" type="button" :class="['grid min-w-0 gap-1 rounded-xl border px-3 py-2 text-left', form.order_no === order.order_no ? 'border-neutral-950 bg-white shadow-sm' : 'border-transparent bg-white/70 hover:border-neutral-300 hover:bg-white']" @click="selectOrder(order.order_no)">
                  <div class="flex min-w-0 items-center justify-between gap-2">
                    <span class="min-w-0 truncate text-xs font-black text-neutral-950">{{ order.order_no }}</span>
                    <span :class="['shrink-0 rounded-full border px-2 py-0.5 text-[11px] font-black', statusClass[order.status] || 'border-neutral-200 bg-white text-neutral-600']">{{ statusText[order.status] || order.status }}</span>
                  </div>
                  <div class="min-w-0 truncate text-xs font-bold text-neutral-500">{{ order.product_name }} · {{ order.plan_name }}</div>
                </button>
              </div>
              <div v-else-if="!ordersLoading" class="rounded-xl border border-dashed border-neutral-300 bg-white px-3 py-4 text-center text-xs font-bold text-neutral-500">当前账号暂无可选订单</div>
            </div>
            <div v-if="selectedOrder" class="min-w-0 rounded-xl border border-neutral-200 bg-white px-3 py-2 text-xs font-bold text-neutral-600">
              已选择 <span class="font-black text-neutral-950">{{ selectedOrder.order_no }}</span>
              <span :class="['ml-2 inline-flex rounded-full border px-2 py-0.5 text-[11px] font-black', statusClass[selectedOrder.status] || 'border-neutral-200 bg-white text-neutral-600']">{{ statusText[selectedOrder.status] || selectedOrder.status }}</span>
            </div>
          </div>
          <label class="grid min-w-0 gap-2 text-sm font-black">分类<AppSelect v-model="form.category" :options="categories" placeholder="请选择工单分类" /></label>
          <label class="grid min-w-0 gap-2 text-sm font-black">优先级<AppSelect v-model="form.priority" :options="priorities" placeholder="请选择优先级" /></label>
        </div>
        <label class="mt-4 grid min-w-0 gap-2 text-sm font-black">问题描述<textarea v-model="form.content" class="min-h-40 w-full min-w-0 rounded-xl border border-neutral-300 px-4 py-3 font-medium outline-none focus:border-neutral-950" maxlength="5000" /></label>
        <label class="mt-4 grid min-w-0 gap-2 text-sm font-black">
          附件
          <input ref="attachmentInput" type="file" multiple class="w-full min-w-0 rounded-xl border border-dashed border-neutral-300 px-4 py-3 text-sm" @change="onFilesChange" />
        </label>
        <div v-if="pendingAttachments.length" class="mt-3 grid gap-3 sm:grid-cols-2">
          <article v-for="attachment in pendingAttachments" :key="attachment.id" class="grid min-w-0 grid-cols-[4rem_minmax(0,1fr)_2rem] items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
            <img v-if="attachment.previewUrl" :src="attachment.previewUrl" :alt="attachment.file.name" class="h-16 w-16 rounded-xl border border-neutral-200 object-cover" />
            <div v-else class="flex h-16 w-16 items-center justify-center rounded-xl border border-neutral-200 bg-white text-xs font-black uppercase text-neutral-500">{{ attachment.file.name.split('.').pop() || 'FILE' }}</div>
            <div class="min-w-0">
              <div class="truncate text-sm font-black text-neutral-950">{{ attachment.file.name }}</div>
              <div class="mt-1 truncate text-xs font-bold text-neutral-500">{{ attachmentSize(attachment.file.size) }} · {{ attachment.file.type || '未知类型' }}</div>
            </div>
            <button type="button" class="flex h-8 w-8 items-center justify-center rounded-full border border-neutral-300 bg-white text-sm font-black hover:border-red-300 hover:text-red-700" @click="removeAttachment(attachment.id)">×</button>
          </article>
        </div>
        <p class="mt-2 text-xs text-neutral-500">单条消息最多 5 个附件，最终安全校验以后端为准。</p>
        <div class="mt-6 flex flex-wrap gap-3"><button type="submit" class="action-pill border border-neutral-950 bg-neutral-950 px-5 py-2 text-sm font-black text-white disabled:opacity-50" :disabled="submitting">{{ submitting ? '提交中...' : '提交工单' }}</button><RouterLink to="/user/tickets" class="action-pill border border-neutral-300 px-5 py-2 text-sm font-black hover:border-neutral-950">返回列表</RouterLink></div>
      </form>
    </div>
  </div>
</template>
