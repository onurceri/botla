import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, QueryWrapper } from '@/test-utils'
import App from '@/App'
import { api } from '@/api/client'

vi.mock('@/features/organization/context/OrganizationContext', () => ({
  useOrganization: () => ({
    currentWorkspace: { id: 'ws-1' },
    isLoading: false,
  }),
  OrganizationProvider: ({ children }: { children: React.ReactNode }) => children,
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

// Auth is cookie based: PrivateRoute trusts a successful GET /api/v1/me,
// not anything in localStorage.
describe('PrivateRoute', () => {
  beforeEach(() => {
    window.localStorage.clear()
    window.history.pushState({}, 'Test page', '/dashboard')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders protected layout when /api/v1/me succeeds', async () => {
    vi.spyOn(api, 'get').mockImplementation(async (url: string) => {
      if (url === '/api/v1/me') {
        return { data: { id: 'u1', email: 'u@example.com', name: 'User' } }
      }
      return { data: {} }
    })
    render(<App />, { wrapper: QueryWrapper })
    expect(await screen.findByText('Dashboard')).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Hoş Geldiniz' })).not.toBeInTheDocument()
  })

  it('redirects to login when /api/v1/me returns 401', async () => {
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
    render(<App />, { wrapper: QueryWrapper })
    expect(await screen.findByRole('heading', { name: 'Hoş Geldiniz' })).toBeInTheDocument()
    expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()
  })
})
