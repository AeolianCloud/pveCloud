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
        component: () => import('../views/auth/index.vue'),
      },
      {
        path: 'user',
        name: 'user-center',
        meta: { requiresAuth: true },
        component: () => import('../views/user-center/index.vue'),
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../views/not-found/index.vue'),
  },
]
