import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { ToastProvider } from '@/components/ui/toast'

let capturedPutBody: any = null
vi.mock('@/api/client', () => {
  const put = vi.fn(async () => ({ data: {} }))
  const get = vi.fn(async (url: string) => {
    if (url === '/api/v1/me') return { data: { subscription_plan: 'pro' } }
    if (url.startsWith('/api/v1/chatbots/')) return { data: {
      id: 'bot1', name: 'B', model: 'gpt-3.5-turbo', theme_color: '#3b82f6', welcome_message: 'Merhaba!',
      position: 'bottom-right', bot_message_color: '#3b82f6', user_message_color: '#3b82f6', bot_message_text_color: '#ffffff', user_message_text_color: '#ffffff',
      chat_font_family: 'Inter, sans-serif', chat_header_color: '#3b82f6', chat_header_text_color: '#ffffff', bot_display_name: 'Bot', secure_embed_enabled: false,
      allowed_domains: '', embed_secret: ''
    } }
    return { data: {} }
  })
  return { api: { get, put } }
})
import { api } from '@/api/client'

describe('Connect tab secure embed UI', () => {
  beforeEach(() => {
    vi.stubGlobal('IntersectionObserver', class { observe(){} disconnect(){} })
  })

  it('paid plan: default payload excludes domain/secret until enabled', async () => {
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/bot1?tab=connect"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    // click save
    const saveBtn = await screen.findByText('Değişiklikleri Kaydet')
    fireEvent.click(saveBtn)
    await waitFor(() => expect((api.put as any).mock.calls.length > 0).toBeTruthy())
    const payload = (api.put as any).mock.calls.at(-1)[1]
    expect(payload.secure_embed_enabled).toBe(false)
    expect(payload.allowed_domains).toBeUndefined()
    expect(payload.embed_secret).toBeUndefined()
  })
})
