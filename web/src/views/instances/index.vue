<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { getInstances, startInstance, stopInstance, type InstanceItem } from '../../api/instance'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'

const confirmDialog = useConfirm()
const toast = useToast()
const loading = ref(false)
const operatingNo = ref('')
const errorMessage = ref('')
const instances = ref<InstanceItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, per_page: 15, status: '' })

const statusText: Record<string, string> = {
  creating: '创建中',
  running: '运行中',
  stopped: '已停止',
  error: '异常',
  releasing: '释放中',
  released: '已释放',
}
const expireStatusText: Record<string, string> = { active: '服务中', expired: '已到期', released: '已释放', unknown: '未开始' }
const statusOptions = [
  { label: '全部', value: '' },
  { label: '创建中', value: 'creating' },
  { label: '运行中', value: 'running' },
  { label: '已停止', value: 'stopped' },
  { label: '异常', value: 'error' },
  { label: '释放中', value: 'releasing' },
  { label: '已释放', value: 'released' },
]

async function loadInstances() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getInstances(query)
    instances.value = data.list
    total.value = data.total
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '实例加载失败')
  } finally {
    loading.value = false
  }
}

async function operate(item: InstanceItem, action: 'start' | 'stop') {
  const label = action === 'start' ? '启动' : '停止'
  const confirmed = await confirmDialog.confirm({
    title: `${label}实例`,
    message: `确认${label}实例 ${item.instance_no}？`,
    confirmText: `确认${label}`,
    cancelText: '取消',
    tone: action === 'stop' ? 'danger' : 'default',
  })
  if (!confirmed) return
  operatingNo.value = item.instance_no
  try {
    if (action === 'start') {
      await startInstance(item.instance_no)
    } else {
      await stopInstance(item.instance_no)
    }
    toast.success(`实例${label}已提交`)
    await loadInstances()
  } catch (err) {
    toast.error(getApiErrorMessage(err, `${label}失败`))
  } finally {
    operatingNo.value = ''
  }
}

onMounted(loadInstances)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="mb-8 flex flex-col justify-between gap-4 border-b border-neutral-200 pb-8 md:flex-row md:items-end">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">实例</p>
          <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">我的实例</h1>
          <p class="mt-3 text-sm text-neutral-500">查看已交付云主机并执行基础电源操作。</p>
        </div>
        <RouterLink to="/user/orders" class="action-pill border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">查看订单</RouterLink>
      </div>

      <div class="mb-6 flex flex-wrap gap-3">
        <button
          v-for="item in statusOptions"
          :key="item.value || 'all'"
          type="button"
          :class="['action-pill border px-4 py-2 text-xs font-black', query.status === item.value ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-neutral-300 text-neutral-700 hover:border-neutral-950']"
          @click="query.status = item.value; query.page = 1; loadInstances()"
        >
          {{ item.label }}
        </button>
      </div>

      <div v-if="loading" class="space-y-3">
        <div v-for="item in 4" :key="item" class="rounded-2xl border border-neutral-200 bg-white p-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_8rem_12rem] lg:items-center">
            <div>
              <div class="skeleton-line h-3 w-36"></div>
              <div class="skeleton-line mt-3 h-5 w-64 max-w-full"></div>
              <div class="skeleton-line mt-3 h-3 w-80 max-w-full"></div>
            </div>
            <div class="skeleton-line h-8 w-24 lg:ml-auto"></div>
            <div class="skeleton-line h-8 w-36 lg:ml-auto"></div>
          </div>
        </div>
      </div>

      <div v-else-if="errorMessage" class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-red-600">实例异常</p>
        <h2 class="mt-3 text-2xl font-black text-neutral-950">实例加载失败</h2>
        <p class="mx-auto mt-3 max-w-xl text-sm leading-6 text-neutral-500">{{ errorMessage }}</p>
        <button type="button" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white" @click="loadInstances">重新加载</button>
      </div>

      <div v-else-if="instances.length" class="space-y-3">
        <article v-for="item in instances" :key="item.instance_no" class="soft-lift rounded-2xl border border-neutral-200 bg-white p-4 sm:p-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_8rem_12rem] lg:items-center">
            <div class="min-w-0">
              <div class="truncate text-[11px] font-black uppercase tracking-[0.14em] text-neutral-500">{{ item.instance_no }}</div>
              <h2 class="mt-1 truncate text-base font-black text-neutral-950 sm:text-lg">{{ item.product_name }} · {{ item.plan_name }}</h2>
              <p class="mt-1 truncate text-xs text-neutral-500 sm:text-sm">{{ item.region_name }} · {{ item.template_name }} · {{ item.created_at }}</p>
              <p class="mt-1 text-xs font-bold text-neutral-500">到期：{{ item.expires_at || '-' }} · {{ expireStatusText[item.expire_status] || item.expire_status }}</p>
            </div>
            <div class="flex items-center justify-between gap-3 lg:block lg:text-right">
              <span class="inline-flex rounded-full border border-neutral-300 px-3 py-1 text-xs font-black">{{ statusText[item.status] }}</span>
            </div>
            <div class="flex flex-wrap gap-2 lg:justify-end">
              <RouterLink :to="`/user/instances/${item.instance_no}`" class="action-pill border border-neutral-950 px-3 py-1.5 text-xs font-black hover:bg-neutral-950 hover:text-white">查看详情</RouterLink>
              <button v-if="item.status === 'stopped'" type="button" class="action-pill border border-emerald-300 px-3 py-1.5 text-xs font-black text-emerald-700 hover:bg-emerald-50 disabled:opacity-50" :disabled="operatingNo === item.instance_no" @click="operate(item, 'start')">启动</button>
              <button v-if="item.status === 'running'" type="button" class="action-pill border border-amber-300 px-3 py-1.5 text-xs font-black text-amber-700 hover:bg-amber-50 disabled:opacity-50" :disabled="operatingNo === item.instance_no" @click="operate(item, 'stop')">停止</button>
            </div>
          </div>
        </article>
      </div>

      <div v-else class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">暂无实例</p>
        <h2 class="mt-3 text-2xl font-black">暂无实例</h2>
        <p class="mt-3 text-sm text-neutral-500">订单交付完成后会在这里展示云主机实例。</p>
        <RouterLink to="/user/orders" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white">查看订单</RouterLink>
      </div>

      <div v-if="total > query.per_page" class="mt-6 flex justify-center gap-3">
        <button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="query.page <= 1" @click="query.page--; loadInstances()">上一页</button>
        <button type="button" class="action-pill border px-4 py-2 text-sm font-black disabled:opacity-40" :disabled="instances.length < query.per_page" @click="query.page++; loadInstances()">下一页</button>
      </div>
    </div>
  </div>
</template>
