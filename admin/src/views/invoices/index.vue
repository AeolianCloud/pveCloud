<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NPagination,
  NSelect,
  NSpace,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import { uploadFile } from '../../api/file-attachment'
import {
  acceptInvoice,
  downloadInvoice,
  getInvoiceDetail,
  getInvoices,
  issueInvoice,
  rejectInvoice,
  updateInvoiceAdminNote,
  type AdminInvoiceDetail,
  type AdminInvoiceItem,
  type InvoiceOrderItem,
} from '../../api/invoice'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'
import { confirm, message } from '../../utils/feedback'
import { hasPermissionCode } from '../../utils/permission'
import InvoiceStatusTag from './components/InvoiceStatusTag.vue'
import type { InvoiceQueryState, IssueFormState, RejectFormState } from './types'

const permissionStore = usePermissionStore()
const loading = ref(false)
const detailLoading = ref(false)
const submitting = ref(false)
const uploading = ref(false)
const downloadingInvoice = ref('')
const invoices = ref<AdminInvoiceItem[]>([])
const total = ref(0)
const dateRange = ref<[number, number] | null>(null)
const detailVisible = ref(false)
const selectedInvoice = ref<AdminInvoiceDetail | null>(null)
const rejectVisible = ref(false)
const issueVisible = ref(false)
const noteVisible = ref(false)
const pdfInput = ref<HTMLInputElement | null>(null)

const query = reactive<InvoiceQueryState>({
  page: 1,
  per_page: 15,
  status: '',
  invoice_no: '',
  order_no: '',
  user_keyword: '',
  title_keyword: '',
})

const rejectForm = reactive<RejectFormState>({ invoice_no: '', reason: '' })
const issueForm = reactive<IssueFormState>({ invoice_no: '', invoice_code: '', invoice_number: '', issued_at: Date.now(), file_id: null, file_name: '' })
const adminNote = ref('')

const canUpdate = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'invoice:update'))
const canReject = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'invoice:reject'))
const canIssue = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'invoice:issue'))

const statusOptions = [
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已开票', value: 'issued' },
  { label: '已驳回', value: 'rejected' },
  { label: '已取消', value: 'cancelled' },
]

const titleTypeText: Record<string, string> = { personal: '个人', company: '企业' }
const orderTypeText: Record<string, string> = { purchase: '新购', renewal: '续费' }

function formatMoney(cents: number, currency = 'CNY') {
  return `${currency} ${(cents / 100).toFixed(2)}`
}

function userLabel(row: AdminInvoiceItem) {
  return row.user.display_name || row.user.username || row.user.email || `用户 ${row.user.id}`
}

function queryParams() {
  const params: Record<string, unknown> = { ...query }
  if (dateRange.value) {
    params.date_from = new Date(dateRange.value[0]).toISOString()
    params.date_to = new Date(dateRange.value[1]).toISOString()
  }
  return params
}

async function loadInvoices() {
  loading.value = true
  try {
    const data = await getInvoices(queryParams())
    invoices.value = data.list
    total.value = data.total
  } catch (error) {
    message.error(error instanceof Error ? error.message : '发票列表加载失败')
  } finally {
    loading.value = false
  }
}

async function openDetail(invoiceNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    selectedInvoice.value = await getInvoiceDetail(invoiceNo)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '发票详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

async function refreshDetail(invoiceNo: string) {
  selectedInvoice.value = await getInvoiceDetail(invoiceNo)
}

function resetQuery() {
  Object.assign(query, { page: 1, per_page: 15, status: '', invoice_no: '', order_no: '', user_keyword: '', title_keyword: '' })
  dateRange.value = null
  void loadInvoices()
}

async function accept(row: AdminInvoiceItem | AdminInvoiceDetail) {
  try {
    await confirm({ title: '受理发票', content: `确认受理发票申请 ${row.invoice_no}？`, positiveText: '受理', negativeText: '取消' })
  } catch {
    return
  }
  submitting.value = true
  try {
    await acceptInvoice(row.invoice_no)
    message.success('已受理')
    await loadInvoices()
    if (selectedInvoice.value?.invoice_no === row.invoice_no) await refreshDetail(row.invoice_no)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '受理失败')
  } finally {
    submitting.value = false
  }
}

function openReject(row: AdminInvoiceItem | AdminInvoiceDetail) {
  rejectForm.invoice_no = row.invoice_no
  rejectForm.reason = ''
  rejectVisible.value = true
}

async function submitReject() {
  const reason = rejectForm.reason.trim()
  if (!reason) {
    message.warning('请填写驳回原因')
    return
  }
  submitting.value = true
  try {
    await rejectInvoice(rejectForm.invoice_no, reason)
    message.success('已驳回')
    rejectVisible.value = false
    await loadInvoices()
    if (selectedInvoice.value?.invoice_no === rejectForm.invoice_no) await refreshDetail(rejectForm.invoice_no)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '驳回失败')
  } finally {
    submitting.value = false
  }
}

function openIssue(row: AdminInvoiceItem | AdminInvoiceDetail) {
  Object.assign(issueForm, { invoice_no: row.invoice_no, invoice_code: '', invoice_number: '', issued_at: Date.now(), file_id: null, file_name: '' })
  issueVisible.value = true
}

function choosePDF() {
  pdfInput.value?.click()
}

async function handlePDFSelected(event: Event) {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (!file) return
  if (file.type && file.type !== 'application/pdf') {
    message.warning('请选择 PDF 文件')
    if (input) input.value = ''
    return
  }
  uploading.value = true
  try {
    const result = await uploadFile(file)
    issueForm.file_id = result.id
    issueForm.file_name = result.original_name
    message.success('PDF 已上传')
  } catch (error) {
    message.error(error instanceof Error ? error.message : 'PDF 上传失败')
  } finally {
    uploading.value = false
    if (input) input.value = ''
  }
}

async function submitIssue() {
  if (!issueForm.invoice_number.trim()) {
    message.warning('请填写发票号码')
    return
  }
  if (!issueForm.issued_at) {
    message.warning('请选择开票时间')
    return
  }
  if (!issueForm.file_id) {
    message.warning('请上传发票 PDF')
    return
  }
  submitting.value = true
  try {
    await issueInvoice(issueForm.invoice_no, {
      invoice_code: issueForm.invoice_code.trim() || null,
      invoice_number: issueForm.invoice_number.trim(),
      issued_at: new Date(issueForm.issued_at).toISOString(),
      file_id: issueForm.file_id,
    })
    message.success('已登记开票')
    issueVisible.value = false
    await loadInvoices()
    if (selectedInvoice.value?.invoice_no === issueForm.invoice_no) await refreshDetail(issueForm.invoice_no)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '登记开票失败')
  } finally {
    submitting.value = false
  }
}

function openNote(row: AdminInvoiceDetail) {
  adminNote.value = row.admin_note || ''
  noteVisible.value = true
}

async function submitNote() {
  if (!selectedInvoice.value) return
  const invoiceNo = selectedInvoice.value.invoice_no
  submitting.value = true
  try {
    await updateInvoiceAdminNote(invoiceNo, adminNote.value.trim() || null)
    message.success('后台备注已更新')
    noteVisible.value = false
    await refreshDetail(invoiceNo)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '备注更新失败')
  } finally {
    submitting.value = false
  }
}

async function download(invoiceNo: string) {
  downloadingInvoice.value = invoiceNo
  try {
    const blob = await downloadInvoice(invoiceNo)
    saveBlob(blob, `${invoiceNo}.pdf`)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '发票 PDF 下载失败')
  } finally {
    downloadingInvoice.value = ''
  }
}

function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function actions(row: AdminInvoiceItem | AdminInvoiceDetail) {
  const nodes = [h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row.invoice_no) }, { default: () => '详情' })]
  if (canUpdate.value && row.status === 'pending') nodes.push(h(NButton, { text: true, type: 'primary', onClick: () => accept(row) }, { default: () => '受理' }))
  if (canReject.value && ['pending', 'processing'].includes(row.status)) nodes.push(h(NButton, { text: true, type: 'error', onClick: () => openReject(row) }, { default: () => '驳回' }))
  if (canIssue.value && row.status === 'processing') nodes.push(h(NButton, { text: true, type: 'success', onClick: () => openIssue(row) }, { default: () => '开票' }))
  if (row.status === 'issued') nodes.push(h(NButton, { text: true, type: 'primary', loading: downloadingInvoice.value === row.invoice_no, onClick: () => download(row.invoice_no) }, { default: () => '下载' }))
  return h(NSpace, { size: 8 }, { default: () => nodes })
}

const columns = computed<DataTableColumns<AdminInvoiceItem>>(() => [
  { key: 'invoice_no', title: '申请编号', minWidth: 170 },
  { key: 'user', title: '用户', minWidth: 150, render: userLabel },
  { key: 'title', title: '抬头', minWidth: 180, ellipsis: { tooltip: true }, render: (row) => `${titleTypeText[row.title_type] || row.title_type} / ${row.title}` },
  { key: 'amount', title: '金额', width: 140, render: (row) => formatMoney(row.amount_cents, row.currency) },
  { key: 'order_count', title: '订单数', width: 90 },
  { key: 'status', title: '状态', width: 100, render: (row) => h(InvoiceStatusTag, { status: row.status }) },
  { key: 'invoice_number', title: '发票号码', minWidth: 150, render: (row) => row.invoice_number || '-' },
  { key: 'created_at', title: '申请时间', minWidth: 170, render: (row) => formatDateTime(row.created_at) },
  { key: 'actions', title: '操作', width: 230, fixed: 'right', render: actions },
])

const orderColumns: DataTableColumns<InvoiceOrderItem> = [
  { key: 'order_no', title: '订单编号', minWidth: 170 },
  { key: 'order_type', title: '类型', width: 90, render: (row) => orderTypeText[row.order_type] || row.order_type },
  { key: 'product_name', title: '产品', minWidth: 140, render: (row) => row.product_name || '-' },
  { key: 'plan_name', title: '套餐', minWidth: 140, render: (row) => row.plan_name || '-' },
  { key: 'amount', title: '金额', width: 130, render: (row) => formatMoney(row.order_amount_cents, row.currency) },
  { key: 'payment_status', title: '支付状态', width: 110 },
  { key: 'paid_at', title: '支付时间', minWidth: 170, render: (row) => formatDateTime(row.paid_at) },
]

onMounted(loadInvoices)
</script>

<template>
  <div class="invoices-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>发票运营</h2>
          <p class="muted">处理用户电子普通发票申请、线下开票登记和 PDF 归档。</p>
        </div>
      </template>

      <NForm inline label-placement="left" class="query-form">
        <NFormItem label="状态">
          <NSelect v-model:value="query.status" clearable :options="statusOptions" placeholder="全部" style="width: 130px" />
        </NFormItem>
        <NFormItem label="编号">
          <NSpace>
            <NInput v-model:value="query.invoice_no" clearable placeholder="申请编号" />
            <NInput v-model:value="query.order_no" clearable placeholder="订单编号" />
          </NSpace>
        </NFormItem>
        <NFormItem label="用户">
          <NInput v-model:value="query.user_keyword" clearable placeholder="用户名/邮箱" />
        </NFormItem>
        <NFormItem label="抬头">
          <NInput v-model:value="query.title_keyword" clearable placeholder="发票抬头" />
        </NFormItem>
        <NFormItem label="申请时间">
          <NDatePicker v-model:value="dateRange" type="datetimerange" clearable style="width: 330px" />
        </NFormItem>
        <NFormItem :show-label="false">
          <NSpace>
            <NButton type="primary" @click="query.page = 1; loadInvoices()">查询</NButton>
            <NButton @click="resetQuery">重置</NButton>
          </NSpace>
        </NFormItem>
      </NForm>

      <NDataTable :loading="loading" :columns="columns" :data="invoices" :row-key="(row: AdminInvoiceItem) => row.invoice_no" :bordered="false" />
      <div class="pagination">
        <NPagination v-model:page="query.page" v-model:page-size="query.per_page" :item-count="total" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="loadInvoices" @update:page-size="loadInvoices" />
      </div>
    </NCard>

    <NDrawer v-model:show="detailVisible" :width="760">
      <NDrawerContent title="发票详情" closable>
        <div v-if="detailLoading" class="muted">加载中</div>
        <template v-else-if="selectedInvoice">
          <NSpace justify="space-between" align="center" class="detail-actions">
            <InvoiceStatusTag :status="selectedInvoice.status" />
            <NSpace>
              <NButton v-if="canUpdate" size="small" @click="openNote(selectedInvoice)">备注</NButton>
              <NButton v-if="canUpdate && selectedInvoice.status === 'pending'" size="small" type="primary" @click="accept(selectedInvoice)">受理</NButton>
              <NButton v-if="canReject && ['pending', 'processing'].includes(selectedInvoice.status)" size="small" type="error" @click="openReject(selectedInvoice)">驳回</NButton>
              <NButton v-if="canIssue && selectedInvoice.status === 'processing'" size="small" type="success" @click="openIssue(selectedInvoice)">开票</NButton>
              <NButton v-if="selectedInvoice.status === 'issued'" size="small" type="primary" :loading="downloadingInvoice === selectedInvoice.invoice_no" @click="download(selectedInvoice.invoice_no)">下载 PDF</NButton>
            </NSpace>
          </NSpace>

          <NDescriptions bordered :column="2" label-placement="left">
            <NDescriptionsItem label="申请编号">{{ selectedInvoice.invoice_no }}</NDescriptionsItem>
            <NDescriptionsItem label="申请金额">{{ formatMoney(selectedInvoice.amount_cents, selectedInvoice.currency) }}</NDescriptionsItem>
            <NDescriptionsItem label="用户">{{ userLabel(selectedInvoice) }}</NDescriptionsItem>
            <NDescriptionsItem label="抬头类型">{{ titleTypeText[selectedInvoice.title_type] }}</NDescriptionsItem>
            <NDescriptionsItem label="发票抬头">{{ selectedInvoice.title }}</NDescriptionsItem>
            <NDescriptionsItem label="税号">{{ selectedInvoice.tax_no || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="邮箱">{{ selectedInvoice.email || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="发票号码">{{ selectedInvoice.invoice_number || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="申请时间">{{ formatDateTime(selectedInvoice.created_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="开票时间">{{ formatDateTime(selectedInvoice.issued_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="驳回原因" :span="2">{{ selectedInvoice.reject_reason || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="用户备注" :span="2">{{ selectedInvoice.remark || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="后台备注" :span="2">{{ selectedInvoice.admin_note || '-' }}</NDescriptionsItem>
          </NDescriptions>

          <h3>订单明细</h3>
          <NDataTable :columns="orderColumns" :data="selectedInvoice.orders" :row-key="(row: InvoiceOrderItem) => row.order_no" :bordered="false" />

          <div v-if="selectedInvoice.file" class="file-line">
            <NTag type="success">{{ selectedInvoice.file.original_name }}</NTag>
            <NButton text type="primary" :loading="downloadingInvoice === selectedInvoice.invoice_no" @click="download(selectedInvoice.invoice_no)">下载 PDF</NButton>
          </div>
        </template>
      </NDrawerContent>
    </NDrawer>

    <NModal v-model:show="rejectVisible" preset="dialog" title="驳回发票" positive-text="驳回" negative-text="取消" :positive-button-props="{ type: 'error', loading: submitting }" @positive-click="submitReject">
      <NInput v-model:value="rejectForm.reason" type="textarea" maxlength="500" show-count :autosize="{ minRows: 3, maxRows: 5 }" placeholder="填写驳回原因" />
    </NModal>

    <NModal v-model:show="issueVisible" preset="dialog" title="登记开票" positive-text="登记" negative-text="取消" :positive-button-props="{ type: 'success', loading: submitting }" @positive-click="submitIssue">
      <NSpace vertical :size="12">
        <NInput v-model:value="issueForm.invoice_code" placeholder="发票代码（可选）" maxlength="64" />
        <NInput v-model:value="issueForm.invoice_number" placeholder="发票号码" maxlength="128" />
        <NDatePicker v-model:value="issueForm.issued_at" type="datetime" clearable style="width: 100%" />
        <NSpace align="center">
          <NButton :loading="uploading" @click="choosePDF">上传 PDF</NButton>
          <NTag v-if="issueForm.file_id" type="success">{{ issueForm.file_name }}</NTag>
        </NSpace>
        <input ref="pdfInput" class="hidden-input" type="file" accept="application/pdf" @change="handlePDFSelected" />
      </NSpace>
    </NModal>

    <NModal v-model:show="noteVisible" preset="dialog" title="后台备注" positive-text="保存" negative-text="取消" :positive-button-props="{ loading: submitting }" @positive-click="submitNote">
      <NInput v-model:value="adminNote" type="textarea" maxlength="1000" show-count :autosize="{ minRows: 4, maxRows: 8 }" />
    </NModal>
  </div>
</template>

<style scoped>
.invoices-page {
  display: grid;
  gap: 16px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
}

.muted {
  margin: 4px 0 0;
  color: var(--text-color-3);
}

.query-form {
  margin-bottom: 16px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.detail-actions {
  margin-bottom: 16px;
}

h3 {
  margin: 20px 0 12px;
  font-size: 16px;
}

.file-line {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
}

.hidden-input {
  display: none;
}
</style>
