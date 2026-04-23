import { createRouter, createWebHistory } from 'vue-router'

import { readStoredToken } from '../lib/http'
import DashboardView from '../views/DashboardPlaceholderPage.vue'
import InstanceManageView from '../views/InstanceManagePage.vue'
import LoginView from '../views/AdminLoginPage.vue'
import OrderManageView from '../views/OrderManagePage.vue'
import ProductManageView from '../views/ProductManagePage.vue'
import TaskManageView from '../views/TaskManagePage.vue'
import UserManageView from '../views/UserManagePlaceholderPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginView },
    { path: '/', component: DashboardView },
    { path: '/users', component: UserManageView },
    { path: '/products', component: ProductManageView },
    { path: '/orders', component: OrderManageView },
    { path: '/instances', component: InstanceManageView },
    { path: '/tasks', component: TaskManageView },
  ],
})

router.beforeEach((to) => {
  if (to.path === '/login') {
    return true
  }

  if (readStoredToken()) {
    return true
  }

  return {
    path: '/login',
    query: { redirect: to.fullPath },
  }
})
