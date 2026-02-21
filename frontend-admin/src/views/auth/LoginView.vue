<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { login } from '@/api/auth'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()

const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur',
  },
  password: {
    required: true,
    message: '请输入密码',
    trigger: 'blur',
  },
}

const formRef = ref()

async function handleLogin() {
  // 表单校验
  await formRef.value?.validate()

  loading.value = true
  try {
    const res = await login(form.username, form.password)
    const { token, refresh_token, user } = res.data.data

    // 保存 access token + refresh token + 用户信息
    // 注意：refresh_token 用于 access token 过期后的自动刷新
    authStore.setTokens(token, refresh_token)
    authStore.user = user

    message.success(`欢迎回来，${user.nickname || user.username}`)
    router.push('/dashboard')
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <!-- 左侧品牌区域 -->
    <div class="brand-panel">
      <!-- 顶部 Logo 区 -->
      <div class="brand-logo">
        <div class="logo-icon">
          <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="4" y="4" width="18" height="18" rx="3" fill="rgba(255,255,255,0.9)" />
            <rect x="26" y="4" width="18" height="18" rx="3" fill="rgba(255,255,255,0.5)" />
            <rect x="4" y="26" width="18" height="18" rx="3" fill="rgba(255,255,255,0.5)" />
            <rect x="26" y="26" width="18" height="18" rx="3" fill="rgba(255,255,255,0.7)" />
          </svg>
        </div>
        <span class="logo-name">pveCloud</span>
      </div>

      <!-- 中央文案 -->
      <div class="brand-content">
        <h2 class="brand-title">云基础设施管理平台</h2>
        <p class="brand-desc">统一管理虚拟机、容器与存储资源，让运维更简单、高效。</p>

        <!-- 特性列表 -->
        <ul class="feature-list">
          <li>
            <span class="feature-dot"></span>
            <span>实时监控集群资源状态</span>
          </li>
          <li>
            <span class="feature-dot"></span>
            <span>自动化运维任务调度</span>
          </li>
          <li>
            <span class="feature-dot"></span>
            <span>多节点统一纳管</span>
          </li>
        </ul>
      </div>

      <!-- 底部版权 -->
      <div class="brand-footer">
        © 2026 pveCloud. All rights reserved.
      </div>
    </div>

    <!-- 右侧登录表单区域 -->
    <div class="form-panel">
      <div class="form-wrapper">
        <div class="form-header">
          <h1 class="form-title">欢迎回来</h1>
          <p class="form-subtitle">请登录您的管理账号</p>
        </div>

        <n-form
          ref="formRef"
          :model="form"
          :rules="rules"
          size="large"
        >
          <n-form-item path="username" label="用户名">
            <n-input
              v-model:value="form.username"
              placeholder="请输入用户名"
              :input-props="{ autocomplete: 'username' }"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <n-icon><person-outline /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <n-form-item path="password" label="密码">
            <n-input
              v-model:value="form.password"
              type="password"
              placeholder="请输入密码"
              show-password-on="click"
              :input-props="{ autocomplete: 'current-password' }"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <n-icon><lock-closed-outline /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <n-button
            type="primary"
            block
            size="large"
            :loading="loading"
            style="margin-top: 8px;"
            @click="handleLogin"
          >
            登 录
          </n-button>
        </n-form>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5'
export default { components: { PersonOutline, LockClosedOutline } }
</script>

<style scoped>
/* ========== 整体布局 ========== */
.login-page {
  height: 100vh;
  display: flex;
  overflow: hidden;
}

/* ========== 左侧品牌区 ========== */
.brand-panel {
  /* 占左侧 45% 宽度 */
  flex: 0 0 45%;
  display: flex;
  flex-direction: column;
  padding: 40px 48px;
  /* 深蓝渐变，科技感 */
  background: linear-gradient(145deg, #1a2740 0%, #0f1c2e 50%, #162136 100%);
  color: #fff;
  position: relative;
  overflow: hidden;
}

/* 装饰性背景圆圈 */
.brand-panel::before,
.brand-panel::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  opacity: 0.06;
  background: #fff;
}

.brand-panel::before {
  width: 500px;
  height: 500px;
  top: -160px;
  right: -160px;
}

.brand-panel::after {
  width: 300px;
  height: 300px;
  bottom: -80px;
  left: -80px;
}

/* Logo 区 */
.brand-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
  z-index: 1;
}

.logo-icon {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
}

.logo-name {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 1px;
  color: #fff;
}

/* 中央文案 */
.brand-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  position: relative;
  z-index: 1;
}

.brand-title {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.3;
  margin-bottom: 16px;
  color: #fff;
}

.brand-desc {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.65);
  line-height: 1.8;
  margin-bottom: 36px;
  max-width: 320px;
}

/* 特性列表 */
.feature-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.feature-list li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.75);
}

.feature-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #4fa8e8;
  flex-shrink: 0;
}

/* 底部版权 */
.brand-footer {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.3);
  position: relative;
  z-index: 1;
}

/* ========== 右侧表单区 ========== */
.form-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f7f8fa;
  padding: 40px;
}

.form-wrapper {
  width: 100%;
  max-width: 360px;
}

.form-header {
  margin-bottom: 36px;
}

.form-title {
  font-size: 26px;
  font-weight: 700;
  color: #18181c;
  margin-bottom: 8px;
}

.form-subtitle {
  font-size: 14px;
  color: #909399;
}

/* ========== 响应式适配（小屏隐藏左侧） ========== */
@media (max-width: 768px) {
  .brand-panel {
    display: none;
  }

  .form-panel {
    background: #fff;
  }
}
</style>
