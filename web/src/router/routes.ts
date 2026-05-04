import type { RouteRecordRaw } from 'vue-router'

import WebLayout from '../layouts/WebLayout.vue'

export const publicRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: WebLayout,
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('../views/home/index.vue'),
      },
      {
        path: 'products',
        name: 'products',
        component: () => import('../views/products/index.vue'),
      },
      {
        path: 'pricing',
        name: 'pricing',
        component: () => import('../views/pricing/index.vue'),
      },
      {
        path: 'login',
        name: 'login',
        meta: { transitionName: 'auth-swap' },
        component: () => import('../views/auth/index.vue'),
      },
      {
        path: 'register',
        name: 'register',
        meta: { transitionName: 'auth-swap' },
        component: () => import('../views/auth/register.vue'),
      },
      {
        path: 'forgot-password',
        name: 'forgot-password',
        meta: { transitionName: 'auth-swap' },
        component: () => import('../views/auth/forgot-password.vue'),
      },
      {
        path: 'reset-password',
        name: 'reset-password',
        meta: { transitionName: 'auth-swap' },
        component: () => import('../views/auth/reset-password.vue'),
      },
      {
        path: 'user',
        name: 'user-center',
        meta: { requiresAuth: true },
        component: () => import('../views/user-center/index.vue'),
      },
      {
        path: 'user/profile',
        name: 'user-profile',
        meta: { requiresAuth: true },
        component: () => import('../views/user-profile/index.vue'),
      },
      {
        path: 'user/real-name',
        name: 'user-real-name',
        meta: { requiresAuth: true },
        component: () => import('../views/real-name/index.vue'),
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../views/not-found/index.vue'),
  },
]
