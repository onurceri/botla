import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage chat error', () => {
  it('shows error message when chat request fails', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: '123', name: 'Bot' } } as any)
    vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))

    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/123"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )

    // Open widget
    await user.click((await screen.findAllByText('Playground'))[0])
    const badge = await screen.findByText('1')
    const openBtn = badge.closest('button') as HTMLButtonElement
    await user.click(openBtn)

    const input = screen.getByPlaceholderText('Mesaj yazın...')
    await user.type(input, 'merhaba')
    await user.keyboard('{Enter}')
    expect(await screen.findByText('Bir hata oluştu.')).toBeInTheDocument()
  })
})

