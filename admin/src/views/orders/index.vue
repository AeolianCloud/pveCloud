<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
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
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import {
  cancelOrder,
  closeOrder,
  getOrderDetail,
  getOrders,
  updateOrderAdminNote,
  type AdminOrderDetail,
  type AdminOrderItem,
} from '../../api/order'
import { usePermissionStore } from '../../store/modules/permission'
import { hasPermissionCode } from '../../utils/permission'
import { confirm, getDialog, message } from '../../utils/feedback'
import { formatDateTime } from '../../utils/datetime'

const permissionStore = usePermissionStore()
const loading = ref(false)
const detailLoading = ref(false)
const detailVisible = ref(false)
const detail = ref<AdminOrderDetail | null>(null)
const noteDraft = ref('')
const orders = ref<AdminOrderItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, status: '', order_no: '', user_keyword: '' })

const canUpdate = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'order:update'))
const canCancel = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'order:cancel'))

const statusText: Record<string, string> = { pending: '待处理', cancelled: '已取消', closed: '已关闭' }
const cycleText: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  semi_yearly: '半年付',
  yearly: '年付',
}

const statusOptions = [
  { label: '待处理', value: 'pending' },
  { label: '已取消', value: 'cancelled' },
  { label: '已关闭', value: 'closed' },
]

const formatMoney = (cents: number) => `¥${(cents / 100).toFixed(2)}`

async function loadOrders() {
  loading.value = true
  try {
    const data = await getOrders(query)
    orders.value = data.list
    total.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '订单加载失败')
  } finally {
    loading.value = false
  }
}

async function openDetail(orderNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    detail.value = await getOrderDetail(orderNo)
    noteDraft.value = detail.value.admin_note || ''
  } catch (err) {
    message.error(err instanceof Error ? err.message : '订单详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

async function saveNote() {
  if (!detail.value) return
  try {
    detail.value = await updateOrderAdminNote(detail.value.order_no, noteDraft.value || null)
    message.success('后台备注已更新')
    await loadOrders()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '备注更新失败')
  }
}

async function updateStatus(action: 'cancel' | 'close', order: AdminOrderItem | AdminOrderDetail) {
  const label = action === 'cancel' ? '取消' : '关闭'
  // 使用一个简易的 prompt：先 confirm 再用第二个对话框收集原因
  const reason = await promptReason(`请输入${label}原因（可选）`, `${label}订单`)
  if (reason === null) return
  try {
    if (action === 'cancel') {
      await cancelOrder(order.order_no, reason)
    } else {
      await closeOrder(order.order_no, reason)
    }
    message.success(`订单已${label}`)
    await loadOrders()
    if (detailVisible.value) await openDetail(order.order_no)
  } catch (err) {
    message.error(err instanceof Error ? err.message : `${label}失败`)
  }
}

function promptReason(content: string, title: string): Promise<string | null> {
  return new Promise((resolve) => {
    const value = ref('')
    const dialog = getDialog()
    const d = dialog.create({
      title,
      showIcon: false,
      content: () =>
        h('div', { style: 'display:grid;gap:8px;' }, [
          h('div', { style: 'color:rgba(15,23,42,0.7);font-size:13px;' }, content),
          h(NInput, {
            value: value.value,
            type: 'textarea',
            rows: 3,
            placeholder: '可选',
            'onUpdate:value': (v: string) => (value.value = v),
          }),
        ]),
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: () => {
        resolve(value.value)
        d.destroy()
      },
      onNegativeClick: () => {
        resolve(null)
        d.destroy()
      },
      onClose: () => {
        resolve(null)
      },
    })
  })
}

// keep confirm import alive (used elsewhere if needed)
void confirm

function resetQuery() {
  Object.assign(query, { page: 1, per_page: 15, status: '', order_no: '', user_keyword: '' })
  loadOrders()
}

const columns = computed<DataTableColumns<AdminOrderItem>>(() => [
  { key: 'order_no', title: '订单号', minWidth: 180 },
  {
    key: 'user',
    title: '用户',
    minWidth: 160,
    render: (row) =>
      h('div', null, [
        h('div', { class: 'strong' }, row.user.username),
        h('div', { class: 'muted' }, row.user.email),
      ]),
  },
  {
    key: 'product',
    title: '产品',
    minWidth: 180,
    render: (row) =>
      h('div', null, [
        h('div', { class: 'strong' }, row.product_name),
        h('div', { class: 'muted' }, `${row.plan_name} · ${cycleText[row.billing_cycle] || row.billing_cycle} · ${row.network_type_name || '-'}`),
      ]),
  },
  {
    key: 'amount',
    title: '金额',
    width: 120,
    render: (row) => formatMoney(row.total_amount_cents),
  },
  {
    key: 'status',
    title: '状态',
    width: 100,
    render: (row) => h(NTag, { size: 'small' }, { default: () => statusText[row.status] || row.status }),
  },
  {
    key: 'created_at',
    title: '创建时间',
    minWidth: 170,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    key: 'actions',
    title: '操作',
    width: 240,
    fixed: 'right',
    render: (row) =>
      h(NSpace, null, {
        default: () => {
          const buttons = [
            h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row.order_no) }, { default: () => '详情' }),
          ]
          if (canCancel.value && row.status === 'pending') {
            buttons.push(
              h(NButton, { text: true, type: 'error', onClick: () => updateStatus('cancel', row) }, { default: () => '取消' }),
            )
          }
          if (canUpdate.value && row.status === 'pending') {
            buttons.push(
              h(NButton, { text: true, type: 'warning', onClick: () => updateStatus('close', row) }, { default: () => '关闭' }),
            )
          }
          return buttons
        },
      }),
  },
])

onMounted(loadOrders)
</script>

<template>
  <div class="orders-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>订单管理</h2>
          <p class="muted">查看和处理用户端创建的订单，不支持后台创建订单。</p>
        </div>
      </template>

      <NForm inline label-placement="left" class="query-form">
        <NFormItem label="状态">
          <NSelect
            v-model:value="query.status"
            :options="statusOptions"
            clearable
            placeholder="全部"
            style="width: 140px"
          />
        </NFormItem>
        <NFormItem label="订单号">
          <NInput v-model:value="query.order_no" clearable placeholder="ORD-" />
        </NFormItem>
        <NFormItem label="用户">
          <NInput v-model:value="query.user_keyword" clearable placeholder="用户名/邮箱" />
        </NFormItem>
        <NFormItem :show-label="false">
          <NSpace>
            <NButton type="primary" @click="query.page = 1; loadOrders()">查询</NButton>
            <NButton @click="resetQuery">重置</NButton>
          </NSpace>
        </NFormItem>
      </NForm>

      <NDataTable
        :loading="loading"
        :columns="columns"
        :data="orders"
        :row-key="(row: AdminOrderItem) => row.order_no"
        :bordered="false"
      />

      <div class="pagination">
        <NPagination
          v-model:page="query.page"
          v-model:page-size="query.per_page"
          :item-count="total"
          show-size-picker
          :page-sizes="[10, 15, 20, 50]"
          @update:page="loadOrders"
          @update:page-size="loadOrders"
        />
      </div>
    </NCard>

    <NDrawer v-model:show="detailVisible" :width="520">
      <NDrawerContent title="订单详情" closable>
        <div v-if="detail" class="detail">
          <h3>{{ detail.order_no }}</h3>
          <NTag>{{ statusText[detail.status] || detail.status }}</NTag>
          <NDescriptions :column="1" bordered class="mt" label-placement="left" size="small">
            <NDescriptionsItem label="用户">{{ detail.user.username }} / {{ detail.user.email }}</NDescriptionsItem>
            <NDescriptionsItem label="产品">{{ detail.product_name }}</NDescriptionsItem>
            <NDescriptionsItem label="套餐">
              {{ detail.plan_name }}（{{ detail.cpu_cores }}核 / {{ Math.round(detail.memory_mb / 1024) }}GB / {{ detail.bandwidth_mbps }}M）
            </NDescriptionsItem>
            <NDescriptionsItem label="地域">{{ detail.region_name }}</NDescriptionsItem>
            <NDescriptionsItem label="系统">{{ detail.template_name }}</NDescriptionsItem>
            <NDescriptionsItem label="金额">{{ formatMoney(detail.total_amount_cents) }}</NDescriptionsItem>
            <NDescriptionsItem label="用户备注">{{ detail.user_note || '-' }}</NDescriptionsItem>
          </NDescriptions>
          <div class="mt">
            <div class="strong">后台备注</div>
            <NInput v-model:value="noteDraft" type="textarea" :rows="4" :disabled="!canUpdate" />
            <NButton v-if="canUpdate" class="mt" type="primary" @click="saveNote">保存备注</NButton>
          </div>
          <div v-if="detail.status === 'pending'" class="mt actions">
            <NSpace>
              <NButton v-if="canCancel" type="error" @click="updateStatus('cancel', detail)">取消订单</NButton>
              <NButton v-if="canUpdate" type="warning" @click="updateStatus('close', detail)">关闭订单</NButton>
            </NSpace>
          </div>
        </div>
        <div v-else-if="detailLoading">加载中...</div>
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
.strong {
  font-weight: 700;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.detail h3 {
  margin: 0 0 12px;
}
.mt {
  margin-top: 16px;
}
.actions {
  display: flex;
  gap: 12px;
}
</style>
