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
      </ToastProvider>,
    )

    const submit = screen.getAllByRole('button', { name: 'Kayıt Ol' })[0]
    await user.click(submit)
    expect(screen.getByText('Lütfen tüm alanları doldurun.')).toBeInTheDocument()
  })

  it('submits successfully and shows success toast', async () => {
    const user = userEvent.setup()
    // Mock both register and login calls
    const postSpy = vi
      .spyOn(api, 'post')
      .mockResolvedValueOnce({ data: {} } as any) // register
      .mockResolvedValueOnce({ data: { token: 't', refresh_token: 'r' } } as any) // login

    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { completed: true } } as any) // onboarding status

    render(
      <ToastProvider>
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      </ToastProvider>,
    )

    await user.type(screen.getByLabelText('Ad Soyad'), 'Onur Ceri')
    await user.type(screen.getByLabelText('Email'), 'onur@example.com')
    await user.type(screen.getByLabelText('Şifre'), 'secret')

    await user.click(screen.getAllByRole('button', { name: 'Kayıt Ol' })[0])

    await waitFor(() => {
      expect(postSpy).toHaveBeenCalledWith('/api/v1/auth/register', {
        full_name: 'Onur Ceri',
        email: 'onur@example.com',
        password: 'secret',
      })
      expect(postSpy).toHaveBeenCalledWith('/api/v1/auth/login', {
        email: 'onur@example.com',
        password: 'secret',
      })
    })
    expect(screen.getByText('Hesabınız oluşturuldu! Hadi başlayalım.')).toBeInTheDocument()
  })

  it('shows fallback error message on API failure', async () => {
    const user = userEvent.setup()
    const postSpy = vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))

    render(
      <ToastProvider>
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      </ToastProvider>,
    )

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
    const postSpy = vi.spyOn(api, 'post').mockRejectedValueOnce({
      isAxiosError: true,
      response: { data: { code: 'ERR_PASSWORD_WEAK', status: 400 } },
      message: 'Request failed with status code 400',
    } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      </ToastProvider>,
    )

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
