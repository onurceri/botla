import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage save error branches', () => {
  it('new chatbot: validate ok but POST fails shows error toast', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/new"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const nameInput = await screen.findByPlaceholderText('Örn: Müşteri Temsilcisi')
    await user.type(nameInput, 'Yeni Bot')
    const createBtn = await screen.findByRole('button', { name: 'Oluştur' })
    await user.click(createBtn)
    const errs1 = await screen.findAllByText('Bir hata oluştu. Lütfen tekrar deneyin.')
    expect(errs1.length).toBeGreaterThan(0)
  })

  it('existing chatbot: PUT fails shows error toast', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/err')) return Promise.resolve({ data: { id: 'err', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/err/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    vi.spyOn(api, 'put').mockRejectedValueOnce(new Error('fail'))
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/err"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const saveBtn = await screen.findByRole('button', { name: 'Değişiklikleri Kaydet' })
    await user.click(saveBtn)
    const errs2 = await screen.findAllByText('Bir hata oluştu. Lütfen tekrar deneyin.')
    expect(errs2.length).toBeGreaterThan(0)
  })
})
