import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from '@/test-utils'
import { render, within, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import DeployTab from '@/features/chatbot/pages/tabs/DeployTab'
import { api } from '@/api/client'

describe('ChatbotDetailPage connect tab secret refresh', () => {
  it('enables secure embed and refreshes secret', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me/plan'))
        return Promise.resolve({
          data: {
            code: 'pro',
            features: {
              security: { secure_embed_enabled: true },
            },
          },
        } as any)
      if (url.includes('/api/v1/me'))
        return Promise.resolve({
          data: {
            subscription_plan: 'pro',
          },
        } as any)
      if (url.includes('/api/v1/chatbots/abc'))
        return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })

    const utils = render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter initialEntries={['/chatbots/abc?tab=deploy']}>
            <Routes>
              <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
                <Route path="deploy" element={<DeployTab />} />
              </Route>
            </Routes>
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )

    const view = within(utils.container)
    const connectTriggers = await view.findAllByRole('link', { name: /Yayınla/ })
    await userEvent.click(connectTriggers[connectTriggers.length - 1])
    await view.findByText('Kodu Web Sitenize Ekleyin')
    const secureToggle = await view.findByLabelText('Güvenli Embed')
    await userEvent.click(secureToggle)

    // Open the advanced section to access the secret/refresh
    const advancedBtn = await screen.findByText(/Gelişmiş: Token Doğrulama/)
    await userEvent.click(advancedBtn)

    const secretInput = await view.findByPlaceholderText('Gizli anahtar henüz oluşturulmadı')
    const before = (secretInput as HTMLInputElement).value
    const refreshBtn = view.getByRole('button', { name: 'Yenile' })
    await userEvent.click(refreshBtn)
    const after = (
      (await view.findByPlaceholderText('Gizli anahtar henüz oluşturulmadı')) as HTMLInputElement
    ).value
    expect(after).not.toEqual(before)
    expect(after.length).toBeGreaterThan(0)
  })
})
