<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  getRealNameStatus,
  submitRealName,
  syncRealName,
  type RealNameApplicationSummary,
  type RealNameStatusResponse,
} from '../../api/real-name'
import { getApiErrorMessage } from '../../api/request'
import type { RealNameConfig } from '../../api/site-config'

const realName = ref('')
const idCard = ref('')
const provider = ref<'alipay' | 'wechat' | 'manual'>('manual')
const loading = ref(false)
const pageLoading = ref(false)
const error = ref('')
const success = ref('')
const status = ref<RealNameStatusResponse['status']>('unverified')
const application = ref<RealNameApplicationSummary | null>(null)
const config = ref<RealNameConfig | null>(null)

const statusText = computed(() => {
  const labels: Record<string, string> = {
    unverified: '未实名',
    pending: '核验中',
    approved: '已通过',
    rejected: '已拒绝',
  }
  return labels[status.value] || status.value
})

const canSubmit = computed(() => {
  if (!config.value?.enabled) {
    return false
  }
  if (status.value === 'unverified') {
    return true
  }
  return status.value === 'rejected' && config.value.resubmit_enabled
})

const providerLabel = (value: string) => {
  const labels: Record<string, string> = {
    alipay: '支付宝实名',
    wechat: '微信实名',
    manual: '人工审核',
  }
  return labels[value] || value
}

const applyStatus = (data: RealNameStatusResponse) => {
  status.value = data.status
  application.value = data.application
  config.value = data.config
  const defaultProvider = data.config.default_provider as 'alipay' | 'wechat' | 'manual'
  provider.value = defaultProvider || 'manual'
}

const loadStatus = async () => {
  pageLoading.value = true
  error.value = ''
  try {
    const data = await getRealNameStatus()
    applyStatus(data)
  } catch (err) {
    error.value = getApiErrorMessage(err, '实名状态加载失败')
  } finally {
    pageLoading.value = false
  }
}

onMounted(loadStatus)

const handleSubmit = async () => {
  if (!realName.value || !idCard.value) {
    error.value = '请填写所有字段'
    return
  }

  loading.value = true
  error.value = ''
  success.value = ''

  try {
    const data = await submitRealName({
      real_name: realName.value,
      id_type: 'id_card',
      id_number: idCard.value,
      provider: provider.value,
    })
    application.value = data.application
    status.value = 'pending'
    success.value = '实名认证申请已提交'
    if (data.provider_action.redirect_url) {
      window.location.href = data.provider_action.redirect_url
    }
  } catch (err) {
    error.value = getApiErrorMessage(err, '提交失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

const handleSync = async () => {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    const data = await syncRealName({ application_no: application.value?.application_no })
    applyStatus(data)
    success.value = '实名状态已刷新'
  } catch (err) {
    error.value = getApiErrorMessage(err, '实名状态同步失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
      <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Real Name</p>
      <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">实名认证</h1>

      <div v-if="error" class="mt-6 rounded-xl border border-neutral-950 bg-neutral-50 p-4 text-sm font-bold text-neutral-950">{{ error }}</div>
      <div v-if="success" class="mt-6 rounded-xl border border-neutral-300 bg-white p-4 text-sm font-bold text-neutral-950">{{ success }}</div>

      <section class="surface-pop mt-8 rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[8px_8px_0_#111]">
        <h2 class="text-xl font-black text-neutral-950">实名状态</h2>
        <div class="mt-5 rounded-2xl border border-neutral-200 bg-neutral-50 p-5">
          <div class="flex items-center gap-4">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl border border-neutral-950 bg-white text-sm font-black">ID</div>
            <div>
              <div class="text-sm font-black text-neutral-950">{{ pageLoading ? '加载中...' : statusText }}</div>
              <div class="mt-1 text-sm text-neutral-500">{{ config?.review_notice || '当前状态以后端返回为准。' }}</div>
            </div>
          </div>
        </div>

        <div v-if="application" class="mt-5 rounded-2xl border border-neutral-200 bg-white p-5 text-sm">
          <div class="flex justify-between gap-4"><span class="text-neutral-500">申请编号</span><span class="font-black">{{ application.application_no }}</span></div>
          <div class="mt-3 flex justify-between gap-4"><span class="text-neutral-500">实名方式</span><span class="font-black">{{ providerLabel(application.verification_provider || provider) }}</span></div>
          <div v-if="application.id_number_masked" class="mt-3 flex justify-between gap-4"><span class="text-neutral-500">证件号码</span><span class="font-black">{{ application.id_number_masked }}</span></div>
          <div v-if="application.failure_reason" class="mt-3 rounded-xl border border-neutral-200 bg-neutral-50 p-3 font-bold text-neutral-950">{{ application.failure_reason }}</div>
        </div>

        <div v-if="status === 'pending'" class="mt-6">
          <button type="button" :disabled="loading" class="w-full rounded-full border border-neutral-950 bg-white py-3 text-sm font-black text-neutral-950 hover:bg-neutral-950 hover:text-white disabled:opacity-50" @click="handleSync">
            {{ loading ? '同步中...' : '同步实名状态' }}
          </button>
        </div>

        <div v-if="config && !config.enabled" class="mt-6 rounded-xl border border-neutral-200 bg-neutral-50 p-4 text-sm font-bold text-neutral-950">实名功能暂未开放</div>

        <form v-if="canSubmit" class="mt-8 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label for="real-name" class="mb-2 block text-sm font-black text-neutral-800">真实姓名</label>
            <input id="real-name" v-model="realName" type="text" class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入真实姓名" />
          </div>
          <div>
            <label for="id-card" class="mb-2 block text-sm font-black text-neutral-800">身份证号码</label>
            <input id="id-card" v-model="idCard" type="text" class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入身份证号码" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">实名方式</label>
            <label v-for="item in config?.allowed_providers || ['manual']" :key="item" class="soft-lift mt-3 flex items-center gap-3 rounded-xl border border-neutral-200 px-4 py-3 text-sm font-bold text-neutral-700">
              <input v-model="provider" name="provider" type="radio" :value="item" class="h-4 w-4 border-neutral-300 text-neutral-950" />
              {{ providerLabel(item) }}
            </label>
          </div>
          <button type="submit" :disabled="loading" class="btn-dark w-full rounded-full border py-3 text-sm font-black disabled:opacity-50">{{ loading ? '提交中...' : '提交实名认证' }}</button>
        </form>
      </section>
    </div>
  </div>
</template>
