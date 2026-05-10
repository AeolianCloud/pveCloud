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
          permission: ['page.system-settings.admin-users', 'page.system-settings.admin-roles', 'page.system-settings.admin-sessions'],
          permissionMode: 'any',
        },
      },
      {
        path: ADMIN_ROUTE_PATH.auditLogs.replace(ADMIN_ROUTE_PATH.systemSettings + '/', ''),
        name: ADMIN_ROUTE_NAME.auditLogs,
        component: () => import('../views/audit-logs/index.vue'),
        meta: {
          title: '日志管理',
          requiresAuth: true,
          permission: ['page.system-settings.audit-logs'],
        },
      },
    ],
  },
  {
    path: ADMIN_ROUTE_PATH.files,
    name: ADMIN_ROUTE_NAME.files,
    component: () => import('../views/file-management/index.vue'),
    meta: {
      title: '附件管理',
      icon: 'FolderOpened',
      requiresAuth: true,
      permission: ['page.file-management'],
    },
  },
  {
    path: ADMIN_ROUTE_PATH.webUsers,
    name: ADMIN_ROUTE_NAME.webUsers,
    component: () => import('../views/web-users/index.vue'),
    meta: {
      title: 'Web 用户管理',
      icon: 'User',
      requiresAuth: true,
      permission: ['page.web-users', 'page.web-user-sessions'],
      permissionMode: 'any',
    },
  },
  {
    path: ADMIN_ROUTE_PATH.realNames,
    name: ADMIN_ROUTE_NAME.realNames,
    component: () => import('../views/real-names/index.vue'),
    meta: {
      title: '实名管理',
      icon: 'Checked',
      requiresAuth: true,
      permission: ['page.real-name-management'],
    },
  },
  {
    path: ADMIN_ROUTE_PATH.products,
    name: ADMIN_ROUTE_NAME.products,
    component: () => import('../views/products/index.vue'),
    meta: {
      title: '产品管理',
      icon: 'Box',
      requiresAuth: true,
      permission: ['page.products'],
    },
  },
  {
    path: ADMIN_ROUTE_PATH.orders,
    name: ADMIN_ROUTE_NAME.orders,
    component: () => import('../views/orders/index.vue'),
    meta: {
      title: '订单管理',
      icon: 'Tickets',
      requiresAuth: true,
      permission: ['page.orders'],
    },
  },
]
