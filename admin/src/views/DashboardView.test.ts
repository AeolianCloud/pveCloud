import { render, screen } from '@testing-library/vue'
import { expect, test } from 'vitest'
import DashboardView from './DashboardView.vue'

test('renders admin dashboard title', () => {
  render(DashboardView)
  expect(screen.getByText('管理仪表盘')).toBeTruthy()
})
