import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import PlaygroundPreview from '../PlaygroundPreview'
import * as chatApi from '@/api/chat'

vi.mock('@/api/chat')

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

  it('renders chat interface with welcome message', () => {
    const { container } = render(<PlaygroundPreview {...defaultProps} />)
    
    expect(screen.getByText('Destek')).toBeInTheDocument()
    expect(screen.getByText('Merhaba! Size nasıl yardımcı olabilirim?')).toBeInTheDocument()
    expect(container.querySelector('.playground-input')).toBeInTheDocument()
  })

  it('displays suggested questions when enabled and no user messages', () => {
    render(
      <PlaygroundPreview
        {...defaultProps}
        suggestionsEnabled={true}
        suggestedQuestions={['Nasıl çalışır?', 'Fiyatlar nedir?']}
      />
    )

    expect(screen.getByText('Nasıl çalışır?')).toBeInTheDocument()
    expect(screen.getByText('Fiyatlar nedir?')).toBeInTheDocument()
  })

  it('does not display suggestions when disabled', () => {
    render(
      <PlaygroundPreview
        {...defaultProps}
        suggestionsEnabled={false}
        suggestedQuestions={['Nasıl çalışır?', 'Fiyatlar nedir?']}
      />
    )

    expect(screen.queryByText('Nasıl çalışır?')).not.toBeInTheDocument()
    expect(screen.queryByText('Fiyatlar nedir?')).not.toBeInTheDocument()
  })

  it('sends message when user types and clicks send', async () => {
    const mockSendChatMessage = vi.mocked(chatApi.sendChatMessage)
    mockSendChatMessage.mockResolvedValue({
      response: 'Test response',
      tokens_used: 10,
      sources_used: [],
    })

    const user = userEvent.setup()
    const { container } = render(<PlaygroundPreview {...defaultProps} />)

    const textarea = container.querySelector('.playground-input') as HTMLTextAreaElement
    const sendButton = screen.getByLabelText('Gönder')

    await user.type(textarea, 'Test message')
    await user.click(sendButton)

    expect(mockSendChatMessage).toHaveBeenCalledWith('123', {
      message: 'Test message',
      session_id: 'test-session-id',
    })

    await waitFor(() => {
      expect(screen.getByText('Test message')).toBeInTheDocument()
      expect(screen.getByText('Test response')).toBeInTheDocument()
    })
  })

  it('displays custom branding when enabled', () => {
    const { container } = render(
      <PlaygroundPreview
        {...defaultProps}
        hideBranding={true}
        customBranding={{ text: 'Custom Brand', link: 'https://example.com' }}
      />
    )

    expect(screen.getByText('Custom Brand')).toBeInTheDocument()
    // Powered by should not appear with custom branding
    const brandingDiv = container.querySelector('.playground-branding')
    expect(brandingDiv?.textContent).toContain('Custom Brand')
    expect(brandingDiv?.textContent).not.toContain('Botla')
  })

  it('displays default Botla branding when not hidden', () => {
    const { container } = render(<PlaygroundPreview {...defaultProps} hideBranding={false} />)

    const brandingDiv = container.querySelector('.playground-branding')
    expect(brandingDiv?.textContent).toContain('Powered by')
    expect(screen.getByText('Botla')).toBeInTheDocument()
  })

  it('enforces character limit', async () => {
    const user = userEvent.setup()
    const { container } = render(<PlaygroundPreview {...defaultProps} />)

    const textarea = container.querySelector('.playground-input') as HTMLTextAreaElement
    const longText = 'a'.repeat(1100)

    await user.type(textarea, longText)

    // Should only accept up to MAX_CHARS (1000)
    expect(textarea.value.length).toBeLessThanOrEqual(1000)
  })
})
