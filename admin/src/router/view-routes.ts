import type { RouteRecordRaw } from 'vue-router'

import { ADMIN_ROUTE_NAME, ADMIN_ROUTE_PATH } from './constants'

export const dashboardViewRoutes: RouteRecordRaw[] = [
  {
    path: ADMIN_ROUTE_PATH.dashboard,
    name: ADMIN_ROUTE_NAME.dashboard,
    component: () => import('../views/dashboard/index.vue'),
    meta: {
      title: '控制台',
      icon: 'Odometer',
      affix: true,
      requiresAuth: true,
      permission: ['dashboard:view'],
    },
  },
]

export const protectedViewRoutes: RouteRecordRaw[] = [...dashboardViewRoutes]
