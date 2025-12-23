import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import LoginPage from '../LoginPage'
import { api } from '@/api/client'

describe('LoginPage', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })
  })

  it('renders form elements', () => {
    render(
      <ToastProvider>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </ToastProvider>,
    )

    expect(screen.getByRole('heading', { name: 'Hoş Geldiniz' })).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Şifre')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Giriş Yap' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Kayıt Olun' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Şifremi unuttum?' })).toBeInTheDocument()
  })

  it('submits and calls signIn with credentials', async () => {
    const user = userEvent.setup()
    render(
      <ToastProvider>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </ToastProvider>,
    )

    const emailInput = screen.getByLabelText('Email')
    const passwordInput = screen.getByLabelText('Şifre')
    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'secret123')
    const submitBtn = screen.getAllByRole('button', { name: 'Giriş Yap' })[0]
    vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { token: 't', refresh_token: 'r' } } as any)
    const setSpy = vi.spyOn(window.localStorage, 'setItem')
    await user.click(submitBtn)
    expect(setSpy).toHaveBeenCalledWith('botla_token', 't')
    expect(setSpy).toHaveBeenCalledWith('botla_refresh_token', 'r')
  })

  it('shows error handling on failed sign in', async () => {
    const user = userEvent.setup()
    const postSpy = vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('invalid'))

    render(
      <ToastProvider>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </ToastProvider>,
    )

    const emailInput = screen.getByLabelText('Email')
    const passwordInput = screen.getByLabelText('Şifre')
    await user.type(emailInput, 'bad@example.com')
    await user.type(passwordInput, 'wrong')
    const submitBtn = screen.getAllByRole('button', { name: 'Giriş Yap' })[0]
    await user.click(submitBtn)
    expect(postSpy).toHaveBeenCalled()
    const setSpy = vi.spyOn(window.localStorage, 'setItem')
    expect(setSpy).not.toHaveBeenCalledWith('botla_token', expect.anything())
  })
})
