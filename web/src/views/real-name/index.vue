<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  getRealNameStatus,
  submitRealName,
  syncRealName,
  type RealNameProviderAction,
  type RealNameStatusResponse,
} from '../../api/real-name'

type RealNameProvider = 'alipay' | 'wechat' | 'manual'

const loading = ref(false)
const submitting = ref(false)
const syncing = ref(false)
const errorMessage = ref('')
const status = ref<RealNameStatusResponse | null>(null)
const lastProviderAction = ref<RealNameProviderAction | null>(null)
const realName = ref('')
const idNumber = ref('')
const provider = ref<RealNameProvider | ''>('')

const config = computed(() => status.value?.config)
const application = computed(() => status.value?.application)
const allowedProviders = computed(() => {
  const providers = config.value?.allowed_providers
  return Array.isArray(providers) ? providers : []
})
const selectedProvider = computed(() => provider.value || config.value?.default_provider || '')
const isManualProvider = computed(() => (application.value?.verification_provider || selectedProvider.value) === 'manual')

const submitBlockers = computed(() => {
  const cfg = config.value
  const blockers: string[] = []
  if (!cfg) return ['实名配置还没有加载完成']
  if (!cfg.enabled) blockers.push('后台尚未开放用户端实名入口')
  if (status.value?.status === 'pending') blockers.push('当前实名正在核验中')
  if (status.value?.status === 'approved') blockers.push('当前账号已通过实名')
  if (status.value?.status === 'rejected' && !cfg.resubmit_enabled) blockers.push('后台不允许实名失败后重新提交')
  if (status.value?.status === 'rejected' && application.value && application.value.submit_attempt >= cfg.max_submit_attempts) {
    blockers.push(`实名提交次数已达上限（最多 ${cfg.max_submit_attempts} 次）`)
  }
  if (!realName.value.trim()) blockers.push('请填写真实姓名')
  if (!idNumber.value.trim()) blockers.push('请填写身份证号码')
  if (allowedProviders.value.length === 0) blockers.push('后台尚未开放可用实名通道')
  else if (!selectedProvider.value) blockers.push('请选择实名方式')
  else if (!allowedProviders.value.includes(selectedProvider.value)) blockers.push('当前默认实名方式不可用，请重新选择')
  if (submitting.value) blockers.push('实名申请正在提交中')
  if (syncing.value) blockers.push('实名结果正在同步中')
  return blockers
})

const canSubmit = computed(() => submitBlockers.value.length === 0)

const statusText = computed(() => {
  switch (status.value?.status) {
    case 'pending': return '核验中'
    case 'approved': return '已通过'
    case 'rejected': return '已拒绝'
    default: return '未实名'
  }
})

const providerText = computed(() => {
  const value = application.value?.verification_provider || selectedProvider.value
  if (value === 'wechat') return '微信'
  if (value === 'alipay') return '支付宝'
  if (value === 'manual') return '人工审核'
  return '未选择'
})

function errorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (error instanceof Error && error.message) return error.message
  return '操作失败，请稍后再试'
}

async function loadStatus() {
  loading.value = true
  errorMessage.value = ''
  try {
    status.value = await getRealNameStatus()
    provider.value = (status.value.config.default_provider as RealNameProvider | '') || ''
  } catch (error) {
    errorMessage.value = errorText(error)
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!canSubmit.value) return
  submitting.value = true
  errorMessage.value = ''
  try {
    const result = await submitRealName({
      real_name: realName.value.trim(),
      id_type: 'id_card',
      id_number: idNumber.value.trim(),
      provider: selectedProvider.value as RealNameProvider,
    })
    lastProviderAction.value = result.provider_action
    if (result.provider_action.redirect_url) {
      window.location.assign(result.provider_action.redirect_url)
      return
    }
    status.value = await getRealNameStatus()
  } catch (error) {
    errorMessage.value = errorText(error)
  } finally {
    submitting.value = false
  }
}

async function handleSync() {
  if (!application.value?.application_no) return
  syncing.value = true
  errorMessage.value = ''
  try {
    status.value = await syncRealName({ application_no: application.value.application_no })
  } catch (error) {
    errorMessage.value = errorText(error)
  } finally {
    syncing.value = false
  }
}

onMounted(loadStatus)
</script>

<template>
  <section class="real-page page-shell">
    <div class="page-hero surface">
      <div>
        <p class="section-label">Real Name</p>
        <h1 class="page-title">个人实名认证</h1>
        <p class="page-copy">实名方式由后台公开配置决定。前端只提交资料、跳转供应商或等待人工审核，最终状态以后端为准。</p>
      </div>
      <span class="status-pill" :data-status="status?.status || 'unverified'">{{ statusText }}</span>
    </div>

    <div v-if="loading" class="loading-panel surface">
      <div class="spinner"></div>
      <span>正在读取实名状态...</span>
    </div>

    <div v-else class="real-grid">
      <aside class="status-card card">
        <div>
          <span class="section-label">Status</span>
          <h2>{{ application?.real_name || '当前实名状态' }}</h2>
          <p v-if="application">证件号码：{{ application.id_number_masked }}</p>
          <p v-else>提交后会根据后台可用通道进入供应商核验或人工审核。</p>
        </div>
        <dl class="status-list">
          <div><dt>购买要求</dt><dd>{{ config?.required_for_order ? '必须实名通过' : '暂不强制' }}</dd></div>
          <div><dt>实名通道</dt><dd>{{ providerText }}</dd></div>
          <div><dt>重提规则</dt><dd>{{ config?.resubmit_enabled ? `最多 ${config.max_submit_attempts} 次` : '不允许重提' }}</dd></div>
        </dl>
        <p v-if="application?.failure_reason" class="notice error">失败原因：{{ application.failure_reason }}</p>
        <p v-if="config?.review_notice" class="notice info">{{ config.review_notice }}</p>
        <p v-if="lastProviderAction?.redirect_url" class="notice info">已创建供应商会话，如未自动跳转，可稍后同步结果。</p>
        <button v-if="status?.status === 'pending' && !isManualProvider" class="btn btn-outline" type="button" :disabled="syncing" @click="handleSync">
          {{ syncing ? '同步中...' : '同步实名结果' }}
        </button>
      </aside>

      <form class="real-form card" @submit.prevent="handleSubmit">
        <template v-if="!config?.enabled">
          <div class="empty-state">
            <h2>实名功能暂未开放</h2>
            <p>后台尚未开启用户端实名提交。</p>
          </div>
        </template>
        <template v-else-if="status?.status === 'pending'">
          <div class="empty-state">
            <h2>实名核验进行中</h2>
            <p>{{ isManualProvider ? '人工审核申请已提交，请等待后台审核结果。' : '完成支付宝或微信侧核验后，返回本页同步结果。' }}</p>
          </div>
        </template>
        <template v-else-if="status?.status === 'approved'">
          <div class="empty-state">
            <h2>实名已通过</h2>
            <p>你已满足购买机器的实名要求。</p>
          </div>
        </template>
        <template v-else>
          <div class="form-heading">
            <span class="section-label">{{ status?.status === 'rejected' ? 'Resubmit' : 'Submit' }}</span>
            <h2>{{ status?.status === 'rejected' ? '重新提交实名' : '提交个人实名' }}</h2>
            <p>当前证件类型固定为身份证，最终校验以后端为准。</p>
          </div>

          <div class="field-grid">
            <label class="field"><span>真实姓名</span><span class="field-control"><input v-model="realName" type="text" placeholder="请输入证件姓名" /></span></label>
            <label class="field"><span>身份证号码</span><span class="field-control"><input v-model="idNumber" type="text" placeholder="18 位身份证号码" /></span></label>
          </div>

          <div class="provider-group">
            <span>实名方式</span>
            <div v-if="allowedProviders.length" class="provider-options">
              <label v-for="item in allowedProviders" :key="item" class="provider-card" :data-active="(provider || config?.default_provider) === item">
                <input v-model="provider" type="radio" :value="item" />
                <b>{{ item === 'manual' ? '人工审核' : item === 'wechat' ? '微信' : '支付宝' }}</b>
                <span>{{ item === 'manual' ? '提交后等待后台审核' : item === 'wechat' ? '跳转微信侧实名核验' : '跳转支付宝身份认证' }}</span>
              </label>
            </div>
            <p v-else class="notice warning">后台尚未开放可用实名通道。</p>
          </div>

          <p v-if="errorMessage" class="notice error">{{ errorMessage }}</p>
          <div v-if="submitBlockers.length" class="requirements notice warning">
            <b>提交条件未满足</b>
            <ul><li v-for="item in submitBlockers" :key="item">{{ item }}</li></ul>
          </div>
          <button class="btn btn-primary" type="submit" :disabled="!canSubmit">
            {{ submitting ? '提交中...' : selectedProvider === 'manual' ? '提交人工审核' : '提交并前往核验' }}
          </button>
        </template>
      </form>
    </div>
  </section>
</template>

<style scoped>
.real-page {
  display: grid;
  gap: 22px;
}

.page-hero {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  padding: clamp(24px, 4vw, 38px);
}

.page-hero .page-copy {
  max-width: 760px;
  margin-top: 14px;
}

.status-pill {
  padding: 8px 12px;
  border-radius: 999px;
  color: var(--c-warning);
  background: var(--c-warning-soft);
  font-weight: 900;
}
.status-pill[data-status='approved'] { color: var(--c-success); background: var(--c-success-soft); }
.status-pill[data-status='rejected'] { color: var(--c-error); background: var(--c-error-soft); }

.loading-panel {
  min-height: 220px;
  display: grid;
  place-items: center;
  gap: 12px;
  color: var(--c-text-2);
}

.real-grid {
  display: grid;
  grid-template-columns: minmax(280px, 0.8fr) minmax(0, 1.2fr);
  gap: 18px;
  align-items: start;
}

.status-card,
.real-form {
  display: grid;
  gap: 18px;
  padding: 22px;
}

.status-card {
  position: sticky;
  top: 92px;
}

.status-card h2,
.form-heading h2,
.empty-state h2 {
  margin-top: 8px;
  font-size: 1.5rem;
  letter-spacing: -0.04em;
}

.status-card p,
.form-heading p,
.empty-state p {
  margin-top: 8px;
  color: var(--c-text-2);
  line-height: 1.7;
}

.status-list {
  display: grid;
  gap: 10px;
}

.status-list div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  background: var(--c-surface-dim);
}

.status-list dt {
  color: var(--c-text-3);
  font-weight: 800;
}

.status-list dd {
  margin: 0;
  font-weight: 800;
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.provider-group {
  display: grid;
  gap: 10px;
  color: var(--c-text-2);
  font-weight: 800;
}

.provider-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.provider-card {
  display: grid;
  gap: 6px;
  padding: 16px;
  border: 1px solid var(--c-border);
  border-radius: 14px;
  background: var(--c-surface-dim);
  cursor: pointer;
}

.provider-card input {
  position: absolute;
  opacity: 0;
}

.provider-card[data-active='true'] {
  border-color: var(--c-primary);
  color: var(--c-primary);
  background: var(--c-primary-soft);
}

.provider-card span {
  color: var(--c-text-2);
  font-size: 0.9rem;
  font-weight: 600;
  line-height: 1.5;
}

.requirements {
  display: grid;
  gap: 8px;
}

.requirements ul {
  padding-left: 18px;
  color: var(--c-text-2);
}

@media (max-width: 920px) {
  .real-grid,
  .field-grid {
    grid-template-columns: 1fr;
  }

  .status-card {
    position: static;
  }

  .page-hero {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 680px) {
  .provider-options {
    grid-template-columns: 1fr;
  }
}
</style>
