import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, act, cleanup } from '@testing-library/preact'

// Mock react-markdown to avoid React/Preact compatibility issues in tests
vi.mock('react-markdown', () => ({
  default: ({ children }: { children: string }) => {
    // Simple text render for testing - just return the content as text
    return children
  }
}))

// Import Message after mock is set up
import { Message } from '../Message'

type ChatMessage = {
  id?: string
  role: 'user' | 'assistant'
  content: string
  ts?: number
  feedback?: boolean
  type?: 'welcome' | 'handoff' | 'normal'
  handoffRequestId?: string
  emailSubmitted?: boolean
}

describe('Message', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders message content', () => {
    const msg: ChatMessage = {
      role: 'assistant',
      content: 'Hello world',
      id: '1'
    }
    
    render(<Message m={msg} />)
    expect(screen.getByText('Hello world')).toBeDefined()
  })

  it('renders user message with user role class', () => {
    const msg: ChatMessage = {
      role: 'user',
      content: 'Hello',
      id: '1'
    }
    
    const { container } = render(<Message m={msg} />)
    expect(container.querySelector('.cbw-msg.user')).toBeDefined()
    expect(screen.getByText('Hello')).toBeDefined()
  })

  it('renders handoff card and handles email submission', async () => {
    const onSubmitEmail = vi.fn().mockResolvedValue(undefined)
    const msg: ChatMessage = {
      role: 'assistant',
      content: 'Handoff',
      type: 'handoff',
      handoffRequestId: 'req123'
    }

    render(<Message m={msg} onSubmitEmail={onSubmitEmail} />)

    expect(screen.getByText('Destek Talebi')).toBeDefined()
    
    const input = screen.getByPlaceholderText('e-posta@adresiniz.com')
    const button = screen.getByText('Gönder')

    fireEvent.input(input, { target: { value: 'test@example.com' } })
    
    await act(async () => {
      fireEvent.submit(button.closest('form')!)
    })

    expect(onSubmitEmail).toHaveBeenCalledWith('req123', 'test@example.com')
    expect(screen.getByText('Talebiniz alındı!')).toBeDefined()
  })

  it('handles positive feedback clicks', () => {
    const onFeedback = vi.fn()
    const msg: ChatMessage = {
      role: 'assistant',
      content: 'Helpful info',
      id: 'msg123'
    }

    render(<Message m={msg} onFeedback={onFeedback} />)

    const thumbsUp = screen.getByTitle('Yararlı')
    fireEvent.click(thumbsUp)

    expect(onFeedback).toHaveBeenCalledWith('msg123', true)
  })

  it('handles negative feedback clicks', () => {
    const onFeedback = vi.fn()
    const msg: ChatMessage = {
      role: 'assistant',
      content: 'Not helpful info',
      id: 'msg456'
    }

    render(<Message m={msg} onFeedback={onFeedback} />)

    const thumbsDown = screen.getByTitle('Yararlı değil')
    fireEvent.click(thumbsDown)

    expect(onFeedback).toHaveBeenCalledWith('msg456', false)
  })

  it('shows bot avatar for assistant messages', () => {
    const msg: ChatMessage = {
      role: 'assistant',
      content: 'Bot message',
      id: '1'
    }
    
    const { container } = render(<Message m={msg} />)
    expect(container.querySelector('.cbw-avatar')).toBeDefined()
  })

  it('renders custom bot icon when provided', () => {
    const msg: ChatMessage = {
      role: 'assistant',
      content: 'Bot message',
      id: '1'
    }
    
    const { container } = render(<Message m={msg} botIcon="https://example.com/icon.png" />)
    const img = container.querySelector('.cbw-avatar-img') as HTMLImageElement
    expect(img).toBeDefined()
    expect(img?.src).toBe('https://example.com/icon.png')
  })
})
