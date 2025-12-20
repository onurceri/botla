import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import DashboardLayout from '../DashboardLayout'
import { OrganizationProvider } from '@/features/organization/context/OrganizationContext'
import { ToastProvider } from '@/components/ui/toast'
import { api } from '@/api/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// Mock api
vi.mock('@/api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  }
}))

// Mock organization api
vi.mock('@/api/organization', () => ({
  getOrganizations: vi.fn().mockResolvedValue([]),
}))

// Mock workspace api
vi.mock('@/api/workspace', () => ({
  getWorkspaces: vi.fn().mockResolvedValue([]),
}))

// Mock plan api
vi.mock('@/api/plan', () => ({
  getPlan: vi.fn().mockResolvedValue({
    limits: { max_chatbots: 5, max_messages: 1000 },
    features: { secure_embed: true },
    available_models: ['gpt-4o-mini']
  }),
}))

describe('DashboardLayout', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })

    // Default mock for me
    vi.mocked(api.get).mockResolvedValue({
        data: { full_name: 'Test User', email: 'test@example.com' }
    })
  })

  const renderWithProviders = (_: React.ReactNode, { initialEntries = ["/"] } = {}) => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } }
    })
    return render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={initialEntries}>
          <ToastProvider>
            <OrganizationProvider>
                 <Routes>
                    <Route path="/" element={<DashboardLayout />}>
                        <Route index element={<div>Dashboard Content</div>} />
                         <Route path="chatbots" element={<div>Chatbots Content</div>} />
                    </Route>
                     <Route path="/login" element={<h1>Login Page</h1>} />
                 </Routes>
            </OrganizationProvider>
          </ToastProvider>
        </MemoryRouter>
      </QueryClientProvider>
    )
  }

  it('toggles sidebar mode and persists to localStorage', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('pinned')
    const setSpy = vi.spyOn(window.localStorage, 'setItem')

    renderWithProviders(null)

    const toggles = screen.getAllByTitle('Sabit → Hover')
    toggles[0].click()
    expect(setSpy).toHaveBeenCalledWith('botla_sidebar_mode', 'hover')
  })

  it('logs out and navigates to login', async () => {
     vi.spyOn(window.localStorage, 'getItem').mockImplementation((key: string) => {
      if (key === 'botla_token') return 'tok'
      return null
    })
    const removeSpy = vi.spyOn(window.localStorage, 'removeItem')

    renderWithProviders(null)

    const logoutBtns = screen.getAllByRole('button', { name: /Çıkış Yap/i })
    logoutBtns[0].click()
    
    expect(removeSpy).toHaveBeenCalledWith('botla_token')
    expect(removeSpy).toHaveBeenCalledWith('botla_refresh_token')
    
    expect(await screen.findByText('Login Page')).toBeInTheDocument()
  })

  it('opens mobile menu overlay and closes on click', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('pinned')
    const { container } = renderWithProviders(null)
    
    const headers = container.querySelectorAll('header')
    const header = headers[0]
    const menuBtn = header.querySelector('button') as HTMLButtonElement
    menuBtn.click()
    await new Promise((r) => setTimeout(r, 0))
    const overlays = container.querySelectorAll('.fixed.inset-0')
    const overlay = overlays[0] as HTMLDivElement
    expect(overlay).not.toBeNull()
    overlay.click()
    await new Promise((r) => setTimeout(r, 0))
    expect(container.querySelector('.fixed.inset-0')).toBeNull()
  })

  it('shows breadcrumb label for Chatbots route', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('pinned')
    
    renderWithProviders(null, { initialEntries: ["/chatbots"] })

    const banners = screen.getAllByRole('banner')
    expect(banners.length).toBeGreaterThan(0)
    // Look for breadcrumb
    expect(screen.getAllByText('Chatbotlar')[0]).toBeInTheDocument()
  })
})
