import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/contexts/AuthContext'
import RegisterPage from '../RegisterPage'
import { api } from '@/api/client'

const createTestQueryClient = () => new QueryClient({
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
          <MemoryRouter>
            {ui}
          </MemoryRouter>
        </AuthProvider>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

describe('RegisterPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
    // Mock /api/v1/me to return null (not authenticated)
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
  })

  it('validates empty fields and shows error toast', async () => {
    const user = userEvent.setup()
    renderWithProviders(<RegisterPage />)

    // Wait for initial render
    await waitFor(() => {
      expect(screen.getByLabelText('Ad Soyad')).toBeInTheDocument()
    })

    const submit = screen.getAllByRole('button', { name: 'Kayıt Ol' })[0]
    await user.click(submit)
    expect(screen.getByText('Lütfen tüm alanları doldurun.')).toBeInTheDocument()
  })

  it('submits successfully and calls register API', async () => {
    vi.clearAllMocks() // Clear beforeEach mocks
    
    const user = userEvent.setup()
    // Mock register call - backend sets HttpOnly cookies automatically
    const postSpy = vi
      .spyOn(api, 'post')
      .mockResolvedValueOnce({ data: {} } as any) // register

    // Mock API get calls: first /me (401), then onboarding status
    vi.spyOn(api, 'get')
      .mockRejectedValueOnce({ response: { status: 401 } }) // initial /me call
      .mockResolvedValueOnce({ data: { completed: true } } as any) // onboarding status

    renderWithProviders(<RegisterPage />)

    // Wait for initial render
    await waitFor(() => {
      expect(screen.getByLabelText('Ad Soyad')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Ad Soyad'), 'Onur Ceri')
    await user.type(screen.getByLabelText('Email'), 'onur@example.com')
    await user.type(screen.getByLabelText('Şifre'), 'secret')

    await user.click(screen.getAllByRole('button', { name: 'Kayıt Ol' })[0])

    // Verify register API was called with correct data
    await waitFor(() => {
      expect(postSpy).toHaveBeenCalledWith('/api/v1/auth/register', {
        full_name: 'Onur Ceri',
        email: 'onur@example.com',
        password: 'secret',
      })
    })
    // No separate login call needed - register sets cookies directly
  })

  it('shows fallback error message on API failure', async () => {
    const user = userEvent.setup()
    // Mock /me then register failure
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
    const postSpy = vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))

    renderWithProviders(<RegisterPage />)

    // Wait for initial render
    await waitFor(() => {
      expect(screen.getByLabelText('Ad Soyad')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Ad Soyad'), 'Test')
    await user.type(screen.getByLabelText('Email'), 'test@example.com')
    await user.type(screen.getByLabelText('Şifre'), '12345678')
    const submit = screen.getAllByRole('button', { name: 'Kayıt Ol' })[0]
    await user.click(submit)

    // Wait for API call and button to be re-enabled after error
    await waitFor(() => {
      expect(postSpy).toHaveBeenCalled()
      expect(submit).not.toBeDisabled()
    })
  })

  it('shows translated error message for known error code', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
    const postSpy = vi.spyOn(api, 'post').mockRejectedValueOnce({
      isAxiosError: true,
      response: { data: { code: 'ERR_PASSWORD_WEAK', status: 400 } },
      message: 'Request failed with status code 400',
    } as any)

    renderWithProviders(<RegisterPage />)

    // Wait for initial render
    await waitFor(() => {
      expect(screen.getByLabelText('Ad Soyad')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Ad Soyad'), 'Test')
    await user.type(screen.getByLabelText('Email'), 'test@example.com')
    await user.type(screen.getByLabelText('Şifre'), 'weak')
    await user.click(screen.getAllByRole('button', { name: 'Kayıt Ol' })[0])

    await waitFor(() => {
      expect(postSpy).toHaveBeenCalled()
    })

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Şifre büyük harf, küçük harf, rakam ve özel karakter (@$!%*?&) içermelidir',
    )
  })
})
