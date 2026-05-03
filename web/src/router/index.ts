import { createRouter, createWebHistory } from 'vue-router'

import { publicRoutes } from './routes'
import { useWebAuthStore } from '../store/modules/auth'

export const router = createRouter({
  history: createWebHistory(),
  routes: publicRoutes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const authStore = useWebAuthStore()
  const loggedIn = await authStore.restore()
  if (to.meta.requiresAuth && !loggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && loggedIn) {
    return { name: 'user-center' }
  }
})
