import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage save success', () => {
  it('enables button after PUT success and shows toast', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/123')) return Promise.resolve({ data: { id: '123', name: 'Var Olan Bot' } } as any)
      if (url.includes('/api/v1/chatbots/123/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    vi.spyOn(api, 'put').mockResolvedValueOnce({ data: {} } as any)
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/123"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const saveBtn = await screen.findByRole('button', { name: 'Değişiklikleri Kaydet' })
    await user.click(saveBtn)
    expect(await screen.findByText('Değişiklikler kaydedildi.')).toBeInTheDocument()
    await waitFor(() => {
      expect(saveBtn).not.toBeDisabled()
    })
  })
})

