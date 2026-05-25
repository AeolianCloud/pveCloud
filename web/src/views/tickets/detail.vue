<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { closeTicket, downloadTicketAttachment, getTicketDetail, replyTicket, type TicketAttachment, type TicketDetail } from '../../api/ticket'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'

const route = useRoute()
const router = useRouter()
const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const ticket = ref<TicketDetail | null>(null)
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
const files = computed(() => pendingAttachments.value.map((item) => item.file))

async function loadDetail() {
  loading.value = true
  errorMessage.value = ''
  try {
    cleanupAttachmentPreviews()
    ticket.value = await getTicketDetail(String(route.params.ticketNo || ''))
    await loadAttachmentPreviews()
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '工单详情加载失败')
  } finally {
    loading.value = false
  }
}

async function loadAttachmentPreviews() {
  if (!ticket.value) return
  const images = ticket.value.messages.flatMap((message) => message.attachments ?? []).filter(isImageAttachment)
  const previews: Record<number, string> = {}
  await Promise.all(images.map(async (file) => {
    try {
      const blob = await downloadTicketAttachment(ticket.value!.ticket_no, file.file_id)
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
  if (!ticket.value || !replyContent.value.trim()) return
  submitting.value = true
  try {
    ticket.value = await replyTicket(ticket.value.ticket_no, replyContent.value.trim(), files.value)
    replyContent.value = ''
    clearPendingAttachments()
    toast.success('回复已发送')
  } catch (err) {
    toast.error(getApiErrorMessage(err, '回复失败'))
  } finally {
    submitting.value = false
  }
}

async function openAttachment(file: TicketAttachment) {
  if (!ticket.value) return
  try {
    const blob = await downloadTicketAttachment(ticket.value.ticket_no, file.file_id)
    const url = URL.createObjectURL(blob)
    attachmentObjectUrls.value.push(url)
    window.open(url, '_blank', 'noopener,noreferrer')
  } catch (err) {
    toast.error(getApiErrorMessage(err, '附件打开失败'))
  }
}

async function saveAttachment(file: TicketAttachment) {
  if (!ticket.value) return
  try {
    const blob = await downloadTicketAttachment(ticket.value.ticket_no, file.file_id)
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

async function closeCurrentTicket() {
  if (!ticket.value) return
  const confirmed = await confirmDialog.confirm({ title: '关闭工单', message: `确认关闭工单 ${ticket.value.ticket_no}？`, confirmText: '确认关闭', cancelText: '继续沟通', tone: 'danger' })
  if (!confirmed) return
  submitting.value = true
  try {
    ticket.value = await closeTicket(ticket.value.ticket_no, '用户关闭工单')
    toast.success('工单已关闭')
  } catch (err) {
    toast.error(getApiErrorMessage(err, '关闭失败'))
  } finally {
    submitting.value = false
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

function tagStyle(color: string | null) {
  return color ? { color, borderColor: color } : undefined
}

function linkLabel(label: string, value: string | null) {
  return value ? `${label} ${value}` : `未关联${label}`
}

function cleanupObjectUrls() {
  attachmentObjectUrls.value.forEach((url) => URL.revokeObjectURL(url))
  attachmentObjectUrls.value = []
}

function cleanupAttachmentPreviews() {
  cleanupObjectUrls()
  attachmentPreviews.value = {}
}

onMounted(loadDetail)
onBeforeUnmount(() => {
  clearPendingAttachments()
  cleanupAttachmentPreviews()
})
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
      <button type="button" class="mb-6 text-sm font-black text-neutral-600 underline hover:text-neutral-950" @click="router.back()">返回</button>
      <div v-if="loading" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">工单详情加载中...</div>
      <div v-else-if="errorMessage" class="rounded-[1.5rem] border border-red-200 bg-red-50 p-6 text-sm font-bold text-red-700">{{ errorMessage }}</div>
      <article v-else-if="ticket" class="rounded-[1.5rem] border border-neutral-200 bg-white p-5 shadow-[8px_8px_0_#111] sm:p-6">
        <div class="grid gap-4 border-b border-neutral-200 pb-5 md:grid-cols-[minmax(0,1fr)_9rem] md:items-start">
          <div class="min-w-0"><p class="truncate text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ ticket.ticket_no }}</p><h1 class="mt-2 text-2xl font-black text-neutral-950">{{ ticket.title }}</h1><p class="mt-2 text-sm text-neutral-500">{{ categoryText[ticket.category] }} · {{ priorityText[ticket.priority] }} · {{ linkLabel('订单', ticket.order_no) }} · {{ linkLabel('实例', ticket.instance_no) }}</p><div v-if="ticket.tags.length" class="mt-3 flex flex-wrap gap-2"><span v-for="tag in ticket.tags" :key="tag.id" class="rounded-full border border-neutral-300 px-3 py-1 text-xs font-black text-neutral-700" :style="tagStyle(tag.color)">{{ tag.name }}</span></div></div>
          <span class="inline-flex justify-center rounded-full border px-3 py-1 text-xs font-black">{{ statusText[ticket.status] }}</span>
        </div>
        <section class="mt-6 grid gap-4">
          <div v-for="message in ticket.messages" :key="message.id" :class="['rounded-2xl border p-4', message.sender_type === 'user' ? 'border-neutral-950 bg-white' : 'border-neutral-200 bg-neutral-50']">
            <div class="flex flex-wrap justify-between gap-2 text-xs font-black text-neutral-500"><span>{{ message.sender_type === 'user' ? '我' : message.sender_name }}</span><span>{{ message.created_at }}</span></div>
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
        <section v-if="ticket.status !== 'closed'" class="mt-6 border-t border-neutral-200 pt-6">
          <h2 class="text-base font-black">继续回复</h2>
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
        <div v-else class="mt-6 rounded-xl bg-neutral-50 p-4 text-sm font-bold text-neutral-600">工单已关闭，不能继续回复。</div>
        <RouterLink to="/user/tickets" class="action-pill mt-6 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">返回工单列表</RouterLink>
      </article>
    </div>
  </div>
</template>
