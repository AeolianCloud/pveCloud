import { createRouter, createWebHistory } from 'vue-router'

import InstanceDetailView from '../views/InstanceDetailView.vue'
import InstanceListView from '../views/InstanceListView.vue'
import LoginView from '../views/LoginView.vue'
import NoticeListView from '../views/NoticeListView.vue'
import OrderListView from '../views/OrderListView.vue'
import PaymentStatusView from '../views/PaymentStatusView.vue'
import ProductDetailView from '../views/ProductDetailView.vue'
import ProductListView from '../views/ProductListView.vue'
import RegisterView from '../views/RegisterView.vue'

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
