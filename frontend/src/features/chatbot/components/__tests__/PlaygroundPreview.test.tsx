import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, cleanup } from '@testing-library/react'
import PlaygroundPreview from '../PlaygroundPreview'

// Mock import.meta.env
vi.stubEnv('VITE_WIDGET_URL', 'http://localhost:5174')
vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:8080')

const defaultProps = {
  id: '123',
  themeColor: '#a78bfa',
  chatHeaderColor: '#3b82f6',
  chatHeaderTextColor: '#ffffff',
  botMessageColor: '#fcfcfd',
  botMessageTextColor: '#030303',
  userMessageColor: '#2e408a',
  userMessageTextColor: '#ffffff',
  chatFontFamily: 'Inter, sans-serif',
  position: 'bottom-right',
  botDisplayName: 'Destek',
  botIcon: '',
  chatBackgroundColor: '#FFF5E6',
  welcomeMessage: 'Merhaba! Size nasıl yardımcı olabilirim?',
  previewOpen: false,
  sessionId: 'test-session-id',
  suggestionsEnabled: false,
  suggestedQuestions: [],
  inputBackgroundColor: '#ffffff',
  inputTextColor: '#000000',
  sendButtonColor: '#3b82f6',
  bubbleRadius: '1rem',
  autoOpen: true,
}

describe('PlaygroundPreview', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders iframe with correct src', () => {
    const { container } = render(<PlaygroundPreview {...defaultProps} />)

    const iframe = container.querySelector('iframe')
    expect(iframe).toBeInTheDocument()
    expect(iframe?.src).toContain('/preview.html')
  })

  it('iframe has correct attributes', () => {
    const { container } = render(<PlaygroundPreview {...defaultProps} />)

    const iframe = container.querySelector('iframe')
    expect(iframe).toHaveAttribute('title', 'Chatbot Önizlemesi')
    expect(iframe).toHaveAttribute(
      'sandbox',
      'allow-scripts allow-same-origin allow-forms allow-popups',
    )
  })

  it('sends config via postMessage when iframe loads', async () => {
    const mockPostMessage = vi.fn()

    const { container } = render(<PlaygroundPreview {...defaultProps} />)
    const iframe = container.querySelector('iframe') as HTMLIFrameElement

    // Simulate iframe with contentWindow
    Object.defineProperty(iframe, 'contentWindow', {
      value: { postMessage: mockPostMessage },
      writable: true,
    })

    // Trigger load event
    iframe.dispatchEvent(new Event('load'))

    // Wait for setTimeout in handleIframeLoad
    await new Promise((resolve) => setTimeout(resolve, 150))

    expect(mockPostMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'WIDGET_CONFIG',
        config: expect.objectContaining({
          'chatbot-id': '123',
          'session-id': 'test-session-id',
          'auto-open': '1',
          color: '#a78bfa',
          position: 'bottom-right',
        }),
      }),
      'http://localhost:5174',
    )
  })

  it('includes suggestions in config when enabled', async () => {
    const mockPostMessage = vi.fn()
    const suggestions = ['Nasıl çalışır?', 'Fiyatlar nedir?']

    const { container } = render(
      <PlaygroundPreview
        {...defaultProps}
        suggestionsEnabled={true}
        suggestedQuestions={suggestions}
      />,
    )

    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    Object.defineProperty(iframe, 'contentWindow', {
      value: { postMessage: mockPostMessage },
      writable: true,
    })

    iframe.dispatchEvent(new Event('load'))
    await new Promise((resolve) => setTimeout(resolve, 150))

    expect(mockPostMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'WIDGET_CONFIG',
        config: expect.objectContaining({
          suggestions: JSON.stringify(suggestions),
        }),
      }),
      'http://localhost:5174',
    )
  })

  it('includes branding config when provided', async () => {
    const mockPostMessage = vi.fn()
    const customBranding = { text: 'Custom Brand', link: 'https://example.com' }

    const { container } = render(
      <PlaygroundPreview {...defaultProps} hideBranding={true} customBranding={customBranding} />,
    )

    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    Object.defineProperty(iframe, 'contentWindow', {
      value: { postMessage: mockPostMessage },
      writable: true,
    })

    iframe.dispatchEvent(new Event('load'))
    await new Promise((resolve) => setTimeout(resolve, 150))

    expect(mockPostMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'WIDGET_CONFIG',
        config: expect.objectContaining({
          'hide-branding': '1',
          'custom-branding': JSON.stringify(customBranding),
        }),
      }),
      'http://localhost:5174',
    )
  })

  it('debounces config updates on prop changes', async () => {
    vi.useFakeTimers()
    const mockPostMessage = vi.fn()

    const { container, rerender } = render(<PlaygroundPreview {...defaultProps} />)
    const iframe = container.querySelector('iframe') as HTMLIFrameElement

    Object.defineProperty(iframe, 'contentWindow', {
      value: { postMessage: mockPostMessage },
      writable: true,
    })

    // Trigger initial load
    iframe.dispatchEvent(new Event('load'))
    vi.advanceTimersByTime(150)

    // Store initial call count (should be 1 from load)
    const initialCallCount = mockPostMessage.mock.calls.length

    // Simulate rapid prop changes (like typing)
    rerender(<PlaygroundPreview {...defaultProps} welcomeMessage="H" />)
    vi.advanceTimersByTime(50)

    rerender(<PlaygroundPreview {...defaultProps} welcomeMessage="He" />)
    vi.advanceTimersByTime(50)

    rerender(<PlaygroundPreview {...defaultProps} welcomeMessage="Hel" />)
    vi.advanceTimersByTime(50)

    rerender(<PlaygroundPreview {...defaultProps} welcomeMessage="Hell" />)
    vi.advanceTimersByTime(50)

    rerender(<PlaygroundPreview {...defaultProps} welcomeMessage="Hello" />)

    // Should not have sent any new messages yet (debounce in progress)
    expect(mockPostMessage.mock.calls.length).toBe(initialCallCount)

    // Wait for debounce to complete (300ms)
    vi.advanceTimersByTime(350)

    // Now should have sent exactly one update (not 5)
    expect(mockPostMessage.mock.calls.length).toBe(initialCallCount + 1)

    // Verify the final message contains the last value
    const lastCall = mockPostMessage.mock.calls[mockPostMessage.mock.calls.length - 1]
    expect(lastCall[0].config['welcome']).toBe('Hello')

    vi.useRealTimers()
  })

  it('never posts config with a wildcard target origin', async () => {
    const mockPostMessage = vi.fn()
    const { container } = render(<PlaygroundPreview {...defaultProps} />)
    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    Object.defineProperty(iframe, 'contentWindow', {
      value: { postMessage: mockPostMessage },
      writable: true,
    })

    iframe.dispatchEvent(new Event('load'))
    await new Promise((resolve) => setTimeout(resolve, 150))

    expect(mockPostMessage).toHaveBeenCalled()
    for (const call of mockPostMessage.mock.calls) {
      expect(call[1]).toBe('http://localhost:5174')
    }
  })

  it.each([
    ['foreign origin', 'https://evil.example', 'iframe'],
    ['lookalike origin', 'http://localhost:5174.evil.example', 'iframe'],
    ['wrong source window', 'http://localhost:5174', 'other'],
  ])('ignores WIDGET_CONFIG_APPLIED from a %s and keeps retrying', async (_label, origin, src) => {
    vi.useFakeTimers()
    const mockPostMessage = vi.fn()
    const { container } = render(<PlaygroundPreview {...defaultProps} />)
    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    const contentWindow = { postMessage: mockPostMessage }
    Object.defineProperty(iframe, 'contentWindow', { value: contentWindow, writable: true })

    iframe.dispatchEvent(new Event('load'))
    vi.advanceTimersByTime(150)
    const afterLoad = mockPostMessage.mock.calls.length

    window.dispatchEvent(
      new MessageEvent('message', {
        origin,
        source: (src === 'iframe' ? contentWindow : window) as unknown as Window,
        data: { type: 'WIDGET_CONFIG_APPLIED' },
      }),
    )

    // Spoofed ack must not stop the retry mechanism
    vi.advanceTimersByTime(2000)
    expect(mockPostMessage.mock.calls.length).toBeGreaterThan(afterLoad)
    vi.useRealTimers()
  })

  it('accepts WIDGET_CONFIG_APPLIED from the preview iframe and stops retrying', async () => {
    vi.useFakeTimers()
    const mockPostMessage = vi.fn()
    const { container } = render(<PlaygroundPreview {...defaultProps} />)
    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    const contentWindow = { postMessage: mockPostMessage }
    Object.defineProperty(iframe, 'contentWindow', { value: contentWindow, writable: true })

    iframe.dispatchEvent(new Event('load'))
    vi.advanceTimersByTime(150)
    const afterLoad = mockPostMessage.mock.calls.length

    window.dispatchEvent(
      new MessageEvent('message', {
        origin: 'http://localhost:5174',
        source: contentWindow as unknown as Window,
        data: { type: 'WIDGET_CONFIG_APPLIED' },
      }),
    )

    vi.advanceTimersByTime(2000)
    expect(mockPostMessage.mock.calls.length).toBe(afterLoad)
    vi.useRealTimers()
  })
})
