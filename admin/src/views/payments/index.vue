<script setup lang="ts">
import {
  NAlert,
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
  NPagination,
  NSelect,
  NSpace,
  NTabPane,
  NTabs,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import {
  createPaymentRefund,
  getPaymentDetail,
  getPayments,
  getRefunds,
  retryPaymentProvision,
  syncPayment,
  type AdminPaymentDetail,
  type AdminPaymentItem,
  type AdminRefundItem,
} from '../../api/payment'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'
import { getDialog, message } from '../../utils/feedback'
import { hasPermissionCode } from '../../utils/permission'
import type { PaymentQueryState, RefundQueryState } from './types'

const permissionStore = usePermissionStore()
const activeTab = ref<'payments' | 'refunds'>('payments')
const loadingPayments = ref(false)
const loadingRefunds = ref(false)
const detailLoading = ref(false)
const payments = ref<AdminPaymentItem[]>([])
const refunds = ref<AdminRefundItem[]>([])
const paymentTotal = ref(0)
const refundTotal = ref(0)
const selectedPayment = ref<AdminPaymentDetail | null>(null)
const detailVisible = ref(false)
const paymentDateRange = ref<[number, number] | null>(null)
const refundDateRange = ref<[number, number] | null>(null)

const paymentQuery = reactive<PaymentQueryState>({
  page: 1,
  per_page: 15,
  provider: '',
  method: '',
  status: '',
  order_no: '',
  payment_no: '',
  user_keyword: '',
})

const refundQuery = reactive<RefundQueryState>({
  page: 1,
  per_page: 15,
  provider: '',
  status: '',
  order_no: '',
  payment_no: '',
  refund_no: '',
})

const canRefund = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'payment:refund'))
const canSync = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'payment:sync'))
const canRetryProvision = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'payment:retry-provision'))

const providerOptions = [
  { label: '支付宝', value: 'alipay' },
  { label: '微信支付', value: 'wechat' },
  { label: '钱包余额', value: 'wallet' },
]

const methodOptions = [
  { label: '支付宝电脑网页', value: 'alipay_page' },
  { label: '支付宝手机网页', value: 'alipay_wap' },
  { label: '微信 Native 扫码', value: 'wechat_native' },
  { label: '微信 H5', value: 'wechat_h5' },
  { label: '钱包余额', value: 'wallet_balance' },
]

const paymentStatusOptions = [
  { label: '待支付', value: 'pending' },
  { label: '已支付', value: 'paid' },
  { label: '已关闭', value: 'closed' },
  { label: '失败', value: 'failed' },
  { label: '已退款', value: 'refunded' },
]

const refundStatusOptions = [
  { label: '处理中', value: 'pending' },
  { label: '已成功', value: 'succeeded' },
  { label: '失败', value: 'failed' },
]

const providerText: Record<string, string> = { alipay: '支付宝', wechat: '微信支付', wallet: '钱包余额' }
const methodText: Record<string, string> = {
  alipay_page: '支付宝电脑网页',
  alipay_wap: '支付宝手机网页',
  wechat_native: '微信 Native 扫码',
  wechat_h5: '微信 H5',
  wallet_balance: '钱包余额',
}
const paymentStatusText: Record<string, string> = { pending: '待支付', paid: '已支付', closed: '已关闭', failed: '失败', refunded: '已退款' }
const refundStatusText: Record<string, string> = { pending: '处理中', succeeded: '已成功', failed: '失败' }
const orderTypeText: Record<string, string> = { purchase: '新购', renewal: '续费' }

function formatMoney(cents: number, currency = 'CNY') {
  return `${currency} ${(cents / 100).toFixed(2)}`
}

function userLabel(row: { user: AdminPaymentItem['user'] }) {
  return row.user.display_name || row.user.username || row.user.email || `用户 ${row.user.id}`
}

function withDateRange(params: Record<string, unknown>, range: [number, number] | null) {
  if (range) {
    params.date_from = new Date(range[0]).toISOString()
    params.date_to = new Date(range[1]).toISOString()
  }
  return params
}

function canRefundPayment(row: AdminPaymentItem | AdminPaymentDetail) {
  // 前端只负责降低误操作概率；订单状态、实例是否已释放和支付状态仍以后端校验为准。
  return canRefund.value && row.status === 'paid' && !(row.order_type === 'purchase' && row.order_status === 'fulfilled')
}

function canRetryPayment(row: AdminPaymentItem | AdminPaymentDetail) {
  return canRetryProvision.value && row.order_type === 'purchase' && row.status === 'paid' && row.order_status === 'error'
}

async function loadPayments() {
  loadingPayments.value = true
  try {
    const data = await getPayments(withDateRange({ ...paymentQuery }, paymentDateRange.value))
    payments.value = data.list
    paymentTotal.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '支付流水加载失败')
  } finally {
    loadingPayments.value = false
  }
}

async function loadRefunds() {
  loadingRefunds.value = true
  try {
    const data = await getRefunds(withDateRange({ ...refundQuery }, refundDateRange.value))
    refunds.value = data.list
    refundTotal.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '退款流水加载失败')
  } finally {
    loadingRefunds.value = false
  }
}

async function openDetail(paymentNo: string) {
  detailLoading.value = true
  detailVisible.value = true
  try {
    selectedPayment.value = await getPaymentDetail(paymentNo)
  } catch (err) {
    message.error(err instanceof Error ? err.message : '支付详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

async function sync(row: AdminPaymentItem | AdminPaymentDetail) {
  try {
    selectedPayment.value = await syncPayment(row.payment_no)
    message.success('渠道状态已同步')
    await loadPayments()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '同步失败')
  }
}

function askRefundReason(row: AdminPaymentItem | AdminPaymentDetail) {
  return new Promise<string | null>((resolve) => {
    const reason = ref('')
    getDialog().warning({
      title: '发起全额退款',
      content: () =>
        h(NSpace, { vertical: true, size: 12 }, {
          default: () => [
            h(NAlert, { type: 'warning', showIcon: true }, { default: () => '一期仅支持全额退款。续费退款会在渠道成功后扣回服务期。' }),
            h(NInput, {
              value: reason.value,
              type: 'textarea',
              placeholder: `填写退款原因，支付 ${row.payment_no}`,
              maxlength: 500,
              showCount: true,
              autosize: { minRows: 3, maxRows: 5 },
              'onUpdate:value': (value: string) => {
                reason.value = value
              },
            }),
          ],
        }),
      positiveText: '确认退款',
      negativeText: '取消',
      onPositiveClick: () => {
        const value = reason.value.trim()
        if (!value) {
          message.warning('请填写退款原因')
          return false
        }
        resolve(value)
      },
      onNegativeClick: () => resolve(null),
      onClose: () => resolve(null),
      onMaskClick: () => resolve(null),
    })
  })
}

async function refund(row: AdminPaymentItem | AdminPaymentDetail) {
  const reason = await askRefundReason(row)
  if (!reason) return
  try {
    await createPaymentRefund(row.payment_no, reason)
    message.success('退款已提交')
    await loadPayments()
    await loadRefunds()
    if (selectedPayment.value?.payment_no === row.payment_no) {
      selectedPayment.value = await getPaymentDetail(row.payment_no)
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : '退款失败')
  }
}

async function retryProvision(row: AdminPaymentItem | AdminPaymentDetail) {
  try {
    selectedPayment.value = await retryPaymentProvision(row.payment_no)
    message.success('自动交付任务已重新入队')
    await loadPayments()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '重试失败')
  }
}

function resetPaymentQuery() {
  Object.assign(paymentQuery, { page: 1, per_page: 15, provider: '', method: '', status: '', order_no: '', payment_no: '', user_keyword: '' })
  paymentDateRange.value = null
  void loadPayments()
}

function resetRefundQuery() {
  Object.assign(refundQuery, { page: 1, per_page: 15, provider: '', status: '', order_no: '', payment_no: '', refund_no: '' })
  refundDateRange.value = null
  void loadRefunds()
}

function paymentTagType(status: string) {
  return status === 'paid' ? 'success' : status === 'failed' ? 'error' : status === 'refunded' ? 'warning' : 'default'
}

function refundTagType(status: string) {
  return status === 'succeeded' ? 'success' : status === 'failed' ? 'error' : 'warning'
}

const paymentColumns = computed<DataTableColumns<AdminPaymentItem>>(() => [
  { key: 'payment_no', title: '支付编号', minWidth: 170 },
  { key: 'order_no', title: '订单编号', minWidth: 170 },
  { key: 'user', title: '用户', minWidth: 150, render: userLabel },
  { key: 'provider', title: '渠道', width: 100, render: (row) => providerText[row.provider] || row.provider },
  { key: 'method', title: '方式', minWidth: 150, render: (row) => methodText[row.method] || row.method },
  { key: 'amount', title: '金额', width: 130, render: (row) => formatMoney(row.amount_cents, row.currency) },
  {
    key: 'status',
    title: '状态',
    width: 100,
    render: (row) => h(NTag, { size: 'small', type: paymentTagType(row.status) }, { default: () => paymentStatusText[row.status] || row.status }),
  },
  { key: 'order_status', title: '订单', width: 130, render: (row) => `${orderTypeText[row.order_type] || row.order_type} / ${row.order_status}` },
  { key: 'created_at', title: '创建时间', minWidth: 170, render: (row) => formatDateTime(row.created_at) },
  {
    key: 'actions',
    title: '操作',
    width: 210,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row.payment_no) }, { default: () => '详情' }),
          canSync.value && ['pending', 'failed'].includes(row.status) ? h(NButton, { text: true, type: 'primary', onClick: () => sync(row) }, { default: () => '同步' }) : null,
          canRefundPayment(row) ? h(NButton, { text: true, type: 'error', onClick: () => refund(row) }, { default: () => '退款' }) : null,
          canRetryPayment(row) ? h(NButton, { text: true, type: 'warning', onClick: () => retryProvision(row) }, { default: () => '重试交付' }) : null,
        ],
      }),
  },
])

const refundColumns = computed<DataTableColumns<AdminRefundItem>>(() => [
  { key: 'refund_no', title: '退款编号', minWidth: 170 },
  { key: 'payment_no', title: '支付编号', minWidth: 170 },
  { key: 'order_no', title: '订单编号', minWidth: 170 },
  { key: 'user', title: '用户', minWidth: 150, render: userLabel },
  { key: 'provider', title: '渠道', width: 100, render: (row) => providerText[row.provider] || row.provider },
  { key: 'amount', title: '金额', width: 130, render: (row) => formatMoney(row.amount_cents, row.currency) },
  {
    key: 'status',
    title: '状态',
    width: 100,
    render: (row) => h(NTag, { size: 'small', type: refundTagType(row.status) }, { default: () => refundStatusText[row.status] || row.status }),
  },
  { key: 'reason', title: '原因', minWidth: 220, ellipsis: { tooltip: true } },
  { key: 'created_at', title: '创建时间', minWidth: 170, render: (row) => formatDateTime(row.created_at) },
  { key: 'completed_at', title: '完成时间', minWidth: 170, render: (row) => formatDateTime(row.completed_at) },
])

onMounted(() => {
  void loadPayments()
  void loadRefunds()
})
</script>

<template>
  <div class="payments-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>支付管理</h2>
          <p class="muted">查看支付与退款流水，并处理渠道同步、全额退款和自动交付失败重试。</p>
        </div>
      </template>

      <NTabs v-model:value="activeTab" type="line" animated>
        <NTabPane name="payments" tab="支付流水">
          <NForm inline label-placement="left" class="query-form">
            <NFormItem label="渠道"><NSelect v-model:value="paymentQuery.provider" :options="providerOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
            <NFormItem label="方式"><NSelect v-model:value="paymentQuery.method" :options="methodOptions" clearable placeholder="全部方式" style="width: 180px" /></NFormItem>
            <NFormItem label="状态"><NSelect v-model:value="paymentQuery.status" :options="paymentStatusOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
            <NFormItem label="编号"><NSpace><NInput v-model:value="paymentQuery.payment_no" clearable placeholder="支付编号" /><NInput v-model:value="paymentQuery.order_no" clearable placeholder="订单编号" /></NSpace></NFormItem>
            <NFormItem label="用户"><NInput v-model:value="paymentQuery.user_keyword" clearable placeholder="用户名/邮箱" /></NFormItem>
            <NFormItem label="创建时间"><NDatePicker v-model:value="paymentDateRange" type="datetimerange" clearable style="width: 330px" /></NFormItem>
            <NFormItem :show-label="false"><NSpace><NButton type="primary" @click="paymentQuery.page = 1; loadPayments()">查询</NButton><NButton @click="resetPaymentQuery">重置</NButton></NSpace></NFormItem>
          </NForm>

          <NDataTable :loading="loadingPayments" :columns="paymentColumns" :data="payments" :row-key="(row: AdminPaymentItem) => row.payment_no" :bordered="false" />
          <div class="pagination">
            <NPagination v-model:page="paymentQuery.page" v-model:page-size="paymentQuery.per_page" :item-count="paymentTotal" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="loadPayments" @update:page-size="loadPayments" />
          </div>
        </NTabPane>

        <NTabPane name="refunds" tab="退款流水">
          <NForm inline label-placement="left" class="query-form">
            <NFormItem label="渠道"><NSelect v-model:value="refundQuery.provider" :options="providerOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
            <NFormItem label="状态"><NSelect v-model:value="refundQuery.status" :options="refundStatusOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
            <NFormItem label="编号"><NSpace><NInput v-model:value="refundQuery.refund_no" clearable placeholder="退款编号" /><NInput v-model:value="refundQuery.payment_no" clearable placeholder="支付编号" /><NInput v-model:value="refundQuery.order_no" clearable placeholder="订单编号" /></NSpace></NFormItem>
            <NFormItem label="创建时间"><NDatePicker v-model:value="refundDateRange" type="datetimerange" clearable style="width: 330px" /></NFormItem>
            <NFormItem :show-label="false"><NSpace><NButton type="primary" @click="refundQuery.page = 1; loadRefunds()">查询</NButton><NButton @click="resetRefundQuery">重置</NButton></NSpace></NFormItem>
          </NForm>

          <NDataTable :loading="loadingRefunds" :columns="refundColumns" :data="refunds" :row-key="(row: AdminRefundItem) => row.refund_no" :bordered="false" />
          <div class="pagination">
            <NPagination v-model:page="refundQuery.page" v-model:page-size="refundQuery.per_page" :item-count="refundTotal" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="loadRefunds" @update:page-size="loadRefunds" />
          </div>
        </NTabPane>
      </NTabs>
    </NCard>

    <NDrawer v-model:show="detailVisible" :width="620" placement="right">
      <NDrawerContent title="支付详情">
        <NAlert v-if="detailLoading" type="info" :show-icon="false">详情加载中...</NAlert>
        <template v-else-if="selectedPayment">
          <NSpace class="detail-actions">
            <NButton v-if="canSync && ['pending', 'failed'].includes(selectedPayment.status)" type="primary" secondary @click="sync(selectedPayment)">同步渠道</NButton>
            <NButton v-if="canRefundPayment(selectedPayment)" type="error" secondary @click="refund(selectedPayment)">发起全额退款</NButton>
            <NButton v-if="canRetryPayment(selectedPayment)" type="warning" secondary @click="retryProvision(selectedPayment)">重试自动交付</NButton>
          </NSpace>
          <NDescriptions :column="1" label-placement="left" bordered size="small">
            <NDescriptionsItem label="支付编号">{{ selectedPayment.payment_no }}</NDescriptionsItem>
            <NDescriptionsItem label="订单编号">{{ selectedPayment.order_no }}</NDescriptionsItem>
            <NDescriptionsItem label="用户">{{ userLabel(selectedPayment) }}</NDescriptionsItem>
            <NDescriptionsItem label="渠道方式">{{ providerText[selectedPayment.provider] }} / {{ methodText[selectedPayment.method] }}</NDescriptionsItem>
            <NDescriptionsItem label="金额">{{ formatMoney(selectedPayment.amount_cents, selectedPayment.currency) }}</NDescriptionsItem>
            <NDescriptionsItem label="支付状态">{{ paymentStatusText[selectedPayment.status] || selectedPayment.status }}</NDescriptionsItem>
            <NDescriptionsItem label="订单状态">{{ orderTypeText[selectedPayment.order_type] || selectedPayment.order_type }} / {{ selectedPayment.order_status }}</NDescriptionsItem>
            <NDescriptionsItem label="渠道交易号">{{ selectedPayment.upstream_trade_no || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="最近错误">{{ selectedPayment.last_error_message || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="过期时间">{{ formatDateTime(selectedPayment.expires_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="完成时间">{{ formatDateTime(selectedPayment.paid_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="创建时间">{{ formatDateTime(selectedPayment.created_at) }}</NDescriptionsItem>
          </NDescriptions>

          <NDescriptions v-if="selectedPayment.refund" class="refund-detail" :column="1" label-placement="left" bordered size="small">
            <NDescriptionsItem label="退款编号">{{ selectedPayment.refund.refund_no }}</NDescriptionsItem>
            <NDescriptionsItem label="退款状态">{{ refundStatusText[selectedPayment.refund.status] || selectedPayment.refund.status }}</NDescriptionsItem>
            <NDescriptionsItem label="退款原因">{{ selectedPayment.refund.reason }}</NDescriptionsItem>
            <NDescriptionsItem label="完成时间">{{ formatDateTime(selectedPayment.refund.completed_at) }}</NDescriptionsItem>
          </NDescriptions>
        </template>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.page-header h2 {
  margin: 0;
  font-size: 20px;
}
.muted {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
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
.refund-detail {
  margin-top: 16px;
}
</style>
