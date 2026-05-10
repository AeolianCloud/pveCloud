import type { App } from 'vue'
import naive from 'naive-ui'

export function setupNaive(app: App) {
  app.use(naive)
}
