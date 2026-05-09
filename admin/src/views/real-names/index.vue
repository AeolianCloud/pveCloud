<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'

import { getRealNameApplication, getRealNameApplications, reviewRealNameApplication, syncRealNameApplication, type RealNameApplicationItem } from '../../api/real-name'
import { hasPermissionCode } from '../../utils/permission'
import { usePermissionStore } from '../../store/modules/permission'

const permissionStore = usePermissionStore()
const loading = ref(false)
const detailLoading = ref(false)
const detailVisible = ref(false)
const applications = ref<RealNameApplicationItem[]>([])
const current = ref<RealNameApplicationItem | null>(null)
const total = ref(0)
const query = reactive({
  page: 1,
  per_page: 15,
  keyword: '',
  status: '',
  provider: '',
  provider_status: '',
})

const statusMap: Record<string, { label: string; type: 'warning' | 'success' | 'danger' }> = {
  pending: { label: '核验中', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已拒绝', type: 'danger' },
}

const providerMap: Record<string, string> = {
  alipay: '支付宝',
  wechat: '微信',
  manual: '人工审核',
}

function canSync() {
  return hasPermissionCode(permissionStore.permissionCodes, 'real-name:sync')
}

function canReview() {
  return hasPermissionCode(permissionStore.permissionCodes, 'real-name:review')
}

function isManualPending(row: RealNameApplicationItem) {
  return row.status === 'pending' && row.verification_provider === 'manual'
}

function isExternalPending(row: RealNameApplicationItem) {
  return row.status === 'pending' && row.verification_provider !== 'manual'
}

function providerLabel(provider: string | null) {
  return provider ? (providerMap[provider] || provider) : '-'
}

async function loadApplications() {
  loading.value = true
  try {
    const data = await getRealNameApplications({
      ...query,
      keyword: query.keyword || undefined,
      status: query.status || undefined,
      provider: query.provider || undefined,
      provider_status: query.provider_status || undefined,
    })
    applications.value = data.list
    total.value = data.total
  } finally {
    loading.value = false
  }
}

async function openDetail(row: RealNameApplicationItem) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    current.value = await getRealNameApplication(row.id)
  } finally {
    detailLoading.value = false
  }
}

async function sync(row: RealNameApplicationItem) {
  try {
    await syncRealNameApplication(row.id)
    ElMessage.success('已触发供应商结果同步')
  } catch (error) {
    const message = error instanceof Error && error.message.trim() ? error.message : '请稍后重试'
    ElMessage.error(`同步失败：${message}`)
    return
  }
  if (current.value?.id === row.id) {
    current.value = await getRealNameApplication(row.id)
  }
  await loadApplications()
}

async function review(row: RealNameApplicationItem, status: 'approved' | 'rejected') {
  let reason = ''
  if (status === 'rejected') {
    try {
      const result = await ElMessageBox.prompt('请输入拒绝原因', '拒绝人工实名申请', {
        confirmButtonText: '拒绝',
        cancelButtonText: '取消',
        inputPattern: /.+/,
        inputErrorMessage: '拒绝原因不能为空',
      })
      reason = result.value.trim()
    } catch {
      return
    }
  } else {
    try {
      await ElMessageBox.confirm('确认通过该人工实名申请？', '通过人工实名申请', {
        confirmButtonText: '通过',
        cancelButtonText: '取消',
        type: 'warning',
      })
    } catch {
      return
    }
  }
  try {
    await reviewRealNameApplication(row.id, { status, reason })
    ElMessage.success(status === 'approved' ? '已通过人工实名申请' : '已拒绝人工实名申请')
  } catch (error) {
    const message = error instanceof Error && error.message.trim() ? error.message : '请稍后重试'
    ElMessage.error(`审核失败：${message}`)
    return
  }
  if (current.value?.id === row.id) {
    current.value = await getRealNameApplication(row.id)
  }
  await loadApplications()
}

onMounted(loadApplications)
</script>

<template>
  <div class="real-name-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <strong>实名管理</strong>
          <span>查看支付宝/微信侧个人实名申请，并在需要时触发结果同步。</span>
        </div>
      </template>
      <div class="toolbar">
        <el-input v-model="query.keyword" clearable placeholder="搜索用户、邮箱、姓名或申请号" style="width: 260px" @keyup.enter="loadApplications" />
        <el-select v-model="query.status" clearable placeholder="实名状态" style="width: 120px">
          <el-option label="核验中" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-select v-model="query.provider" clearable placeholder="供应商" style="width: 120px">
          <el-option label="支付宝" value="alipay" />
          <el-option label="微信" value="wechat" />
          <el-option label="人工审核" value="manual" />
        </el-select>
        <el-input v-model="query.provider_status" clearable placeholder="供应商状态" style="width: 140px" @keyup.enter="loadApplications" />
        <el-button type="primary" @click="loadApplications">查询</el-button>
      </div>
      <el-table v-loading="loading" :data="applications" border>
        <el-table-column prop="application_no" label="申请号" min-width="170" />
        <el-table-column label="用户" min-width="160">
          <template #default="{ row }">{{ row.user.username }} / {{ row.user.email }}</template>
        </el-table-column>
        <el-table-column prop="real_name" label="真实姓名" width="120" />
        <el-table-column prop="id_number_masked" label="证件号码" min-width="180" />
        <el-table-column label="供应商" width="100">
          <template #default="{ row }">{{ providerLabel(row.verification_provider) }}</template>
        </el-table-column>
        <el-table-column prop="provider_status" label="供应商状态" min-width="120" />
        <el-table-column label="实名状态" width="110">
          <template #default="{ row }"><el-tag :type="statusMap[row.status]?.type">{{ statusMap[row.status]?.label || row.status }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" min-width="170" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="isExternalPending(row) && canSync()" link type="success" @click="sync(row)">同步结果</el-button>
            <el-button v-if="isManualPending(row) && canReview()" link type="success" @click="review(row, 'approved')">通过</el-button>
            <el-button v-if="isManualPending(row) && canReview()" link type="danger" @click="review(row, 'rejected')">拒绝</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="query.page" v-model:page-size="query.per_page" :total="total" layout="total, prev, pager, next, sizes" @current-change="loadApplications" @size-change="loadApplications" />
    </el-card>

    <el-drawer v-model="detailVisible" title="实名申请详情" size="520px">
      <div v-loading="detailLoading" v-if="current" class="detail-list">
        <p><b>申请号：</b>{{ current.application_no }}</p>
        <p><b>用户：</b>{{ current.user.username }} / {{ current.user.email }}</p>
        <p><b>真实姓名：</b>{{ current.real_name }}</p>
        <p><b>证件类型：</b>{{ current.id_type }}</p>
        <p><b>证件号码：</b>{{ current.id_number_masked }}</p>
        <p><b>实名供应商：</b>{{ providerLabel(current.verification_provider) }}</p>
        <p><b>供应商会话：</b>{{ current.provider_application_id || '-' }}</p>
        <p><b>供应商状态：</b>{{ current.provider_status || '-' }}</p>
        <p><b>供应商结果码：</b>{{ current.provider_result_code || '-' }}</p>
        <p><b>供应商结果说明：</b>{{ current.provider_result_message || '-' }}</p>
        <p><b>链路号：</b>{{ current.provider_trace_id || '-' }}</p>
        <p><b>实名状态：</b>{{ statusMap[current.status]?.label || current.status }}</p>
        <p v-if="current.failure_reason"><b>失败原因：</b>{{ current.failure_reason }}</p>
        <p><b>核验开始时间：</b>{{ current.provider_started_at || '-' }}</p>
        <p><b>核验完成时间：</b>{{ current.provider_finished_at || '-' }}</p>
        <div v-if="isExternalPending(current) && canSync()" class="drawer-actions">
          <el-button type="success" @click="sync(current)">同步供应商结果</el-button>
        </div>
        <div v-if="isManualPending(current) && canReview()" class="drawer-actions">
          <el-button type="success" @click="review(current, 'approved')">通过人工审核</el-button>
          <el-button type="danger" @click="review(current, 'rejected')">拒绝人工审核</el-button>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.real-name-page { display: grid; gap: 16px; }
.card-header { display: flex; align-items: baseline; gap: 12px; }
.card-header span { color: var(--el-text-color-secondary); font-size: 13px; }
.toolbar { display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 16px; }
.detail-list { display: grid; gap: 10px; }
.detail-list p { margin: 0; }
.drawer-actions { margin-top: 24px; display: flex; gap: 12px; }
</style>
