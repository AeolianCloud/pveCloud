import type { RouteRecordRaw } from 'vue-router'

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/home/index.vue'),
    meta: {
      title: '首页',
    },
  },
  {
    path: '/products',
    name: 'products',
    component: () => import('../views/products/index.vue'),
    meta: {
      title: '产品展示',
    },
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/auth/login.vue'),
    meta: {
      title: '登录',
    },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/auth/register.vue'),
    meta: {
      title: '注册',
    },
  },
  {
    path: '/forgot-password',
    name: 'forgot-password',
    component: () => import('../views/auth/forgot-password.vue'),
    meta: {
      title: '忘记密码',
    },
  },
  {
    path: '/reset-password',
    name: 'reset-password',
    component: () => import('../views/auth/reset-password.vue'),
    meta: {
      title: '重置密码',
    },
  },
  {
    path: '/user',
    name: 'user-center',
    component: () => import('../views/user-center/index.vue'),
    meta: {
      title: '用户中心',
      requiresAuth: true,
    },
  },
  {
    path: '/user/profile',
    name: 'user-profile',
    component: () => import('../views/user-profile/index.vue'),
    meta: {
      title: '账号资料',
      requiresAuth: true,
    },
  },
  {
    path: '/user/real-name',
    name: 'real-name',
    component: () => import('../views/real-name/index.vue'),
    meta: {
      title: '实名认证',
      requiresAuth: true,
    },
  },
  {
    path: '/user/orders',
    name: 'orders',
    component: () => import('../views/orders/index.vue'),
    meta: {
      title: '订单管理',
      requiresAuth: true,
    },
  },
  {
    path: '/user/orders/:orderNo',
    name: 'order-detail',
    component: () => import('../views/orders/detail.vue'),
    meta: {
      title: '订单详情',
      requiresAuth: true,
    },
  },
  {
    path: '/user/payments/:paymentNo',
    name: 'payment',
    component: () => import('../views/payments/detail.vue'),
    meta: {
      title: '订单支付',
      requiresAuth: true,
    },
  },
  {
    path: '/user/instances',
    name: 'instances',
    component: () => import('../views/instances/index.vue'),
    meta: {
      title: '实例管理',
      requiresAuth: true,
    },
  },
  {
    path: '/user/instances/:instanceNo',
    name: 'instance-detail',
    component: () => import('../views/instances/detail.vue'),
    meta: {
      title: '实例详情',
      requiresAuth: true,
    },
  },
  {
    path: '/user/tickets',
    name: 'tickets',
    component: () => import('../views/tickets/index.vue'),
    meta: {
      title: '我的工单',
      requiresAuth: true,
    },
  },
  {
    path: '/user/tickets/new',
    name: 'ticket-create',
    component: () => import('../views/tickets/create.vue'),
    meta: {
      title: '提交工单',
      requiresAuth: true,
    },
  },
  {
    path: '/user/tickets/:ticketNo',
    name: 'ticket-detail',
    component: () => import('../views/tickets/detail.vue'),
    meta: {
      title: '工单详情',
      requiresAuth: true,
    },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../views/not-found/index.vue'),
    meta: {
      title: '页面未找到',
    },
  },
]
