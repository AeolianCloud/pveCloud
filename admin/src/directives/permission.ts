import type { App, Directive } from 'vue'

import { usePermissionStore } from '../store/modules/permission'

const permissionDirective: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    updateVisibility(el, binding.value)
  },
  updated(el, binding) {
    updateVisibility(el, binding.value)
  },
}

export function setupPermissionDirective(app: App) {
  app.directive('permission', permissionDirective)
}

function updateVisibility(el: HTMLElement, required?: string | string[]) {
  const permissionStore = usePermissionStore()
  el.style.display = permissionStore.hasPermission(required) ? '' : 'none'
}
