import { createApp } from 'vue'

import App from './App.vue'
import { setupPermissionDirective } from './directives/permission'
import { setupElement } from './plugins/element'
import { router } from './router'
import { pinia } from './store'
import './permission'
import './styles/index.css'

const app = createApp(App)

app.use(pinia)
app.use(router)
setupElement(app)
setupPermissionDirective(app)
app.mount('#app')
