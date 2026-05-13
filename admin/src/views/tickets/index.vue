<script setup lang="ts">
import {
  AttachOutline,
  ChatbubbleEllipsesOutline,
  CheckmarkDoneOutline,
  CloseCircleOutline,
  DocumentTextOutline,
  DownloadOutline,
  MailOpenOutline,
  PersonOutline,
} from '@vicons/ionicons5'
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NInput,
  NPagination,
  NSelect,
  NSpace,
  NScrollbar,
  NIcon,
  NEmpty,
  NTimeline,
  NTimelineItem,
  NTag,
  NUpload,
  type DataTableColumns,
  type UploadFileInfo,
} from 'naive-ui'
import { computed, h, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

import {
  closeTicket,
  downloadTicketAttachment,
  getTicketDetail,
  getTickets,
  replyTicket,
  type AdminTicketAttachment,
  type AdminTicketDetail,
  type AdminTicketItem,
  type AdminTicketMessage,
} from '../../api/ticket'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'
import { message } from '../../utils/feedback'
import { hasPermissionCode } from '../../utils/permission'

const permissionStore = usePermissionStore()
const loading = ref(false)
const detailLoading = ref(false)
const detailVisible = ref(false)
const submitting = ref(false)
const tickets = ref<AdminTicketItem[]>([])
const detail = ref<AdminTicketDetail | null>(null)
const total = ref(0)
const replyContent = ref('')
const uploadFiles = ref<UploadFileInfo[]>([])
const closeReason = ref('')
const query = reactive({ page: 1, per_page: 15, status: '', category: '', priority: '', ticket_no: '', order_no: '', user_keyword: '' })
const attachmentPreviews = ref<Record<number, string>>({})
const attachmentObjectUrls = ref<string[]>([])

const canReply = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:reply'))
const canClose = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:close'))
const selectedClosed = computed(() => detail.value?.status === 'closed')
const detailStatusMeta = computed(() => {
  if (!detail.value) return { label: '-', type: 'default' as const, icon: CheckmarkDoneOutline }
  if (detail.value.status === 'closed') return { label: '已关闭', type: 'default' as const, icon: CloseCircleOutline }
  if (detail.value.status === 'waiting_user') return { label: '待用户反馈', type: 'warning' as const, icon: MailOpenOutline }
  return { label: '待后台处理', type: 'info' as const, icon: ChatbubbleEllipsesOutline }
})
const detailCategoryLabel = computed(() => (detail.value ? categoryText[detail.value.category] || detail.value.category : '-'))
const detailPriorityLabel = computed(() => (detail.value ? priorityText[detail.value.priority] || detail.value.priority : '-'))
const detailMessageCount = computed(() => detail.value?.messages.length || 0)
const detailAttachmentCount = computed(() => detail.value?.messages.reduce((count, item) => count + (item.attachments?.length || 0), 0) || 0)
const replyDisabled = computed(() => !replyContent.value.trim() || submitting.value || selectedClosed.value || !canReply.value)
const closeDisabled = computed(() => submitting.value || selectedClosed.value || !canClose.value)
const detailUserLabel = computed(() => detail.value ? `${detail.value.user.username} / ${detail.value.user.email}` : '-')
const detailOrderLabel = computed(() => detail.value?.order_no || '-')
const detailClosedAtLabel = computed(() => formatDateTime(detail.value?.closed_at))

const statusText: Record<string, string> = { waiting_admin: '待后台处理', waiting_user: '待用户反馈', closed: '已关闭' }
const categoryText: Record<string, string> = { account: '账号问题', order: '订单问题', product: '产品咨询', technical: '技术支持', billing: '账务问题', other: '其它问题' }
const priorityText: Record<string, string> = { low: '低', normal: '普通', high: '高', urgent: '紧急' }

const statusOptions = [
  { label: '待后台处理', value: 'waiting_admin' },
  { label: '待用户反馈', value: 'waiting_user' },
  { label: '已关闭', value: 'closed' },
]
const categoryOptions = Object.entries(categoryText).map(([value, label]) => ({ value, label }))
const priorityOptions = Object.entries(priorityText).map(([value, label]) => ({ value, label }))

async function loadTickets() {
  loading.value = true
  try {
    const data = await getTickets(query)
    tickets.value = data.list
    total.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '工单加载失败')
  } finally {
    loading.value = false
  }
}

async function openDetail(ticketNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    cleanupAttachmentPreviews()
    detail.value = await getTicketDetail(ticketNo)
    await loadAttachmentPreviews()
    replyContent.value = ''
    uploadFiles.value = []
    closeReason.value = ''
  } catch (err) {
    message.error(err instanceof Error ? err.message : '工单详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

async function loadAttachmentPreviews() {
  if (!detail.value) return
  const images = detail.value.messages.flatMap((item) => item.attachments ?? []).filter(isImageAttachment)
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

async function submitReply() {
  if (!detail.value || !replyContent.value.trim()) return
  submitting.value = true
  try {
    const files = uploadFiles.value.map((item) => item.file).filter((file): file is File => Boolean(file))
    detail.value = await replyTicket(detail.value.ticket_no, replyContent.value.trim(), files)
    replyContent.value = ''
    uploadFiles.value = []
    message.success('回复已发送')
    await loadTickets()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '回复失败')
  } finally {
    submitting.value = false
  }
}

async function submitClose() {
  if (!detail.value) return
  submitting.value = true
  try {
    detail.value = await closeTicket(detail.value.ticket_no, closeReason.value || undefined)
    closeReason.value = ''
    message.success('工单已关闭')
    await loadTickets()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '关闭失败')
  } finally {
    submitting.value = false
  }
}

function resetQuery() {
  Object.assign(query, { page: 1, per_page: 15, status: '', category: '', priority: '', ticket_no: '', order_no: '', user_keyword: '' })
  loadTickets()
}

function attachmentSize(size: number) {
  if (size >= 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  if (size >= 1024) return `${Math.round(size / 1024)} KB`
  return `${size} B`
}

function isImageAttachment(file: AdminTicketAttachment | null | undefined) {
  if (!file) return false
  return file.mime_type.startsWith('image/')
}

function attachmentType(file: AdminTicketAttachment | null | undefined) {
  if (!file) return 'file'
  return file.extension || file.mime_type.split('/').pop() || 'file'
}

function attachmentLabel(file: AdminTicketAttachment | null | undefined) {
  if (!file) return '-'
  return `${attachmentSize(file.size)} · ${file.mime_type}`
}

async function openAttachment(file: AdminTicketAttachment) {
  if (!detail.value) return
  try {
    const blob = await downloadTicketAttachment(detail.value.ticket_no, file.file_id)
    const url = URL.createObjectURL(blob)
    attachmentObjectUrls.value.push(url)
    window.open(url, '_blank', 'noopener,noreferrer')
  } catch (err) {
    message.error(err instanceof Error ? err.message : '附件打开失败')
  }
}

async function saveAttachment(file: AdminTicketAttachment) {
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
    message.error(err instanceof Error ? err.message : '附件下载失败')
  }
}

function cleanupObjectUrls() {
  attachmentObjectUrls.value.forEach((url) => URL.revokeObjectURL(url))
  attachmentObjectUrls.value = []
}

function cleanupAttachmentPreviews() {
  cleanupObjectUrls()
  attachmentPreviews.value = {}
}

function resetDetailState() {
  detail.value = null
  replyContent.value = ''
  uploadFiles.value = []
  closeReason.value = ''
  cleanupAttachmentPreviews()
}

function handleDetailVisibleChange(show: boolean) {
  detailVisible.value = show
  if (!show) {
    resetDetailState()
  }
}

function renderMessage(messageItem: AdminTicketMessage) {
  return h('div', { class: ['ticket-message', messageItem.sender_type === 'admin' ? 'ticket-message--admin' : ''] }, [
    h('div', { class: 'ticket-message__meta' }, [
      h('div', { class: 'ticket-message__sender' }, [
        h(NIcon, { size: 14 }, { default: () => h(messageItem.sender_type === 'admin' ? CheckmarkDoneOutline : PersonOutline) }),
        h('span', null, messageItem.sender_name),
      ]),
      h('div', { class: 'ticket-message__time' }, formatDateTime(messageItem.created_at)),
    ]),
    h('div', { class: 'ticket-message__content' }, messageItem.content),
    (messageItem.attachments ?? []).length
      ? h('div', { class: 'ticket-message__attachments' }, (messageItem.attachments ?? []).map((file, index) =>
          h('div', { class: 'ticket-attachment' }, [
            h('button', { class: 'ticket-attachment__preview', type: 'button', disabled: !file, onClick: () => file && openAttachment(file) }, [
              isImageAttachment(file) && attachmentPreviews.value[file.file_id]
                ? h('img', { src: attachmentPreviews.value[file.file_id], alt: file?.original_name || '附件' })
                : h('span', null, attachmentType(file)),
            ]),
            h('div', { class: 'ticket-attachment__body' }, [
              h('div', { class: 'ticket-attachment__name' }, file?.original_name || '未命名附件'),
              h('div', { class: 'ticket-attachment__meta' }, attachmentLabel(file)),
              h('div', { class: 'ticket-attachment__actions' }, [
                h(
                  NButton,
                  { size: 'tiny', secondary: true, disabled: !file, onClick: () => file && openAttachment(file) },
                  { icon: () => h(NIcon, { size: 14 }, { default: () => h(DocumentTextOutline) }), default: () => '预览' },
                ),
                h(
                  NButton,
                  { size: 'tiny', secondary: true, disabled: !file, onClick: () => file && saveAttachment(file) },
                  { icon: () => h(NIcon, { size: 14 }, { default: () => h(DownloadOutline) }), default: () => '下载' },
                ),
              ]),
            ]),
          ]),
        ))
      : null,
  ])
}

const columns = computed<DataTableColumns<AdminTicketItem>>(() => [
  { key: 'ticket_no', title: '工单号', minWidth: 170 },
  {
    key: 'title',
    title: '标题',
    minWidth: 220,
    render: (row) => h('div', null, [h('div', { class: 'strong' }, row.title), h('div', { class: 'muted' }, row.order_no ? `订单 ${row.order_no}` : '未关联订单')]),
  },
  {
    key: 'user',
    title: '用户',
    minWidth: 160,
    render: (row) => h('div', null, [h('div', { class: 'strong' }, row.user.username), h('div', { class: 'muted' }, row.user.email)]),
  },
  { key: 'category', title: '分类', width: 110, render: (row) => categoryText[row.category] || row.category },
  { key: 'priority', title: '优先级', width: 90, render: (row) => priorityText[row.priority] || row.priority },
  { key: 'status', title: '状态', width: 120, render: (row) => h(NTag, { size: 'small' }, { default: () => statusText[row.status] || row.status }) },
  { key: 'last_message_at', title: '最近消息', minWidth: 160, render: (row) => formatDateTime(row.last_message_at) },
  {
    key: 'actions',
    title: '操作',
    width: 90,
    fixed: 'right',
    render: (row) => h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row.ticket_no) }, { default: () => '详情' }),
  },
])

onMounted(loadTickets)
onBeforeUnmount(cleanupAttachmentPreviews)
</script>

<template>
  <div class="tickets-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>工单管理</h2>
          <p class="muted">查看用户提交的工单，进入详情后处理回复、关闭和附件。</p>
        </div>
      </template>

      <NForm inline label-placement="left" class="query-form">
        <NFormItem label="状态"><NSelect v-model:value="query.status" :options="statusOptions" clearable placeholder="全部" style="width: 140px" /></NFormItem>
        <NFormItem label="分类"><NSelect v-model:value="query.category" :options="categoryOptions" clearable placeholder="全部" style="width: 130px" /></NFormItem>
        <NFormItem label="优先级"><NSelect v-model:value="query.priority" :options="priorityOptions" clearable placeholder="全部" style="width: 110px" /></NFormItem>
        <NFormItem label="工单号"><NInput v-model:value="query.ticket_no" clearable placeholder="TIC-" /></NFormItem>
        <NFormItem label="订单号"><NInput v-model:value="query.order_no" clearable placeholder="ORD-" /></NFormItem>
        <NFormItem label="用户"><NInput v-model:value="query.user_keyword" clearable placeholder="用户名/邮箱" /></NFormItem>
        <NFormItem :show-label="false">
          <NSpace><NButton type="primary" @click="query.page = 1; loadTickets()">查询</NButton><NButton @click="resetQuery">重置</NButton></NSpace>
        </NFormItem>
      </NForm>

      <NDataTable :loading="loading" :columns="columns" :data="tickets" :row-key="(row: AdminTicketItem) => row.ticket_no" :bordered="false" />
      <div class="pagination"><NPagination v-model:page="query.page" v-model:page-size="query.per_page" :item-count="total" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="loadTickets" @update:page-size="loadTickets" /></div>
    </NCard>

    <NDrawer :show="detailVisible" placement="right" :width="1040" :auto-focus="false" @update:show="handleDetailVisibleChange">
      <NDrawerContent closable :title="detail?.title || '工单详情'">
        <NScrollbar v-if="detailLoading" class="ticket-detail-shell">
          <div class="ticket-detail-skeleton">
            <div class="skeleton-card skeleton-card--lg" />
            <div class="skeleton-grid">
              <div class="skeleton-card" />
              <div class="skeleton-card" />
              <div class="skeleton-card" />
              <div class="skeleton-card" />
            </div>
          </div>
        </NScrollbar>

        <div v-else-if="detail" class="ticket-detail-shell">
          <section class="ticket-hero">
            <div class="ticket-hero__main">
              <div class="ticket-hero__title-row">
                <div>
                  <div class="ticket-hero__eyebrow">工单详情</div>
                  <h3>{{ detail.title }}</h3>
                  <div class="ticket-hero__meta">{{ detail.ticket_no }} · {{ detailUserLabel }}</div>
                </div>
                <NTag :type="detailStatusMeta.type" round size="small">
                  <template #icon>
                    <NIcon :size="14">
                      <component :is="detailStatusMeta.icon" />
                    </NIcon>
                  </template>
                  {{ detailStatusMeta.label }}
                </NTag>
              </div>

              <div class="ticket-hero__stats">
                <div class="stat-card">
                  <div class="stat-card__label">消息</div>
                  <div class="stat-card__value">{{ detailMessageCount }}</div>
                </div>
                <div class="stat-card">
                  <div class="stat-card__label">附件</div>
                  <div class="stat-card__value">{{ detailAttachmentCount }}</div>
                </div>
                <div class="stat-card">
                  <div class="stat-card__label">分类</div>
                  <div class="stat-card__value">{{ detailCategoryLabel }}</div>
                </div>
                <div class="stat-card">
                  <div class="stat-card__label">优先级</div>
                  <div class="stat-card__value">{{ detailPriorityLabel }}</div>
                </div>
              </div>
            </div>
            <div class="ticket-hero__aside">
              <div class="ticket-hero__aside-label">最近更新时间</div>
              <div class="ticket-hero__aside-value">{{ formatDateTime(detail.last_message_at) }}</div>
              <div class="ticket-hero__aside-note">创建于 {{ formatDateTime(detail.created_at) }}</div>
            </div>
          </section>

          <section class="ticket-detail-grid">
            <NCard :bordered="false" class="detail-panel detail-panel--summary">
              <div class="panel-head">
                <div class="panel-head__title">
                  <NIcon :size="16">
                    <PersonOutline />
                  </NIcon>
                  <span>工单信息</span>
                </div>
              </div>
              <NDescriptions :column="1" label-placement="left" size="small">
                <NDescriptionsItem label="工单号">{{ detail.ticket_no }}</NDescriptionsItem>
                <NDescriptionsItem label="用户">{{ detail.user.username }}</NDescriptionsItem>
                <NDescriptionsItem label="邮箱">{{ detail.user.email }}</NDescriptionsItem>
                <NDescriptionsItem label="订单号">{{ detailOrderLabel }}</NDescriptionsItem>
                <NDescriptionsItem label="创建时间">{{ formatDateTime(detail.created_at) }}</NDescriptionsItem>
                <NDescriptionsItem label="关闭时间">{{ detailClosedAtLabel }}</NDescriptionsItem>
                <NDescriptionsItem label="关闭原因">{{ detail.close_reason || '-' }}</NDescriptionsItem>
              </NDescriptions>
            </NCard>

            <NCard :bordered="false" class="detail-panel detail-panel--content">
              <div class="panel-head">
                <div class="panel-head__title">
                  <NIcon :size="16">
                    <DocumentTextOutline />
                  </NIcon>
                  <span>消息时间线</span>
                </div>
              </div>

              <NTimeline v-if="detail.messages.length" class="ticket-timeline">
                <NTimelineItem v-for="item in detail.messages" :key="item.id" :type="item.sender_type === 'admin' ? 'info' : 'default'">
                  <template #icon>
                    <NIcon :size="14">
                      <component :is="item.sender_type === 'admin' ? CheckmarkDoneOutline : PersonOutline" />
                    </NIcon>
                  </template>
                  <div class="ticket-message">
                    <div class="ticket-message__meta">
                      <div class="ticket-message__sender">{{ item.sender_name }}</div>
                      <div class="ticket-message__time">{{ formatDateTime(item.created_at) }}</div>
                    </div>
                    <div class="ticket-message__content">{{ item.content }}</div>
                    <div v-if="(item.attachments ?? []).length" class="ticket-message__attachments">
                      <div v-for="(file, index) in (item.attachments ?? [])" :key="file?.file_id || index" class="ticket-attachment">
                        <button class="ticket-attachment__preview" type="button" @click="file && openAttachment(file)">
                          <img v-if="isImageAttachment(file) && file && attachmentPreviews[file.file_id]" :src="attachmentPreviews[file.file_id]" :alt="file.original_name" />
                          <span v-else>{{ attachmentType(file) }}</span>
                        </button>
                        <div class="ticket-attachment__body">
                          <div class="ticket-attachment__name">{{ file?.original_name || '未命名附件' }}</div>
                          <div class="ticket-attachment__meta">{{ attachmentLabel(file) }}</div>
                          <div class="ticket-attachment__actions">
                            <NButton size="tiny" secondary :disabled="!file" @click="file && openAttachment(file)">
                              <template #icon>
                                <NIcon :size="14">
                                  <DocumentTextOutline />
                                </NIcon>
                              </template>
                              预览
                            </NButton>
                            <NButton size="tiny" secondary :disabled="!file" @click="file && saveAttachment(file)">
                              <template #icon>
                                <NIcon :size="14">
                                  <DownloadOutline />
                                </NIcon>
                              </template>
                              下载
                            </NButton>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </NTimelineItem>
              </NTimeline>
              <NEmpty v-else description="暂无消息记录" />
            </NCard>
          </section>

          <section class="ticket-actions">
            <NCard v-if="!selectedClosed && canReply" :bordered="false" class="detail-panel">
              <div class="panel-head">
                <div class="panel-head__title">
                  <NIcon :size="16">
                    <ChatbubbleEllipsesOutline />
                  </NIcon>
                  <span>回复工单</span>
                </div>
                <div class="panel-head__hint">回复后会同步回工单消息时间线</div>
              </div>
              <NInput v-model:value="replyContent" type="textarea" :rows="5" placeholder="输入回复内容" />
              <div class="reply-tools">
                <NUpload v-model:file-list="uploadFiles" multiple :max="5" :default-upload="false">
                  <NButton secondary>
                    <template #icon>
                      <NIcon :size="14">
                        <AttachOutline />
                      </NIcon>
                    </template>
                    选择附件
                  </NButton>
                </NUpload>
                <div class="reply-tools__meta">最多 5 个附件</div>
              </div>
              <div class="action-row">
                <NButton type="primary" :loading="submitting" :disabled="replyDisabled" @click="submitReply">
                  <template #icon>
                    <NIcon :size="14">
                      <ChatbubbleEllipsesOutline />
                    </NIcon>
                  </template>
                  发送回复
                </NButton>
              </div>
            </NCard>

            <NCard v-if="!selectedClosed && canClose" :bordered="false" class="detail-panel detail-panel--warning">
              <div class="panel-head">
                <div class="panel-head__title">
                  <NIcon :size="16">
                    <CloseCircleOutline />
                  </NIcon>
                  <span>关闭工单</span>
                </div>
                <div class="panel-head__hint">关闭后不可继续回复</div>
              </div>
              <NInput v-model:value="closeReason" type="textarea" :rows="4" placeholder="关闭原因（可选）" />
              <div class="action-row action-row--end">
                <NButton type="warning" :loading="submitting" :disabled="closeDisabled" @click="submitClose">
                  <template #icon>
                    <NIcon :size="14">
                      <CheckmarkDoneOutline />
                    </NIcon>
                  </template>
                  关闭工单
                </NButton>
              </div>
            </NCard>
          </section>
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.page-header h2 {
  margin: 0;
  color: #0f172a;
  font-size: 20px;
}

.page-header p {
  margin: 4px 0 0;
}

.muted {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}

.query-form {
  margin-bottom: 16px;
}

.strong {
  font-weight: 700;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

:global(.n-drawer-container .n-drawer-body-content-wrapper) {
  background: #f8fafc;
}

.ticket-detail-shell {
  display: grid;
  gap: 16px;
  max-height: calc(100vh - 112px);
  padding-right: 8px;
}

.ticket-detail-skeleton {
  display: grid;
  gap: 16px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.skeleton-card {
  height: 140px;
  border-radius: 12px;
  background: linear-gradient(90deg, #e5e7eb 0%, #f1f5f9 50%, #e5e7eb 100%);
  background-size: 200% 100%;
  animation: skeleton-shift 1.2s ease-in-out infinite;
}

.skeleton-card--lg {
  height: 188px;
}

@keyframes skeleton-shift {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.ticket-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 240px;
  gap: 16px;
  align-items: stretch;
}

.ticket-hero__main,
.ticket-hero__aside,
.detail-panel {
  border-radius: 12px;
}

.ticket-hero__main {
  padding: 20px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  border: 1px solid #e2e8f0;
}

.ticket-hero__title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.ticket-hero__eyebrow {
  color: rgba(15, 23, 42, 0.5);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0;
}

.ticket-hero h3 {
  margin: 6px 0 0;
  color: #0f172a;
  font-size: 22px;
  line-height: 1.3;
}

.ticket-hero__meta {
  margin-top: 6px;
  color: rgba(15, 23, 42, 0.62);
  font-size: 13px;
}

.ticket-hero__stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 20px;
}

.stat-card {
  min-width: 0;
  padding: 12px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #fff;
}

.stat-card__label {
  color: rgba(15, 23, 42, 0.52);
  font-size: 12px;
  font-weight: 700;
}

.stat-card__value {
  margin-top: 6px;
  overflow: hidden;
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ticket-hero__aside {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  padding: 20px;
  border: 1px solid #e2e8f0;
  background: #fff;
}

.ticket-hero__aside-label {
  color: rgba(15, 23, 42, 0.5);
  font-size: 12px;
  font-weight: 700;
}

.ticket-hero__aside-value {
  color: #0f172a;
  font-size: 16px;
  font-weight: 700;
}

.ticket-hero__aside-note {
  color: rgba(15, 23, 42, 0.58);
  font-size: 12px;
}

.ticket-detail-grid {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.ticket-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.detail-panel {
  padding: 16px;
  border: 1px solid #e2e8f0;
  background: #fff;
}

.detail-panel--summary {
  position: sticky;
  top: 0;
}

.detail-panel--content {
  min-width: 0;
}

.detail-panel--warning {
  border-color: #fde68a;
  background: #fffbeb;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.panel-head__title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #0f172a;
  font-size: 14px;
  font-weight: 800;
}

.panel-head__hint {
  color: rgba(15, 23, 42, 0.52);
  font-size: 12px;
}

.ticket-timeline {
  padding-top: 4px;
}

.ticket-message {
  padding: 12px 0 2px;
}

.ticket-message__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: rgba(15, 23, 42, 0.58);
  font-size: 12px;
}

.ticket-message__sender {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #0f172a;
  font-weight: 700;
}

.ticket-message__time {
  white-space: nowrap;
}

.ticket-message__content {
  margin-top: 10px;
  padding: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #fff;
  white-space: pre-wrap;
  color: #0f172a;
  line-height: 1.7;
}

.ticket-message--admin .ticket-message__content {
  background: #f8fafc;
}

.ticket-message__attachments {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.ticket-attachment {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr);
  gap: 10px;
  padding: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #fff;
}

.ticket-attachment__preview {
  width: 56px;
  height: 56px;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
  color: rgba(15, 23, 42, 0.58);
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
}

.ticket-attachment__preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ticket-attachment__body {
  min-width: 0;
}

.ticket-attachment__name {
  overflow: hidden;
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ticket-attachment__meta {
  margin-top: 4px;
  overflow: hidden;
  color: rgba(15, 23, 42, 0.58);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ticket-attachment__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.reply-tools {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 12px;
}

.reply-tools__meta {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}

.action-row {
  display: flex;
  justify-content: flex-start;
  margin-top: 14px;
}

.action-row--end {
  justify-content: flex-end;
}

@media (max-width: 1280px) {
  .ticket-hero {
    grid-template-columns: 1fr;
  }

  .ticket-detail-grid {
    grid-template-columns: 1fr;
  }

  .detail-panel--summary {
    position: static;
  }
}

@media (max-width: 960px) {
  .ticket-hero__stats,
  .ticket-actions,
  .skeleton-grid {
    grid-template-columns: 1fr 1fr;
  }

  .ticket-hero__title-row,
  .panel-head,
  .reply-tools {
    align-items: stretch;
    flex-direction: column;
  }
}

@media (max-width: 640px) {
  .ticket-hero__stats,
  .ticket-actions,
  .skeleton-grid {
    grid-template-columns: 1fr;
  }

  .ticket-message__meta {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
