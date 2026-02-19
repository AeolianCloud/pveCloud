import { createRouter, createWebHistory } from 'vue-router';
import { useAdminStore } from '../stores/user';
import LoginPage from '../pages/LoginPage.vue';
import AdminLayout from '../layouts/AdminLayout.vue';
import DashboardPage from '../pages/DashboardPage.vue';
import UserManagementPage from '../pages/UserManagementPage.vue';
import ProductManagementPage from '../pages/ProductManagementPage.vue';
import OrderManagementPage from '../pages/OrderManagementPage.vue';
import TicketManagementPage from '../pages/TicketManagementPage.vue';
import NodeMonitorPage from '../pages/NodeMonitorPage.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginPage },
    {
      path: '/',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/dashboard' },
        { path: 'dashboard', component: DashboardPage },
        { path: 'users', component: UserManagementPage },
        { path: 'products', component: ProductManagementPage },
        { path: 'orders', component: OrderManagementPage },
        { path: 'tickets', component: TicketManagementPage },
        { path: 'nodes', component: NodeMonitorPage },
      ],
    },
  ],
});

router.beforeEach((to) => {
  const adminStore = useAdminStore();
  if (to.meta.requiresAuth && !adminStore.isLoggedIn) {
    return '/login';
  }
  return true;
});

export default router;
