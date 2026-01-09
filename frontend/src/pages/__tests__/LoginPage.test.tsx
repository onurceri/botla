import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
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

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
    // Mock /api/v1/me to return null (not authenticated)
    vi.spyOn(api, 'get').mockRejectedValue({ response: { status: 401 } })
  })

  it('renders form elements', async () => {
    renderWithProviders(<LoginPage />)

    // Wait for loading state to complete
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Hoş Geldiniz' })).toBeInTheDocument()
    })
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Şifre')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Giriş Yap' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Kayıt Olun' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Şifremi unuttum?' })).toBeInTheDocument()
  })

  it('submits and calls login API (cookies set by backend)', async () => {
    const user = userEvent.setup()
    const postSpy = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { token: 't', refresh_token: 'r' } } as any)
    
    renderWithProviders(<LoginPage />)

    // Wait for initial render
    await waitFor(() => {
      expect(screen.getByLabelText('Email')).toBeInTheDocument()
    })

    const emailInput = screen.getByLabelText('Email')
    const passwordInput = screen.getByLabelText('Şifre')
    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'secret123')
    const submitBtn = screen.getAllByRole('button', { name: 'Giriş Yap' })[0]
    await user.click(submitBtn)
    
    // Verify API was called - cookies are set by backend, not localStorage
    expect(postSpy).toHaveBeenCalledWith('/api/v1/auth/login', { email: 'test@example.com', password: 'secret123' })
  })

  it('shows error handling on failed sign in', async () => {
    const user = userEvent.setup()
    const postSpy = vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('invalid'))

    renderWithProviders(<LoginPage />)

    // Wait for initial render
    await waitFor(() => {
      expect(screen.getByLabelText('Email')).toBeInTheDocument()
    })

    const emailInput = screen.getByLabelText('Email')
    const passwordInput = screen.getByLabelText('Şifre')
    await user.type(emailInput, 'bad@example.com')
    await user.type(passwordInput, 'wrong')
    const submitBtn = screen.getAllByRole('button', { name: 'Giriş Yap' })[0]
    await user.click(submitBtn)
    expect(postSpy).toHaveBeenCalled()
    
    // Verify error toast is shown
    expect(await screen.findByText('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.')).toBeInTheDocument()
  })
})
