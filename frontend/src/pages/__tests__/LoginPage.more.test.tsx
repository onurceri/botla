import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import LoginPage from '../LoginPage'
import { api } from '@/api/client'

describe('LoginPage extra tests', () => {
  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
    window.localStorage.clear()
  })

  it('shows error toast when fields are empty', async () => {
    const user = userEvent.setup()
    render(
      <ToastProvider>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </ToastProvider>,
    )
    const submit = screen.getByRole('button', { name: 'Giriş Yap' })
    await user.click(submit)
    expect(await screen.findByText('Lütfen tüm alanları doldurun.')).toBeInTheDocument()
  })

  it('logs in successfully and re-enables button', async () => {
    const user = userEvent.setup()
    const postSpy = vi
      .spyOn(api, 'post')
      .mockResolvedValueOnce({ data: { token: 't', refresh_token: 'r' } } as any)
    render(
      <ToastProvider>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </ToastProvider>,
    )
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
    vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))
    render(
      <ToastProvider>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </ToastProvider>,
    )
    await user.type(screen.getByLabelText('Email'), 'e@e.com')
    await user.type(screen.getByLabelText('Şifre'), 'p')
    const submitBtns = screen.getAllByRole('button', { name: 'Giriş Yap' })
    await user.click(submitBtns[0])
    expect(
      await screen.findByText('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.'),
    ).toBeInTheDocument()
  })
})
