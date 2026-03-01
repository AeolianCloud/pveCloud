<script setup lang="ts">
import { h, ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import type { Component } from 'vue'
import { GridOutline, PersonOutline, ChevronDownOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { useAuthStore } from '@/store/auth'
import { updateAdminUser } from '@/api/adminUser'
import { getMyMenus } from '@/api/menu'
import type { AdminMenuNode } from '@/types'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

// 当前页面标题，从路由 meta 取
const pageTitle = computed(() => (route.meta.title as string) || '')

// ── 菜单 ──────────────────────────────────────────────────

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// 动态菜单：由后端按当前用户权限裁剪后下发（/menus/my）
const menuTree = ref<AdminMenuNode[]>([])

function iconComponent(name?: string | null): Component | null {
  // icon 字段是后端与前端的“约定字符串”，不直接传 SVG/组件，避免接口携带实现细节。
  switch (name) {
    case 'dashboard':
      return GridOutline
    case 'system':
      return PersonOutline
    default:
      return null
  }
}

function toMenuOptions(nodes: AdminMenuNode[]): MenuOption[] {
  return nodes.map((n) => {
    const icon = iconComponent(n.icon)
    const children = n.children?.length ? toMenuOptions(n.children) : undefined

    // 约定：叶子节点 key 使用 path（便于高亮与跳转）；目录节点使用稳定的虚拟 key。
    const key = n.path ? n.path : `dir:${n.id}`

    const opt: MenuOption = {
      label: n.title,
      key,
      ...(icon ? { icon: renderIcon(icon) } : {}),
      ...(children ? { children } : {}),
    }
    return opt
  })
}

const menuOptions = computed<MenuOption[]>(() => {
  // 防御性兜底：即使菜单表为空，也至少给一个控制台入口，避免“登录后无路可走”。
  const fallback: MenuOption[] = [{
    label: '控制台',
    key: '/dashboard',
    icon: renderIcon(GridOutline),
  }]

  if (!menuTree.value || menuTree.value.length === 0) return fallback
  return toMenuOptions(menuTree.value)
})

const activeMenuKey = computed(() => route.path ?? '')

// 菜单点击 → 跳转路由（只有 key 为 /xxx 才认为是可跳转的叶子节点）
function handleMenuUpdate(key: string) {
  if (typeof key === 'string' && key.startsWith('/')) {
    router.push(key)
  }
}

async function loadMenus() {
  try {
    const res = await getMyMenus()
    menuTree.value = res.data.data
  } catch (err: unknown) {
    // 菜单加载失败不阻塞布局渲染，但会导致侧边栏展示兜底菜单。
    message.error(err instanceof Error ? err.message : '菜单加载失败')
    menuTree.value = []
  }
}

onMounted(() => {
  loadMenus()
})

// ── 用户下拉菜单 ──────────────────────────────────────────

const userMenuOptions = [
  { label: '修改密码', key: 'change-password' },
  { label: '退出登录', key: 'logout' },
]

// ── 修改密码弹窗 ──────────────────────────────────────────
const pwdModalVisible = ref(false)
const pwdLoading = ref(false)
const pwdFormRef = ref()
const pwdForm = ref({ oldPassword: '', newPassword: '', confirmPassword: '' })

const pwdRules = {
  oldPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '新密码至少 6 位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule: unknown, value: string) =>
        value === pwdForm.value.newPassword || new Error('两次输入的密码不一致'),
      trigger: 'blur',
    },
  ],
}

function openChangePwd() {
  pwdForm.value = { oldPassword: '', newPassword: '', confirmPassword: '' }
  pwdModalVisible.value = true
}

async function handleChangePwd() {
  await pwdFormRef.value?.validate()
  if (!authStore.user) return
  pwdLoading.value = true
  try {
    await updateAdminUser(authStore.user.id, { password: pwdForm.value.newPassword })
    message.success('密码修改成功，请重新登录')
    pwdModalVisible.value = false
    // 密码变更后退出登录，强制重新认证
    await authStore.logout()
    router.push('/login')
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '修改失败')
  } finally {
    pwdLoading.value = false
  }
}

async function handleUserMenu(key: string) {
  if (key === 'change-password') {
    openChangePwd()
  } else if (key === 'logout') {
    await authStore.logout()
    message.success('已退出登录')
    router.push('/login')
  }
}
</script>

<template>
  <n-layout style="height: 100vh" has-sider>
    <!-- 侧边栏 -->
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="220"
      show-trigger="bar"
      :native-scrollbar="false"
    >
      <!-- Logo 区 -->
      <div class="sider-logo">
        <div class="sider-logo-icon">
          <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="4" y="4" width="18" height="18" rx="3" fill="#4fa8e8" />
            <rect x="26" y="4" width="18" height="18" rx="3" fill="#4fa8e8" opacity="0.6" />
            <rect x="4" y="26" width="18" height="18" rx="3" fill="#4fa8e8" opacity="0.6" />
            <rect x="26" y="26" width="18" height="18" rx="3" fill="#4fa8e8" opacity="0.85" />
          </svg>
        </div>
        <span class="sider-logo-text">pveCloud</span>
      </div>

      <!-- 导航菜单 -->
      <n-menu
        :options="menuOptions"
        :indent="18"
        :value="activeMenuKey"
        @update:value="handleMenuUpdate"
      />
    </n-layout-sider>

    <!-- 右侧主体 -->
    <n-layout>
      <!-- 顶部栏 -->
      <n-layout-header bordered class="header">
        <div class="header-left">
          <span class="header-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <n-dropdown trigger="hover" :options="userMenuOptions" @select="handleUserMenu">
            <div class="user-area">
              <n-avatar
                round
                size="small"
                color="#4fa8e8"
                style="color: #fff; font-size: 13px; font-weight: 600;"
              >
                {{ authStore.user?.nickname?.charAt(0)?.toUpperCase() || 'A' }}
              </n-avatar>
              <span class="user-name">{{ authStore.user?.nickname || authStore.user?.username }}</span>
              <n-icon size="14" style="color: #c2c2c2;"><ChevronDownOutline /></n-icon>
            </div>
          </n-dropdown>
        </div>
      </n-layout-header>

      <!-- 内容区 -->
      <n-layout-content class="main-content" :native-scrollbar="false">
        <router-view v-slot="{ Component, route }">
          <Transition name="page" mode="out-in">
            <component :is="Component" :key="route.path" />
          </Transition>
        </router-view>
      </n-layout-content>
    </n-layout>
  </n-layout>

  <!-- 修改密码弹窗 -->
  <n-modal
    v-model:show="pwdModalVisible"
    title="修改密码"
    preset="card"
    style="width: 420px;"
  >
    <n-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-placement="left" label-width="90">
      <n-form-item label="新密码" path="newPassword">
        <n-input
          v-model:value="pwdForm.newPassword"
          type="password"
          show-password-on="click"
          placeholder="至少 6 位"
        />
      </n-form-item>
      <n-form-item label="确认新密码" path="confirmPassword">
        <n-input
          v-model:value="pwdForm.confirmPassword"
          type="password"
          show-password-on="click"
          placeholder="再次输入新密码"
        />
      </n-form-item>
    </n-form>
    <template #footer>
      <n-space justify="end">
        <n-button @click="pwdModalVisible = false">取消</n-button>
        <n-button type="primary" :loading="pwdLoading" @click="handleChangePwd">
          确认修改
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<style scoped>
/* ========== 侧边栏 Logo ========== */
.sider-logo {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  border-bottom: 1px solid #efeff5;
  overflow: hidden;
  white-space: nowrap;
}

.sider-logo-icon {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}

.sider-logo-text {
  font-size: 16px;
  font-weight: 700;
  color: #18181c;
  letter-spacing: 0.5px;
}

/* ========== 顶部栏 ========== */
.header {
  height: 56px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-title {
  font-size: 15px;
  font-weight: 600;
  color: #18181c;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-area {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.15s;
}

.user-area:hover {
  background: #f5f5f5;
}

.user-name {
  font-size: 13px;
  color: #333;
}

/* ========== 内容区 ========== */
.main-content {
  padding: 24px;
  background: #f7f8fa;
  height: calc(100vh - 56px);
  overflow-y: auto;
}

/* ========== 页面切换过渡 ========== */
.page-enter-active,
.page-leave-active {
  transition: opacity 0.15s ease;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
}
</style>
