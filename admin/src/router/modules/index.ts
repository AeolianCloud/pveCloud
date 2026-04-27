import type { RouteRecordRaw } from 'vue-router'

import { dashboardRoute } from './dashboard'

export const protectedRoutes: RouteRecordRaw[] = [dashboardRoute]
