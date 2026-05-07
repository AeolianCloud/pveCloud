<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  getRealNameStatus,
  submitRealName,
  syncRealName,
  type RealNameProviderAction,
  type RealNameStatusResponse,
} from '../../api/real-name'

const loading = ref(false)
const submitting = ref(false)
const syncing = ref(false)
const errorMessage = ref('')
const status = ref<RealNameStatusResponse | null>(null)
const lastProviderAction = ref<RealNameProviderAction | null>(null)
const realName = ref('')
const idNumber = ref('')
const provider = ref<'alipay' | 'wechat' | ''>('')

const config = computed(() => status.value?.config)
const application = computed(() => status.value?.application)
const canSubmit = computed(() => {
  const cfg = config.value
  if (!cfg || !cfg.enabled || submitting.value || syncing.value || status.value?.status === 'pending' || status.value?.status === 'approved') return false
  if (!realName.value.trim() || !idNumber.value.trim()) return false
  return Boolean(provider.value || cfg.default_provider)
})

const statusText = computed(() => {
  switch (status.value?.status) {
    case 'pending': return '核验中'
    case 'approved': return '已通过'
    case 'rejected': return '已拒绝'
    default: return '未实名'
  }
})

const providerText = computed(() => {
  const value = application.value?.verification_provider || provider.value
  if (value === 'wechat') return '微信'
  if (value === 'alipay') return '支付宝'
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
    provider.value = (status.value.config.default_provider as 'alipay' | 'wechat' | '') || ''
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
      provider: (provider.value || config.value?.default_provider || '') as 'alipay' | 'wechat',
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
  <section class="real-name-page content-page">
    <div class="real-name-shell">
      <div class="real-hero">
        <div>
          <p class="eyebrow">REAL NAME VERIFICATION</p>
          <h1>购买机器前完成个人实名</h1>
          <p>实名由支付宝或微信侧通道完成，页面只负责提交资料、跳转供应商和同步结果，最终状态以后端为准。</p>
        </div>
        <div class="hero-badge-card">
          <span>购买门禁</span>
          <strong>实名通过后开放购买</strong>
        </div>
      </div>

      <div v-if="loading" class="loading-card">正在读取实名状态...</div>
      <div v-else class="real-grid">
        <aside class="status-card">
          <span class="status-pill" :data-status="status?.status || 'unverified'">{{ statusText }}</span>
          <div class="status-main">
            <span>当前状态</span>
            <h2>{{ application?.real_name || '个人实名状态' }}</h2>
            <p v-if="application">证件号码：{{ application.id_number_masked }}</p>
            <p v-else>提交实名资料后会跳转支付宝或微信侧完成核验。</p>
          </div>
          <div class="status-list">
            <p><b>购买要求</b><span>{{ config?.required_for_order ? '必须实名通过' : '暂不强制' }}</span></p>
            <p><b>重提规则</b><span>{{ config?.resubmit_enabled ? `最多 ${config.max_submit_attempts} 次` : '不允许重提' }}</span></p>
            <p><b>实名通道</b><span>{{ providerText }}</span></p>
          </div>
          <p v-if="application?.failure_reason" class="notice error">失败原因：{{ application.failure_reason }}</p>
          <p v-if="config?.review_notice" class="notice info">{{ config.review_notice }}</p>
          <p v-if="lastProviderAction?.redirect_url" class="notice info">已创建供应商会话，如未自动跳转，可重试同步当前结果。</p>
          <div class="status-actions">
            <button v-if="status?.status === 'pending'" class="btn btn-outline" type="button" :disabled="syncing" @click="handleSync">
              {{ syncing ? '同步中...' : '同步实名结果' }}
            </button>
            <RouterLink v-if="status?.status === 'approved'" class="btn btn-primary" to="/pricing">返回价格页</RouterLink>
          </div>
        </aside>

        <form class="real-form" @submit.prevent="handleSubmit">
          <template v-if="!config?.enabled">
            <div class="empty-state">
              <h3>实名功能暂未开放</h3>
              <p>后台尚未开启实名提交，请稍后再试。</p>
            </div>
          </template>
          <template v-else-if="status?.status === 'pending'">
            <div class="empty-state">
              <h3>实名核验进行中</h3>
              <p>完成支付宝或微信侧核验后，返回本页并点击“同步实名结果”。</p>
            </div>
          </template>
          <template v-else-if="status?.status === 'approved'">
            <div class="empty-state empty-state--success">
              <h3>实名已通过</h3>
              <p>你已满足购买机器的实名要求。</p>
            </div>
          </template>
          <template v-else>
            <div class="form-heading">
              <span>{{ status?.status === 'rejected' ? 'RESUBMIT' : 'SUBMIT' }}</span>
              <h3>{{ status?.status === 'rejected' ? '重新提交实名' : '提交个人实名' }}</h3>
              <p>请填写与身份证一致的姓名和证件号码，并选择后台开放的实名供应商。</p>
            </div>

            <div class="field-grid">
              <label class="field">
                <span>真实姓名</span>
                <input v-model="realName" type="text" placeholder="请输入证件姓名" />
              </label>
              <label class="field">
                <span>身份证号码</span>
                <input v-model="idNumber" type="text" placeholder="18 位身份证号码" />
              </label>
            </div>

            <div class="provider-group">
              <span>实名供应商</span>
              <div class="provider-options">
                <label v-for="item in config?.allowed_providers || []" :key="item" class="provider-card" :data-active="(provider || config?.default_provider) === item">
                  <input v-model="provider" type="radio" :value="item" />
                  <b>{{ item === 'wechat' ? '微信' : '支付宝' }}</b>
                  <span>{{ item === 'wechat' ? '跳转微信侧实名核验' : '跳转支付宝身份认证' }}</span>
                </label>
              </div>
            </div>

            <p v-if="errorMessage" class="notice error">{{ errorMessage }}</p>
            <button class="btn btn-primary submit-btn" type="submit" :disabled="!canSubmit">{{ submitting ? '提交中...' : '提交并前往核验' }}</button>
          </template>
        </form>
      </div>
    </div>
  </section>
</template>

<style scoped>
.real-name-page {
  position: relative;
  min-height: calc(100vh - 96px);
  overflow: hidden;
}

.real-name-page::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(circle at 12% 0%, rgba(59, 130, 246, 0.2), transparent 28%),
    radial-gradient(circle at 92% 12%, rgba(16, 185, 129, 0.16), transparent 30%);
}

.real-name-shell {
  position: relative;
  width: min(1160px, calc(100% - 40px));
  margin: 0 auto;
  padding: clamp(26px, 5vw, 58px) 0 72px;
  display: grid;
  gap: 24px;
}

.real-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 24px;
  align-items: stretch;
  padding: clamp(28px, 4vw, 44px);
  border: 1px solid var(--c-border);
  border-radius: 34px;
  background:
    linear-gradient(135deg, rgba(59, 130, 246, 0.16), rgba(16, 185, 129, 0.08)),
    var(--c-card);
  box-shadow: var(--shadow-lg);
}

.eyebrow {
  margin-bottom: 14px;
  color: var(--c-primary);
  font-size: 0.76rem;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.real-hero h1 {
  max-width: 720px;
  margin-bottom: 16px;
  font-size: clamp(2.35rem, 5vw, 4.4rem);
  line-height: 0.98;
  letter-spacing: -0.065em;
}

.real-hero p {
  max-width: 680px;
  color: var(--c-text-2);
  line-height: 1.8;
}

.hero-badge-card {
  display: grid;
  align-content: end;
  gap: 12px;
  min-height: 180px;
  padding: 24px;
  border: 1px solid var(--c-border);
  border-radius: 26px;
  background:
    radial-gradient(circle at 100% 0%, rgba(16, 185, 129, 0.28), transparent 44%),
    rgba(255, 255, 255, 0.04);
}

.hero-badge-card span {
  width: fit-content;
  padding: 7px 11px;
  border-radius: 999px;
  color: var(--c-success);
  background: var(--c-success-soft);
  font-size: 0.8rem;
  font-weight: 900;
}

.hero-badge-card strong {
  font-size: 1.45rem;
  line-height: 1.2;
}

.loading-card,
.status-card,
.real-form {
  border: 1px solid var(--c-border);
  border-radius: 28px;
  background: var(--c-card);
  box-shadow: var(--shadow-sm);
}

.loading-card {
  padding: 32px;
  color: var(--c-text-2);
}

.real-grid {
  display: grid;
  grid-template-columns: minmax(300px, 0.82fr) minmax(0, 1.18fr);
  gap: 18px;
  align-items: start;
}

.status-card {
  position: sticky;
  top: 96px;
  display: grid;
  gap: 20px;
  padding: 26px;
}

.status-pill {
  width: fit-content;
  padding: 7px 12px;
  border-radius: 999px;
  color: var(--c-warning);
  background: var(--c-warning-soft);
  font-size: 0.82rem;
  font-weight: 900;
}

.status-pill[data-status='approved'] {
  color: var(--c-success);
  background: var(--c-success-soft);
}

.status-pill[data-status='rejected'] {
  color: var(--c-error);
  background: var(--c-error-soft);
}

.status-main span {
  display: block;
  margin-bottom: 8px;
  color: var(--c-text-3);
  font-size: 0.78rem;
  font-weight: 900;
  letter-spacing: 0.12em;
}

.status-main h2 {
  margin-bottom: 10px;
  font-size: 1.65rem;
  letter-spacing: -0.05em;
}

.status-main p,
.empty-state p,
.form-heading p {
  color: var(--c-text-2);
  line-height: 1.75;
}

.status-list {
  display: grid;
  gap: 10px;
  padding: 16px;
  border: 1px solid var(--c-border-light);
  border-radius: 20px;
  background: var(--c-surface-dim);
}

.status-list p {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin: 0;
  color: var(--c-text-2);
}

.status-list b {
  color: var(--c-text);
}

.status-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.notice {
  margin: 0;
  padding: 12px 14px;
  border-radius: 16px;
  line-height: 1.65;
}

.notice.error {
  color: var(--c-error);
  background: var(--c-error-soft);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.notice.info {
  color: var(--c-primary-h);
  background: var(--c-primary-soft);
  border: 1px solid rgba(59, 130, 246, 0.22);
}

.real-form {
  display: grid;
  gap: 20px;
  padding: clamp(24px, 4vw, 34px);
}

.form-heading {
  display: grid;
  gap: 8px;
}

.form-heading span {
  color: var(--c-text-3);
  font-size: 0.76rem;
  font-weight: 900;
  letter-spacing: 0.14em;
}

.form-heading h3,
.empty-state h3 {
  font-size: 1.65rem;
  letter-spacing: -0.05em;
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.field {
  display: grid;
  gap: 8px;
  color: var(--c-text-2);
  font-weight: 800;
}

.field input {
  min-height: 52px;
  padding: 0 15px;
  border: 1px solid var(--c-border);
  border-radius: 16px;
  color: var(--c-text);
  background: var(--c-surface-dim);
}

.provider-group {
  display: grid;
  gap: 12px;
}

.provider-group > span {
  color: var(--c-text-2);
  font-weight: 800;
}

.provider-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.provider-card {
  position: relative;
  display: grid;
  gap: 8px;
  min-height: 132px;
  padding: 18px;
  border: 1px solid var(--c-border);
  border-radius: 20px;
  background: var(--c-surface-dim);
  cursor: pointer;
}

.provider-card[data-active='true'] {
  border-color: rgba(59, 130, 246, 0.4);
  background:
    radial-gradient(circle at 100% 0%, rgba(59, 130, 246, 0.12), transparent 44%),
    var(--c-surface-dim);
}

.provider-card input {
  position: absolute;
  inset: 0;
  opacity: 0;
}

.provider-card b {
  font-size: 1rem;
}

.provider-card span {
  color: var(--c-text-3);
  line-height: 1.6;
}

.submit-btn {
  justify-self: start;
}

.btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.empty-state {
  min-height: 320px;
  display: grid;
  align-content: center;
  justify-items: start;
  gap: 12px;
}

.empty-state--success h3 {
  color: var(--c-success);
}

@media (max-width: 960px) {
  .real-hero,
  .real-grid {
    grid-template-columns: 1fr;
  }

  .status-card {
    position: static;
  }
}

@media (max-width: 680px) {
  .real-name-shell {
    width: min(100% - 28px, 1160px);
    padding-top: 24px;
  }

  .real-hero,
  .real-form,
  .status-card {
    border-radius: 24px;
  }

  .field-grid,
  .provider-options {
    grid-template-columns: 1fr;
  }

  .hero-badge-card {
    min-height: 150px;
  }
}
</style>
