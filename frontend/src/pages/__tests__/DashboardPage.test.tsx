import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import DashboardPage from '../DashboardPage'
import { api } from '@/api/client'
import { getAnalytics } from '@/api/analytics'

vi.mock('@/api/analytics', () => ({
  getAnalytics: vi.fn(),
}))

const mockWorkspace = { id: 'ws1', name: 'Test Workspace' }

vi.mock('@/features/organization/context/OrganizationContext', () => ({
  useOrganization: () => ({
    organizations: [],
    currentOrganization: { id: 'org1', name: 'Test Org' },
    workspaces: [],
    currentWorkspace: mockWorkspace,
    isLoading: false,
    selectOrganization: vi.fn(),
    selectWorkspace: vi.fn(),
  }),
  OrganizationProvider: ({ children }: any) => <>{children}</>,
}))

describe('DashboardPage', () => {
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

  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
  })

  it('renders stats and recent bots from API', async () => {
    const chart = Array.from({ length: 7 }).map((_, i) => ({
      date: new Date(2024, 0, i + 1).toISOString(),
      conversations: i + 1,
      messages: (i + 1) * 2,
    }))
    vi.mocked(getAnalytics).mockResolvedValue(chart as any)
    vi.mocked(api.get).mockResolvedValue({ data: [
      { id: 1, name: 'Bot A', model: 'gpt-4o' },
      { id: 2, name: 'Bot B', model: 'gpt-4.1' },
      { id: 3, name: 'Bot C', model: 'gpt-3.5' },
    ] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    expect(screen.getByText('Yükleniyor...')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Dashboard' })).toBeInTheDocument()
    })

    expect(screen.getByText('Toplam Konuşma')).toBeInTheDocument()
    expect(screen.getByText('Toplam Mesaj')).toBeInTheDocument()
    expect(screen.getByText('Harcanan Token')).toBeInTheDocument()
    expect(screen.getByText('Memnuniyet')).toBeInTheDocument()
    expect(screen.getByText('Bot A')).toBeInTheDocument()
    expect(screen.getByText('Bot B')).toBeInTheDocument()
  })

  it('handles fetch errors and shows empty chart state', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.mocked(getAnalytics).mockRejectedValue(new Error('network'))
    vi.mocked(api.get).mockResolvedValue({ data: [] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Henüz aktivite yok')).toBeInTheDocument()
    })

    expect(errSpy).toHaveBeenCalled()

    expect(screen.queryByText('Yükleniyor...')).not.toBeInTheDocument()
    expect(screen.getByText('Henüz aktivite yok')).toBeInTheDocument()
  })

  it('shows date range and growth badge when data exists', async () => {
    const chart = Array.from({ length: 7 }).map((_, i) => ({
      date: new Date(2024, 4, i + 10).toISOString(),
      conversations: 5,
      messages: 10,
    }))
    vi.mocked(getAnalytics).mockResolvedValue(chart as any)
    vi.mocked(api.get).mockResolvedValue({ data: [
      { id: 1, name: 'Bot A', model: 'gpt-4o' },
      { id: 2, name: 'Bot B', model: 'gpt-4.1' },
      { id: 3, name: 'Bot C', model: 'gpt-3.5' },
      { id: 4, name: 'Bot D', model: 'gpt-4o-mini' },
    ] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    const rangeText = await screen.findByText((t) => t.includes('10') && t.includes('16'))
    expect(rangeText).toBeInTheDocument()
    expect(screen.getAllByText('+12.5%').length).toBeGreaterThan(0)

    expect(screen.queryByText('Bot D')).not.toBeInTheDocument()
  })

  it('shows default header when chartData empty and empty bots message', async () => {
    vi.mocked(getAnalytics).mockResolvedValue([] as any)
    vi.mocked(api.get).mockResolvedValue({ data: [] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    const header = await screen.findByText('Son 7 Gün')
    expect(header).toBeInTheDocument()
    expect(screen.getAllByText('Henüz aktivite yok').length).toBeGreaterThan(0)
    
    await waitFor(() => {
        expect(screen.getAllByText('Henüz bir bot oluşturmadınız.').length).toBeGreaterThan(0)
    })
  })

  it('renders XAxis tick labels with Turkish short month', async () => {
    const chart = [
      { date: new Date(2024, 4, 10).toISOString(), conversations: 1, messages: 2 },
      { date: new Date(2024, 4, 11).toISOString(), conversations: 1, messages: 2 },
    ]
    vi.mocked(getAnalytics).mockResolvedValue(chart as any)
    vi.mocked(api.get).mockResolvedValue({ data: [{ id: 1, name: 'Bot', model: 'gpt-4o' }] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    expect(
      await screen.findByText((t) => /\b(10|11)\s*May\b/i.test(t))
    ).toBeInTheDocument()
  })
})
