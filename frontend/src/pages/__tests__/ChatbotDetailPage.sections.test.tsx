import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import PlaygroundTab from '@/features/chatbot/pages/tabs/PlaygroundTab'
import { api } from '@/api/client'

vi.mock('@/features/organization/context/OrganizationContext', () => ({
  useOrganization: () => ({
    currentWorkspace: { id: 'ws-1' },
    isLoading: false
  }),
  OrganizationProvider: ({ children }: any) => children
}))

describe('ChatbotDetailPage sections toggle', () => {
  it('collapses and expands Identity/Appearance/Colors sections', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc')) return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const user = userEvent.setup()
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/abc?tab=playground"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
              <Route path="playground" element={<PlaygroundTab />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const playTrigger = await screen.findByRole('link', { name: /Görünüm/ })
    await user.click(playTrigger)
    const identityBtn = await screen.findByRole('button', { name: /Kimlik/ })
    await user.click(identityBtn)
    expect(screen.queryByLabelText('Bot Görünen Adı')).not.toBeInTheDocument()
    await user.click(identityBtn)
    expect(await screen.findByLabelText('Bot Görünen Adı')).toBeInTheDocument()

    // Appearance
    const appearanceBtn = screen.getByRole('button', { name: /Görünüm/ })
    appearanceBtn.click()
    expect(await screen.findByLabelText('Konum')).toBeInTheDocument()

    // Colors
    const colorsBtn = screen.getByRole('button', { name: /Renkler/ })
    colorsBtn.click()
    expect(await screen.findByText('Header Yazı')).toBeInTheDocument()
  })
})
