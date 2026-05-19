<script setup lang="ts">
import {
  NButton,
  NCard,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NTabPane,
  NTable,
  NTag,
  NTabs,
} from 'naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'

import {
  createInstanceMapping,
  getInstanceDetail,
  getInstanceMappings,
  getInstances,
  getPveNodeVMs,
  getPveNodes,
  getPveStorage,
  releaseInstance,
  startInstance,
  stopInstance,
  syncInstance,
  updateInstanceExpiresAt,
  updateInstanceMapping,
  type InstanceDetail,
  type InstanceItem,
  type InstanceMappingItem,
  type InstanceMappingPayload,
  type PveNode,
  type PveStorage,
  type PveVM,
} from '../../api/instance'
import { usePermissionStore } from '../../store/modules/permission'
import { formatDateTime } from '../../utils/datetime'
import { confirm, message } from '../../utils/feedback'
import { hasPermissionCode } from '../../utils/permission'
import InstancesTab from './components/InstancesTab.vue'
import McpResourcesTab from './components/McpResourcesTab.vue'
import ProvisionMappingsTab from './components/ProvisionMappingsTab.vue'
import {
  instanceStatusText,
  makeEmptyMappingForm,
  operationActionText,
  operationStatusText,
  type InstanceTabKey,
  type MappingDialogMode,
} from './types'

const permissionStore = usePermissionStore()

const activeTab = ref<InstanceTabKey>('instances')
const instanceLoading = ref(false)
const mappingLoading = ref(false)
const mcpLoading = ref(false)
const detailLoading = ref(false)
const detailVisible = ref(false)
const expiresAtVisible = ref(false)
const mappingVisible = ref(false)
const mappingMode = ref<MappingDialogMode>('create')
const mappingEditId = ref<number | null>(null)
const selectedNode = ref('')

const instances = ref<InstanceItem[]>([])
const instanceTotal = ref(0)
const mappings = ref<InstanceMappingItem[]>([])
const mappingTotal = ref(0)
const detail = ref<InstanceDetail | null>(null)
const pveNodes = ref<PveNode[]>([])
const pveStorage = ref<PveStorage[]>([])
const pveVMs = ref<PveVM[]>([])
const mappingForm = reactive<InstanceMappingPayload>(makeEmptyMappingForm())
const expiresAtValue = ref<number | null>(null)

const instanceQuery = reactive({ page: 1, per_page: 15, status: '', instance_no: '', order_no: '', user_keyword: '', date_from: '', date_to: '' })
const mappingQuery = reactive({ page: 1, per_page: 15, status: '', plan_no: '', region_no: '', template_no: '', network_type_no: '' })

const canProvision = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'instance:provision'))
const canOperate = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'instance:operate'))
const canRelease = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'instance:release'))
const canSync = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'instance:sync'))
const canRenew = computed(() => hasPermissionCode(permissionStore.permissionCodes, 'instance:renew'))

const mappingStatusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
]

const memoryText = (mb: number) => (mb >= 1024 ? `${Math.round(mb / 1024)}GB` : `${mb}MB`)

async function loadInstances() {
  instanceLoading.value = true
  try {
    const data = await getInstances(instanceQuery)
    instances.value = data.list
    instanceTotal.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '实例加载失败')
  } finally {
    instanceLoading.value = false
  }
}

async function loadMappings() {
  mappingLoading.value = true
  try {
    const data = await getInstanceMappings(mappingQuery)
    mappings.value = data.list
    mappingTotal.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '交付映射加载失败')
  } finally {
    mappingLoading.value = false
  }
}

async function openDetail(instanceNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    detail.value = await getInstanceDetail(instanceNo)
  } catch (err) {
    message.error(err instanceof Error ? err.message : '实例详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

async function operateInstance(action: 'start' | 'stop' | 'release' | 'sync', item: InstanceItem) {
  const labelMap = { start: '开机', stop: '关机', release: '释放', sync: '同步' }
  const label = labelMap[action]
  if (action === 'release') {
    try {
      await confirm({ title: '释放实例', content: `确认释放实例 ${item.instance_no}？释放会删除上游虚拟机。`, type: 'error', positiveText: '确认释放' })
    } catch {
      return
    }
  }
  try {
    const api = action === 'start' ? startInstance : action === 'stop' ? stopInstance : action === 'release' ? releaseInstance : syncInstance
    const updated = await api(item.instance_no)
    message.success(`实例${label}已提交`)
    await loadInstances()
    if (detailVisible.value) detail.value = updated
  } catch (err) {
    message.error(err instanceof Error ? err.message : `${label}失败`)
  }
}

function openExpiresAtModal() {
  if (!detail.value) return
  expiresAtValue.value = detail.value.expires_at ? new Date(detail.value.expires_at).getTime() : Date.now() + 30 * 24 * 60 * 60 * 1000
  expiresAtVisible.value = true
}

async function updateExpiresAt() {
  if (!detail.value || !expiresAtValue.value) {
    message.error('请选择新的到期时间')
    return false
  }
  if (expiresAtValue.value <= Date.now()) {
    message.error('到期时间必须晚于当前时间')
    return false
  }
  try {
    detail.value = await updateInstanceExpiresAt(detail.value.instance_no, new Date(expiresAtValue.value).toISOString())
    message.success('实例到期时间已更新')
    expiresAtVisible.value = false
    await loadInstances()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '到期时间更新失败')
  }
}

function resetInstanceQuery() {
  Object.assign(instanceQuery, { page: 1, per_page: 15, status: '', instance_no: '', order_no: '', user_keyword: '', date_from: '', date_to: '' })
  void loadInstances()
}

function resetMappingQuery() {
  Object.assign(mappingQuery, { page: 1, per_page: 15, status: '', plan_no: '', region_no: '', template_no: '', network_type_no: '' })
  void loadMappings()
}

function normalizeOptional(value: string | null | undefined) {
  const text = String(value || '').trim()
  return text ? text : null
}

function toMappingPayload(): InstanceMappingPayload {
  return {
    ...mappingForm,
    mapping_no: String(mappingForm.mapping_no || '').trim(),
    product_no: normalizeOptional(mappingForm.product_no),
    plan_no: String(mappingForm.plan_no || '').trim(),
    region_no: String(mappingForm.region_no || '').trim(),
    template_no: String(mappingForm.template_no || '').trim(),
    network_type_no: String(mappingForm.network_type_no || '').trim(),
    node: String(mappingForm.node || '').trim(),
    storage: String(mappingForm.storage || '').trim(),
    disk_source: String(mappingForm.disk_source || '').trim(),
    disk_format: normalizeOptional(mappingForm.disk_format),
    disk_interface: normalizeOptional(mappingForm.disk_interface),
    snippets_storage: normalizeOptional(mappingForm.snippets_storage),
    ci_user: normalizeOptional(mappingForm.ci_user),
    ssh_keys: normalizeOptional(mappingForm.ssh_keys),
    ip_config0: normalizeOptional(mappingForm.ip_config0),
    nameserver: normalizeOptional(mappingForm.nameserver),
    search_domain: normalizeOptional(mappingForm.search_domain),
    ci_packages: normalizeOptional(mappingForm.ci_packages),
    apt_mirror: normalizeOptional(mappingForm.apt_mirror),
    remark: normalizeOptional(mappingForm.remark),
  }
}

function openCreateMapping() {
  mappingMode.value = 'create'
  mappingEditId.value = null
  Object.assign(mappingForm, makeEmptyMappingForm())
  mappingVisible.value = true
}

function openEditMapping(item: InstanceMappingItem) {
  mappingMode.value = 'edit'
  mappingEditId.value = item.id
  Object.assign(mappingForm, {
    mapping_no: item.mapping_no,
    product_no: item.product_no,
    plan_no: item.plan_no,
    region_no: item.region_no,
    template_no: item.template_no,
    network_type_no: item.network_type_no,
    node: item.node,
    storage: item.storage,
    disk_source: item.disk_source,
    disk_format: item.disk_format,
    disk_interface: item.disk_interface,
    snippets_storage: item.snippets_storage,
    ci_user: item.ci_user,
    ssh_keys: item.ssh_keys,
    ip_config0: item.ip_config0,
    nameserver: item.nameserver,
    search_domain: item.search_domain,
    ci_packages: item.ci_packages,
    apt_mirror: item.apt_mirror,
    vmid_start: item.vmid_start,
    vmid_end: item.vmid_end,
    next_vmid: item.next_vmid,
    status: item.status,
    remark: item.remark,
  })
  mappingVisible.value = true
}

async function saveMapping() {
  try {
    const payload = toMappingPayload()
    if (mappingMode.value === 'edit' && mappingEditId.value != null) {
      await updateInstanceMapping(mappingEditId.value, payload)
    } else {
      await createInstanceMapping(payload)
    }
    message.success('交付映射已保存')
    mappingVisible.value = false
    await loadMappings()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '交付映射保存失败')
  }
}

async function loadPveNodes() {
  mcpLoading.value = true
  try {
    pveNodes.value = await getPveNodes()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '节点加载失败')
  } finally {
    mcpLoading.value = false
  }
}

async function loadPveStorage() {
  mcpLoading.value = true
  try {
    pveStorage.value = await getPveStorage()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '存储加载失败')
  } finally {
    mcpLoading.value = false
  }
}

async function loadPveVMs() {
  if (!selectedNode.value.trim()) {
    message.warning('请先输入节点名称')
    return
  }
  mcpLoading.value = true
  try {
    pveVMs.value = await getPveNodeVMs(selectedNode.value.trim())
  } catch (err) {
    message.error(err instanceof Error ? err.message : '虚拟机列表加载失败')
  } finally {
    mcpLoading.value = false
  }
}

onMounted(() => {
  void loadInstances()
  void loadMappings()
})
</script>

<template>
  <div class="instances-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>实例管理</h2>
          <p class="muted">维护交付映射、查看实例、同步虚拟化状态并执行当前开放的电源操作。</p>
        </div>
      </template>

      <NTabs v-model:value="activeTab" type="line" animated>
        <NTabPane name="instances" tab="实例列表">
          <InstancesTab
            :loading="instanceLoading"
            :items="instances"
            :total="instanceTotal"
            :query="instanceQuery"
            :can-operate="canOperate"
            :can-release="canRelease"
            :can-sync="canSync"
            @search="loadInstances"
            @reset="resetInstanceQuery"
            @detail="openDetail"
            @start="operateInstance('start', $event)"
            @stop="operateInstance('stop', $event)"
            @release="operateInstance('release', $event)"
            @sync="operateInstance('sync', $event)"
          />
        </NTabPane>
        <NTabPane name="mappings" tab="交付映射">
          <ProvisionMappingsTab
            :loading="mappingLoading"
            :items="mappings"
            :total="mappingTotal"
            :query="mappingQuery"
            :can-provision="canProvision"
            @search="loadMappings"
            @reset="resetMappingQuery"
            @create="openCreateMapping"
            @edit="openEditMapping"
          />
        </NTabPane>
        <NTabPane name="mcp" tab="MCP 只读资源">
          <McpResourcesTab
            v-model:selected-node="selectedNode"
            :loading="mcpLoading"
            :nodes="pveNodes"
            :storage="pveStorage"
            :vms="pveVMs"
            @load-nodes="loadPveNodes"
            @load-storage="loadPveStorage"
            @load-vms="loadPveVMs"
          />
        </NTabPane>
      </NTabs>
    </NCard>

    <NDrawer v-model:show="detailVisible" :width="620">
      <NDrawerContent title="实例详情" closable>
        <div v-if="detail" class="detail">
          <div class="detail-head">
            <div>
              <h3>{{ detail.instance_no }}</h3>
              <p class="muted">{{ detail.product_name }} · {{ detail.plan_name }}</p>
            </div>
            <NTag :type="detail.status === 'error' ? 'error' : detail.status === 'running' ? 'success' : 'default'">
              {{ instanceStatusText[detail.status] || detail.status }}
            </NTag>
          </div>
          <NDescriptions :column="1" bordered class="mt" label-placement="left" size="small">
            <NDescriptionsItem label="用户">{{ detail.user.username }} / {{ detail.user.email }}</NDescriptionsItem>
            <NDescriptionsItem label="订单">{{ detail.order_no }}</NDescriptionsItem>
            <NDescriptionsItem label="规格">{{ detail.cpu_cores }} 核 / {{ memoryText(detail.memory_mb) }} / {{ detail.system_disk_gb + detail.data_disk_gb }}GB / {{ detail.bandwidth_mbps }}M</NDescriptionsItem>
            <NDescriptionsItem label="地域">{{ detail.region_name }}</NDescriptionsItem>
            <NDescriptionsItem label="系统">{{ detail.template_name }} · {{ detail.os_distribution }} {{ detail.os_version }}</NDescriptionsItem>
            <NDescriptionsItem label="上游资源">{{ detail.external_node }} / {{ detail.external_vmid }}</NDescriptionsItem>
            <NDescriptionsItem label="服务开始">{{ formatDateTime(detail.service_started_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="到期时间">{{ formatDateTime(detail.expires_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="到期提醒">{{ formatDateTime(detail.expire_notice_sent_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="自动释放计划">{{ formatDateTime(detail.expire_release_scheduled_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="因到期释放">{{ formatDateTime(detail.expire_released_at) }}</NDescriptionsItem>
            <NDescriptionsItem label="最近续费">
              <span v-if="detail.latest_renewal_order">
                {{ detail.latest_renewal_order.order_no }} / {{ detail.latest_renewal_order.payment_status }}
              </span>
              <span v-else>-</span>
            </NDescriptionsItem>
            <NDescriptionsItem label="最近错误">{{ detail.last_error_message || '-' }}</NDescriptionsItem>
          </NDescriptions>
          <div class="detail-actions">
            <NSpace>
              <NButton v-if="canOperate && detail.status === 'stopped'" type="success" @click="operateInstance('start', detail)">开机</NButton>
              <NButton v-if="canOperate && detail.status === 'running'" type="warning" @click="operateInstance('stop', detail)">关机</NButton>
              <NButton v-if="canSync" @click="operateInstance('sync', detail)">同步</NButton>
              <NButton v-if="canRenew && detail.status !== 'released'" @click="openExpiresAtModal">调整到期</NButton>
              <NButton v-if="canRelease && detail.status !== 'released' && detail.status !== 'releasing'" type="error" @click="operateInstance('release', detail)">释放</NButton>
            </NSpace>
          </div>
          <h4 class="mt">操作记录</h4>
          <NTable size="small" :bordered="false">
            <thead><tr><th>操作</th><th>状态</th><th>创建时间</th><th>错误</th></tr></thead>
            <tbody>
              <tr v-for="op in detail.operations" :key="op.operation_no">
                <td>{{ operationActionText[op.action] || op.action }}</td>
                <td>{{ operationStatusText[op.status] || op.status }}</td>
                <td>{{ formatDateTime(op.created_at) }}</td>
                <td>{{ op.error_message || '-' }}</td>
              </tr>
              <tr v-if="detail.operations.length === 0"><td colspan="4">暂无操作记录</td></tr>
            </tbody>
          </NTable>
        </div>
        <div v-else-if="detailLoading">加载中...</div>
      </NDrawerContent>
    </NDrawer>

    <NModal
      v-model:show="expiresAtVisible"
      preset="dialog"
      title="调整到期时间"
      positive-text="保存"
      negative-text="取消"
      @positive-click="updateExpiresAt"
    >
      <NDatePicker v-model:value="expiresAtValue" type="datetime" clearable style="width: 100%" />
    </NModal>

    <NDrawer v-model:show="mappingVisible" :width="720">
      <NDrawerContent :title="mappingMode === 'create' ? '新增交付映射' : '编辑交付映射'" closable>
        <NForm label-placement="left" label-width="110" class="mapping-form">
          <NFormItem label="映射编号"><NInput v-model:value="mappingForm.mapping_no" placeholder="留空自动生成" /></NFormItem>
          <NFormItem label="产品编号"><NInput v-model:value="mappingForm.product_no" placeholder="可选" /></NFormItem>
          <NFormItem label="套餐编号"><NInput v-model:value="mappingForm.plan_no" placeholder="必填" /></NFormItem>
          <NFormItem label="地域编号"><NInput v-model:value="mappingForm.region_no" placeholder="必填" /></NFormItem>
          <NFormItem label="模板编号"><NInput v-model:value="mappingForm.template_no" placeholder="必填" /></NFormItem>
          <NFormItem label="网络类型"><NInput v-model:value="mappingForm.network_type_no" placeholder="留空表示不限" /></NFormItem>
          <NFormItem label="节点"><NInput v-model:value="mappingForm.node" placeholder="节点名称" /></NFormItem>
          <NFormItem label="存储"><NInput v-model:value="mappingForm.storage" placeholder="目标存储池" /></NFormItem>
          <NFormItem label="磁盘来源"><NInput v-model:value="mappingForm.disk_source" placeholder="存储池:路径" /></NFormItem>
          <NFormItem label="磁盘格式"><NInput v-model:value="mappingForm.disk_format" placeholder="磁盘格式，可选" /></NFormItem>
          <NFormItem label="磁盘接口"><NInput v-model:value="mappingForm.disk_interface" placeholder="磁盘接口，可选" /></NFormItem>
          <NFormItem label="片段存储"><NInput v-model:value="mappingForm.snippets_storage" placeholder="可选" /></NFormItem>
          <NFormItem label="默认用户"><NInput v-model:value="mappingForm.ci_user" placeholder="可选" /></NFormItem>
          <NFormItem label="SSH 公钥"><NInput v-model:value="mappingForm.ssh_keys" type="textarea" :rows="3" placeholder="可选" /></NFormItem>
          <NFormItem label="网络配置"><NInput v-model:value="mappingForm.ip_config0" placeholder="ip=dhcp 或静态配置，可选" /></NFormItem>
          <NFormItem label="DNS"><NInput v-model:value="mappingForm.nameserver" placeholder="可选" /></NFormItem>
          <NFormItem label="搜索域"><NInput v-model:value="mappingForm.search_domain" placeholder="可选" /></NFormItem>
          <NFormItem label="软件包"><NInput v-model:value="mappingForm.ci_packages" type="textarea" :rows="2" placeholder="请输入字符串数组，例如常用初始化软件包" /></NFormItem>
          <NFormItem label="镜像源"><NInput v-model:value="mappingForm.apt_mirror" placeholder="可选" /></NFormItem>
          <NFormItem label="编号范围">
            <NSpace>
              <NInputNumber v-model:value="mappingForm.vmid_start" :min="1" placeholder="起始" />
              <NInputNumber v-model:value="mappingForm.vmid_end" :min="1" placeholder="结束" />
              <NInputNumber v-model:value="mappingForm.next_vmid" :min="1" placeholder="下一个" />
            </NSpace>
          </NFormItem>
          <NFormItem label="状态"><NSelect v-model:value="mappingForm.status" :options="mappingStatusOptions" /></NFormItem>
          <NFormItem label="备注"><NInput v-model:value="mappingForm.remark" type="textarea" :rows="3" placeholder="可选" /></NFormItem>
        </NForm>
        <template #footer>
          <NSpace justify="end">
            <NButton @click="mappingVisible = false">取消</NButton>
            <NButton type="primary" @click="saveMapping">保存</NButton>
          </NSpace>
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
.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.detail-head h3 {
  margin: 0 0 6px;
}
.mt {
  margin-top: 16px;
}
.detail-actions {
  margin-top: 16px;
}
.mapping-form {
  padding-right: 8px;
}
</style>

<style>
.instances-page .query-form {
  margin-bottom: 16px;
}
.instances-page .strong {
  font-weight: 700;
}
.instances-page .muted {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}
.instances-page .pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.instances-page .toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 8px;
}
.instances-page .mcp-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}
@media (max-width: 1100px) {
  .instances-page .toolbar {
    flex-direction: column;
  }
  .instances-page .mcp-grid {
    grid-template-columns: 1fr;
  }
}
</style>
