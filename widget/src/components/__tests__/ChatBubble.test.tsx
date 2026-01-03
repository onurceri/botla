/**
 * ChatBubble Component Unit Tests
 */

import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/preact'
import { ChatBubble } from '../ChatBubble'

describe('ChatBubble', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders with default color when no icon', () => {
    const { container } = render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={() => {}} 
      />
    )
    
    const bubble = container.querySelector('.cbw-bubble')
    expect(bubble).not.toBeNull()
    
    // Check that background color is applied inline
    const style = bubble?.getAttribute('style')
    expect(style).toContain('background')
  })

  it('renders with custom icon when provided', () => {
    render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={() => {}}
        icon="https://example.com/bot-icon.png"
      />
    )
    
    const icon = screen.getByRole('img', { hidden: true }) as HTMLImageElement
    expect(icon).not.toBeNull()
    expect(icon.src).toBe('https://example.com/bot-icon.png')
  })

  it('applies has-icon class when icon is provided', () => {
    const { container } = render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={() => {}}
        icon="https://example.com/icon.png"
      />
    )
    
    const bubble = container.querySelector('.cbw-bubble.has-icon')
    expect(bubble).not.toBeNull()
  })

  it('does not apply has-icon class when no icon', () => {
    const { container } = render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={() => {}}
      />
    )
    
    const bubble = container.querySelector('.cbw-bubble.has-icon')
    expect(bubble).toBeNull()
  })

  it('shows unread badge when count > 0', () => {
    render(
      <ChatBubble 
        color="#6366f1" 
        unread={5} 
        onClick={() => {}}
      />
    )
    
    const badge = screen.getByText('5')
    expect(badge).not.toBeNull()
    expect(badge.classList.contains('cbw-badge')).toBe(true)
  })

  it('shows correct unread count for large numbers', () => {
    render(
      <ChatBubble 
        color="#6366f1" 
        unread={99} 
        onClick={() => {}}
      />
    )
    
    const badge = screen.getByText('99')
    expect(badge).not.toBeNull()
  })

  it('does not show badge when unread is 0', () => {
    const { container } = render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={() => {}}
      />
    )
    
    const badge = container.querySelector('.cbw-badge')
    expect(badge).toBeNull()
  })

  it('calls onClick when clicked', () => {
    const onClick = vi.fn()
    render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={onClick}
      />
    )
    
    const button = screen.getByRole('button')
    button.click()
    
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('has correct accessibility label', () => {
    render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={() => {}}
      />
    )
    
    const button = screen.getByRole('button')
    expect(button.getAttribute('aria-label')).toBe('Sohbeti aç')
  })

  it('renders default SVG icon when no icon provided', () => {
    const { container } = render(
      <ChatBubble 
        color="#6366f1" 
        unread={0} 
        onClick={() => {}}
      />
    )
    
    // Default icon should contain chat bubble SVG path
    const svg = container.querySelector('.cbw-bubble svg')
    expect(svg).not.toBeNull()
    expect(svg?.innerHTML).toContain('path')
  })
})
