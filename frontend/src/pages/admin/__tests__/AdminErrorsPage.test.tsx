import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AdminErrorsPage } from '../AdminErrorsPage'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as adminApi from '@/api/admin'

// Mock Radix Select
vi.mock('@radix-ui/react-select', () => ({
  Root: ({ children, onValueChange, value }: any) => (
    <div data-testid="mock-select" onClick={() => onValueChange?.('error')}>{children}</div>
  ),
  Trigger: ({ children }: any) => <button data-testid="severity-select">{children}</button>,
  Value: ({ placeholder }: any) => <span>{placeholder}</span>,
  Portal: ({ children }: any) => <>{children}</>,
  Content: ({ children }: any) => <div>{children}</div>,
  Viewport: ({ children }: any) => <div>{children}</div>,
  Item: ({ children, value }: any) => <div role="option" data-value={value}>{children}</div>,
  ItemText: ({ children }: any) => <>{children}</>,
  ItemIndicator: () => null,
  Group: ({ children }: any) => <div>{children}</div>,
  Label: ({ children }: any) => <div>{children}</div>,
  Separator: () => <hr />,
  ScrollUpButton: () => null,
  ScrollDownButton: () => null,
  Icon: () => null,
}))

// Mock the API
vi.mock('@/api/admin', () => ({
  getErrors: vi.fn(),
  getErrorStats: vi.fn(),
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

describe('AdminErrorsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(adminApi.getErrors).mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      per_page: 20
    })
    vi.mocked(adminApi.getErrorStats).mockResolvedValue({
      critical: 0,
      error: 0,
      warning: 0,
      info: 1
    })
  })

  it('renders page title and description', async () => {
    render(<AdminErrorsPage />, { wrapper: createWrapper() })
    expect(await screen.findByText('Hata Kayıtları')).toBeInTheDocument()
    expect(screen.getByText(/Sistem genelinde oluşan hata ve uyarıları izle/i)).toBeInTheDocument()
  })

  it('renders error list and stats', async () => {
    vi.mocked(adminApi.getErrorStats).mockResolvedValue({
      critical: 5,
      error: 10,
      warning: 2,
      info: 1
    })

    vi.mocked(adminApi.getErrors).mockResolvedValue({
      data: [
        {
          id: 'err-1',
          severity: 'error',
          error_type: 'api_error',
          message: 'Database connection failed',
          created_at: new Date().toISOString(),
          request_method: 'GET',
          request_path: '/api/v1/test'
        }
      ],
      total: 1,
      page: 1,
      per_page: 20
    })

    render(<AdminErrorsPage />, { wrapper: createWrapper() })

    // High level stats
    expect(await screen.findByText('15')).toBeInTheDocument() // 5 + 10
    expect(screen.getByText('2')).toBeInTheDocument()
    
    // Error entry
    expect(screen.getByText('Database connection failed')).toBeInTheDocument()
  })

  it('handles severity filter change', async () => {
    const user = userEvent.setup()
    render(<AdminErrorsPage />, { wrapper: createWrapper() })

    // With our mock, clicking the trigger triggers onValueChange('error')
    const selectTrigger = await screen.findByTestId('severity-select')
    await user.click(selectTrigger)

    expect(adminApi.getErrors).toHaveBeenCalledWith('error', 0, 20)
  })

  it('opens detail dialog on row click', async () => {
    const user = userEvent.setup()
    vi.mocked(adminApi.getErrors).mockResolvedValue({
      data: [
        {
          id: 'err-1',
          severity: 'critical',
          error_type: 'panic',
          message: 'unexpected nil pointer',
          stack_trace: 'main.go:42...',
          created_at: new Date().toISOString(),
        }
      ],
      total: 1,
      page: 1,
      per_page: 20
    })

    render(<AdminErrorsPage />, { wrapper: createWrapper() })

    const row = await screen.findByText('unexpected nil pointer')
    await user.click(row)

    expect(await screen.findByText('Stack Trace')).toBeInTheDocument()
    expect(screen.getByText('main.go:42...')).toBeInTheDocument()
  })
})
