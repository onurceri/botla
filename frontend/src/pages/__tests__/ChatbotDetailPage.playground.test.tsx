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

describe('ChatbotDetailPage playground', () => {
  it('sends chat message and renders assistant reply', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: '123', name: 'Bot' } } as any)
    vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { response: 'Merhaba' } } as any)

    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/123"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
              <Route path="playground" element={<PlaygroundTab />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )

    const playTabs = screen.getAllByRole('link', { name: /Görünüm/i })
    await user.click(playTabs[playTabs.length - 1])
    const badge = await screen.findByText('1')
    const openBtn = badge.closest('button') as HTMLButtonElement
    await user.click(openBtn)
    const inputs = screen.getAllByPlaceholderText('Mesaj yazın...')
    const input = inputs[inputs.length - 1]
    await user.type(input, 'selam')
    await user.keyboard('{Enter}')
    expect(await screen.findByText('Merhaba')).toBeInTheDocument()
  })

  it('blocks rapid sends while loading and disables input', async () => {
    const user = userEvent.setup()
    vi.restoreAllMocks()
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: '123', name: 'Bot' } } as any)
    const postSpy = vi.spyOn(api, 'post').mockImplementation(async () => {
      await new Promise((r) => setTimeout(r, 300))
      return { data: { response: 'Merhaba' } } as any
    })

    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/123"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
              <Route path="playground" element={<PlaygroundTab />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )

    const playTabs2 = screen.getAllByRole('link', { name: /Görünüm/i })
    await user.click(playTabs2[playTabs2.length - 1])
    const badge = await screen.findByText('1')
    const openBtn = badge.closest('button') as HTMLButtonElement
    await user.click(openBtn)
    const poweredBy = await screen.findAllByText(/Powered by/i)
    const panel = poweredBy[poweredBy.length - 1]
    const input = panel.parentElement!.querySelector('input[placeholder="Mesaj yazın..."]') as HTMLInputElement
    await user.type(input, 'selam')
    input.focus()
    await user.keyboard('{Enter}')
    expect(input).toBeDisabled()
    expect(postSpy).toHaveBeenCalledTimes(1)
    expect(input).toBeDisabled()
    expect(await screen.findByText('Merhaba')).toBeInTheDocument()
  })

  it('guards against empty or whitespace-only messages', async () => {
    const user = userEvent.setup()
    vi.restoreAllMocks()
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: '123', name: 'Bot' } } as any)
    const postSpy = vi.spyOn(api, 'post').mockResolvedValue({ data: { response: 'Merhaba' } } as any)

    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/123"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
              <Route path="playground" element={<PlaygroundTab />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )

    const playTabs = screen.getAllByRole('link', { name: /Görünüm/i })
    await user.click(playTabs[playTabs.length - 1])
    const badge = await screen.findByText('1')
    const openBtn = badge.closest('button') as HTMLButtonElement
    await user.click(openBtn)
    const inputs2 = screen.getAllByPlaceholderText('Mesaj yazın...')
    const input = inputs2[inputs2.length - 1]

    // Empty
    await user.keyboard('{Enter}')
    // Whitespace
    await user.type(input, '   ')
    await user.keyboard('{Enter}')

    expect(postSpy).not.toHaveBeenCalled()
    expect(screen.queryByText('Bir hata oluştu.')).not.toBeInTheDocument()
  })
})
