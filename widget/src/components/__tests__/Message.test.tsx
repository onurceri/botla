import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, act, cleanup } from '@testing-library/preact'
import { Message } from '@botla/ui-shared'
import type { ChatMessage } from '@botla/ui-shared'

describe('Message', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders markdown content correctly', () => {
    const msg: ChatMessage = {
      role: 'assistant',
      content: '**Bold** and *Italic*',
      id: '1'
    }
    
    render(<Message message={msg} />)
    
    const strong = screen.getByText('Bold')
    expect(strong.tagName).toBe('STRONG')
    
    const em = screen.getByText('Italic')
    expect(em.tagName).toBe('EM')
  })

  it('renders user message differently', () => {
    const msg: ChatMessage = {
      role: 'user',
      content: 'Hello',
      id: '1'
    }
    
    const { container } = render(<Message message={msg} />)
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

    render(<Message message={msg} onSubmitEmail={onSubmitEmail} />)

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

  it('handles feedback clicks', () => {
    const onFeedback = vi.fn()
    const msg: ChatMessage = {
      role: 'assistant',
      content: 'Helpful info',
      id: 'msg123'
    }

    render(<Message message={msg} onFeedback={onFeedback} />)

    const thumbsUp = screen.getByTitle('Yararlı')
    fireEvent.click(thumbsUp)

    expect(onFeedback).toHaveBeenCalledWith('msg123', true)
  })
})
