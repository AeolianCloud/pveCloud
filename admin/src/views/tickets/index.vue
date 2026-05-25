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
  PeopleOutline,
} from '@vicons/ionicons5'
import {
  NButton,
  NCard,
  NDataTable,
  NDivider,
  NForm,
  NFormItem,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NInput,
  NInputNumber,
  NModal,
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
import { useRouter } from 'vue-router'

import {
  addTicketCollaborator,
  addTicketInternalNote,
  assignTicket,
  closeTicket,
  createTicketTag,
  downloadTicketAttachment,
  getAssigneeCandidates,
  getTicketDetail,
  getTicketTags,
  getTickets,
  removeTicketCollaborator,
  replaceTicketTags,
  replyTicket,
  updateTicketTag,
  upgradeTicketPriority,
  type AdminTicketAttachment,
  type AdminTicketDetail,
  type AdminTicketItem,
  type TicketAdminSummary,
  type TicketPriority,
  type TicketTagItem,
  type TicketTagStatus,
  type TicketTagVisibility,
} from '../../api/ticket'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'
import { message } from '../../utils/feedback'
import { hasPermissionCode } from '../../utils/permission'

const permissionStore = usePermissionStore()
const router = useRouter()
const loading = ref(false)
const detailLoading = ref(false)
const detailVisible = ref(false)
const tagManagerVisible = ref(false)
const submitting = ref(false)
const tagSubmitting = ref(false)
const tickets = ref<AdminTicketItem[]>([])
const detail = ref<AdminTicketDetail | null>(null)
const assigneeOptions = ref<Array<{ label: string; value: number }>>([])
const tagOptions = ref<Array<{ label: string; value: number; item: TicketTagItem }>>([])
const tagList = ref<TicketTagItem[]>([])
const total = ref(0)
const replyContent = ref('')
const uploadFiles = ref<UploadFileInfo[]>([])
const closeReason = ref('')
const assignForm = reactive<{ assignee_admin_id: number | null; reason: string }>({ assignee_admin_id: null, reason: '' })
const collaboratorAdminId = ref<number | null>(null)
const internalNoteContent = ref('')
const priorityForm = reactive<{ priority: TicketPriority | ''; reason: string }>({ priority: '', reason: '' })
const selectedTagIds = ref<number[]>([])
const tagForm = reactive<{ id: number | null; name: string; color: string; visibility: TicketTagVisibility; status: TicketTagStatus; sort_order: number }>({ id: null, name: '', color: '', visibility: 'internal', status: 'active', sort_order: 0 })
const query = reactive({ page: 1, per_page: 15, status: '', category: '', priority: '', ticket_no: '', order_no: '', instance_no: '', user_keyword: '', assignee_admin_id: null as number | null, tag_id: null as number | null, sla_status: '' })
const attachmentPreviews = ref<Record<number, string>>({})
const attachmentObjectUrls = ref<string[]>([])

const canReply = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:reply'))
const canClose = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:close'))
const canAssign = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:assign'))
const canCollaborate = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:collaborate'))
const canNote = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:note'))
const canPriority = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:priority'))
const canTag = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:tag'))
const canTagManage = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'ticket:tag-manage'))
const canViewInstances = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'page.instances'))
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
const detailInstanceLabel = computed(() => detail.value?.instance_no || '-')
const detailClosedAtLabel = computed(() => formatDateTime(detail.value?.closed_at))
const detailAssigneeLabel = computed(() => adminLabel(detail.value?.assignee))
const detailSlaLabel = computed(() => detail.value ? slaStatusText[detail.value.sla.status] || detail.value.sla.status : '-')

const statusText: Record<string, string> = { waiting_admin: '待后台处理', waiting_user: '待用户反馈', closed: '已关闭' }
const categoryText: Record<string, string> = { account: '账号问题', order: '订单问题', product: '产品咨询', technical: '技术支持', billing: '账务问题', other: '其它问题' }
const priorityText: Record<string, string> = { low: '低', normal: '普通', high: '高', urgent: '紧急' }
const slaStatusText: Record<string, string> = { normal: '正常', first_response_overdue: '首响逾期', resolution_overdue: '解决逾期' }
const eventText: Record<string, string> = { assign: '指派', transfer: '转派', collaborator_add: '添加协作者', collaborator_remove: '移除协作者', internal_note: '内部备注', priority_upgrade: '优先级升级', tags_replace: '更新标签', admin_reply: '回复', admin_close: '关闭' }

const statusOptions = [
  { label: '待后台处理', value: 'waiting_admin' },
  { label: '待用户反馈', value: 'waiting_user' },
  { label: '已关闭', value: 'closed' },
]
const categoryOptions = Object.entries(categoryText).map(([value, label]) => ({ value, label }))
const priorityOptions = Object.entries(priorityText).map(([value, label]) => ({ value, label }))
const slaStatusOptions = Object.entries(slaStatusText).map(([value, label]) => ({ value, label }))
const tagVisibilityOptions = [{ label: '公开', value: 'public' }, { label: '内部', value: 'internal' }]
const tagStatusOptions = [{ label: '启用', value: 'active' }, { label: '停用', value: 'disabled' }]
const defaultTagColor = '#2563eb'
const tagColorPresets = [
  { label: '蓝色', value: '#2563eb' },
  { label: '绿色', value: '#16a34a' },
  { label: '红色', value: '#dc2626' },
  { label: '橙色', value: '#ea580c' },
  { label: '紫色', value: '#9333ea' },
  { label: '青色', value: '#0891b2' },
  { label: '灰色', value: '#64748b' },
  { label: '黑色', value: '#0f172a' },
]
const activeTagColor = computed(() => normalizeHexColor(tagForm.color))
const tagColorPreview = computed(() => activeTagColor.value || defaultTagColor)

async function loadTickets() {
  loading.value = true
  try {
    const params = { ...query, assignee_admin_id: query.assignee_admin_id || undefined, tag_id: query.tag_id || undefined }
    const data = await getTickets(params)
    tickets.value = data.list
    total.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '工单加载失败')
  } finally {
    loading.value = false
  }
}

async function loadAssignees() {
  if (!canAssign.value && !canCollaborate.value) return
  try {
    const data = await getAssigneeCandidates({ page: 1, per_page: 100 })
    assigneeOptions.value = data.list.map((item) => ({ label: adminLabel(item), value: item.id }))
  } catch {
    assigneeOptions.value = []
  }
}

async function loadTags() {
  try {
    const data = await getTicketTags({ page: 1, per_page: 100 })
    tagList.value = data.list
    tagOptions.value = data.list.filter((item) => item.status === 'active').map((item) => ({ label: `${item.name}${item.visibility === 'public' ? ' · 公开' : ' · 内部'}`, value: item.id, item }))
  } catch (err) {
    tagList.value = []
    tagOptions.value = []
    message.error(err instanceof Error ? err.message : '标签加载失败')
  }
}

async function openDetail(ticketNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    cleanupAttachmentPreviews()
    detail.value = await getTicketDetail(ticketNo)
    assignForm.assignee_admin_id = detail.value.assignee?.id ?? null
    selectedTagIds.value = detail.value.tags.map((tag) => tag.id)
    priorityForm.priority = ''
    priorityForm.reason = ''
    internalNoteContent.value = ''
    collaboratorAdminId.value = null
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
  Object.assign(query, { page: 1, per_page: 15, status: '', category: '', priority: '', ticket_no: '', order_no: '', instance_no: '', user_keyword: '', assignee_admin_id: null, tag_id: null, sla_status: '' })
  loadTickets()
}

async function submitAssign() {
  if (!detail.value || !assignForm.assignee_admin_id) return
  submitting.value = true
  try {
    detail.value = await assignTicket(detail.value.ticket_no, assignForm.assignee_admin_id, assignForm.reason || undefined)
    assignForm.reason = ''
    message.success('处理人已更新')
    await loadTickets()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '指派失败')
  } finally {
    submitting.value = false
  }
}

async function submitCollaborator() {
  if (!detail.value || !collaboratorAdminId.value) return
  submitting.value = true
  try {
    detail.value = await addTicketCollaborator(detail.value.ticket_no, collaboratorAdminId.value)
    collaboratorAdminId.value = null
    message.success('协作者已添加')
  } catch (err) {
    message.error(err instanceof Error ? err.message : '添加协作者失败')
  } finally {
    submitting.value = false
  }
}

async function submitRemoveCollaborator(adminId: number) {
  if (!detail.value) return
  submitting.value = true
  try {
    detail.value = await removeTicketCollaborator(detail.value.ticket_no, adminId)
    message.success('协作者已移除')
  } catch (err) {
    message.error(err instanceof Error ? err.message : '移除协作者失败')
  } finally {
    submitting.value = false
  }
}

async function submitInternalNote() {
  if (!detail.value || !internalNoteContent.value.trim()) return
  submitting.value = true
  try {
    detail.value = await addTicketInternalNote(detail.value.ticket_no, internalNoteContent.value.trim())
    internalNoteContent.value = ''
    message.success('内部备注已追加')
  } catch (err) {
    message.error(err instanceof Error ? err.message : '追加内部备注失败')
  } finally {
    submitting.value = false
  }
}

async function submitPriorityUpgrade() {
  if (!detail.value || !priorityForm.priority || !priorityForm.reason.trim()) return
  submitting.value = true
  try {
    detail.value = await upgradeTicketPriority(detail.value.ticket_no, priorityForm.priority, priorityForm.reason.trim())
    priorityForm.priority = ''
    priorityForm.reason = ''
    message.success('优先级已升级')
    await loadTickets()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '优先级升级失败')
  } finally {
    submitting.value = false
  }
}

async function submitTags() {
  if (!detail.value) return
  submitting.value = true
  try {
    detail.value = await replaceTicketTags(detail.value.ticket_no, selectedTagIds.value)
    message.success('标签已更新')
    await loadTickets()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '标签更新失败')
  } finally {
    submitting.value = false
  }
}

function resetTagForm() {
  Object.assign(tagForm, { id: null, name: '', color: '', visibility: 'internal', status: 'active', sort_order: 0 })
}

function editTag(tag: TicketTagItem) {
  Object.assign(tagForm, { id: tag.id, name: tag.name, color: tag.color || '', visibility: tag.visibility, status: tag.status, sort_order: tag.sort_order })
}

function normalizeHexColor(color: string) {
  const value = color.trim()
  return /^#[0-9a-fA-F]{6}$/.test(value) ? value : ''
}

function setTagColor(color: string) {
  tagForm.color = color
}

function handleTagColorInput(event: Event) {
  tagForm.color = (event.target as HTMLInputElement).value
}

async function submitTagForm() {
  if (!tagForm.name.trim()) return
  tagSubmitting.value = true
  try {
    const payload = { name: tagForm.name.trim(), color: tagForm.color || null, visibility: tagForm.visibility, status: tagForm.status, sort_order: tagForm.sort_order }
    let saved: TicketTagItem
    if (tagForm.id) {
      saved = await updateTicketTag(tagForm.id, payload)
      message.success('标签已更新')
    } else {
      saved = await createTicketTag(payload)
      if (detail.value && saved.status === 'active' && !selectedTagIds.value.includes(saved.id)) {
        selectedTagIds.value = [...selectedTagIds.value, saved.id]
      }
      message.success(detail.value ? '标签已创建并选中，请保存标签' : '标签已创建')
    }
    resetTagForm()
    await loadTags()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '标签保存失败')
  } finally {
    tagSubmitting.value = false
  }
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
  assignForm.assignee_admin_id = null
  assignForm.reason = ''
  collaboratorAdminId.value = null
  internalNoteContent.value = ''
  priorityForm.priority = ''
  priorityForm.reason = ''
  selectedTagIds.value = []
  cleanupAttachmentPreviews()
}

function handleDetailVisibleChange(show: boolean) {
  detailVisible.value = show
  if (!show) {
    resetDetailState()
  }
}

function adminLabel(admin: TicketAdminSummary | null | undefined) {
  if (!admin) return '未指派'
  return admin.display_name || admin.username || `#${admin.id}`
}

function tagStyle(tag: TicketTagItem) {
  return tag.color ? { color: tag.color, borderColor: tag.color } : undefined
}

function ticketLinkSummary(row: AdminTicketItem) {
  const links = []
  links.push(row.order_no ? `订单 ${row.order_no}` : '未关联订单')
  links.push(row.instance_no ? `实例 ${row.instance_no}` : '未关联实例')
  return links.join(' · ')
}

function openInstanceFromTicket(instanceNo: string) {
  void router.push({ path: '/instances', query: { instance_no: instanceNo } })
}

function priorityRank(priority: string) {
  return { low: 1, normal: 2, high: 3, urgent: 4 }[priority] || 0
}

function upgradePriorityOptions(current: string) {
  return priorityOptions.filter((item) => priorityRank(item.value) > priorityRank(current))
}

const columns = computed<DataTableColumns<AdminTicketItem>>(() => [
  { key: 'ticket_no', title: '工单号', minWidth: 170 },
  {
    key: 'title',
    title: '标题',
    minWidth: 220,
    render: (row) => h('div', null, [h('div', { class: 'strong' }, row.title), h('div', { class: 'muted' }, ticketLinkSummary(row))]),
  },
  {
    key: 'user',
    title: '用户',
    minWidth: 160,
    render: (row) => h('div', null, [h('div', { class: 'strong' }, row.user.username), h('div', { class: 'muted' }, row.user.email)]),
  },
  { key: 'category', title: '分类', width: 110, render: (row) => categoryText[row.category] || row.category },
  { key: 'priority', title: '优先级', width: 90, render: (row) => priorityText[row.priority] || row.priority },
  { key: 'assignee', title: '处理人', width: 120, render: (row) => adminLabel(row.assignee) },
  { key: 'tags', title: '标签', minWidth: 150, render: (row) => h('div', { class: 'ticket-tags' }, row.tags?.length ? row.tags.map((tag) => h(NTag, { size: 'small', bordered: true, style: tagStyle(tag) }, { default: () => tag.name })) : h('span', { class: 'muted' }, '-')) },
  { key: 'sla', title: 'SLA', width: 110, render: (row) => h(NTag, { size: 'small', type: row.sla.status === 'normal' ? 'success' : 'error' }, { default: () => slaStatusText[row.sla.status] || row.sla.status }) },
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

onMounted(() => {
  loadTickets()
  loadAssignees()
  loadTags()
})
onBeforeUnmount(cleanupAttachmentPreviews)
</script>

<template>
  <div class="tickets-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>工单管理</h2>
          <p class="muted">查看用户提交的工单，进入详情后处理回复、协作、标签和内部备注。</p>
        </div>
      </template>

      <NForm inline label-placement="left" class="query-form">
        <NFormItem label="状态"><NSelect v-model:value="query.status" :options="statusOptions" clearable placeholder="全部" style="width: 140px" /></NFormItem>
        <NFormItem label="分类"><NSelect v-model:value="query.category" :options="categoryOptions" clearable placeholder="全部" style="width: 130px" /></NFormItem>
        <NFormItem label="优先级"><NSelect v-model:value="query.priority" :options="priorityOptions" clearable placeholder="全部" style="width: 110px" /></NFormItem>
        <NFormItem label="处理人"><NSelect v-model:value="query.assignee_admin_id" :options="assigneeOptions" clearable filterable placeholder="全部" style="width: 140px" /></NFormItem>
        <NFormItem label="标签"><NSelect v-model:value="query.tag_id" :options="tagOptions" clearable filterable placeholder="全部" style="width: 150px" /></NFormItem>
        <NFormItem label="SLA"><NSelect v-model:value="query.sla_status" :options="slaStatusOptions" clearable placeholder="全部" style="width: 130px" /></NFormItem>
        <NFormItem label="工单号"><NInput v-model:value="query.ticket_no" clearable placeholder="TIC-" /></NFormItem>
        <NFormItem label="订单号"><NInput v-model:value="query.order_no" clearable placeholder="ORD-" /></NFormItem>
        <NFormItem label="实例号"><NInput v-model:value="query.instance_no" clearable placeholder="INS-" /></NFormItem>
        <NFormItem label="用户"><NInput v-model:value="query.user_keyword" clearable placeholder="用户名/邮箱" /></NFormItem>
        <NFormItem :show-label="false">
          <NSpace><NButton type="primary" @click="query.page = 1; loadTickets()">查询</NButton><NButton @click="resetQuery">重置</NButton></NSpace>
        </NFormItem>
        <NFormItem v-if="canTagManage" :show-label="false">
          <NButton secondary @click="tagManagerVisible = true">标签字典</NButton>
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
                <div class="stat-card">
                  <div class="stat-card__label">处理人</div>
                  <div class="stat-card__value">{{ detailAssigneeLabel }}</div>
                </div>
                <div class="stat-card">
                  <div class="stat-card__label">SLA</div>
                  <div class="stat-card__value">{{ detailSlaLabel }}</div>
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
                <NDescriptionsItem label="实例编号">
                  <NButton v-if="detail.instance_no && canViewInstances" text type="primary" @click="openInstanceFromTicket(detail.instance_no)">{{ detail.instance_no }}</NButton>
                  <span v-else>{{ detailInstanceLabel }}</span>
                </NDescriptionsItem>
                <NDescriptionsItem label="处理人">{{ detailAssigneeLabel }}</NDescriptionsItem>
                <NDescriptionsItem label="首响截止">{{ formatDateTime(detail.sla.first_response_due_at) }}</NDescriptionsItem>
                <NDescriptionsItem label="解决截止">{{ formatDateTime(detail.sla.resolution_due_at) }}</NDescriptionsItem>
                <NDescriptionsItem label="创建时间">{{ formatDateTime(detail.created_at) }}</NDescriptionsItem>
                <NDescriptionsItem label="关闭时间">{{ detailClosedAtLabel }}</NDescriptionsItem>
                <NDescriptionsItem label="关闭原因">{{ detail.close_reason || '-' }}</NDescriptionsItem>
              </NDescriptions>
              <div class="tag-list">
                <NTag v-for="tag in detail.tags" :key="tag.id" size="small" :style="tagStyle(tag)">{{ tag.name }}</NTag>
                <span v-if="!detail.tags.length" class="muted">暂无标签</span>
              </div>
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
            <NCard v-if="canAssign || canCollaborate || canTag || canPriority || canNote" :bordered="false" class="detail-panel detail-panel--ops">
              <div class="panel-head">
                <div class="panel-head__title">
                  <NIcon :size="16">
                    <PeopleOutline />
                  </NIcon>
                  <span>内部协作</span>
                </div>
              </div>

              <div class="ops-grid">
                <div v-if="canAssign && !selectedClosed" class="ops-block">
                  <div class="ops-block__title">处理人</div>
                  <NSelect v-model:value="assignForm.assignee_admin_id" :options="assigneeOptions" clearable filterable placeholder="选择处理人" />
                  <NInput v-model:value="assignForm.reason" class="mt-8" placeholder="指派/转派原因（可选）" />
                  <NButton class="mt-8" size="small" type="primary" :loading="submitting" :disabled="!assignForm.assignee_admin_id" @click="submitAssign">保存处理人</NButton>
                </div>

                <div v-if="canTag" class="ops-block">
                  <div class="ops-block__title">标签</div>
                  <NSelect v-model:value="selectedTagIds" :options="tagOptions" multiple clearable filterable placeholder="选择标签" />
                  <NSpace class="mt-8">
                    <NButton size="small" type="primary" :loading="submitting" @click="submitTags">保存标签</NButton>
                    <NButton v-if="canTagManage" size="small" secondary @click="tagManagerVisible = true">新建标签</NButton>
                  </NSpace>
                </div>

                <div v-if="canPriority && !selectedClosed" class="ops-block">
                  <div class="ops-block__title">优先级升级</div>
                  <NSelect v-model:value="priorityForm.priority" :options="upgradePriorityOptions(detail.priority)" clearable placeholder="选择更高优先级" />
                  <NInput v-model:value="priorityForm.reason" class="mt-8" placeholder="升级原因" />
                  <NButton class="mt-8" size="small" type="warning" :loading="submitting" :disabled="!priorityForm.priority || !priorityForm.reason.trim()" @click="submitPriorityUpgrade">升级优先级</NButton>
                </div>

                <div v-if="canCollaborate && !selectedClosed" class="ops-block">
                  <div class="ops-block__title">协作者</div>
                  <NSelect v-model:value="collaboratorAdminId" :options="assigneeOptions" clearable filterable placeholder="选择协作者" />
                  <NButton class="mt-8" size="small" type="primary" :loading="submitting" :disabled="!collaboratorAdminId" @click="submitCollaborator">添加协作者</NButton>
                  <div class="tag-list">
                    <NTag v-for="item in detail.collaborators" :key="item.id" closable @close="submitRemoveCollaborator(item.id)">{{ adminLabel(item) }}</NTag>
                    <span v-if="!detail.collaborators.length" class="muted">暂无协作者</span>
                  </div>
                </div>
              </div>

              <NDivider v-if="canNote" />
              <div v-if="canNote" class="ops-block">
                <div class="ops-block__title">内部备注</div>
                <NInput v-model:value="internalNoteContent" type="textarea" :rows="3" placeholder="仅管理端可见" />
                <NButton class="mt-8" size="small" type="primary" :loading="submitting" :disabled="!internalNoteContent.trim()" @click="submitInternalNote">追加备注</NButton>
                <div class="note-list">
                  <div v-for="note in detail.internal_notes" :key="note.id" class="note-item">
                    <div class="note-item__meta">{{ adminLabel(note.admin) }} · {{ formatDateTime(note.created_at) }}</div>
                    <div class="note-item__body">{{ note.content }}</div>
                  </div>
                  <span v-if="!detail.internal_notes.length" class="muted">暂无内部备注</span>
                </div>
              </div>

              <NDivider />
              <div class="ops-block">
                <div class="ops-block__title">操作历史</div>
                <div class="note-list">
                  <div v-for="event in detail.events" :key="event.id" class="note-item">
                    <div class="note-item__meta">{{ eventText[event.event_type] || event.event_type }} · {{ event.actor?.username || '-' }} · {{ formatDateTime(event.created_at) }}</div>
                    <div v-if="event.remark" class="note-item__body">{{ event.remark }}</div>
                  </div>
                  <span v-if="!detail.events.length" class="muted">暂无操作历史</span>
                </div>
              </div>
            </NCard>

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

    <NModal v-model:show="tagManagerVisible" preset="card" title="工单标签字典" style="width: min(760px, 92vw)" @after-enter="loadTags">
      <div class="tag-manager">
        <NForm label-placement="top" class="tag-form">
          <div class="tag-form__main">
            <NFormItem label="名称"><NInput v-model:value="tagForm.name" placeholder="例如：售后跟进" clearable /></NFormItem>
            <NFormItem label="颜色">
              <div class="tag-color-field">
                <div class="tag-color-picker">
                  <span class="tag-color-preview" :style="{ backgroundColor: tagColorPreview }"></span>
                  <input class="tag-color-input" type="color" :value="tagColorPreview" aria-label="选择标签颜色" @input="handleTagColorInput" />
                  <NInput v-model:value="tagForm.color" placeholder="#2563eb" clearable />
                </div>
                <div class="tag-color-swatches" aria-label="常用标签颜色">
                  <button
                    v-for="preset in tagColorPresets"
                    :key="preset.value"
                    class="tag-color-swatch"
                    :class="{ 'tag-color-swatch--active': activeTagColor === preset.value }"
                    :style="{ backgroundColor: preset.value }"
                    type="button"
                    :title="preset.label"
                    :aria-label="preset.label"
                    @click="setTagColor(preset.value)"
                  ></button>
                </div>
              </div>
            </NFormItem>
          </div>
          <div class="tag-form__meta">
            <NFormItem label="可见性"><NSelect v-model:value="tagForm.visibility" :options="tagVisibilityOptions" /></NFormItem>
            <NFormItem label="状态"><NSelect v-model:value="tagForm.status" :options="tagStatusOptions" /></NFormItem>
            <NFormItem label="排序"><NInputNumber v-model:value="tagForm.sort_order" :min="0" :max="9999" style="width: 100%" /></NFormItem>
          </div>
          <div class="tag-form__actions">
            <NButton type="primary" :loading="tagSubmitting" :disabled="!tagForm.name.trim()" @click="submitTagForm">{{ tagForm.id ? '保存标签' : '创建标签' }}</NButton>
            <NButton @click="resetTagForm">清空</NButton>
          </div>
        </NForm>
        <div class="tag-table">
          <div v-for="tag in tagList" :key="tag.id" class="tag-row">
            <div class="tag-row__name"><NTag :style="tagStyle(tag)">{{ tag.name }}</NTag></div>
            <span class="tag-row__cell">{{ tag.visibility === 'public' ? '公开' : '内部' }}</span>
            <span class="tag-row__cell">{{ tag.status === 'active' ? '启用' : '停用' }}</span>
            <span class="tag-row__cell">排序 {{ tag.sort_order }}</span>
            <NButton size="tiny" secondary @click="editTag(tag)">编辑</NButton>
          </div>
        </div>
      </div>
    </NModal>
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
  grid-template-columns: repeat(6, minmax(0, 1fr));
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
  grid-template-columns: minmax(0, 1fr);
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

.detail-panel--ops {
  border-color: #bfdbfe;
}

.ops-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.ops-block {
  min-width: 0;
}

.ops-block__title {
  margin-bottom: 8px;
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
}

.mt-8 {
  margin-top: 8px;
}

.ticket-tags,
.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.note-list {
  display: grid;
  gap: 8px;
  margin-top: 10px;
}

.note-item {
  padding: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
}

.note-item__meta {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
  font-weight: 700;
}

.note-item__body {
  margin-top: 4px;
  color: #0f172a;
  font-size: 13px;
  white-space: pre-wrap;
}

.tag-manager {
  display: grid;
  gap: 16px;
}

.tag-form {
  display: grid;
  gap: 12px;
  padding: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
}

.tag-form :deep(.n-form-item) {
  margin-bottom: 0;
}

.tag-form__main {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
  gap: 12px;
}

.tag-color-field {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.tag-color-picker {
  display: grid;
  grid-template-columns: 32px 42px minmax(0, 1fr);
  gap: 8px;
  align-items: center;
}

.tag-color-preview {
  width: 32px;
  height: 32px;
  border: 1px solid rgba(15, 23, 42, 0.14);
  border-radius: 8px;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.42);
}

.tag-color-input {
  width: 42px;
  height: 32px;
  padding: 0;
  overflow: hidden;
  border: 1px solid #dbe3ef;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
}

.tag-color-input::-webkit-color-swatch-wrapper {
  padding: 0;
}

.tag-color-input::-webkit-color-swatch {
  border: 0;
}

.tag-color-input::-moz-color-swatch {
  border: 0;
}

.tag-color-swatches {
  display: grid;
  grid-template-columns: repeat(8, 24px);
  gap: 8px;
}

.tag-color-swatch {
  width: 24px;
  height: 24px;
  border: 2px solid #fff;
  border-radius: 999px;
  box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.16);
  cursor: pointer;
}

.tag-color-swatch:hover,
.tag-color-swatch:focus-visible {
  box-shadow: 0 0 0 2px #fff, 0 0 0 4px rgba(37, 99, 235, 0.36);
  outline: none;
}

.tag-color-swatch--active {
  box-shadow: 0 0 0 2px #fff, 0 0 0 4px rgba(15, 23, 42, 0.32);
}

.tag-form__meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.tag-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.tag-table {
  display: grid;
  gap: 8px;
  max-height: 360px;
  overflow: auto;
}

.tag-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 72px 72px 88px 64px;
  gap: 12px;
  align-items: center;
  min-height: 48px;
  padding: 8px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #fff;
}

.tag-row__name {
  min-width: 0;
}

.tag-row__cell {
  color: rgba(15, 23, 42, 0.7);
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
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

  .tag-form__main,
  .tag-form__meta {
    grid-template-columns: 1fr;
  }

  .tag-row {
    grid-template-columns: minmax(0, 1fr) 64px;
  }

  .tag-row__cell {
    display: none;
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

  .tag-color-picker {
    grid-template-columns: 32px 42px minmax(0, 1fr);
  }

  .tag-color-swatches {
    grid-template-columns: repeat(4, 24px);
  }

  .ticket-message__meta {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
