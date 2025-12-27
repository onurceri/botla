import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { sendMessage, MAX_MESSAGE_LENGTH } from '../api/chat'

describe('Widget Edge Cases', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('handles network failure gracefully', async () => {
    vi.spyOn(global, 'fetch').mockRejectedValue(new Error('Network error'))
    
    const result = await sendMessage('test-bot', 'Hello')
    
    expect(result.error).toBeDefined()
    expect(result.error?.toLowerCase()).toContain('network')
  })

  it('handles fetch failed error', async () => {
    vi.spyOn(global, 'fetch').mockRejectedValue(new Error('Failed to fetch'))
    
    const result = await sendMessage('test-bot', 'Hello')
    
    expect(result.error).toBeDefined()
    expect(result.error?.toLowerCase()).toContain('failed')
  })

  it('handles 404 chatbot not found', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: 'ERR_NOT_FOUND' }),
    } as Response)
    
    const result = await sendMessage('invalid-id', 'Hello')
    
    expect(result.error).toBeDefined()
    expect(result.error).toBe('ERR_NOT_FOUND')
  })

  it('handles 404 with JSON parse failure', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 404,
      json: () => Promise.reject(new Error('Invalid JSON')),
    } as Response)
    
    const result = await sendMessage('invalid-id', 'Hello')
    
    expect(result.error).toBeDefined()
    expect(result.error).toBe('ERR_NOT_FOUND')
  })

  it('handles rate limiting (429)', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 429,
      json: () => Promise.resolve({ error: 'ERR_RATE_LIMITED' }),
    } as Response)
    
    const result = await sendMessage('bot-id', 'Hello')
    
    expect(result.error).toBeDefined()
    expect(result.rateLimited).toBe(true)
  })

  it('handles rate limiting with JSON parse failure', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 429,
      json: () => Promise.reject(new Error('Invalid JSON')),
    } as Response)
    
    const result = await sendMessage('bot-id', 'Hello')
    
    expect(result.rateLimited).toBe(true)
    expect(result.error).toBe('ERR_RATE_LIMITED')
  })

  it('handles other HTTP errors', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.resolve({ error: 'Internal Server Error' }),
    } as Response)
    
    const result = await sendMessage('bot-id', 'Hello')
    
    expect(result.error).toBeDefined()
    expect(result.error).toBe('HTTP 500')
  })

  it('truncates overly long messages', async () => {
    const longMessage = 'a'.repeat(10000)
    
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ 
        response: 'OK',
        message_id: 'msg-123'
      }),
    } as Response)
    
    const result = await sendMessage('bot-id', longMessage)
    
    // Message should be truncated to MAX_MESSAGE_LENGTH
    expect(result.message?.length).toBeLessThanOrEqual(MAX_MESSAGE_LENGTH)
    expect(result.message?.length).toBe(MAX_MESSAGE_LENGTH)
    expect(result.error).toBeUndefined()
  })

  it('preserves messages under the limit', async () => {
    const normalMessage = 'Hello, this is a normal message'
    
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ 
        response: 'Hi there!',
        message_id: 'msg-456'
      }),
    } as Response)
    
    const result = await sendMessage('bot-id', normalMessage)
    
    expect(result.message).toBe(normalMessage)
    expect(result.response).toBe('Hi there!')
    expect(result.error).toBeUndefined()
  })

  it('includes session and embed token in request', async () => {
    let lastRequestBody = ''
    let lastHeaders: Record<string, string> = {}
    
    vi.spyOn(global, 'fetch').mockImplementation(async (_, init) => {
      lastRequestBody = (init as RequestInit).body as string
      lastHeaders = (init as RequestInit).headers as Record<string, string>
      return {
        ok: true,
        json: () => Promise.resolve({ response: 'OK', message_id: 'msg-789' }),
      } as Response
    })
    
    await sendMessage('bot-id', 'Hello', {
      sessionId: 'session-123',
      embedToken: 'token-abc',
      captchaToken: 'captcha-xyz'
    })
    
    const body = JSON.parse(lastRequestBody)
    expect(body.session_id).toBe('session-123')
    expect(body.captcha_token).toBe('captcha-xyz')
    expect(lastHeaders['X-Embed-Token']).toBe('token-abc')
  })

  it('uses correct API URL with custom base', async () => {
    let capturedUrl: string | null = null
    
    vi.spyOn(global, 'fetch').mockImplementation(async (url) => {
      capturedUrl = url as string
      return {
        ok: true,
        json: () => Promise.resolve({ response: 'OK', message_id: 'msg-000' }),
      } as Response
    })
    
    await sendMessage('my-chatbot', 'Hello', {
      apiBase: 'https://api.botla.app'
    })
    
    expect(capturedUrl).toBe('https://api.botla.app/api/v1/public/chatbots/my-chatbot/chat')
  })

  it('handles successful response with handoff', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ 
        response: 'Connecting you to a human...',
        message_id: 'msg-handoff',
        handoff_request_id: 'handoff-123'
      }),
    } as Response)
    
    const result = await sendMessage('bot-id', 'I need human help')
    
    expect(result.response).toBe('Connecting you to a human...')
    expect(result.handoffRequestId).toBe('handoff-123')
    expect(result.error).toBeUndefined()
  })

  it('handles non-Error exceptions', async () => {
    vi.spyOn(global, 'fetch').mockRejectedValue('String error')
    
    const result = await sendMessage('bot-id', 'Hello')
    
    expect(result.error).toBeDefined()
  })
})
