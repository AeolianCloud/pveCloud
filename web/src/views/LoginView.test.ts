import { render, screen } from '@testing-library/vue'
import { expect, test } from 'vitest'
import LoginView from './LoginView.vue'

test('renders login form fields', () => {
  render(LoginView)
  expect(screen.getByLabelText('手机号')).toBeTruthy()
  expect(screen.getByLabelText('密码')).toBeTruthy()
})
