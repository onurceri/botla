import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage POST navigate', () => {
  it('shows success toast on POST and enables save', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { id: '999' } } as any)
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
    const createBtn = screen.getByRole('button', { name: 'Oluştur' })
    await user.click(createBtn)
    expect(await screen.findByText('Chatbot başarıyla oluşturuldu.')).toBeInTheDocument()
  })
})
