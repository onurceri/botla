import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/contexts/AuthContext'
import LoginPage from '../LoginPage'
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

describe('LoginPage extra tests', () => {
  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
    window.localStorage.clear()
  })

  it('shows error toast when fields are empty', async () => {
    // Mock /api/v1/me to return null (not authenticated)
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
    
    const user = userEvent.setup()
    renderWithProviders(<LoginPage />)
    const submit = screen.getByRole('button', { name: 'Giriş Yap' })
    await user.click(submit)
    expect(await screen.findByText('Lütfen tüm alanları doldurun.')).toBeInTheDocument()
  })

  it('logs in successfully and re-enables button', async () => {
    const user = userEvent.setup()
    // Mock /api/v1/me for initial load (not authenticated)
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
    const postSpy = vi
      .spyOn(api, 'post')
      .mockResolvedValueOnce({ data: { token: 't', refresh_token: 'r' } } as any)
    
    renderWithProviders(<LoginPage />)
    await user.type(screen.getByLabelText('Email'), 'e@e.com')
    await user.type(screen.getByLabelText('Şifre'), 'p')
    const submitBtns = screen.getAllByRole('button', { name: 'Giriş Yap' })
    const submit = submitBtns[0]
    await user.click(submit)
    expect(postSpy).toHaveBeenCalledWith('/api/v1/auth/login', { email: 'e@e.com', password: 'p' })
    await waitFor(() => {
      expect(submit).not.toBeDisabled()
    })
  })

  it('shows error toast when login fails', async () => {
    const user = userEvent.setup()
    // Mock /api/v1/me for initial load (not authenticated)
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
    vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))
    
    renderWithProviders(<LoginPage />)
    await user.type(screen.getByLabelText('Email'), 'e@e.com')
    await user.type(screen.getByLabelText('Şifre'), 'p')
    const submitBtns = screen.getAllByRole('button', { name: 'Giriş Yap' })
    await user.click(submitBtns[0])
    expect(
      await screen.findByText('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.'),
    ).toBeInTheDocument()
  })
})
