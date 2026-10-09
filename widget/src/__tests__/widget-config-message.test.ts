/**
 * Integration test for the widget's WIDGET_CONFIG listener (Playground support).
 */
import { describe, it, expect, vi, beforeAll, afterAll, beforeEach } from 'vitest'

const fetchMock = vi.fn((input: RequestInfo | URL) => {
  void input
  return Promise.resolve(new Response(JSON.stringify({}), { status: 200 }))
})

function fetchedUrls(): string[] {
  return fetchMock.mock.calls.map((c) => String(c[0]))
}

function sendConfig(origin: string, config: Record<string, string>, source: MessageEventSource | null = window) {
  window.dispatchEvent(new MessageEvent('message', { data: { type: 'WIDGET_CONFIG', config }, origin, source }))
}

async function flush() {
  await new Promise((r) => setTimeout(r, 50))
}

async function expectFetched(prefix: string) {
  await vi.waitFor(() => expect(fetchedUrls().some((u) => u.startsWith(prefix))).toBe(true), {
    timeout: 3000,
  })
}

describe('widget WIDGET_CONFIG listener', () => {
  let widget: typeof import('../widget')

  beforeAll(async () => {
    vi.stubGlobal('fetch', fetchMock)
    // Pin the trusted dashboard origin so a developer's widget/.env
    // (e.g. VITE_DASHBOARD_URL=http://localhost:5173) cannot change the result.
    vi.stubEnv('VITE_DASHBOARD_URL', 'https://botla.app')
    window.__CBW_PARAMS = { 'skip-auto-mount': '1' }
    widget = await import('../widget')
  })

  afterAll(() => {
    widget.unmount()
    vi.unstubAllGlobals()
    vi.unstubAllEnvs()
  })

  beforeEach(() => {
    vi.spyOn(console, 'warn').mockImplementation(() => {})
  })

  it('applies config only from the dashboard origin and validates api-base', async () => {
    window.__CBW_PARAMS = {
      'chatbot-id': 'bot1',
      'preview-mode': '1',
      'api-base': 'https://api.botla.app',
    }
    widget.mount()
    await expectFetched('https://api.botla.app/api/v1/public/chatbots/bot1')

    // Lookalike origin that a prefix check would have accepted
    fetchMock.mockClear()
    sendConfig('https://botla.app.evil.com', { 'chatbot-id': 'bot1', 'api-base': 'https://evil.example' })
    await flush()
    expect(fetchedUrls().some((u) => u.includes('evil.example'))).toBe(false)

    // Trusted origin but not the parent window
    sendConfig('https://botla.app', { 'chatbot-id': 'bot1', 'api-base': 'https://evil.example' }, null)
    await flush()
    expect(fetchedUrls().some((u) => u.includes('evil.example'))).toBe(false)

    // Trusted origin, insecure api-base is ignored (keeps current base)
    sendConfig('https://botla.app', { 'chatbot-id': 'bot2', 'api-base': 'http://evil.example' })
    await expectFetched('https://api.botla.app/api/v1/public/chatbots/bot2')
    expect(fetchedUrls().some((u) => u.includes('evil.example'))).toBe(false)

    // Trusted origin with a valid https api-base is applied
    fetchMock.mockClear()
    sendConfig('https://botla.app', { 'chatbot-id': 'bot3', 'api-base': 'https://api2.botla.app/' })
    await expectFetched('https://api2.botla.app/api/v1/public/chatbots/bot3')
  })
})
