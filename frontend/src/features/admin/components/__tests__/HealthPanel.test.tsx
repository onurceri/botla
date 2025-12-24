import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { HealthPanel } from '../HealthPanel'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { getDetailedHealth } from '@/api/admin'

// Mock the API
vi.mock('@/api/admin', () => ({
  getDetailedHealth: vi.fn(),
}))

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 0,
        gcTime: 0,
      },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('HealthPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders loading state initially', () => {
    vi.mocked(getDetailedHealth).mockReturnValue(new Promise(() => {}))
    
    render(<HealthPanel />, { wrapper: createWrapper() })
    
    expect(screen.getByText(/System Health/i)).toBeInTheDocument()
  })

  it('renders healthy status and dependencies', async () => {
    vi.mocked(getDetailedHealth).mockResolvedValue({
      status: 'healthy',
      version: '1.0.0',
      uptime: '10d',
      environment: 'production',
      dependencies: [
        { name: 'database', status: 'ok', latency_ms: 5, checked_at: new Date().toISOString() },
        { name: 'redis', status: 'ok', latency_ms: 2, checked_at: new Date().toISOString() },
      ],
    })

    render(<HealthPanel />, { wrapper: createWrapper() })

    expect(await screen.findAllByText(/HEALTHY/i)).not.toHaveLength(0)
    expect(screen.getByText('database')).toBeInTheDocument()
    expect(screen.getByText('redis')).toBeInTheDocument()
    expect(screen.getByText(/5ms/)).toBeInTheDocument()
    expect(screen.getByText(/2ms/)).toBeInTheDocument()
  })

  it('renders degraded status correctly', async () => {
    vi.mocked(getDetailedHealth).mockResolvedValue({
      status: 'degraded',
      version: '1.0.0',
      uptime: '10d',
      environment: 'production',
      dependencies: [
        { name: 'database', status: 'ok', latency_ms: 5, checked_at: new Date().toISOString() },
        { name: 'openai', status: 'degraded', latency_ms: 2000, message: 'High latency', checked_at: new Date().toISOString() },
      ],
    })

    render(<HealthPanel />, { wrapper: createWrapper() })

    expect(await screen.findAllByText(/DEGRADED/i)).not.toHaveLength(0)
    expect(screen.getByText('openai')).toBeInTheDocument()
    expect(screen.getByText(/High latency/)).toBeInTheDocument()
  })
})
