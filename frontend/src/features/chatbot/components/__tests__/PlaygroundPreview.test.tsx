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
    expect(iframe).toHaveAttribute('title', 'Chatbot Preview')
    expect(iframe).toHaveAttribute('sandbox', 'allow-scripts allow-same-origin allow-forms allow-popups')
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
    await new Promise(resolve => setTimeout(resolve, 150))
    
    expect(mockPostMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'WIDGET_CONFIG',
        config: expect.objectContaining({
          'chatbot-id': '123',
          'session-id': 'test-session-id',
          'auto-open': '1',
          'color': '#a78bfa',
          'position': 'bottom-right',
        }),
      }),
      '*'
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
      />
    )
    
    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    Object.defineProperty(iframe, 'contentWindow', {
      value: { postMessage: mockPostMessage },
      writable: true,
    })
    
    iframe.dispatchEvent(new Event('load'))
    await new Promise(resolve => setTimeout(resolve, 150))
    
    expect(mockPostMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'WIDGET_CONFIG',
        config: expect.objectContaining({
          suggestions: JSON.stringify(suggestions),
        }),
      }),
      '*'
    )
  })

  it('includes branding config when provided', async () => {
    const mockPostMessage = vi.fn()
    const customBranding = { text: 'Custom Brand', link: 'https://example.com' }
    
    const { container } = render(
      <PlaygroundPreview
        {...defaultProps}
        hideBranding={true}
        customBranding={customBranding}
      />
    )
    
    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    Object.defineProperty(iframe, 'contentWindow', {
      value: { postMessage: mockPostMessage },
      writable: true,
    })
    
    iframe.dispatchEvent(new Event('load'))
    await new Promise(resolve => setTimeout(resolve, 150))
    
    expect(mockPostMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'WIDGET_CONFIG',
        config: expect.objectContaining({
          'hide-branding': '1',
          'custom-branding': JSON.stringify(customBranding),
        }),
      }),
      '*'
    )
  })
})
