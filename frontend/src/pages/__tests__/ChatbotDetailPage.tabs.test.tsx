import { describe, it, expect, vi } from 'vitest'
import { render, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { ToastProvider } from '@/components/ui/toast'
import { api } from '@/api/client'

describe('ChatbotDetailPage tab query param', () => {
  it('opens sources tab on trigger click', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc')) return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const utils = render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/abc"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const view = within(utils.container)
    const user = userEvent.setup()
    const trigger = await view.findByRole('tab', { name: /Veri Kaynakları/ })
    await user.click(trigger)
    expect(await view.findByText('Bilgi Bankası')).toBeInTheDocument()
    expect(view.getByText('Henüz kaynak eklenmemiş')).toBeInTheDocument()
  })

  it('opens playground tab on trigger click', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc')) return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const utils2 = render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/abc"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const view2 = within(utils2.container)
    const user = userEvent.setup()
    const trigger = await view2.findByRole('tab', { name: /Playground/ })
    await user.click(trigger)
    const openBtn = await view2.findByLabelText('Sohbeti aç')
    await user.click(openBtn)
    expect(await view2.findByText('Powered by Botla')).toBeInTheDocument()
  })
})
