import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from '@/test-utils'
import { render, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { ToastProvider } from '@/components/ui/toast'
import { api } from '@/api/client'

vi.mock('@/features/organization/context/OrganizationContext', () => ({
  useOrganization: () => ({
    currentWorkspace: { id: 'ws-1' },
    isLoading: false,
  }),
  OrganizationProvider: ({ children }: any) => children,
}))

describe('ChatbotDetailPage tab query param', () => {
  it('opens sources tab on trigger click', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me'))
        return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc'))
        return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const utils = render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter initialEntries={['/chatbots/abc']}>
            <Routes>
              <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
                <Route path="sources" element={<div>Bilgi Bankası</div>} />
                <Route
                  path="playground"
                  element={<div aria-label="Sohbeti aç">Powered by Botla</div>}
                />
              </Route>
            </Routes>
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )
    const view = within(utils.container)
    const user = userEvent.setup()
    const triggers = await view.findAllByRole('link', { name: /Kaynaklar/ })
    await user.click(triggers[0])
    expect(await view.findByText('Bilgi Bankası')).toBeInTheDocument()
    // expect(view.getByText('Henüz kaynak eklenmemiş')).toBeInTheDocument()
  })

  it('opens playground tab on trigger click', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me'))
        return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc'))
        return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const utils2 = render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter initialEntries={['/chatbots/abc']}>
            <Routes>
              <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
                <Route path="sources" element={<div>Bilgi Bankası</div>} />
                <Route
                  path="playground"
                  element={<div aria-label="Sohbeti aç">Powered by Botla</div>}
                />
              </Route>
            </Routes>
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )
    const view2 = within(utils2.container)
    const user = userEvent.setup()
    const triggers2 = await view2.findAllByRole('link', { name: /Görünüm ve Test/i })
    await user.click(triggers2[0])
    const openBtn = await view2.findByLabelText('Sohbeti aç')
    await user.click(openBtn)
    expect(await view2.findByText('Powered by Botla')).toBeInTheDocument()
  })
})
