import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from "@/test-utils"
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage chat handler', () => {
  it('adds assistant message on success', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc')) return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { response: 'Selam!' } } as any)
    render(
      <QueryWrapper>
        <MemoryRouter initialEntries={["/chatbots/abc"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </QueryWrapper>
    )
    const btns = await screen.findAllByLabelText('Test Chat Send')
    await userEvent.click(btns[0])
    const assists = await screen.findAllByTestId('chat-last-assistant')
    expect(assists.some(el => el.textContent === 'Selam!')).toBe(true)
  })

  it('adds error assistant message on failure', async () => {
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/xyz')) return Promise.resolve({ data: { id: 'xyz', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/xyz/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))
    render(
      <QueryWrapper>
        <MemoryRouter initialEntries={["/chatbots/xyz"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </QueryWrapper>
    )
    const btns = await screen.findAllByLabelText('Test Chat Send')
    await userEvent.click(btns[0])
    const assists = await screen.findAllByTestId('chat-last-assistant')
    expect(assists.some(el => el.textContent === 'Bir hata oluştu.')).toBe(true)
  })
})
