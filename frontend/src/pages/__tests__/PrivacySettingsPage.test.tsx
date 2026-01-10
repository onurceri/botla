import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/contexts/AuthContext'
import PrivacySettingsPage from '../PrivacySettingsPage'
import { api } from '@/api/client'
import { privacy } from '@/i18n/privacy'

const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })

const renderWithProviders = (ui: React.ReactElement) => {
  const queryClient = createTestQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <AuthProvider>
          <MemoryRouter>{ui}</MemoryRouter>
        </AuthProvider>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

describe('PrivacySettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
  })

  it('renders page and fetches consents', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: true, analytics: false, personalization: true, third_party: false },
        })
      }
      if (url.includes('/api/v1/me/privacy/requests')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: privacy.page.title })).toBeInTheDocument()
    })

    expect(screen.getByText(privacy.consents.title)).toBeInTheDocument()
    expect(screen.getByText(privacy.export.title)).toBeInTheDocument()
    expect(screen.getByText(privacy.correction.title)).toBeInTheDocument()
    expect(screen.getByText(privacy.delete.title)).toBeInTheDocument()
  })

  it('fetches existing privacy requests and displays export status', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'export',
                status: 'pending',
                created_at: '2024-01-01T00:00:00Z',
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me/privacy/requests/req-1') {
        return Promise.resolve({
          data: {
            id: 'req-1',
            request_type: 'export',
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
          },
        })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByText(new RegExp(privacy.status.pending))).toBeInTheDocument()
    })
  })

  it('disables download button when pending export request exists', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'export',
                status: 'pending',
                created_at: '2024-01-01T00:00:00Z',
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me/privacy/requests/req-1') {
        return Promise.resolve({
          data: {
            id: 'req-1',
            request_type: 'export',
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
          },
        })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const downloadButton = screen.getByRole('button', { name: privacy.export.button })
    await waitFor(() => {
      expect(downloadButton).toBeDisabled()
    })
  })

  it('handles 409 conflict error when duplicate export request', async () => {
    const user = userEvent.setup()

    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('/api/v1/me/privacy/requests')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    vi.spyOn(api, 'post').mockRejectedValueOnce({
      response: { status: 409 },
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const downloadButton = screen.getByRole('button', { name: privacy.export.button })
    await user.click(downloadButton)

    await waitFor(() => {
      expect(screen.getByText(privacy.toast.exportActiveExists)).toBeInTheDocument()
    })
  })

  it('handles 429 rate limit error for export requests', async () => {
    const user = userEvent.setup()

    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('/api/v1/me/privacy/requests')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    vi.spyOn(api, 'post').mockRejectedValueOnce({
      response: { status: 429 },
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const downloadButton = screen.getByRole('button', { name: privacy.export.button })
    await user.click(downloadButton)

    await waitFor(() => {
      expect(screen.getByText(privacy.toast.exportRateLimit)).toBeInTheDocument()
    })
  })

  it('disables export button when rate limited by recent completed export', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        const completedTime = new Date()
        completedTime.setHours(completedTime.getHours() - 12)
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'export',
                status: 'completed',
                completed_at: completedTime.toISOString(),
                created_at: completedTime.toISOString(),
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const downloadButton = screen.getByRole('button', { name: privacy.export.button })
    await waitFor(() => {
      expect(downloadButton).toBeDisabled()
    })

    await waitFor(() => {
      expect(screen.getByText(/Sonraki dışa aktarma:/i)).toBeInTheDocument()
    })
  })

  it('displays export history table with pagination', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'export',
                status: 'completed',
                created_at: '2024-01-01T00:00:00Z',
                completed_at: '2024-01-01T01:00:00Z',
              },
              {
                id: 'req-2',
                request_type: 'export',
                status: 'completed',
                created_at: '2024-01-02T00:00:00Z',
                completed_at: '2024-01-02T01:00:00Z',
              },
            ],
            total: 2,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.export.history.title)).toBeInTheDocument()
    })

    expect(screen.getByText(privacy.export.history.status)).toBeInTheDocument()
    expect(screen.getByText(privacy.export.history.date)).toBeInTheDocument()
    expect(screen.getByText(privacy.export.history.actions)).toBeInTheDocument()
  })

  it('allows pagination through export history', async () => {
    const user = userEvent.setup()

    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export&page=1')) {
        return Promise.resolve({
          data: {
            data: Array.from({ length: 10 }, (_, i) => ({
              id: `req-${i}`,
              request_type: 'export',
              status: 'completed',
              created_at: '2024-01-01T00:00:00Z',
              completed_at: '2024-01-01T01:00:00Z',
            })),
            total: 15,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url.includes('type=export&page=2')) {
        return Promise.resolve({
          data: {
            data: Array.from({ length: 5 }, (_, i) => ({
              id: `req-${10 + i}`,
              request_type: 'export',
              status: 'completed',
              created_at: '2024-01-01T00:00:00Z',
              completed_at: '2024-01-01T01:00:00Z',
            })),
            total: 15,
            page: 2,
            limit: 10,
          },
        })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.export.history.title)).toBeInTheDocument()
    })

    const nextButton = screen.getByRole('button', { name: /Sonraki/i })
    await waitFor(() => {
      expect(nextButton).not.toBeDisabled()
    })

    await user.click(nextButton)

    await waitFor(() => {
      expect(screen.getByText('Sayfa 2 / 2')).toBeInTheDocument()
    })
  })

  it('deletes privacy request successfully', async () => {
    const user = userEvent.setup()

    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'export',
                status: 'completed',
                created_at: '2024-01-01T00:00:00Z',
                completed_at: '2024-01-01T01:00:00Z',
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    vi.spyOn(api, 'delete').mockResolvedValueOnce({})

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.export.history.title)).toBeInTheDocument()
    })

    const deleteButton = screen.getByRole('button')
    await user.click(deleteButton)

    const confirmButton = screen.getByRole('button', { name: privacy.export.history.delete })
    await user.click(confirmButton)

    await waitFor(() => {
      expect(screen.getByText(privacy.export.history.deleteSuccess)).toBeInTheDocument()
    })
  })

  it('displays correction history table separate from export history', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'export',
                status: 'completed',
                created_at: '2024-01-01T00:00:00Z',
                completed_at: '2024-01-01T01:00:00Z',
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-2',
                request_type: 'correction',
                status: 'completed',
                reason: 'E-posta adresim yanlış girilmiş',
                created_at: '2024-01-02T00:00:00Z',
                completed_at: '2024-01-02T01:00:00Z',
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.export.history.title)).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByText(privacy.correction.history.title)).toBeInTheDocument()
    })

    expect(screen.getByText(privacy.export.history.status)).toBeInTheDocument()
    expect(screen.getByText(privacy.correction.history.status)).toBeInTheDocument()
  })

  it('opens detail modal when clicking view button on correction request', async () => {
    const user = userEvent.setup()
    const testReason = 'E-posta adresim yanlış girilmiş'

    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url.includes('type=correction')) {
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'correction',
                status: 'pending',
                reason: testReason,
                created_at: '2024-01-02T00:00:00Z',
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.correction.history.title)).toBeInTheDocument()
    })

    await waitFor(() => {
      const buttons = screen.getAllByRole('button')
      const eyeButton = buttons.find(btn => btn.querySelector('.lucide-eye'))
      expect(eyeButton).toBeTruthy()
    })

    const buttons = screen.getAllByRole('button')
    const eyeButton = buttons.find(btn => btn.querySelector('.lucide-eye'))!
    await user.click(eyeButton)

    await waitFor(() => {
      expect(screen.getByText(privacy.correction.history.detailsTitle)).toBeInTheDocument()
    })
    expect(screen.getByText(testReason)).toBeInTheDocument()
  })

  it('disables correction button when rate limited by recent completed correction', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('type=export')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url.includes('type=correction')) {
        const completedTime = new Date()
        completedTime.setHours(completedTime.getHours() - 12)
        return Promise.resolve({
          data: {
            data: [
              {
                id: 'req-1',
                request_type: 'correction',
                status: 'completed',
                completed_at: completedTime.toISOString(),
                created_at: completedTime.toISOString(),
              },
            ],
            total: 1,
            page: 1,
            limit: 10,
          },
        })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const correctionButton = screen.getByRole('button', { name: privacy.correction.button })
    await waitFor(() => {
      expect(correctionButton).toBeDisabled()
    })

    await waitFor(() => {
      expect(screen.getByText(/Sonraki düzeltme:/i)).toBeInTheDocument()
    })
  })

  it('handles 429 rate limit error for correction requests', async () => {
    const user = userEvent.setup()

    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url === '/api/v1/me/privacy/consents') {
        return Promise.resolve({
          data: { marketing: false, analytics: false, personalization: false, third_party: false },
        })
      }
      if (url.includes('/api/v1/me/privacy/requests')) {
        return Promise.resolve({ data: { data: [], total: 0, page: 1, limit: 10 } })
      }
      if (url === '/api/v1/me') {
        return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
      }
      return Promise.reject(new Error('Not found'))
    })

    vi.spyOn(api, 'post').mockRejectedValueOnce({
      response: { status: 429 },
    })

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const textarea = screen.getByPlaceholderText(privacy.correction.placeholder)
    await user.type(textarea, 'Please correct my data')

    const correctionButton = screen.getByRole('button', { name: privacy.correction.button })
    await user.click(correctionButton)

    await waitFor(() => {
      expect(screen.getByText(privacy.correction.rateLimit.message)).toBeInTheDocument()
    })
  })
})
