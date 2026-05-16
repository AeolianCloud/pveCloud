import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { installClientErrorReporter } from './utils/client-log'
import './styles/index.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

installClientErrorReporter()
app.mount('#app')
