import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AdminQueuesPage } from '../AdminQueuesPage'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as adminApi from '@/api/admin'
import { ToastProvider } from '@/components/ui/toast'

// Mock the API
vi.mock('@/api/admin', () => ({
  getQueues: vi.fn(),
  getStuckJobs: vi.fn(),
  retryJob: vi.fn(),
  deleteJob: vi.fn(),
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
    <ToastProvider>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </ToastProvider>
  )
}

describe('AdminQueuesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(adminApi.getQueues).mockResolvedValue([])
    vi.mocked(adminApi.getStuckJobs).mockResolvedValue([])
  })

  it('renders page title and description', async () => {
    render(<AdminQueuesPage />, { wrapper: createWrapper() })
    
    expect(screen.getByText('Kuyruk Yönetimi')).toBeInTheDocument()
    expect(screen.getByText(/Bileşenlerin işleme kuyruklarını/i)).toBeInTheDocument()
  })

  it('renders queue stats', async () => {
    vi.mocked(adminApi.getQueues).mockResolvedValue([
      {
        queue_name: 'source_processing',
        pending_count: 10,
        processing_count: 2,
        failed_count: 5,
        oldest_pending: new Date().toISOString(),
      }
    ])

    render(<AdminQueuesPage />, { wrapper: createWrapper() })

    expect(await screen.findByText(/source processing/i)).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('renders stuck jobs table', async () => {
    vi.mocked(adminApi.getStuckJobs).mockResolvedValue([
      {
        id: 'job-1',
        queue_name: 'source_processing',
        status: 'processing',
        started_at: new Date().toISOString(),
        stuck_duration: '45m',
        error_message: 'Database timeout',
      }
    ])

    render(<AdminQueuesPage />, { wrapper: createWrapper() })

    expect(await screen.findByText('job-1')).toBeInTheDocument()
    expect(screen.getByText('45m')).toBeInTheDocument()
    expect(screen.getByText('Database timeout')).toBeInTheDocument()
  })

  it('handles retry job action', async () => {
    const user = userEvent.setup()
    vi.mocked(adminApi.getStuckJobs).mockResolvedValue([
      {
        id: 'job-1',
        queue_name: 'test_queue',
        status: 'stuck',
        started_at: new Date().toISOString(),
        stuck_duration: '1h',
      }
    ])
    vi.mocked(adminApi.retryJob).mockResolvedValue({ status: 'ok' })

    render(<AdminQueuesPage />, { wrapper: createWrapper() })

    const retryButton = await screen.findByTitle('Tekrar Dene')
    await user.click(retryButton)

    expect(adminApi.retryJob).toHaveBeenCalledWith('job-1')
    expect(await screen.findByText(/Görev başarıyla bekleme kuyruğuna alındı/i)).toBeInTheDocument()
  })

  it('handles delete job action', async () => {
    const user = userEvent.setup()
    vi.mocked(adminApi.getStuckJobs).mockResolvedValue([
      {
        id: 'job-1',
        queue_name: 'test_queue',
        status: 'stuck',
        started_at: new Date().toISOString(),
        stuck_duration: '1h',
      }
    ])
    vi.mocked(adminApi.deleteJob).mockResolvedValue({ status: 'ok' })

    render(<AdminQueuesPage />, { wrapper: createWrapper() })

    const deleteButton = await screen.findByTitle('Sil')
    await user.click(deleteButton)

    expect(adminApi.deleteJob).toHaveBeenCalledWith('job-1')
    expect(await screen.findByText(/Görev kuyruktan başarıyla kaldırıldı/i)).toBeInTheDocument()
  })
})
