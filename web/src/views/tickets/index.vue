<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getApiErrorMessage } from '../../api/request'
import {
  closeTicket,
  downloadTicketAttachment,
  getTicketDetail,
  getTickets,
  replyTicket,
  type TicketAttachment,
  type TicketDetail,
  type TicketItem,
} from '../../api/ticket'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'

const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const detailLoading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const tickets = ref<TicketItem[]>([])
const detail = ref<TicketDetail | null>(null)
const detailVisible = ref(false)
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, status: '' })
const replyContent = ref('')
const attachmentInput = ref<HTMLInputElement | null>(null)
const pendingAttachments = ref<PendingAttachment[]>([])
const attachmentPreviews = ref<Record<number, string>>({})
const attachmentObjectUrls = ref<string[]>([])
let attachmentSeq = 0

interface PendingAttachment {
  id: number
  file: File
  previewUrl: string | null
}

const statusText: Record<string, string> = { waiting_admin: '等待客服处理', waiting_user: '等待用户反馈', closed: '已关闭' }
const categoryText: Record<string, string> = { account: '账号问题', order: '订单问题', product: '产品咨询', technical: '技术支持', billing: '账务问题', other: '其它问题' }
const priorityText: Record<string, string> = { low: '低', normal: '普通', high: '高', urgent: '紧急' }
const statusClass: Record<string, string> = {
  waiting_admin: 'border-amber-200 bg-amber-50 text-amber-700',
  waiting_user: 'border-sky-200 bg-sky-50 text-sky-700',
  closed: 'border-neutral-200 bg-neutral-100 text-neutral-700',
}
const priorityClass: Record<string, string> = {
  low: 'border-neutral-200 bg-neutral-50 text-neutral-600',
  normal: 'border-blue-200 bg-blue-50 text-blue-700',
  high: 'border-orange-200 bg-orange-50 text-orange-700',
  urgent: 'border-red-200 bg-red-50 text-red-700',
}
const replyFiles = computed(() => pendingAttachments.value.map((item) => item.file))

async function loadTickets() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getTickets(query)
    tickets.value = data.list
    total.value = data.total
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '工单加载失败')
  } finally {
    loading.value = false
  }
}

function formatDateTime(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const pad = (num: number) => String(num).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function tagStyle(color: string | null) {
  return color ? { color, borderColor: color } : undefined
}

async function openDetail(ticketNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  replyContent.value = ''
  clearPendingAttachments()
  cleanupAttachmentPreviews()
  try {
    detail.value = await getTicketDetail(ticketNo)
    await loadAttachmentPreviews()
  } catch (err) {
    toast.error(getApiErrorMessage(err, '工单详情加载失败'))
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

function closeDetail() {
  detailVisible.value = false
  detail.value = null
  replyContent.value = ''
  clearPendingAttachments()
  cleanupAttachmentPreviews()
}

async function loadAttachmentPreviews() {
  if (!detail.value) return
  const images = detail.value.messages.flatMap((message) => message.attachments ?? []).filter(isImageAttachment)
  const previews: Record<number, string> = {}
  await Promise.all(images.map(async (file) => {
    try {
      const blob = await downloadTicketAttachment(detail.value!.ticket_no, file.file_id)
      const url = URL.createObjectURL(blob)
      attachmentObjectUrls.value.push(url)
      previews[file.file_id] = url
    } catch {
      previews[file.file_id] = ''
    }
  }))
  attachmentPreviews.value = previews
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

function removePendingAttachment(id: number) {
  const item = pendingAttachments.value.find((attachment) => attachment.id === id)
  if (item?.previewUrl) URL.revokeObjectURL(item.previewUrl)
  pendingAttachments.value = pendingAttachments.value.filter((attachment) => attachment.id !== id)
  if (attachmentInput.value) attachmentInput.value.value = ''
}

function clearPendingAttachments() {
  pendingAttachments.value.forEach((attachment) => {
    if (attachment.previewUrl) URL.revokeObjectURL(attachment.previewUrl)
  })
  pendingAttachments.value = []
}

async function submitReply() {
  if (!detail.value || !replyContent.value.trim()) return
  submitting.value = true
  try {
    detail.value = await replyTicket(detail.value.ticket_no, replyContent.value.trim(), replyFiles.value)
    replyContent.value = ''
    clearPendingAttachments()
    cleanupAttachmentPreviews()
    await loadAttachmentPreviews()
    await loadTickets()
    toast.success('回复已发送')
  } catch (err) {
    toast.error(getApiErrorMessage(err, '回复失败'))
  } finally {
    submitting.value = false
  }
}

async function closeCurrentTicket() {
  if (!detail.value) return
  const confirmed = await confirmDialog.confirm({
    title: '关闭工单',
    message: `确认关闭工单 ${detail.value.ticket_no}？`,
    confirmText: '确认关闭',
    cancelText: '继续沟通',
    tone: 'danger',
  })
  if (!confirmed) return
  submitting.value = true
  try {
    detail.value = await closeTicket(detail.value.ticket_no, '用户关闭工单')
    await loadTickets()
    toast.success('工单已关闭')
  } catch (err) {
    toast.error(getApiErrorMessage(err, '关闭失败'))
  } finally {
    submitting.value = false
  }
}

async function openAttachment(file: TicketAttachment) {
  if (!detail.value) return
  try {
    const blob = await downloadTicketAttachment(detail.value.ticket_no, file.file_id)
    const url = URL.createObjectURL(blob)
    attachmentObjectUrls.value.push(url)
    window.open(url, '_blank', 'noopener,noreferrer')
  } catch (err) {
    toast.error(getApiErrorMessage(err, '附件打开失败'))
  }
}

async function saveAttachment(file: TicketAttachment) {
  if (!detail.value) return
  try {
    const blob = await downloadTicketAttachment(detail.value.ticket_no, file.file_id)
    const url = URL.createObjectURL(blob)
    attachmentObjectUrls.value.push(url)
    const link = document.createElement('a')
    link.href = url
    link.download = file.original_name
    link.click()
  } catch (err) {
    toast.error(getApiErrorMessage(err, '附件下载失败'))
  }
}

function attachmentSize(size: number) {
  if (size >= 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  if (size >= 1024) return `${Math.round(size / 1024)} KB`
  return `${size} B`
}

function isImageAttachment(file: TicketAttachment) {
  return file.mime_type.startsWith('image/')
}

function attachmentType(file: TicketAttachment) {
  return file.extension || file.mime_type.split('/').pop() || 'file'
}

function cleanupObjectUrls() {
  attachmentObjectUrls.value.forEach((url) => URL.revokeObjectURL(url))
  attachmentObjectUrls.value = []
}

function cleanupAttachmentPreviews() {
  cleanupObjectUrls()
  attachmentPreviews.value = {}
}

onMounted(loadTickets)
onBeforeUnmount(() => {
  clearPendingAttachments()
  cleanupAttachmentPreviews()
})
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="mb-8 flex flex-col justify-between gap-4 border-b border-neutral-200 pb-8 md:flex-row md:items-end">
        <div><p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Tickets</p><h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">我的工单</h1><p class="mt-3 text-sm text-neutral-500">围绕账号、订单、产品和技术问题与客服沟通。</p></div>
        <RouterLink to="/user/tickets/new" class="action-pill border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">提交工单</RouterLink>
      </div>
      <div class="mb-6 flex flex-wrap gap-3">
        <button v-for="item in [{ label: '全部', value: '' }, { label: '等待客服处理', value: 'waiting_admin' }, { label: '等待用户反馈', value: 'waiting_user' }, { label: '已关闭', value: 'closed' }]" :key="item.value || 'all'" type="button" :class="['action-pill border px-4 py-2 text-xs font-black', query.status === item.value ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-neutral-300 text-neutral-700 hover:border-neutral-950']" @click="query.status = item.value; query.page = 1; loadTickets()">{{ item.label }}</button>
      </div>
      <div v-if="loading" class="space-y-3">
        <div v-for="item in 4" :key="item" class="rounded-2xl border border-neutral-200 bg-white p-5"><div class="skeleton-line h-3 w-36"></div><div class="skeleton-line mt-3 h-5 w-72 max-w-full"></div><div class="skeleton-line mt-3 h-3 w-96 max-w-full"></div></div>
      </div>
      <div v-else-if="errorMessage" class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-red-600">Tickets Error</p><h2 class="mt-3 text-2xl font-black text-neutral-950">工单加载失败</h2><p class="mx-auto mt-3 max-w-xl text-sm leading-6 text-neutral-500">{{ errorMessage }}</p><button type="button" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white" @click="loadTickets">重新加载</button>
      </div>
      <div v-else-if="tickets.length" class="space-y-3">
        <article v-for="ticket in tickets" :key="ticket.ticket_no" class="soft-lift rounded-2xl border border-neutral-200 bg-white p-4 sm:p-5">
          <div class="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,1fr)_12rem_6.5rem] lg:items-center">
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <span class="truncate text-[11px] font-black uppercase tracking-[0.14em] text-neutral-500">{{ ticket.ticket_no }}</span>
                <span :class="['shrink-0 rounded-full border px-2 py-0.5 text-[11px] font-black', priorityClass[ticket.priority] || 'border-neutral-200 bg-neutral-50 text-neutral-600']">{{ priorityText[ticket.priority] }}</span>
                <span v-for="tag in ticket.tags" :key="tag.id" class="shrink-0 rounded-full border border-neutral-300 px-2 py-0.5 text-[11px] font-black text-neutral-700" :style="tagStyle(tag.color)">{{ tag.name }}</span>
              </div>
              <h2 class="mt-2 line-clamp-2 text-base font-black leading-6 text-neutral-950 sm:text-lg">{{ ticket.title }}</h2>
              <div class="mt-2 flex min-w-0 flex-wrap gap-x-3 gap-y-1 text-xs font-bold text-neutral-500 sm:text-sm">
                <span>{{ categoryText[ticket.category] }}</span>
                <span class="min-w-0 truncate">{{ ticket.order_no ? `订单 ${ticket.order_no}` : '未关联订单' }}</span>
              </div>
            </div>
            <div class="grid gap-2 lg:justify-items-end">
              <span :class="['inline-flex w-fit rounded-full border px-3 py-1 text-xs font-black', statusClass[ticket.status] || 'border-neutral-300 text-neutral-700']">{{ statusText[ticket.status] }}</span>
              <div class="text-xs font-bold text-neutral-500 lg:text-right">
                <div>最近消息</div>
                <time :datetime="ticket.last_message_at" class="mt-0.5 block text-neutral-700">{{ formatDateTime(ticket.last_message_at) }}</time>
              </div>
            </div>
            <div class="flex lg:justify-end">
              <button type="button" class="action-pill w-full border border-neutral-950 px-3 py-2 text-xs font-black hover:bg-neutral-950 hover:text-white sm:w-auto" @click="openDetail(ticket.ticket_no)">查看详情</button>
            </div>
          </div>
        </article>
      </div>
      <div v-else class="state-panel p-8 text-center"><p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">No Tickets</p><h2 class="mt-3 text-2xl font-black">暂无工单</h2><p class="mt-3 text-sm text-neutral-500">有问题时可以提交工单与客服沟通。</p><RouterLink to="/user/tickets/new" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">提交工单</RouterLink></div>
      <div v-if="total > query.per_page" class="mt-6 flex justify-center gap-3"><button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="query.page <= 1" @click="query.page--; loadTickets()">上一页</button><button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="tickets.length < query.per_page" @click="query.page++; loadTickets()">下一页</button></div>
    </div>

    <Teleport to="body">
      <div v-if="detailVisible" class="fixed inset-0 z-50 flex items-end justify-center bg-neutral-950/50 p-0 sm:items-center sm:p-6" @click.self="closeDetail">
        <section class="max-h-[92vh] w-full max-w-4xl overflow-hidden rounded-t-[1.5rem] bg-white shadow-2xl sm:rounded-[1.5rem]">
          <div class="flex items-start justify-between gap-4 border-b border-neutral-200 p-5 sm:p-6">
            <div class="min-w-0">
              <p class="truncate text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ detail?.ticket_no || 'Ticket Detail' }}</p>
              <h2 class="mt-2 line-clamp-2 text-2xl font-black text-neutral-950">{{ detail?.title || '工单详情' }}</h2>
              <p v-if="detail" class="mt-2 text-sm text-neutral-500">{{ categoryText[detail.category] }} · {{ priorityText[detail.priority] }} · {{ detail.order_no || '未关联订单' }}</p>
            </div>
            <button type="button" class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-neutral-300 text-lg font-black hover:border-neutral-950" @click="closeDetail">×</button>
          </div>

          <div class="max-h-[calc(92vh-8rem)] overflow-y-auto p-5 sm:p-6">
            <div v-if="detailLoading" class="rounded-2xl border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">工单详情加载中...</div>
            <div v-else-if="detail" class="grid gap-5">
              <div class="flex flex-wrap items-center gap-2">
                <span :class="['inline-flex rounded-full border px-3 py-1 text-xs font-black', statusClass[detail.status] || 'border-neutral-300 text-neutral-700']">{{ statusText[detail.status] }}</span>
                <span v-for="tag in detail.tags" :key="tag.id" class="inline-flex rounded-full border border-neutral-300 px-3 py-1 text-xs font-black text-neutral-700" :style="tagStyle(tag.color)">{{ tag.name }}</span>
                <span class="text-xs font-bold text-neutral-500">创建：{{ formatDateTime(detail.created_at) }}</span>
                <span class="text-xs font-bold text-neutral-500">最近消息：{{ formatDateTime(detail.last_message_at) }}</span>
              </div>

              <section class="grid gap-4">
                <div v-for="message in detail.messages" :key="message.id" :class="['rounded-2xl border p-4', message.sender_type === 'user' ? 'border-neutral-950 bg-white' : 'border-neutral-200 bg-neutral-50']">
                  <div class="flex flex-wrap justify-between gap-2 text-xs font-black text-neutral-500"><span>{{ message.sender_type === 'user' ? '我' : message.sender_name }}</span><span>{{ formatDateTime(message.created_at) }}</span></div>
                  <p class="mt-3 whitespace-pre-wrap text-sm leading-6 text-neutral-800">{{ message.content }}</p>
                  <div v-if="(message.attachments ?? []).length" class="mt-3 grid gap-3 sm:grid-cols-2">
                    <article v-for="file in (message.attachments ?? [])" :key="file.file_id" class="grid min-w-0 grid-cols-[4rem_minmax(0,1fr)] gap-3 rounded-2xl border border-neutral-200 bg-white p-3">
                      <button type="button" class="h-16 w-16 overflow-hidden rounded-xl border border-neutral-200 bg-neutral-50" @click="openAttachment(file)">
                        <img v-if="isImageAttachment(file) && attachmentPreviews[file.file_id]" :src="attachmentPreviews[file.file_id]" :alt="file.original_name" class="h-full w-full object-cover" />
                        <span v-else class="flex h-full w-full items-center justify-center text-xs font-black uppercase text-neutral-500">{{ attachmentType(file) }}</span>
                      </button>
                      <div class="min-w-0">
                        <div class="truncate text-sm font-black text-neutral-950">{{ file.original_name }}</div>
                        <div class="mt-1 truncate text-xs font-bold text-neutral-500">{{ attachmentSize(file.size) }} · {{ file.mime_type }}</div>
                        <div class="mt-2 flex flex-wrap gap-2">
                          <button type="button" class="rounded-full border border-neutral-300 px-3 py-1 text-xs font-black text-neutral-700 hover:border-neutral-950" @click="openAttachment(file)">打开</button>
                          <button type="button" class="rounded-full border border-neutral-300 px-3 py-1 text-xs font-black text-neutral-700 hover:border-neutral-950" @click="saveAttachment(file)">下载</button>
                        </div>
                      </div>
                    </article>
                  </div>
                </div>
              </section>

              <section v-if="detail.status !== 'closed'" class="border-t border-neutral-200 pt-5">
                <h3 class="text-base font-black">继续回复</h3>
                <textarea v-model="replyContent" class="mt-3 min-h-32 w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" maxlength="5000" />
                <input ref="attachmentInput" type="file" multiple class="mt-3 block w-full rounded-xl border border-dashed border-neutral-300 px-4 py-3 text-sm" @change="onFilesChange" />
                <div v-if="pendingAttachments.length" class="mt-3 grid gap-3 sm:grid-cols-2">
                  <article v-for="attachment in pendingAttachments" :key="attachment.id" class="grid min-w-0 grid-cols-[4rem_minmax(0,1fr)_2rem] items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                    <img v-if="attachment.previewUrl" :src="attachment.previewUrl" :alt="attachment.file.name" class="h-16 w-16 rounded-xl border border-neutral-200 object-cover" />
                    <div v-else class="flex h-16 w-16 items-center justify-center rounded-xl border border-neutral-200 bg-white text-xs font-black uppercase text-neutral-500">{{ attachment.file.name.split('.').pop() || 'FILE' }}</div>
                    <div class="min-w-0">
                      <div class="truncate text-sm font-black text-neutral-950">{{ attachment.file.name }}</div>
                      <div class="mt-1 truncate text-xs font-bold text-neutral-500">{{ attachmentSize(attachment.file.size) }} · {{ attachment.file.type || '未知类型' }}</div>
                    </div>
                    <button type="button" class="flex h-8 w-8 items-center justify-center rounded-full border border-neutral-300 bg-white text-sm font-black hover:border-red-300 hover:text-red-700" @click="removePendingAttachment(attachment.id)">×</button>
                  </article>
                </div>
                <div class="mt-4 flex flex-wrap gap-3"><button type="button" class="action-pill border border-neutral-950 bg-neutral-950 px-5 py-2 text-sm font-black text-white disabled:opacity-50" :disabled="submitting || !replyContent.trim()" @click="submitReply">发送回复</button><button type="button" class="action-pill border border-red-300 px-5 py-2 text-sm font-black text-red-700 hover:bg-red-50" :disabled="submitting" @click="closeCurrentTicket">关闭工单</button></div>
              </section>
              <div v-else class="rounded-xl bg-neutral-50 p-4 text-sm font-bold text-neutral-600">工单已关闭，不能继续回复。</div>
            </div>
          </div>
        </section>
      </div>
    </Teleport>
  </div>
</template>
