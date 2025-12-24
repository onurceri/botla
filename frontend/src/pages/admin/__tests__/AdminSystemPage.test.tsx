import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AdminSystemPage } from '../AdminSystemPage'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as adminApi from '@/api/admin'

// Mock the API
vi.mock('@/api/admin', () => ({
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

describe('AdminSystemPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders page title and description', () => {
     vi.mocked(adminApi.getDetailedHealth).mockReturnValue(new Promise(() => {}))
    
    render(<AdminSystemPage />, { wrapper: createWrapper() })
    
    expect(screen.getByText('Sistem Durumu')).toBeInTheDocument()
    expect(screen.getByText(/Sistem bileşenlerinin sağlık durumunu/i)).toBeInTheDocument()
  })

  it('renders health details from API', async () => {
    const mockData: adminApi.DetailedHealth = {
      status: 'healthy',
      version: '1.2.3',
      uptime: '5d 4h',
      environment: 'production',
      dependencies: [
        {
          name: 'database',
          status: 'ok',
          latency_ms: 25,
          checked_at: new Date().toISOString(),
        },
        {
          name: 'redis',
          status: 'degraded',
          latency_ms: 150,
          checked_at: new Date().toISOString(),
          message: 'High latency detected',
        }
      ],
    }
    vi.mocked(adminApi.getDetailedHealth).mockResolvedValue(mockData)

    render(<AdminSystemPage />, { wrapper: createWrapper() })

    // Check overall status
    expect(await screen.findByText('healthy')).toBeInTheDocument()
    expect(screen.getByText('1.2.3')).toBeInTheDocument()
    expect(screen.getByText('5d 4h')).toBeInTheDocument()
    expect(screen.getByText('production')).toBeInTheDocument()

    // Check dependencies
    expect(screen.getByText('database')).toBeInTheDocument()
    expect(screen.getByText('OK')).toBeInTheDocument()
    expect(screen.getByText('25ms')).toBeInTheDocument()

    expect(screen.getByText('redis')).toBeInTheDocument()
    expect(screen.getByText('DEGRADED')).toBeInTheDocument()
    expect(screen.getByText('150ms')).toBeInTheDocument()
    expect(screen.getByText('High latency detected')).toBeInTheDocument()
  })

  it('handles refresh button click', async () => {
    const user = userEvent.setup()
    vi.mocked(adminApi.getDetailedHealth).mockResolvedValue({
      status: 'healthy',
      version: '1.2.3',
      uptime: '5d 4h',
      environment: 'production',
      dependencies: [],
    })

    render(<AdminSystemPage />, { wrapper: createWrapper() })

    const refreshButton = await screen.findByRole('button', { name: /yenile/i })
    await user.click(refreshButton)

    await waitFor(() => {
      expect(adminApi.getDetailedHealth).toHaveBeenCalledTimes(2)
    }, { timeout: 2000 })
  })

  it('shows loading state', () => {
    vi.mocked(adminApi.getDetailedHealth).mockReturnValue(new Promise(() => {}))
    
    render(<AdminSystemPage />, { wrapper: createWrapper() })
    
    expect(screen.getByText('Yükleniyor...')).toBeInTheDocument()
  })
})
