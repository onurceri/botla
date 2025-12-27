import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { ChatDrawer } from '../ChatDrawer'
import type { ChatMessage } from '../../types'

describe('ChatDrawer Component', () => {
  const defaultProps = {
    messages: [],
    loading: false,
    input: '',
    setInput: vi.fn(),
    onSend: vi.fn(),
    onClose: vi.fn(),
  }

  it('renders chat drawer with default bot name', () => {
    render(<ChatDrawer {...defaultProps} />)

    expect(screen.getByText('Chatbot')).toBeInTheDocument()
  })

  it('renders chat drawer with custom bot name', () => {
    render(<ChatDrawer {...defaultProps} botName="Custom Bot" />)

    expect(screen.getByText('Custom Bot')).toBeInTheDocument()
  })

  it('renders all messages', () => {
    const messages: ChatMessage[] = [
      { id: '1', role: 'user', content: 'Hello', ts: Date.now() },
      { id: '2', role: 'assistant', content: 'Hi there!', ts: Date.now() },
    ]

    render(<ChatDrawer {...defaultProps} messages={messages} />)

    expect(screen.getByText('Hello')).toBeInTheDocument()
    expect(screen.getByText('Hi there!')).toBeInTheDocument()
  })

  it('calls setInput when textarea value changes', () => {
    const setInput = vi.fn()

    render(<ChatDrawer {...defaultProps} setInput={setInput} />)

    const textarea = screen.getByPlaceholderText('Mesaj yazın...')
    fireEvent.change(textarea, { target: { value: 'Test message' } })

    expect(setInput).toHaveBeenCalledWith('Test message')
  })

  it('calls onSend when send button is clicked', () => {
    const onSend = vi.fn()

    render(<ChatDrawer {...defaultProps} input="Test" onSend={onSend} />)

    const sendBtn = screen.getByLabelText('Gönder')
    fireEvent.click(sendBtn)

    expect(onSend).toHaveBeenCalledTimes(1)
  })

  it('calls onSend when Enter is pressed without Shift', () => {
    const onSend = vi.fn()

    render(<ChatDrawer {...defaultProps} input="Test" onSend={onSend} />)

    const textarea = screen.getByPlaceholderText('Mesaj yazın...')
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: false })

    expect(onSend).toHaveBeenCalledTimes(1)
  })

  it('does not call onSend when Enter is pressed with Shift', () => {
    const onSend = vi.fn()

    render(<ChatDrawer {...defaultProps} input="Test" onSend={onSend} />)

    const textarea = screen.getByPlaceholderText('Mesaj yazın...')
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: true })

    expect(onSend).not.toHaveBeenCalled()
  })

  it('disables send button when input is empty', () => {
    render(<ChatDrawer {...defaultProps} input="" />)

    const sendBtn = screen.getByLabelText('Gönder')
    expect(sendBtn).toBeDisabled()
  })

  it('disables send button and textarea when loading', () => {
    render(<ChatDrawer {...defaultProps} loading={true} input="Test" />)

    const sendBtn = screen.getByLabelText('Gönder')
    const textarea = screen.getByPlaceholderText('Mesaj yazın...')

    expect(sendBtn).toBeDisabled()
    expect(textarea).toBeDisabled()
  })

  it('shows loading indicator when loading is true', () => {
    const { container } = render(<ChatDrawer {...defaultProps} loading={true} />)

    expect(container.querySelector('.cbw-loading-row')).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn()

    render(<ChatDrawer {...defaultProps} onClose={onClose} />)

    const closeBtn = screen.getByLabelText('Kapat')
    fireEvent.click(closeBtn)

    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('displays suggestions when no user messages exist', () => {
    const messages: ChatMessage[] = []
    const suggestions = ['Question 1', 'Question 2']
    const onPickSuggestion = vi.fn()

    render(
      <ChatDrawer
        {...defaultProps}
        messages={messages}
        suggestions={suggestions}
        onPickSuggestion={onPickSuggestion}
      />
    )

    expect(screen.getByText('Question 1')).toBeInTheDocument()
  })

  it('hides suggestions after user sends a message', () => {
    const messages: ChatMessage[] = [{ id: '1', role: 'user', content: 'Hello', ts: Date.now() }]
    const suggestions = ['Question 1', 'Question 2']

    render(<ChatDrawer {...defaultProps} messages={messages} suggestions={suggestions} />)

    expect(screen.queryByText('Question 1')).not.toBeInTheDocument()
  })

  it('shows character count', () => {
    render(<ChatDrawer {...defaultProps} input="Hello" maxChars={1000} />)

    expect(screen.getByText('5 / 1000')).toBeInTheDocument()
  })

  it('prevents input beyond max characters', () => {
    const setInput = vi.fn()

    render(<ChatDrawer {...defaultProps} setInput={setInput} maxChars={10} />)

    const textarea = screen.getByPlaceholderText('Mesaj yazın...')
    fireEvent.change(textarea, { target: { value: '12345678901' } }) // 11 chars

    // Should not call setInput because it exceeds maxChars
    expect(setInput).not.toHaveBeenCalled()
  })

  it('shows default branding when not hidden', () => {
    render(<ChatDrawer {...defaultProps} hideBranding={false} />)

    expect(screen.getByText(/Powered by/)).toBeInTheDocument()
    expect(screen.getByText('Botla')).toBeInTheDocument()
  })

  it('shows custom branding when provided', () => {
    const customBranding = {
      text: 'Powered by Custom',
      link: 'https://custom.com',
    }

    render(<ChatDrawer {...defaultProps} hideBranding={true} customBranding={customBranding} />)

    expect(screen.getByText('Powered by Custom')).toBeInTheDocument()
    expect(screen.getByText('Powered by Custom')).toHaveAttribute('href', 'https://custom.com')
  })

  it('hides branding when hideBranding is true and no custom branding', () => {
    render(<ChatDrawer {...defaultProps} hideBranding={true} />)

    expect(screen.queryByText(/Powered by/)).not.toBeInTheDocument()
  })

  it('applies preview mode class when isPreviewMode is true', () => {
    const { container } = render(<ChatDrawer {...defaultProps} isPreviewMode={true} />)

    expect(container.querySelector('.cbw-preview-panel')).toBeInTheDocument()
  })

  it('displays bot icon when provided', () => {
    render(<ChatDrawer {...defaultProps} botIcon="https://example.com/bot.png" />)

    const icon = screen.getAllByAltText('')[0]
    expect(icon).toHaveAttribute('src', 'https://example.com/bot.png')
  })
})
