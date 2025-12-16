import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from "@/test-utils"
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage POST navigate', () => {
  it('shows success toast on POST and enables save', async () => {
    const user = userEvent.setup()
    
    // Mock localStorage and Organizations
    Object.defineProperty(window, 'localStorage', {
      value: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn() },
      writable: true
    })
    
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
       if (url.includes('/api/v1/organizations')) {
         return Promise.resolve({ data: [{ id: 'org1', name: 'Test Org' }] } as any)
       }
       if (url.includes('/api/v1/workspaces')) {
         return Promise.resolve({ data: [{ id: 'ws1', name: 'Test WS' }] } as any)
       }
       return Promise.resolve({ data: {} } as any)
    })

    vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { id: '999' } } as any)
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
    const nameInput = await screen.findByPlaceholderText('Örn: Müşteri Temsilcisi')
    await user.type(nameInput, 'Yeni Bot')
    const createBtn = screen.getByRole('button', { name: 'Oluştur' })
    await user.click(createBtn)
    expect(await screen.findByText('Chatbot başarıyla oluşturuldu.')).toBeInTheDocument()
  })
})
