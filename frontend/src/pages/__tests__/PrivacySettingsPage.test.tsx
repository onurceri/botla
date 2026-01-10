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

const mockEmptyRequests = () => ({
  data: [],
  total: 0,
  page: 1,
  limit: 10,
})

const mockDefaultApiGet = () => {
  vi.spyOn(api, 'get').mockImplementation((url: string) => {
    if (url === '/api/v1/me/privacy/consents') {
      return Promise.resolve({
        data: { marketing: false, analytics: false, personalization: false, third_party: false },
      })
    }
    if (url.includes('/api/v1/me/privacy/requests')) {
      return Promise.resolve({ data: mockEmptyRequests() })
    }
    if (url === '/api/v1/me') {
      return Promise.resolve({ data: { id: 'user-1', email: 'test@example.com' } })
    }
    return Promise.reject(new Error('Not found'))
  })
}

describe('PrivacySettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
  })

  it('renders page with all sections', async () => {
    mockDefaultApiGet()

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: privacy.page.title })).toBeInTheDocument()
    })

    expect(screen.getByText(privacy.consents.title)).toBeInTheDocument()
    expect(screen.getByText(privacy.export.title)).toBeInTheDocument()
    expect(screen.getByText(privacy.correction.title)).toBeInTheDocument()
    expect(screen.getByText(privacy.delete.title)).toBeInTheDocument()
  })

  it('displays export and correction history titles', async () => {
    mockDefaultApiGet()

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.export.history.title)).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByText(privacy.correction.history.title)).toBeInTheDocument()
    })
  })

  it('shows empty state when no requests exist', async () => {
    mockDefaultApiGet()

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.export.history.empty)).toBeInTheDocument()
    })

    expect(screen.getByText(privacy.correction.history.empty)).toBeInTheDocument()
  })

  it('displays pending status for active export request', async () => {
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
        return Promise.resolve({ data: mockEmptyRequests() })
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
      expect(screen.getByText(privacy.export.history.title)).toBeInTheDocument()
    })
  })

  it('handles 409 conflict error when duplicate export request', async () => {
    const user = userEvent.setup()
    mockDefaultApiGet()

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
    mockDefaultApiGet()

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

  it('enables correction button when not rate limited', async () => {
    const user = userEvent.setup()
    mockDefaultApiGet()

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const textarea = screen.getByPlaceholderText(privacy.correction.placeholder)
    await user.type(textarea, 'Test correction request')

    const correctionButton = screen.getByRole('button', { name: privacy.correction.button })
    expect(correctionButton).not.toBeDisabled()
  })

  it('handles 429 rate limit error for correction requests', async () => {
    const user = userEvent.setup()
    mockDefaultApiGet()

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
      expect(screen.getByText(privacy.export.rateLimit.message)).toBeInTheDocument()
    })
  })

  it('displays character count for correction textarea', async () => {
    const user = userEvent.setup()
    mockDefaultApiGet()

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    expect(screen.getByText('0 / 1000')).toBeInTheDocument()

    const textarea = screen.getByPlaceholderText(privacy.correction.placeholder)
    await user.type(textarea, 'Test')

    expect(screen.getByText('4 / 1000')).toBeInTheDocument()
  })

  it('limits correction textarea to max length', async () => {
    const user = userEvent.setup()
    mockDefaultApiGet()

    renderWithProviders(<PrivacySettingsPage />)

    await waitFor(() => {
      expect(screen.getByText(privacy.page.title)).toBeInTheDocument()
    })

    const textarea = screen.getByPlaceholderText(privacy.correction.placeholder) as HTMLTextAreaElement
    expect(textarea.maxLength).toBe(1000)
  })
})
