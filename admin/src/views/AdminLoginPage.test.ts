import { fireEvent, render, screen } from '@testing-library/vue'
import { createPinia } from 'pinia'
import { beforeEach, expect, test, vi } from 'vitest'
import { createRouter, createWebHistory } from 'vue-router'

import AdminLoginPage from './AdminLoginPage.vue'

beforeEach(() => {
  vi.restoreAllMocks()
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        token: 'token-1',
        subject_id: 9001,
        subject_type: 'admin',
      }),
    }),
  )
})

test('submits admin credentials to the real auth route', async () => {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/login', component: AdminLoginPage },
      { path: '/products', component: { template: '<div>products</div>' } },
    ],
  })
  await router.push('/login')
  await router.isReady()

  render(AdminLoginPage, {
    global: {
      plugins: [createPinia(), router],
    },
  })

  await fireEvent.update(screen.getByLabelText('用户名'), 'admin')
  await fireEvent.update(screen.getByLabelText('密码'), 'secret')
  await fireEvent.click(screen.getByRole('button', { name: '登录后台' }))

  expect(fetch).toHaveBeenCalledWith(
    '/admin-api/auth/login',
    expect.objectContaining({
      method: 'POST',
    }),
  )
})
