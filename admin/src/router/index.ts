import { createRouter, createWebHistory } from 'vue-router'

import AdminLayout from '../components/AdminLayout.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import ForbiddenPage from '../pages/ForbiddenPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import AuditLogPage from '../pages/AuditLogPage.vue'
import AdminUserPage from '../pages/AdminUserPage.vue'
import AdminRolePage from '../pages/AdminRolePage.vue'
import AdminSessionPage from '../pages/AdminSessionPage.vue'
import SystemConfigPage from '../pages/SystemConfigPage.vue'
import { useAuthStore } from '../stores/auth'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      path: '/login',
      name: 'login',
      component: LoginPage,
      meta: {
        guestOnly: true,
      },
    },
    {
      path: '/',
      component: AdminLayout,
      meta: {
        requiresAuth: true,
      },
      children: [
        {
          path: 'dashboard',
          name: 'dashboard',
          component: DashboardPage,
          meta: {
            title: '控制台',
            requiresAuth: true,
            permissionCode: 'dashboard:view',
          },
        },
        {
          path: '403',
          name: 'forbidden',
          component: ForbiddenPage,
          meta: {
            title: '无权访问',
            requiresAuth: true,
          },
        },
        {
          path: 'admin-users',
          name: 'admin_users',
          component: AdminUserPage,
          meta: {
            title: '管理员账号',
            requiresAuth: true,
            permissionCode: 'admin:manage',
          },
        },
        {
          path: 'admin-roles',
          name: 'admin_roles',
          component: AdminRolePage,
          meta: {
            title: '角色权限',
            requiresAuth: true,
            permissionCode: 'admin:manage',
          },
        },
        {
          path: 'admin-sessions',
          name: 'admin_sessions',
          component: AdminSessionPage,
          meta: {
            title: '登录会话',
            requiresAuth: true,
            permissionCode: 'admin:manage',
          },
        },
        {
          path: 'system-configs',
          name: 'system_configs',
          component: SystemConfigPage,
          meta: {
            title: '系统设置',
            requiresAuth: true,
            permissionCode: 'system:update',
          },
        },
        {
          path: 'audit-logs',
          name: 'audit_logs',
          component: AuditLogPage,
          meta: {
            title: '审计日志',
            requiresAuth: true,
            permissionCode: 'audit:view',
          },
        },
        {
          path: 'risk-logs',
          name: 'risk_logs',
          component: AuditLogPage,
          meta: {
            title: '高危操作日志',
            requiresAuth: true,
            permissionCode: 'audit:view',
          },
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)
  let permissionCode: unknown
  for (const record of to.matched) {
    if (record.meta.permissionCode) {
      permissionCode = record.meta.permissionCode
    }
  }

  if (auth.hasLocalToken && !auth.restored) {
    await auth.restore()
  }

  if (requiresAuth && !auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guestOnly && auth.hasLocalToken && !auth.isLoggedIn) {
    await auth.restore()
  }
  if (to.meta.guestOnly && auth.isLoggedIn) {
    return { name: 'dashboard' }
  }
  if (requiresAuth && typeof permissionCode === 'string' && !auth.hasPermission(permissionCode)) {
    return { name: 'forbidden' }
  }
  return true
})
