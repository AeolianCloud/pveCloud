import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/403',
      name: 'Forbidden',
      component: () => import('@/views/error/ForbiddenView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/layouts/DefaultLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard',
        },
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
          meta: { title: '控制台' },
        },
        {
          path: 'system/admin-users',
          name: 'AdminUsers',
          component: () => import('@/views/system/AdminUsersView.vue'),
          meta: { title: '管理员账号', permission: 'admin:list' },
        },
        {
          path: 'system/roles',
          name: 'Roles',
          component: () => import('@/views/system/RolesView.vue'),
          meta: { title: '角色管理', permission: 'role:list' },
        },
        {
          path: 'system/login-logs',
          name: 'LoginLogs',
          component: () => import('@/views/system/LoginLogsView.vue'),
          meta: { title: '登录日志', permission: 'log:list' },
        },
        {
          path: 'system/op-logs',
          name: 'OpLogs',
          component: () => import('@/views/system/OpLogsView.vue'),
          meta: { title: '操作日志', permission: 'op:list' },
        },
      ],
    },
    // 未匹配路由重定向到首页
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

// 全局路由守卫：鉴权 + 权限检查
router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  // 未登录跳转登录页
  if (to.meta.requiresAuth !== false && !authStore.isLoggedIn) {
    return { name: 'Login' }
  }

  // 已登录访问登录页，跳转首页
  if (to.name === 'Login' && authStore.isLoggedIn) {
    return { path: '/dashboard' }
  }

  // 已登录但 user 尚未加载（刷新页面场景），先拉取用户信息
  // 这样后续权限检查和菜单渲染都能拿到完整的 roles + permissions 数据
  if (authStore.isLoggedIn && !authStore.user) {
    try {
      await authStore.fetchUser()
    } catch {
      // token 已失效，清除登录态跳登录页
      authStore.logout()
      return { name: 'Login' }
    }
  }

  // 权限检查：路由 meta 有 permission 字段时才校验
  if (to.meta.permission && !authStore.hasPermission(to.meta.permission as string)) {
    return { path: '/403' }
  }
})

export default router
