import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from '../LoginPage'

const signInMock = vi.fn()

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({ signIn: signInMock }),
}))

describe('LoginPage', () => {
  beforeEach(() => {
    signInMock.mockReset()
  })

  it('renders form elements', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    expect(screen.getByRole('heading', { name: 'Giriş Yap' })).toBeInTheDocument()
    expect(screen.getByPlaceholderText('E-posta')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Şifre')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Giriş' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Kaydol' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Şifremi unuttum' })).toBeInTheDocument()
  })

  it('submits and calls signIn with credentials', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const emailInput = screen.getAllByPlaceholderText('E-posta')[0]
    const passwordInput = screen.getAllByPlaceholderText('Şifre')[0]
    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'secret123')
    const submitBtn = screen.getAllByRole('button', { name: 'Giriş' })[0]
    await user.click(submitBtn)

    expect(signInMock).toHaveBeenCalledWith('test@example.com', 'secret123')
  })

  it('shows error message on failed sign in', async () => {
    const user = userEvent.setup()
    signInMock.mockRejectedValueOnce(new Error('invalid'))

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const emailInput = screen.getAllByPlaceholderText('E-posta')[0]
    const passwordInput = screen.getAllByPlaceholderText('Şifre')[0]
    await user.type(emailInput, 'bad@example.com')
    await user.type(passwordInput, 'wrong')
    const submitBtn = screen.getAllByRole('button', { name: 'Giriş' })[0]
    await user.click(submitBtn)

    expect(await screen.findByText('Giriş başarısız. Bilgilerinizi kontrol edin.')).toBeInTheDocument()
  })
})
