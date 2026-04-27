import type { RouteRecordRaw } from 'vue-router'

import Layout from '../layouts/index.vue'
import { ADMIN_ROUTE_NAME, ADMIN_ROUTE_PATH } from './constants'

export const staticRoutes: RouteRecordRaw[] = [
  {
    path: ADMIN_ROUTE_PATH.login,
    name: ADMIN_ROUTE_NAME.login,
    component: () => import('../views/login/index.vue'),
    meta: {
      guestOnly: true,
      hidden: true,
      title: '登录',
    },
  },
  {
    path: ADMIN_ROUTE_PATH.forbidden,
    component: Layout,
    children: [
      {
        path: '',
        name: ADMIN_ROUTE_NAME.forbidden,
        component: () => import('../views/error-page/403.vue'),
        meta: {
          requiresAuth: true,
          hidden: true,
          title: '无权访问',
        },
      },
    ],
  },
]
