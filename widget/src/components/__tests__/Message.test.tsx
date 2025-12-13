import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/preact'
import { Message } from '../Message'

describe('Message', () => {
  it('renders markdown content correctly', () => {
    const msg = {
      role: 'assistant',
      content: '**Bold** and *Italic*',
      id: '1'
    } as const
    
    render(<Message m={msg} />)
    
    // Check if Strong element is rendered
    const strong = screen.getByText('Bold')
    expect(strong.tagName).toBe('STRONG')
    
    // Check if Em element is rendered
    const em = screen.getByText('Italic')
    expect(em.tagName).toBe('EM')
  })

  it('renders links correctly', () => {
    const msg = {
      role: 'assistant',
      content: '[Link](https://example.com)',
      id: '2'
    } as const
    
    render(<Message m={msg} />)
    
    const link = screen.getByRole('link', { name: 'Link' })
    expect(link).toBeDefined()
    expect(link.getAttribute('href')).toBe('https://example.com')
  })

  it('renders lists correctly', () => {
     const msg = {
      role: 'assistant',
      content: '- Item 1\n- Item 2',
      id: '3'
    } as const

    render(<Message m={msg} />)
    
    expect(screen.getByText('Item 1')).toBeDefined()
    expect(screen.getByText('Item 2')).toBeDefined()
    // List items are usually li
    const items = screen.getAllByRole('listitem')
    expect(items.length).toBe(2)
  })
})
