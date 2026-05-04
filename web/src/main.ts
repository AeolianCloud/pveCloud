import { createApp } from 'vue'

import App from './App.vue'
import { router } from './router'
import { pinia } from './store'
import { useWebAuthStore } from './store/modules/auth'
import { resolveWebRedirect, webAuthUnauthorizedEvent } from './utils/web-auth'
import './styles/index.css'

const app = createApp(App)

function setupUnauthorizedHandler() {
  let redirecting = false

  window.addEventListener(webAuthUnauthorizedEvent, () => {
    const authStore = useWebAuthStore(pinia)
    authStore.handleUnauthorized()

    const currentRoute = router.currentRoute.value
    if (!currentRoute.meta.requiresAuth || currentRoute.name === 'login' || redirecting) {
      return
    }

    redirecting = true
    void router
      .replace({
        name: 'login',
        query: { redirect: resolveWebRedirect(currentRoute.fullPath) },
      })
      .finally(() => {
        redirecting = false
      })
  })
}

app.use(pinia)
setupUnauthorizedHandler()
app.use(router)
app.mount('#app')
