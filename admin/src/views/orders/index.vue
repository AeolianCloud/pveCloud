<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { cancelOrder, closeOrder, getOrderDetail, getOrders, updateOrderAdminNote, type AdminOrderDetail, type AdminOrderItem } from '../../api/order'
import { usePermissionStore } from '../../store/modules/permission'
import { hasPermissionCode } from '../../utils/permission'

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
const cycleText: Record<string, string> = { monthly: '月付', quarterly: '季付', semi_yearly: '半年付', yearly: '年付' }

const formatMoney = (cents: number) => `¥${(cents / 100).toFixed(2)}`

async function loadOrders() {
  loading.value = true
  try {
    const data = await getOrders(query)
    orders.value = data.list
    total.value = data.total
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '订单加载失败')
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
    ElMessage.error(err instanceof Error ? err.message : '订单详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

async function saveNote() {
  if (!detail.value) return
  try {
    detail.value = await updateOrderAdminNote(detail.value.order_no, noteDraft.value || null)
    ElMessage.success('后台备注已更新')
    await loadOrders()
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '备注更新失败')
  }
}

async function updateStatus(action: 'cancel' | 'close', order: AdminOrderItem | AdminOrderDetail) {
  const label = action === 'cancel' ? '取消' : '关闭'
  try {
    const { value } = await ElMessageBox.prompt(`请输入${label}原因`, `${label}订单`, { inputType: 'textarea', inputPlaceholder: '可选' })
    if (action === 'cancel') {
      await cancelOrder(order.order_no, value)
    } else {
      await closeOrder(order.order_no, value)
    }
    ElMessage.success(`订单已${label}`)
    await loadOrders()
    if (detailVisible.value) await openDetail(order.order_no)
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(err instanceof Error ? err.message : `${label}失败`)
  }
}

function resetQuery() {
  Object.assign(query, { page: 1, per_page: 15, status: '', order_no: '', user_keyword: '' })
  loadOrders()
}

onMounted(loadOrders)
</script>

<template>
  <div class="orders-page">
    <el-card shadow="never">
      <template #header>
        <div class="page-header"><div><h2>订单管理</h2><p>查看和处理用户端创建的订单，不支持后台创建订单。</p></div></div>
      </template>
      <el-form :inline="true" :model="query" class="query-form">
        <el-form-item label="状态"><el-select v-model="query.status" clearable placeholder="全部" style="width: 140px"><el-option label="待处理" value="pending" /><el-option label="已取消" value="cancelled" /><el-option label="已关闭" value="closed" /></el-select></el-form-item>
        <el-form-item label="订单号"><el-input v-model="query.order_no" clearable placeholder="ORD-" /></el-form-item>
        <el-form-item label="用户"><el-input v-model="query.user_keyword" clearable placeholder="用户名/邮箱" /></el-form-item>
        <el-form-item><el-button type="primary" @click="query.page = 1; loadOrders()">查询</el-button><el-button @click="resetQuery">重置</el-button></el-form-item>
      </el-form>
      <el-table v-loading="loading" :data="orders" row-key="order_no">
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column label="用户" min-width="160"><template #default="{ row }"><div class="strong">{{ row.user.username }}</div><div class="muted">{{ row.user.email }}</div></template></el-table-column>
        <el-table-column label="产品" min-width="180"><template #default="{ row }"><div class="strong">{{ row.product_name }}</div><div class="muted">{{ row.plan_name }} · {{ cycleText[row.billing_cycle] || row.billing_cycle }}</div></template></el-table-column>
        <el-table-column label="金额" width="120"><template #default="{ row }">{{ formatMoney(row.total_amount_cents) }}</template></el-table-column>
        <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag>{{ statusText[row.status] || row.status }}</el-tag></template></el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="170" />
        <el-table-column label="操作" width="220" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openDetail(row.order_no)">详情</el-button><el-button v-if="canCancel && row.status === 'pending'" link type="danger" @click="updateStatus('cancel', row)">取消</el-button><el-button v-if="canUpdate && row.status === 'pending'" link type="warning" @click="updateStatus('close', row)">关闭</el-button></template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination v-model:current-page="query.page" v-model:page-size="query.per_page" layout="total, sizes, prev, pager, next" :total="total" @change="loadOrders" /></div>
    </el-card>
    <el-drawer v-model="detailVisible" title="订单详情" size="520px">
      <div v-loading="detailLoading" v-if="detail" class="detail">
        <h3>{{ detail.order_no }}</h3><el-tag>{{ statusText[detail.status] || detail.status }}</el-tag>
        <el-descriptions :column="1" border class="mt"><el-descriptions-item label="用户">{{ detail.user.username }} / {{ detail.user.email }}</el-descriptions-item><el-descriptions-item label="产品">{{ detail.product_name }}</el-descriptions-item><el-descriptions-item label="套餐">{{ detail.plan_name }}（{{ detail.cpu_cores }}核 / {{ Math.round(detail.memory_mb / 1024) }}GB / {{ detail.bandwidth_mbps }}M）</el-descriptions-item><el-descriptions-item label="地域">{{ detail.region_name }}</el-descriptions-item><el-descriptions-item label="系统">{{ detail.template_name }}</el-descriptions-item><el-descriptions-item label="金额">{{ formatMoney(detail.total_amount_cents) }}</el-descriptions-item><el-descriptions-item label="用户备注">{{ detail.user_note || '-' }}</el-descriptions-item></el-descriptions>
        <div class="mt"><div class="strong">后台备注</div><el-input v-model="noteDraft" type="textarea" :rows="4" :disabled="!canUpdate" /><el-button v-if="canUpdate" class="mt" type="primary" @click="saveNote">保存备注</el-button></div>
        <div v-if="detail.status === 'pending'" class="mt actions"><el-button v-if="canCancel" type="danger" @click="updateStatus('cancel', detail)">取消订单</el-button><el-button v-if="canUpdate" type="warning" @click="updateStatus('close', detail)">关闭订单</el-button></div>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-header h2 { margin: 0; font-size: 20px; }
.page-header p, .muted { color: var(--el-text-color-secondary); font-size: 12px; }
.query-form { margin-bottom: 16px; }
.strong { font-weight: 700; }
.pagination { display: flex; justify-content: flex-end; margin-top: 16px; }
.detail h3 { margin: 0 0 12px; }
.mt { margin-top: 16px; }
.actions { display: flex; gap: 12px; }
</style>
