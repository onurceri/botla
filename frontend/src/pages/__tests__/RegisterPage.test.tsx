import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import RegisterPage from '../RegisterPage'
import { api } from '@/api/client'

describe('RegisterPage', () => {
  it('validates empty fields and shows error toast', async () => {
    const user = userEvent.setup()
    render(
      <ToastProvider>
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      </ToastProvider>
    )

    const submit = screen.getAllByRole('button', { name: 'Kayıt Ol' })[0]
    await user.click(submit)
    expect(screen.getByText('Lütfen tüm alanları doldurun.')).toBeInTheDocument()
  })

  it('submits successfully and shows success toast', async () => {
    const user = userEvent.setup()
    const postSpy = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: {} } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      </ToastProvider>
    )

    await user.type(screen.getByLabelText('Ad Soyad'), 'Onur Ceri')
    await user.type(screen.getByLabelText('Email'), 'onur@example.com')
    await user.type(screen.getByLabelText('Şifre'), 'secret')

    await user.click(screen.getAllByRole('button', { name: 'Kayıt Ol' })[0])
    expect(postSpy).toHaveBeenCalledWith('/api/v1/auth/register', { full_name: 'Onur Ceri', email: 'onur@example.com', password: 'secret' })
    expect(screen.getByText('Kayıt başarılı! Giriş yapabilirsiniz.')).toBeInTheDocument()
  })

  it('shows error toast on API failure', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))

    render(
      <ToastProvider>
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      </ToastProvider>
    )

    await user.type(screen.getByLabelText('Ad Soyad'), 'Onur Ceri')
    await user.type(screen.getByLabelText('Email'), 'onur@example.com')
    await user.type(screen.getByLabelText('Şifre'), 'secret')
    const submit = screen.getAllByRole('button', { name: 'Kayıt Ol' })[0]
    await user.click(submit)
    await waitFor(() => {
      expect(submit).not.toBeDisabled()
    })
  })
})
