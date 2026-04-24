import { createRouter, createWebHistory } from 'vue-router'

import { readStoredToken } from '../lib/http'
import InstanceDetailView from '../views/InstanceDetailPage.vue'
import InstanceListView from '../views/InstanceListPage.vue'
import LoginView from '../views/LoginPage.vue'
import NoticeListView from '../views/NoticeListPage.vue'
import OrderListView from '../views/OrderListPage.vue'
import PaymentStatusView from '../views/PaymentStatusPage.vue'
import ProductDetailView from '../views/ProductDetailPage.vue'
import ProductListView from '../views/ProductListPage.vue'
import RegisterView from '../views/RegisterPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/products' },
    { path: '/login', component: LoginView },
    { path: '/register', component: RegisterView },
    { path: '/products', component: ProductListView },
    { path: '/products/:id', component: ProductDetailView },
    { path: '/orders', component: OrderListView },
    { path: '/payment/:paymentOrderNo', component: PaymentStatusView },
    { path: '/instances', component: InstanceListView },
    { path: '/instances/:id', component: InstanceDetailView },
    { path: '/notices', component: NoticeListView },
  ],
})

router.beforeEach((to) => {
  const requiresAuth = to.path === '/orders' || to.path.startsWith('/instances') || to.path === '/notices'
  if (!requiresAuth) {
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
