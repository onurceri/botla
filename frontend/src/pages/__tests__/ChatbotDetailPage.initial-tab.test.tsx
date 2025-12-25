import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from '@/test-utils'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import ConnectTab from '@/features/chatbot/pages/tabs/ConnectTab'
import { api } from '@/api/client'

describe('ChatbotDetailPage initial tab selection', () => {
  it('reads ?tab=connect from window.location and shows embed panel', async () => {
    window.localStorage.clear()

    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/organizations'))
        return Promise.resolve({ data: [{ id: 'org1', name: 'Test Org' }] } as any)
      if (url.includes('/workspaces'))
        return Promise.resolve({ data: [{ id: 'ws1', name: 'Test WS' }] } as any)
      if (url.includes('/api/v1/me'))
        return Promise.resolve({
          data: {
            subscription_plan: 'pro',
            config: { security: { secure_embed_enabled: true } },
          },
        } as any)
      if (url.includes('/api/v1/chatbots/abc'))
        return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    Object.defineProperty(window, 'location', {
      value: new URL('http://localhost/chatbots/abc?tab=connect'),
    })
    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter initialEntries={['/chatbots/abc']}>
            <Routes>
              <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
                <Route path="connect" element={<ConnectTab />} />
              </Route>
            </Routes>
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )
    expect(await screen.findByText('Kodu Web Sitenize Ekleyin')).toBeInTheDocument()
  })
})
