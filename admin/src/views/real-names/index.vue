<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NInput,
  NPagination,
  NSelect,
  NSpace,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import {
  getRealNameApplication,
  getRealNameApplications,
  reviewRealNameApplication,
  syncRealNameApplication,
  type RealNameApplicationItem,
} from '../../api/real-name'
import { hasPermissionCode } from '../../utils/permission'
import { usePermissionStore } from '../../store/modules/permission'
import { confirm, getDialog, message } from '../../utils/feedback'
import { formatDateTime } from '../../utils/datetime'

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

const statusMap: Record<string, { label: string; type: 'warning' | 'success' | 'error' }> = {
  pending: { label: '核验中', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已拒绝', type: 'error' },
}

const providerMap: Record<string, string> = {
  alipay: '支付宝',
  wechat: '微信',
  manual: '人工审核',
}

const statusOptions = [
  { label: '核验中', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已拒绝', value: 'rejected' },
]

const providerOptions = [
  { label: '支付宝', value: 'alipay' },
  { label: '微信', value: 'wechat' },
  { label: '人工审核', value: 'manual' },
]

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
  return provider ? providerMap[provider] || provider : '-'
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
    message.success('已触发供应商结果同步')
  } catch (error) {
    const msg = error instanceof Error && error.message.trim() ? error.message : '请稍后重试'
    message.error(`同步失败：${msg}`)
    return
  }
  if (current.value?.id === row.id) {
    current.value = await getRealNameApplication(row.id)
  }
  await loadApplications()
}

function promptReason(title: string, content: string): Promise<string | null> {
  return new Promise((resolve) => {
    const value = ref('')
    const dialog = getDialog()
    const d = dialog.create({
      title,
      content: () =>
        h('div', { style: 'display:grid;gap:8px;' }, [
          h('div', { style: 'color: rgba(15,23,42,0.7); font-size: 13px' }, content),
          h(NInput, {
            value: value.value,
            type: 'textarea',
            rows: 3,
            placeholder: '请输入原因',
            'onUpdate:value': (v: string) => (value.value = v),
          }),
        ]),
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: () => {
        if (!value.value.trim()) {
          message.warning('原因不能为空')
          return false
        }
        resolve(value.value.trim())
        d.destroy()
      },
      onNegativeClick: () => {
        resolve(null)
        d.destroy()
      },
      onClose: () => resolve(null),
    })
  })
}

async function review(row: RealNameApplicationItem, status: 'approved' | 'rejected') {
  let reason = ''
  if (status === 'rejected') {
    const r = await promptReason('拒绝人工实名申请', '请输入拒绝原因')
    if (r === null) return
    reason = r
  } else {
    try {
      await confirm({
        title: '通过人工实名申请',
        content: '确认通过该人工实名申请？',
        type: 'warning',
        positiveText: '通过',
      })
    } catch {
      return
    }
  }
  try {
    await reviewRealNameApplication(row.id, { status, reason })
    message.success(status === 'approved' ? '已通过人工实名申请' : '已拒绝人工实名申请')
  } catch (error) {
    const msg = error instanceof Error && error.message.trim() ? error.message : '请稍后重试'
    message.error(`审核失败：${msg}`)
    return
  }
  if (current.value?.id === row.id) {
    current.value = await getRealNameApplication(row.id)
  }
  await loadApplications()
}

const columns = computed<DataTableColumns<RealNameApplicationItem>>(() => [
  { key: 'application_no', title: '申请号', minWidth: 170 },
  {
    key: 'user',
    title: '用户',
    minWidth: 180,
    render: (row) => `${row.user.username} / ${row.user.email}`,
  },
  { key: 'real_name', title: '真实姓名', width: 120 },
  { key: 'id_number_masked', title: '证件号码', minWidth: 180 },
  {
    key: 'provider',
    title: '供应商',
    width: 100,
    render: (row) => providerLabel(row.verification_provider),
  },
  { key: 'provider_status', title: '供应商状态', minWidth: 140 },
  {
    key: 'status',
    title: '实名状态',
    width: 110,
    render: (row) =>
      h(
        NTag,
        { type: statusMap[row.status]?.type || 'default', size: 'small' },
        { default: () => statusMap[row.status]?.label || row.status },
      ),
  },
  {
    key: 'created_at',
    title: '提交时间',
    minWidth: 170,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    key: 'actions',
    title: '操作',
    width: 220,
    fixed: 'right',
    render: (row) => {
      const buttons: any[] = [
        h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row) }, { default: () => '详情' }),
      ]
      if (isExternalPending(row) && canSync()) {
        buttons.push(
          h(NButton, { text: true, type: 'success', onClick: () => sync(row) }, { default: () => '同步结果' }),
        )
      }
      if (isManualPending(row) && canReview()) {
        buttons.push(
          h(NButton, { text: true, type: 'success', onClick: () => review(row, 'approved') }, { default: () => '通过' }),
          h(NButton, { text: true, type: 'error', onClick: () => review(row, 'rejected') }, { default: () => '拒绝' }),
        )
      }
      return h(NSpace, { size: 6 }, { default: () => buttons })
    },
  },
])

onMounted(loadApplications)
</script>

<template>
  <div class="real-name-page">
    <NCard :bordered="false">
      <template #header>
        <div class="card-header">
          <strong>实名管理</strong>
          <span>查看支付宝/微信侧个人实名申请，并在需要时触发结果同步。</span>
        </div>
      </template>

      <div class="toolbar">
        <NInput
          v-model:value="query.keyword"
          clearable
          placeholder="搜索用户、邮箱、姓名或申请号"
          style="width: 260px"
          @keyup.enter="loadApplications"
        />
        <NSelect v-model:value="query.status" :options="statusOptions" clearable placeholder="实名状态" style="width: 140px" />
        <NSelect v-model:value="query.provider" :options="providerOptions" clearable placeholder="供应商" style="width: 140px" />
        <NInput
          v-model:value="query.provider_status"
          clearable
          placeholder="供应商状态"
          style="width: 140px"
          @keyup.enter="loadApplications"
        />
        <NButton type="primary" @click="loadApplications">查询</NButton>
      </div>

      <NDataTable
        :loading="loading"
        :columns="columns"
        :data="applications"
        :row-key="(row: RealNameApplicationItem) => row.id"
        :bordered="false"
        striped
      />

      <div class="pagination">
        <NPagination
          v-model:page="query.page"
          v-model:page-size="query.per_page"
          :item-count="total"
          :page-sizes="[15, 30, 50, 100]"
          show-size-picker
          @update:page="loadApplications"
          @update:page-size="loadApplications"
        />
      </div>
    </NCard>

    <NDrawer v-model:show="detailVisible" :width="520">
      <NDrawerContent title="实名申请详情" closable>
        <div v-if="current" class="detail-list">
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
          <p><b>核验开始时间：</b>{{ formatDateTime(current.provider_started_at) }}</p>
          <p><b>核验完成时间：</b>{{ formatDateTime(current.provider_finished_at) }}</p>
          <div v-if="isExternalPending(current) && canSync()" class="drawer-actions">
            <NButton type="success" @click="sync(current)">同步供应商结果</NButton>
          </div>
          <div v-if="isManualPending(current) && canReview()" class="drawer-actions">
            <NSpace>
              <NButton type="success" @click="review(current, 'approved')">通过人工审核</NButton>
              <NButton type="error" @click="review(current, 'rejected')">拒绝人工审核</NButton>
            </NSpace>
          </div>
        </div>
        <div v-else-if="detailLoading">加载中...</div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.real-name-page {
  display: grid;
  gap: 16px;
}
.card-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.card-header span {
  color: rgba(15, 23, 42, 0.55);
  font-size: 13px;
}
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.detail-list {
  display: grid;
  gap: 10px;
}
.detail-list p {
  margin: 0;
}
.drawer-actions {
  margin-top: 24px;
  display: flex;
  gap: 12px;
}
</style>
