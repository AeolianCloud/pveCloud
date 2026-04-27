import { createRouter, createWebHistory } from 'vue-router'

import { ADMIN_ROUTE_PATH } from './constants'
import { protectedRoutes } from './modules'
import { staticRoutes } from './static'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    ...staticRoutes,
    ...protectedRoutes,
    {
      path: '/:pathMatch(.*)*',
      redirect: ADMIN_ROUTE_PATH.dashboard,
      meta: {
        hidden: true,
      },
    },
  ],
  scrollBehavior() {
    return { left: 0, top: 0 }
  },
})
