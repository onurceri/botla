import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { ChatBubble } from '../ChatBubble'

describe('ChatBubble Component', () => {
  it('renders with default icon when no custom icon is provided', () => {
    const onClick = vi.fn()

    const { container } = render(<ChatBubble color="#007bff" onClick={onClick} />)

    expect(container.querySelector('svg')).toBeInTheDocument()
    expect(screen.getByLabelText('Sohbeti aç')).toBeInTheDocument()
  })

  it('renders with custom icon when provided', () => {
    const onClick = vi.fn()

    render(<ChatBubble color="#007bff" onClick={onClick} icon="https://example.com/icon.png" />)

    const img = screen.getByAltText('')
    expect(img).toHaveAttribute('src', 'https://example.com/icon.png')
    expect(img).toHaveClass('cbw-bubble-icon')
  })

  it('applies background color when no icon is provided', () => {
    const onClick = vi.fn()

    render(<ChatBubble color="#ff0000" onClick={onClick} />)

    const button = screen.getByLabelText('Sohbeti aç')
    expect(button).toHaveStyle({ background: '#ff0000' })
  })

  it('does not apply background color when icon is provided', () => {
    const onClick = vi.fn()

    render(<ChatBubble color="#ff0000" onClick={onClick} icon="https://example.com/icon.png" />)

    const button = screen.getByLabelText('Sohbeti aç')
    expect(button).not.toHaveStyle({ background: '#ff0000' })
  })

  it('shows unread badge when unread count is greater than 0', () => {
    const onClick = vi.fn()

    render(<ChatBubble color="#007bff" onClick={onClick} unread={5} />)

    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getByLabelText('Okunmamış 5')).toBeInTheDocument()
  })

  it('does not show unread badge when unread count is 0', () => {
    const onClick = vi.fn()

    render(<ChatBubble color="#007bff" onClick={onClick} unread={0} />)

    expect(screen.queryByText(/Okunmamış/)).not.toBeInTheDocument()
  })

  it('calls onClick when button is clicked', () => {
    const onClick = vi.fn()

    render(<ChatBubble color="#007bff" onClick={onClick} />)

    const button = screen.getByLabelText('Sohbeti aç')
    fireEvent.click(button)

    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('adds has-icon class when icon is provided', () => {
    const onClick = vi.fn()

    render(<ChatBubble color="#007bff" onClick={onClick} icon="https://example.com/icon.png" />)

    const button = screen.getByLabelText('Sohbeti aç')
    expect(button).toHaveClass('has-icon')
  })

  it('applies custom class names', () => {
    const onClick = vi.fn()

    render(
      <ChatBubble
        color="#007bff"
        onClick={onClick}
        unread={3}
        classNames={{
          button: 'custom-button',
          badge: 'custom-badge',
        }}
      />
    )

    const button = screen.getByLabelText('Sohbeti aç')
    expect(button).toHaveClass('custom-button')

    const badge = screen.getByText('3')
    expect(badge).toHaveClass('custom-badge')
  })
})
