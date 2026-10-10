import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/preact'

vi.mock('react-markdown', () => ({
  default: ({ children }: { children: string }) => children
}))

import { WidgetApp } from '../widgetApp'
import { ChatDrawer } from '../components/ChatDrawer'
import { Message } from '../components/Message'
import { Suggestions } from '../components/Suggestions'

const BOT_ID = '11111111-1111-1111-1111-111111111111'

function mockConfig(config: Record<string, unknown>) {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => config }))
}

async function openWidget(props: { lang?: string } = {}) {
  render(<WidgetApp chatbotId={BOT_ID} {...props} />)
  await waitFor(() => expect(fetch).toHaveBeenCalled())
  // Let the config promise resolve before opening
  await new Promise((r) => setTimeout(r, 0))
  fireEvent.click(document.querySelector('.cbw-bubble')!)
}

describe('widget UI language', () => {
  beforeEach(() => {
    const store = new Map<string, string>()
    vi.stubGlobal('localStorage', {
      getItem: (k: string) => store.get(k) ?? null,
      setItem: (k: string, v: string) => { store.set(k, v) },
      removeItem: (k: string) => { store.delete(k) },
      clear: () => store.clear(),
    })
  })
  afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
  })

  it('uses English when the bot language is English', async () => {
    mockConfig({ language_code: 'en-US' })
    await openWidget()
    expect(await screen.findByPlaceholderText('Type a message...')).toBeDefined()
    expect(screen.getByLabelText('Close')).toBeDefined()
    expect(screen.getByLabelText('Close chat')).toBeDefined()
    expect(screen.getByLabelText('New conversation')).toBeDefined()
  })

  it('uses Turkish when the bot language is Turkish', async () => {
    mockConfig({ language_code: 'tr' })
    await openWidget()
    expect(await screen.findByPlaceholderText('Mesaj yazın...')).toBeDefined()
    expect(screen.getByLabelText('Sohbeti kapat')).toBeDefined()
  })

  it('falls back to Turkish when the language is missing or unknown', async () => {
    mockConfig({ language_code: 'de' })
    await openWidget()
    expect(await screen.findByPlaceholderText('Mesaj yazın...')).toBeDefined()
  })

  it('lets the lang embed param override the bot language', async () => {
    mockConfig({ language_code: 'tr' })
    await openWidget({ lang: 'en' })
    expect(await screen.findByPlaceholderText('Type a message...')).toBeDefined()
  })

  it('shows the chat error message in the bot language', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: async () => ({ language_code: 'en' }) })
      .mockResolvedValueOnce({ ok: false, status: 500, json: async () => ({}) })
    vi.stubGlobal('fetch', fetchMock)
    render(<WidgetApp chatbotId={BOT_ID} />)
    await new Promise((r) => setTimeout(r, 0))
    fireEvent.click(document.querySelector('.cbw-bubble')!)
    const input = await screen.findByPlaceholderText('Type a message...')
    fireEvent.input(input, { target: { value: 'hi' } })
    fireEvent.click(screen.getByLabelText('Send'))
    expect(await screen.findByText('Something went wrong, please try again.', {}, { timeout: 2000 })).toBeDefined()
  })
})

describe('components in English', () => {
  afterEach(() => cleanup())

  it('ChatDrawer renders English labels and the reset dialog', () => {
    render(
      <ChatDrawer messages={[]} loading={false} input="" setInput={() => {}} onSend={() => {}} onClose={() => {}}
        onResetSession={() => {}} showResetConfirm onResetConfirm={() => {}} onResetCancel={() => {}} lang="en" />
    )
    expect(screen.getByPlaceholderText('Type a message...')).toBeDefined()
    expect(screen.getByLabelText('Send')).toBeDefined()
    expect(screen.getByText('Your current conversation will be deleted.')).toBeDefined()
    expect(screen.getByText('Start')).toBeDefined()
    expect(screen.getByText('Cancel')).toBeDefined()
  })

  it('Message renders the English handoff form and feedback titles', () => {
    render(<Message m={{ role: 'assistant', content: 'x', type: 'handoff', handoffRequestId: 'r1' }} onSubmitEmail={async () => {}} lang="en" />)
    expect(screen.getByText('Support Request')).toBeDefined()
    expect(screen.getByPlaceholderText('you@example.com')).toBeDefined()
    expect(screen.getByText('Send')).toBeDefined()
    cleanup()
    render(<Message m={{ id: 'm1', role: 'assistant', content: 'x' }} onFeedback={() => {}} lang="en" />)
    expect(screen.getByTitle('Helpful')).toBeDefined()
    expect(screen.getByTitle('Not helpful')).toBeDefined()
  })

  it('Suggestions renders English labels', () => {
    render(<Suggestions items={['a', 'b']} disabled={false} onPick={() => {}} lang="en" />)
    expect(screen.getByLabelText('Suggested questions')).toBeDefined()
    expect(screen.getByLabelText('Previous question')).toBeDefined()
    expect(screen.getByLabelText('Next question')).toBeDefined()
    expect(screen.getByText(/SUGGESTED QUESTIONS/)).toBeDefined()
  })
})
