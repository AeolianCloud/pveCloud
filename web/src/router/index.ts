import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { routes } from './routes'

const router = createRouter({
  history: createWebHistory(),
  routes,
})

const guestOnlyRoutes = ['login', 'register']

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth) {
    const restored = authStore.isAuthenticated || await authStore.restoreAuth()
    if (!restored) {
      return {
        path: '/login',
        query: { redirect: to.fullPath },
      }
    }
  }

  if (guestOnlyRoutes.includes(String(to.name))) {
    const restored = authStore.isAuthenticated || await authStore.restoreAuth()
    if (restored) {
      return '/user'
    }
  }
})

export default router
