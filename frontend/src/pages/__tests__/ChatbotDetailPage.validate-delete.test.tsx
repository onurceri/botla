import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage validate & delete branches', () => {
  afterEach(() => {
    cleanup()
  })
  it('shows error toast when saving without name (new)', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'post')
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/new"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const createBtn = await screen.findByRole('button', { name: 'Oluştur' })
    await user.click(createBtn)
    expect(await screen.findByText('Lütfen bir bot ismi girin.')).toBeInTheDocument()
    expect((api.post as any).mock?.calls?.length || 0).toBe(0)
  })

  it('does not delete when confirm is cancelled', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/xyz')) return Promise.resolve({ data: { id: 'xyz', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/xyz/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const delSpy = vi.spyOn(api, 'delete')
    vi.spyOn(global, 'confirm').mockReturnValue(false as any)
    const utils = render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/xyz"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const view = within(utils.container)
    const deleteBtn = await view.findByLabelText('Sil')
    await user.click(deleteBtn)
    expect(delSpy).not.toHaveBeenCalled()
  })

  it('shows error toast when delete fails', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/xyz')) return Promise.resolve({ data: { id: 'xyz', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/xyz/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    vi.spyOn(global, 'confirm').mockReturnValue(true as any)
    vi.spyOn(api, 'delete').mockRejectedValueOnce(new Error('fail'))
    const utils2 = render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/xyz"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const view2 = within(utils2.container)
    const deleteBtn = await view2.findByLabelText('Sil')
    await user.click(deleteBtn)
    expect(await view2.findByText('Silme işlemi başarısız oldu.')).toBeInTheDocument()
  })
})
