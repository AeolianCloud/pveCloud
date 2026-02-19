import { createRouter, createWebHistory } from 'vue-router';
import { useUserStore } from '../stores/user';
import HomePage from '../pages/public/HomePage.vue';
import ProductListPage from '../pages/public/ProductListPage.vue';
import ProductDetailPage from '../pages/public/ProductDetailPage.vue';
import LoginPage from '../pages/public/LoginPage.vue';
import RegisterPage from '../pages/public/RegisterPage.vue';
import ConsoleLayout from '../layouts/ConsoleLayout.vue';
import InstanceListPage from '../pages/user/InstanceListPage.vue';
import InstanceDetailPage from '../pages/user/InstanceDetailPage.vue';
import OrderFlowPage from '../pages/user/OrderFlowPage.vue';
import WalletPage from '../pages/user/WalletPage.vue';
import TicketListPage from '../pages/user/TicketListPage.vue';
import TicketDetailPage from '../pages/user/TicketDetailPage.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomePage },
    { path: '/products', component: ProductListPage },
    { path: '/products/:id', component: ProductDetailPage },
    { path: '/login', component: LoginPage },
    { path: '/register', component: RegisterPage },
    {
      path: '/console',
      component: ConsoleLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/console/instances' },
        { path: 'instances', component: InstanceListPage },
        { path: 'instances/:id', component: InstanceDetailPage },
        { path: 'order-flow', component: OrderFlowPage },
        { path: 'wallet', component: WalletPage },
        { path: 'tickets', component: TicketListPage },
        { path: 'tickets/:id', component: TicketDetailPage },
      ],
    },
  ],
});

router.beforeEach((to) => {
  const userStore = useUserStore();
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    return '/login';
  }
  return true;
});

export default router;
