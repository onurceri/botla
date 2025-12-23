import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from '@/test-utils'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage save error branches', () => {
  it('new chatbot: validate ok but POST fails shows error toast', async () => {
    const user = userEvent.setup()

    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    try {
      Object.defineProperty(window, 'localStorage', {
        value: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn() },
        writable: true,
      })

      vi.spyOn(api, 'get').mockImplementation((url: string) => {
        if (url.includes('/api/v1/organizations'))
          return Promise.resolve({ data: [{ id: 'org1', name: 'Test Org' }] } as any)
        if (url.includes('/workspaces'))
          return Promise.resolve({ data: [{ id: 'ws1', name: 'Test WS' }] } as any)
        return Promise.resolve({ data: {} } as any)
      })

      vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))
      render(
        <QueryWrapper>
          <ToastProvider>
            <MemoryRouter initialEntries={['/chatbots/new']}>
              <Routes>
                <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
              </Routes>
            </MemoryRouter>
          </ToastProvider>
        </QueryWrapper>,
      )
      const nameInput = await screen.findByPlaceholderText('Örn: Müşteri Temsilcisi')
      await user.type(nameInput, 'Yeni Bot')
      const createBtn = await screen.findByRole('button', { name: 'Oluştur' })
      await user.click(createBtn)
      const errs1 = await screen.findAllByText('Bir hata oluştu. Lütfen tekrar deneyin.')
      expect(errs1.length).toBeGreaterThan(0)
    } finally {
      consoleErrorSpy.mockRestore()
    }
  })
})
