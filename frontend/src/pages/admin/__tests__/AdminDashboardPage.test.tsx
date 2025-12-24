import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { AdminDashboardPage } from '../AdminDashboardPage'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as adminApi from '@/api/admin'

// Mock the API
vi.mock('@/api/admin', () => ({
  getOverviewStats: vi.fn(),
  getDetailedHealth: vi.fn(),
}))

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('AdminDashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    
    // Default health mock to avoid errors in HealthPanel
    vi.mocked(adminApi.getDetailedHealth).mockResolvedValue({
      status: 'healthy',
      version: '1.0.0',
      uptime: '10d',
      environment: 'production',
      dependencies: [],
    })
  })

  it('renders page title and description', () => {
    vi.mocked(adminApi.getOverviewStats).mockReturnValue(new Promise(() => {}))
    
    render(<AdminDashboardPage />, { wrapper: createWrapper() })
    
    expect(screen.getByText('Genel Bakış')).toBeInTheDocument()
    expect(screen.getByText(/Platform genel istatistikleri/i)).toBeInTheDocument()
  })

  it('renders stats from API', async () => {
    vi.mocked(adminApi.getOverviewStats).mockResolvedValue({
      total_users: 1500,
      total_organizations: 150,
      total_chatbots: 300,
      total_messages: 50000,
      users_today: 10,
      conversations_today: 500,
    })

    render(<AdminDashboardPage />, { wrapper: createWrapper() })

    // Check stats cards
    const expectedUsers = (1500).toLocaleString()
    const expectedMessages = (50000).toLocaleString()
    
    expect(await screen.findByText(expectedUsers)).toBeInTheDocument()
    expect(screen.getByText('150')).toBeInTheDocument()
    expect(screen.getByText('300')).toBeInTheDocument()
    expect(screen.getByText(expectedMessages)).toBeInTheDocument()
    
    // Check subtitles
    expect(screen.getByText('+10 today')).toBeInTheDocument()
    expect(screen.getByText('+500 today')).toBeInTheDocument()
  })

  it('handles empty stats gracefully', async () => {
    vi.mocked(adminApi.getOverviewStats).mockResolvedValue({
      total_users: 0,
      total_organizations: 0,
      total_chatbots: 0,
      total_messages: 0,
    })

    render(<AdminDashboardPage />, { wrapper: createWrapper() })

    expect(await screen.findAllByText('0')).toHaveLength(4)
  })
})
