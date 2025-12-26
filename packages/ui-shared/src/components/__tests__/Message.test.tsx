import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { Message } from '../Message'
import type { ChatMessage } from '../../types'

describe('Message Component', () => {
  it('renders user message with correct styling', () => {
    const message: ChatMessage = {
      id: '1',
      role: 'user',
      content: 'Hello, bot!',
      ts: Date.now(),
    }

    render(<Message message={message} />)

    expect(screen.getByText('Hello, bot!')).toBeInTheDocument()
    const row = screen.getByText('Hello, bot!').closest('.cbw-msg-row')
    expect(row).toHaveClass('user')
  })

  it('renders assistant message with correct styling', () => {
    const message: ChatMessage = {
      id: '2',
      role: 'assistant',
      content: 'Hi there!',
      ts: Date.now(),
    }

    render(<Message message={message} />)

    expect(screen.getByText('Hi there!')).toBeInTheDocument()
    const row = screen.getByText('Hi there!').closest('.cbw-msg-row')
    expect(row).toHaveClass('assistant')
  })

  it('renders feedback buttons for assistant messages when onFeedback is provided', () => {
    const message: ChatMessage = {
      id: '3',
      role: 'assistant',
      content: 'How can I help?',
      ts: Date.now(),
    }

    const onFeedback = vi.fn()
    render(<Message message={message} onFeedback={onFeedback} />)

    const feedbackButtons = screen.getAllByRole('button')
    expect(feedbackButtons).toHaveLength(2) // positive and negative
  })

  it('calls onFeedback with correct parameters when feedback button is clicked', () => {
    const message: ChatMessage = {
      id: '4',
      role: 'assistant',
      content: 'Test message',
      ts: Date.now(),
    }

    const onFeedback = vi.fn()
    render(<Message message={message} onFeedback={onFeedback} />)

    const positiveBtn = screen.getByTitle('Yararlı')
    fireEvent.click(positiveBtn)

    expect(onFeedback).toHaveBeenCalledWith('4', true)
  })

  it('does not render feedback buttons for user messages', () => {
    const message: ChatMessage = {
      id: '5',
      role: 'user',
      content: 'User message',
      ts: Date.now(),
    }

    const onFeedback = vi.fn()
    render(<Message message={message} onFeedback={onFeedback} />)

    expect(screen.queryByTitle('Yararlı')).not.toBeInTheDocument()
  })

  it('renders handoff card when type is handoff', () => {
    const message: ChatMessage = {
      id: '6',
      role: 'assistant',
      content: 'Handoff message',
      type: 'handoff',
      handoffRequestId: 'req-123',
      ts: Date.now(),
    }

    const onSubmitEmail = vi.fn()
    render(<Message message={message} onSubmitEmail={onSubmitEmail} />)

    expect(screen.getByText('Destek Talebi')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('e-posta@adresiniz.com')).toBeInTheDocument()
  })

  it('displays bot icon when provided', () => {
    const message: ChatMessage = {
      id: '7',
      role: 'assistant',
      content: 'Message with icon',
      ts: Date.now(),
    }

    render(<Message message={message} botIcon="https://example.com/bot.png" />)

    const icon = screen.getByAltText('')
    expect(icon).toHaveAttribute('src', 'https://example.com/bot.png')
  })

  it('renders markdown content correctly', () => {
    const message: ChatMessage = {
      id: '8',
      role: 'assistant',
      content: '**Bold text** and *italic text*',
      ts: Date.now(),
    }

    const { container } = render(<Message message={message} />)

    expect(container.querySelector('strong')).toBeInTheDocument()
    expect(container.querySelector('em')).toBeInTheDocument()
  })

  it('shows submitted state for handoff after email submission', () => {
    const message: ChatMessage = {
      id: '9',
      role: 'assistant',
      content: 'Handoff',
      type: 'handoff',
      handoffRequestId: 'req-456',
      emailSubmitted: true,
      ts: Date.now(),
    }

    const onSubmitEmail = vi.fn()
    render(<Message message={message} onSubmitEmail={onSubmitEmail} />)

    expect(screen.getByText('Talebiniz alındı!')).toBeInTheDocument()
    expect(screen.queryByPlaceholderText('e-posta@adresiniz.com')).not.toBeInTheDocument()
  })

  it('applies custom class names', () => {
    const message: ChatMessage = {
      id: '10',
      role: 'assistant',
      content: 'Custom classes',
      ts: Date.now(),
    }

    const { container } = render(
      <Message
        message={message}
        classNames={{
          row: 'custom-row',
          bubble: 'custom-bubble',
        }}
      />
    )

    expect(container.querySelector('.custom-row')).toBeInTheDocument()
    expect(container.querySelector('.custom-bubble')).toBeInTheDocument()
  })
})
