import { fireEvent, render, screen } from '@testing-library/vue'
import { createPinia } from 'pinia'
import { beforeEach, expect, test, vi } from 'vitest'
import { createRouter, createWebHistory } from 'vue-router'

import LoginPage from './LoginPage.vue'

beforeEach(() => {
  vi.restoreAllMocks()
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        token: 'token-1',
        subject_id: 1001,
        subject_type: 'user',
      }),
    }),
  )
})

test('submits login credentials to the real auth route', async () => {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/login', component: LoginPage },
      { path: '/products', component: { template: '<div>products</div>' } },
    ],
  })
  await router.push('/login')
  await router.isReady()

  render(LoginPage, {
    global: {
      plugins: [createPinia(), router],
    },
  })

  await fireEvent.update(screen.getByLabelText('手机号'), '13800000000')
  await fireEvent.update(screen.getByLabelText('密码'), 'secret')
  await fireEvent.click(screen.getByRole('button', { name: '立即登录' }))

  expect(fetch).toHaveBeenCalledWith(
    '/api/auth/login',
    expect.objectContaining({
      method: 'POST',
    }),
  )
})
