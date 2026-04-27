export type RoutePermission = string | string[]

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    icon?: string
    affix?: boolean
    hidden?: boolean
    guestOnly?: boolean
    requiresAuth?: boolean
    permission?: RoutePermission
  }
}
