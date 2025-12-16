import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from '@/App'

vi.mock('@/features/organization/context/OrganizationContext', () => ({
  useOrganization: () => ({
    currentWorkspace: { id: 'ws-1' },
    isLoading: false,
  }),
  OrganizationProvider: ({ children }: any) => children,
}))

vi.mock('@/components/layout/DashboardLayout', async () => {
  const React = await import('react')
  const { Outlet } = await import('react-router-dom')
  return {
    default: () => React.createElement(React.Fragment, null, React.createElement(Outlet)),
  }
})

vi.mock('@/pages/DashboardPage', async () => {
  const React = await import('react')
  return { default: () => React.createElement('div', null, 'Dashboard') }
})

describe('PrivateRoute', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })
  })
  it('renders protected layout when authenticated', () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('tok')
    render(<App />)
    expect(screen.queryByRole('heading', { name: /Giriş Yap/i })).not.toBeInTheDocument()
  })
})
