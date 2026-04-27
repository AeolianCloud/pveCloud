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
      permission: ['page.dashboard'],
    },
  },
  {
    path: ADMIN_ROUTE_PATH.systemSettings,
    name: ADMIN_ROUTE_NAME.systemSettings,
    redirect: ADMIN_ROUTE_PATH.systemSettingsConfig,
    meta: {
      title: '系统设置',
      icon: 'Setting',
      requiresAuth: true,
      alwaysShow: true,
    },
    children: [
      {
        path: ADMIN_ROUTE_PATH.systemSettingsConfig.replace(ADMIN_ROUTE_PATH.systemSettings + '/', ''),
        name: ADMIN_ROUTE_NAME.systemSettingsConfig,
        component: () => import('../views/system-settings/index.vue'),
        meta: {
          title: '系统配置',
          requiresAuth: true,
          permission: ['page.system-settings.config'],
        },
      },
      {
        path: ADMIN_ROUTE_PATH.adminUsers.replace(ADMIN_ROUTE_PATH.systemSettings + '/', ''),
        name: ADMIN_ROUTE_NAME.adminUsers,
        component: () => import('../views/admin-users/index.vue'),
        meta: {
          title: '管理员设置',
          requiresAuth: true,
          permission: ['page.system-settings.admin-users', 'page.system-settings.admin-roles'],
          permissionMode: 'any',
        },
      },
    ],
  },
]
