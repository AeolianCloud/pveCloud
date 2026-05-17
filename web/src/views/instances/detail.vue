<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { getInstanceDetail, startInstance, stopInstance, type InstanceDetail } from '../../api/instance'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'

const route = useRoute()
const router = useRouter()
const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const operating = ref(false)
const errorMessage = ref('')
const instance = ref<InstanceDetail | null>(null)

const statusText: Record<string, string> = {
  creating: '创建中',
  running: '运行中',
  stopped: '已停止',
  error: '异常',
  releasing: '释放中',
  released: '已释放',
}
const actionText: Record<string, string> = { provision: '交付', start: '启动', stop: '停止', release: '释放', sync: '同步' }
const operationStatusText: Record<string, string> = { running: '执行中', succeeded: '成功', failed: '失败' }
const formatMemory = (mb: number) => mb >= 1024 ? `${Math.round(mb / 1024)}GB` : `${mb}MB`

async function loadDetail() {
  loading.value = true
  errorMessage.value = ''
  try {
    instance.value = await getInstanceDetail(String(route.params.instanceNo || ''))
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '实例详情加载失败')
  } finally {
    loading.value = false
  }
}

async function operate(action: 'start' | 'stop') {
  if (!instance.value) return
  const label = action === 'start' ? '启动' : '停止'
  const confirmed = await confirmDialog.confirm({
    title: `${label}实例`,
    message: `确认${label}实例 ${instance.value.instance_no}？`,
    confirmText: `确认${label}`,
    cancelText: '取消',
    tone: action === 'stop' ? 'danger' : 'default',
  })
  if (!confirmed) return
  operating.value = true
  try {
    instance.value = action === 'start' ? await startInstance(instance.value.instance_no) : await stopInstance(instance.value.instance_no)
    toast.success(`实例${label}已提交`)
  } catch (err) {
    toast.error(getApiErrorMessage(err, `${label}失败`))
  } finally {
    operating.value = false
  }
}

onMounted(loadDetail)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
      <button type="button" class="mb-6 text-sm font-black text-neutral-600 underline hover:text-neutral-950" @click="router.back()">返回</button>

      <div v-if="loading" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">实例详情加载中...</div>
      <div v-else-if="errorMessage" class="rounded-[1.5rem] border border-red-200 bg-red-50 p-6 text-sm font-bold text-red-700">{{ errorMessage }}</div>

      <article v-else-if="instance" class="rounded-[1.5rem] border border-neutral-200 bg-white p-5 shadow-[8px_8px_0_#111] sm:p-6">
        <div class="grid gap-4 border-b border-neutral-200 pb-5 md:grid-cols-[minmax(0,1fr)_10rem] md:items-start">
          <div class="min-w-0">
            <p class="truncate text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ instance.instance_no }}</p>
            <h1 class="mt-2 text-2xl font-black text-neutral-950">{{ instance.product_name }} · {{ instance.plan_name }}</h1>
            <p class="mt-2 text-sm text-neutral-500">只开放启动和停止，不提供通用虚拟化运维入口。</p>
          </div>
          <div class="flex items-center justify-between gap-3 md:block md:text-right">
            <span class="inline-flex rounded-full border px-3 py-1 text-xs font-black">{{ statusText[instance.status] }}</span>
          </div>
        </div>

        <dl class="mt-6 grid gap-3 md:grid-cols-2">
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">订单编号</dt><dd class="mt-1 text-sm font-black">{{ instance.order_no }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">销售地域</dt><dd class="mt-1 text-sm font-black">{{ instance.region_name }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">系统模板</dt><dd class="mt-1 text-sm font-black">{{ instance.template_name }}</dd></div>
          <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">创建时间</dt><dd class="mt-1 text-sm font-black">{{ instance.created_at }}</dd></div>
        </dl>

        <section class="mt-6">
          <h2 class="text-base font-black">配置快照</h2>
          <div class="mt-3 grid gap-2 text-sm md:grid-cols-4">
            <div class="rounded-xl border p-3">{{ instance.cpu_cores }} 核 CPU</div>
            <div class="rounded-xl border p-3">{{ formatMemory(instance.memory_mb) }} 内存</div>
            <div class="rounded-xl border p-3">{{ instance.system_disk_gb + instance.data_disk_gb }}GB 磁盘</div>
            <div class="rounded-xl border p-3">{{ instance.bandwidth_mbps }}M 带宽</div>
          </div>
        </section>

        <section class="mt-6">
          <h2 class="text-base font-black">最近操作</h2>
          <div v-if="instance.operations.length" class="mt-3 divide-y divide-neutral-200 rounded-xl border border-neutral-200">
            <div v-for="operation in instance.operations" :key="operation.operation_no" class="grid gap-2 p-3 text-sm md:grid-cols-[8rem_8rem_minmax(0,1fr)]">
              <span class="font-black">{{ actionText[operation.action] || operation.action }}</span>
              <span class="text-neutral-600">{{ operationStatusText[operation.status] || operation.status }}</span>
              <span class="text-neutral-500">{{ operation.created_at }}</span>
            </div>
          </div>
          <p v-else class="mt-3 rounded-xl bg-neutral-50 p-3 text-sm text-neutral-500">暂无操作记录</p>
        </section>

        <section v-if="instance.status === 'error'" class="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm font-bold text-red-700">
          实例状态异常，请通过工单联系后台处理。
        </section>

        <div class="mt-6 flex flex-wrap gap-3">
          <button v-if="instance.status === 'stopped'" type="button" class="action-pill border border-emerald-300 px-5 py-2 text-sm font-black text-emerald-700 hover:bg-emerald-50 disabled:opacity-50" :disabled="operating" @click="operate('start')">启动实例</button>
          <button v-if="instance.status === 'running'" type="button" class="action-pill border border-amber-300 px-5 py-2 text-sm font-black text-amber-700 hover:bg-amber-50 disabled:opacity-50" :disabled="operating" @click="operate('stop')">停止实例</button>
          <RouterLink to="/user/instances" class="action-pill border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">返回实例列表</RouterLink>
        </div>
      </article>
    </div>
  </div>
</template>
