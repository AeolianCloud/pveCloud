import { createRouter, createWebHistory } from 'vue-router'

import AdminLayout from '../components/AdminLayout.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import ForbiddenPage from '../pages/ForbiddenPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import PlaceholderPage from '../pages/PlaceholderPage.vue'
import { useAuthStore } from '../stores/auth'

const protectedPlaceholderRoutes = [
  { path: 'instances', name: 'instances', title: '云服务器', permissionCode: 'instance:view' },
  { path: 'products', name: 'products', title: '产品套餐', permissionCode: 'product:update' },
  { path: 'orders', name: 'orders', title: '订单管理', permissionCode: 'order:view' },
  { path: 'users', name: 'users', title: '客户管理', permissionCode: 'user:view' },
  { path: 'tickets', name: 'tickets', title: '工单服务', permissionCode: 'ticket:reply' },
  { path: 'payments', name: 'payments', title: '财务中心', permissionCode: 'payment:view' },
  { path: 'audit', name: 'audit', title: '资源监控', permissionCode: 'audit:view' },
  { path: 'admins', name: 'admins', title: '营销活动', permissionCode: 'admin:manage' },
  { path: 'system', name: 'system', title: '系统设置', permissionCode: 'system:update' },
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
