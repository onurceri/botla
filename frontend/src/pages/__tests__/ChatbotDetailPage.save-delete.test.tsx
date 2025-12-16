import { describe, it, expect, vi, afterEach } from 'vitest'
import { QueryWrapper } from "@/test-utils"
import { render, screen, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage save/delete', () => {
  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
  })
  it('creates new chatbot on valid form and shows success toast', async () => {
    const user = userEvent.setup()
    
    Object.defineProperty(window, 'localStorage', {
      value: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn() },
      writable: true
    })
    
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
       if (url.includes('/api/v1/organizations')) return Promise.resolve({ data: [{ id: 'org1', name: 'Test Org' }] } as any)
       if (url.includes('/workspaces')) return Promise.resolve({ data: [{ id: 'ws1', name: 'Test WS' }] } as any)
       return Promise.resolve({ data: {} } as any)
    })

    const postSpy = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { id: '999' } } as any)

    render(
      <QueryWrapper>
        <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/new"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
            <Route path="/dashboard/chatbots" element={<div />} />
            <Route path="/dashboard/chatbots/:id" element={<div />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
      </QueryWrapper>
    )

    const nameInput = screen.getByPlaceholderText('Örn: Müşteri Temsilcisi')
    await user.type(nameInput, 'Yeni Bot')
    const createBtn = screen.getByRole('button', { name: 'Oluştur' })
    await user.click(createBtn)

    expect(postSpy).toHaveBeenCalledWith('/api/v1/chatbots', expect.any(Object))
    expect(await screen.findByText('Chatbot başarıyla oluşturuldu.')).toBeInTheDocument()
  })

  it('deletes existing chatbot and shows success toast', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/123/sources')) return Promise.resolve({ data: [] } as any)
      if (url.includes('/api/v1/chatbots/123')) return Promise.resolve({ data: { id: '123', name: 'Var Olan Bot' } } as any)
      return Promise.resolve({ data: {} } as any)
    })
    vi.spyOn(window, 'confirm').mockReturnValue(true as any)
    const delSpy = vi.spyOn(api, 'delete').mockResolvedValueOnce({ data: {} } as any)

    render(
      <QueryWrapper>
        <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/123"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
            <Route path="/dashboard/chatbots" element={<div />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
      </QueryWrapper>
    )

    await screen.findByText('Var Olan Bot')
    const deleteBtn = document.querySelector('button[aria-label="Sil"]') as HTMLButtonElement
    expect(deleteBtn).toBeTruthy()
    await user.click(deleteBtn)
    expect(delSpy).toHaveBeenCalledWith('/api/v1/chatbots/123')
    expect(await screen.findByText('Chatbot silindi.')).toBeInTheDocument()
  })
})
