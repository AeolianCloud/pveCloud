<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'

import { getRealNameApplication, getRealNameApplications, reviewRealNameApplication, type RealNameApplicationItem } from '../../api/real-name'
import { hasPermissionCode } from '../../utils/permission'
import { usePermissionStore } from '../../store/modules/permission'

const permissionStore = usePermissionStore()
const loading = ref(false)
const detailLoading = ref(false)
const detailVisible = ref(false)
const applications = ref<RealNameApplicationItem[]>([])
const current = ref<RealNameApplicationItem | null>(null)
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, keyword: '', status: '' })

const statusMap: Record<string, { label: string; type: 'warning' | 'success' | 'danger' }> = {
  pending: { label: '待审核', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已拒绝', type: 'danger' },
}

function canReview() {
  return hasPermissionCode(permissionStore.permissionCodes, 'real-name:review')
}

async function loadApplications() {
  loading.value = true
  try {
    const data = await getRealNameApplications({ ...query, keyword: query.keyword || undefined, status: query.status || undefined })
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

async function review(row: RealNameApplicationItem, status: 'approved' | 'rejected') {
  let reason = ''
  if (status === 'rejected') {
    const result = await ElMessageBox.prompt('请输入拒绝原因', '拒绝实名申请', { inputType: 'textarea', inputValidator: value => Boolean(value && value.trim()), inputErrorMessage: '拒绝原因不能为空' })
    reason = result.value
  } else {
    await ElMessageBox.confirm('确认通过该实名申请？', '审核确认', { type: 'warning' })
  }
  await reviewRealNameApplication(row.id, status, reason)
  ElMessage.success('审核已提交')
  detailVisible.value = false
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
          <span>审核用户购买机器前需要完成的个人实名申请</span>
        </div>
      </template>
      <div class="toolbar">
        <el-input v-model="query.keyword" clearable placeholder="搜索用户、邮箱、姓名或申请号" style="width: 280px" @keyup.enter="loadApplications" />
        <el-select v-model="query.status" clearable placeholder="状态" style="width: 140px">
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-button type="primary" @click="loadApplications">查询</el-button>
      </div>
      <el-table v-loading="loading" :data="applications" border>
        <el-table-column prop="application_no" label="申请号" min-width="170" />
        <el-table-column label="用户" min-width="160">
          <template #default="{ row }">{{ row.user.username }} / {{ row.user.email }}</template>
        </el-table-column>
        <el-table-column prop="real_name" label="真实姓名" width="120" />
        <el-table-column prop="id_number_masked" label="证件号码" min-width="180" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }"><el-tag :type="statusMap[row.status]?.type">{{ statusMap[row.status]?.label || row.status }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" min-width="170" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 'pending' && canReview()" link type="success" @click="review(row, 'approved')">通过</el-button>
            <el-button v-if="row.status === 'pending' && canReview()" link type="danger" @click="review(row, 'rejected')">拒绝</el-button>
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
        <p><b>状态：</b>{{ statusMap[current.status]?.label || current.status }}</p>
        <p v-if="current.reject_reason"><b>拒绝原因：</b>{{ current.reject_reason }}</p>
        <p><b>人像面：</b>{{ current.id_card_front_file?.original_name || '未上传' }}</p>
        <p><b>国徽面：</b>{{ current.id_card_back_file?.original_name || '未上传' }}</p>
        <p><b>手持证件：</b>{{ current.hold_card_file?.original_name || '未上传' }}</p>
        <div v-if="current.status === 'pending' && canReview()" class="drawer-actions">
          <el-button type="success" @click="review(current, 'approved')">通过</el-button>
          <el-button type="danger" @click="review(current, 'rejected')">拒绝</el-button>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.real-name-page { display: grid; gap: 16px; }
.card-header { display: flex; align-items: baseline; gap: 12px; }
.card-header span { color: var(--el-text-color-secondary); font-size: 13px; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
.detail-list { display: grid; gap: 10px; }
.detail-list p { margin: 0; }
.drawer-actions { margin-top: 24px; display: flex; gap: 12px; }
</style>
