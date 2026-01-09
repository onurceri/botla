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
      if (url === '/api/v1/me/privacy/requests') {
        return Promise.resolve({ data: { data: [], total: 0 } })
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
      if (url === '/api/v1/me/privacy/requests') {
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
          },
        })
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
      if (url === '/api/v1/me/privacy/requests') {
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
          },
        })
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
      if (url === '/api/v1/me/privacy/requests') {
        return Promise.resolve({ data: { data: [], total: 0 } })
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
})
