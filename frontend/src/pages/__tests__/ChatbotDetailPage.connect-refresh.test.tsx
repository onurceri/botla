import { describe, it, expect, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage connect tab secret refresh', () => {
  it('enables secure embed and refreshes secret', async () => {
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
    await view.findByText('Web Sitenize Ekleyin')
    const secureToggle = await view.findByLabelText('Güvenli Embed')
    secureToggle.click()
    const secretInput = await view.findByPlaceholderText('Gizli anahtar')
    const before = (secretInput as HTMLInputElement).value
    const refreshBtn = view.getByRole('button', { name: 'Yenile' })
    refreshBtn.click()
    const after = (await view.findByPlaceholderText('Gizli anahtar') as HTMLInputElement).value
    expect(after).not.toEqual(before)
    expect(after.length).toBeGreaterThan(0)
  })
})
