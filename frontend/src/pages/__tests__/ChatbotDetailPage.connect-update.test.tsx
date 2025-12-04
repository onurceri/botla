import { describe, it, expect, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage connect tab domain/secret updates', () => {
  it('updates allowedDomains and embedSecret inputs', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc')) return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const utils = render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/abc?tab=connect"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const view = within(utils.container)
    const connectTrigger = await view.findByRole('tab', { name: /Entegrasyon/ })
    await userEvent.click(connectTrigger)
    const toggle = await view.findByLabelText('Güvenli Embed')
    await userEvent.click(toggle)
    const domainsInput = await view.findByPlaceholderText('example.com, another.com')
    await userEvent.clear(domainsInput)
    await userEvent.type(domainsInput, 'a.com, b.com')
    expect((domainsInput as HTMLInputElement).value).toBe('a.com, b.com')
    const secretInput = await view.findByPlaceholderText('Gizli anahtar')
    await userEvent.clear(secretInput)
    await userEvent.type(secretInput, 'secret-2')
    expect((secretInput as HTMLInputElement).value).toBe('secret-2')
  })
})
