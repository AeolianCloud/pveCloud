import { createRouter, createWebHistory } from 'vue-router'

import DashboardView from '../views/DashboardView.vue'
import InstanceManageView from '../views/InstanceManageView.vue'
import LoginView from '../views/LoginView.vue'
import OrderManageView from '../views/OrderManageView.vue'
import ProductManageView from '../views/ProductManageView.vue'
import TaskManageView from '../views/TaskManageView.vue'
import UserManageView from '../views/UserManageView.vue'

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
