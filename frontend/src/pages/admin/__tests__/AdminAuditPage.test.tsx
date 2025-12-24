import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AdminAuditPage } from '../AdminAuditPage'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as adminApi from '@/api/admin'

// Mock the API
vi.mock('@/api/admin', () => ({
  listAuditLogs: vi.fn(),
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

describe('AdminAuditPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(adminApi.listAuditLogs).mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      per_page: 20
    })
  })

  it('renders page title and empty state', async () => {
    render(<AdminAuditPage />, { wrapper: createWrapper() })
    
    expect(await screen.findByText('Denetim Günlüğü')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText(/Henüz herhangi bir admin işlemi kaydedilmemiş/i)).toBeInTheDocument()
    })
  })

  it('renders audit log list', async () => {
    vi.mocked(adminApi.listAuditLogs).mockResolvedValue({
      data: [
        {
          id: 'log-1',
          admin_user_id: 'admin-1',
          action: 'update_user',
          target_type: 'user',
          details: { is_platform_admin: true },
          created_at: new Date().toISOString(),
          ip_address: '127.0.0.1',
          user_agent: 'Mozilla/5.0'
        }
      ],
      total: 1,
      page: 1,
      per_page: 20
    })

    render(<AdminAuditPage />, { wrapper: createWrapper() })

    // Find the action text (which is rendered in uppercase via CSS, but lowercase in DOM)
    await waitFor(() => {
      expect(screen.getByText(/update user/i)).toBeInTheDocument()
      expect(screen.getByText('admin-1')).toBeInTheDocument()
    })
  })

  it('opens detail dialog on eye button click', async () => {
    const user = userEvent.setup()
    vi.mocked(adminApi.listAuditLogs).mockResolvedValue({
      data: [
        {
          id: 'log-1',
          admin_user_id: 'admin-1',
          action: 'delete_chatbot',
          target_type: 'chatbot',
          details: { name: 'Old Bot' },
          created_at: new Date().toISOString(),
          ip_address: '127.0.0.1',
          user_agent: 'Chrome'
        }
      ],
      total: 1,
      page: 1,
      per_page: 20
    })

    render(<AdminAuditPage />, { wrapper: createWrapper() })

    const detailButton = await screen.findByTestId('audit-detail-button')
    await user.click(detailButton)

    await waitFor(() => {
      expect(screen.getByText('Audit Kaydı Detayları')).toBeInTheDocument()
      expect(screen.getByText(/delete chatbot/i)).toBeInTheDocument()
      expect(screen.getByText('Chrome')).toBeInTheDocument()
    })
  })
})
