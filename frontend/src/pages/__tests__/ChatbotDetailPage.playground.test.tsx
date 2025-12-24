import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import ChatbotDetailPage from '../ChatbotDetailPage'
import PlaygroundTab from '@/features/chatbot/pages/tabs/PlaygroundTab'
import { api } from '@/api/client'
import { QueryWrapper } from '@/test-utils'

vi.mock('@/features/organization/context/OrganizationContext', () => ({
  useOrganization: () => ({
    currentWorkspace: { id: 'ws-1' },
    isLoading: false,
  }),
  OrganizationProvider: ({ children }: any) => children,
}))

// Mock PlaygroundPreview to avoid iframe issues and use the mocked WidgetApp
vi.mock('@/features/chatbot/components/PlaygroundPreview', async () => {
  const React = await import('react')
  const { api } = await import('@/api/client')

  const FakeWidgetApp = () => {
    const [messages, setMessages] = React.useState<string[]>([])
    const [loading, setLoading] = React.useState(false)
    const [input, setInput] = React.useState('')

    const handleSend = async () => {
      if (!input.trim()) return
      setLoading(true)
      try {
        const { data } = await api.post('/api/v1/chatbots/123/chat', { message: input })
        setMessages((prev) => [...prev, data.response])
      } catch (e) {
        console.error(e)
      }
      setLoading(false)
      setInput('')
    }

    return (
      <div>
        {messages.map((m, i) => (
          <div key={i}>{m}</div>
        ))}
        <input
          placeholder="Mesaj yazın..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          disabled={loading}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
        />
        <div>Powered by Botla</div>
      </div>
    )
  }

  return {
    default: () => <FakeWidgetApp />,
  }
})

describe('ChatbotDetailPage playground', () => {
  it('sends chat message and renders assistant reply', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: '123', name: 'Bot' } } as any)
    vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { response: 'Merhaba' } } as any)

    render(
      <QueryWrapper>
        <MemoryRouter initialEntries={['/chatbots/123']}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
              <Route path="playground" element={<PlaygroundTab />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </QueryWrapper>,
    )

    const playTabs = await screen.findAllByRole('link', { name: /Görünüm ve Test/i })
    await user.click(playTabs[playTabs.length - 1])
    expect(await screen.findByRole('heading', { name: /Görünüm ve Test/i })).toBeInTheDocument()
    const inputs = screen.getAllByPlaceholderText('Mesaj yazın...')
    const input = inputs[inputs.length - 1]
    await user.type(input, 'selam')
    await user.keyboard('{Enter}')
    expect(await screen.findByText('Merhaba')).toBeInTheDocument()
  })

  it('blocks rapid sends while loading and disables input', async () => {
    const user = userEvent.setup()
    vi.clearAllMocks()
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: '123', name: 'Bot' } } as any)
    const postSpy = vi.spyOn(api, 'post').mockImplementation(async () => {
      await new Promise((r) => setTimeout(r, 300))
      return { data: { response: 'Merhaba' } } as any
    })

    render(
      <QueryWrapper>
        <MemoryRouter initialEntries={['/chatbots/123']}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
              <Route path="playground" element={<PlaygroundTab />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </QueryWrapper>,
    )

    const playTabs2 = await screen.findAllByRole('link', { name: /Görünüm ve Test/i })
    await user.click(playTabs2[playTabs2.length - 1])
    const poweredBy = await screen.findAllByText(/Powered by/i)
    const panel = poweredBy[poweredBy.length - 1]
    const input = panel.parentElement!.querySelector(
      'input[placeholder="Mesaj yazın..."]',
    ) as HTMLInputElement
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
    vi.clearAllMocks()
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: '123', name: 'Bot' } } as any)
    const postSpy = vi
      .spyOn(api, 'post')
      .mockResolvedValue({ data: { response: 'Merhaba' } } as any)

    render(
      <QueryWrapper>
        <MemoryRouter initialEntries={['/chatbots/123']}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
              <Route path="playground" element={<PlaygroundTab />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </QueryWrapper>,
    )

    const playTabs3 = await screen.findAllByRole('link', { name: /Görünüm ve Test/i })
    await user.click(playTabs3[playTabs3.length - 1])
    const inputs2 = screen.getAllByPlaceholderText('Mesaj yazın...')
    const input = inputs2[inputs2.length - 1]
    await user.type(input, '   ')
    await user.keyboard('{Enter}')
    expect(postSpy).not.toHaveBeenCalled()
  })
})
