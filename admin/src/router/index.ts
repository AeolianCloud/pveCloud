import { createRouter, createWebHistory } from 'vue-router'

import AdminLayout from '../components/AdminLayout.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import ForbiddenPage from '../pages/ForbiddenPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import PlaceholderPage from '../pages/PlaceholderPage.vue'
import { useAuthStore } from '../stores/auth'

const protectedPlaceholderRoutes = [
  { path: 'users', name: 'users', title: '用户', permissionCode: 'user:view' },
  { path: 'products', name: 'products', title: '产品配置', permissionCode: 'product:update' },
  { path: 'orders', name: 'orders', title: '订单', permissionCode: 'order:view' },
  { path: 'payments', name: 'payments', title: '支付与钱包', permissionCode: 'payment:view' },
  { path: 'instances', name: 'instances', title: '实例', permissionCode: 'instance:view' },
  { path: 'tickets', name: 'tickets', title: '工单', permissionCode: 'ticket:reply' },
  { path: 'admins', name: 'admins', title: '管理员', permissionCode: 'admin:manage' },
  { path: 'system', name: 'system', title: '系统设置', permissionCode: 'system:update' },
  { path: 'audit', name: 'audit', title: '审计日志', permissionCode: 'audit:view' },
]

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
            title: '首页',
            requiresAuth: true,
            permissionCode: 'dashboard:view',
          },
        },
        {
          path: '403',
          name: 'forbidden',
          component: ForbiddenPage,
          meta: {
            title: '无权限',
            requiresAuth: true,
          },
        },
        ...protectedPlaceholderRoutes.map((item) => ({
          path: item.path,
          name: item.name,
          component: PlaceholderPage,
          meta: {
            title: item.title,
            requiresAuth: true,
            permissionCode: item.permissionCode,
          },
        })),
      ],
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)
  let permissionCode: unknown
  for (const record of to.matched) {
    if (record.meta.permissionCode) {
      permissionCode = record.meta.permissionCode
    }
  }

  if (requiresAuth && !auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guestOnly && auth.isLoggedIn) {
    return { name: 'dashboard' }
  }
  if (requiresAuth && typeof permissionCode === 'string' && !auth.hasPermission(permissionCode)) {
    return { name: 'forbidden' }
  }
  return true
})
