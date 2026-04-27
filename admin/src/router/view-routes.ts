import type { RouteRecordRaw } from 'vue-router'

import { ADMIN_ROUTE_NAME, ADMIN_ROUTE_PATH } from './constants'

/**
 * 受保护子路由 —— 挂在 Layout 下，菜单也从这里生成。
 * 新增页面只需往这里加，再在 modules/ 下注册对应的顶级路由即可。
 */
export const viewRoutes: RouteRecordRaw[] = [
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
