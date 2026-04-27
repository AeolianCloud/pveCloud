import type { RouteRecordRaw } from 'vue-router'

import Layout from '../../layouts/index.vue'
import { ADMIN_ROUTE_PATH } from '../constants'
import { viewRoutes } from '../view-routes'

export const dashboardRoute: RouteRecordRaw = {
  path: ADMIN_ROUTE_PATH.root,
  component: Layout,
  redirect: ADMIN_ROUTE_PATH.dashboard,
  children: viewRoutes,
}
