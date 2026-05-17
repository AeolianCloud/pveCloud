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
        redirect: ADMIN_ROUTE_PATH.adminOperationLogs,
        meta: {
          title: '日志管理',
          requiresAuth: true,
          hidden: true,
          permission: ['page.logs.admin-operations', 'page.system-settings.audit-logs'],
          permissionMode: 'any',
        },
      },
    ],
  },
  {
    path: ADMIN_ROUTE_PATH.logs,
    name: ADMIN_ROUTE_NAME.logs,
    redirect: ADMIN_ROUTE_PATH.adminOperationLogs,
    meta: {
      title: '日志管理中心',
      icon: 'DocumentText',
      requiresAuth: true,
      alwaysShow: true,
      permission: ['page.logs'],
    },
    children: [
      {
        path: ADMIN_ROUTE_PATH.adminOperationLogs.replace(ADMIN_ROUTE_PATH.logs + '/', ''),
        name: ADMIN_ROUTE_NAME.adminOperationLogs,
        component: () => import('../views/audit-logs/index.vue'),
        meta: {
          title: '操作审计',
          requiresAuth: true,
          permission: ['page.logs.admin-operations', 'page.system-settings.audit-logs'],
          permissionMode: 'any',
        },
      },
      {
        path: ADMIN_ROUTE_PATH.adminSecurityLogs.replace(ADMIN_ROUTE_PATH.logs + '/', ''),
        name: ADMIN_ROUTE_NAME.adminSecurityLogs,
        component: () => import('../views/audit-logs/index.vue'),
        meta: {
          title: '登录安全',
          requiresAuth: true,
          permission: ['page.logs.admin-security'],
        },
      },
      {
        path: ADMIN_ROUTE_PATH.userSecurityLogs.replace(ADMIN_ROUTE_PATH.logs + '/', ''),
        name: ADMIN_ROUTE_NAME.userSecurityLogs,
        component: () => import('../views/logs/index.vue'),
        meta: {
          title: '用户安全日志',
          requiresAuth: true,
          permission: ['page.logs.user-security'],
        },
      },
      {
        path: ADMIN_ROUTE_PATH.userBusinessLogs.replace(ADMIN_ROUTE_PATH.logs + '/', ''),
        name: ADMIN_ROUTE_NAME.userBusinessLogs,
        component: () => import('../views/logs/index.vue'),
        meta: {
          title: '用户业务日志',
          requiresAuth: true,
          permission: ['page.logs.user-business'],
        },
      },
      {
        path: ADMIN_ROUTE_PATH.frontendErrorLogs.replace(ADMIN_ROUTE_PATH.logs + '/', ''),
        name: ADMIN_ROUTE_NAME.frontendErrorLogs,
        component: () => import('../views/logs/index.vue'),
        meta: {
          title: '前端错误日志',
          requiresAuth: true,
          permission: ['page.logs.frontend-errors'],
        },
      },
      {
        path: ADMIN_ROUTE_PATH.backendRuntimeLogs.replace(ADMIN_ROUTE_PATH.logs + '/', ''),
        name: ADMIN_ROUTE_NAME.backendRuntimeLogs,
        component: () => import('../views/logs/index.vue'),
        meta: {
          title: '后端运行日志',
          requiresAuth: true,
          permission: ['page.logs.backend-runtime'],
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
  {
    path: ADMIN_ROUTE_PATH.instances,
    name: ADMIN_ROUTE_NAME.instances,
    component: () => import('../views/instances/index.vue'),
    meta: {
      title: '实例管理',
      icon: 'Server',
      requiresAuth: true,
      permission: ['page.instances'],
    },
  },
  {
    path: ADMIN_ROUTE_PATH.tickets,
    name: ADMIN_ROUTE_NAME.tickets,
    component: () => import('../views/tickets/index.vue'),
    meta: {
      title: '工单管理',
      icon: 'Chatbubbles',
      requiresAuth: true,
      permission: ['page.tickets'],
    },
  },
]
