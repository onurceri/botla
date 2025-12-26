import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { LoadingIndicator } from '../LoadingIndicator'

describe('LoadingIndicator Component', () => {
  it('renders loading dots', () => {
    const { container } = render(<LoadingIndicator />)

    const dots = container.querySelectorAll('.cbw-loading-dot')
    expect(dots).toHaveLength(3)
  })

  it('renders default bot icon when no custom icon is provided', () => {
    const { container } = render(<LoadingIndicator />)

    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('renders custom bot icon when provided', () => {
    render(<LoadingIndicator botIcon="https://example.com/bot.png" />)

    const icon = screen.getByAltText('')
    expect(icon).toHaveAttribute('src', 'https://example.com/bot.png')
    expect(icon).toHaveClass('cbw-avatar-img')
  })

  it('applies custom class names', () => {
    const { container } = render(
      <LoadingIndicator
        classNames={{
          row: 'custom-row',
          avatar: 'custom-avatar',
          bubble: 'custom-bubble',
          dot: 'custom-dot',
        }}
      />
    )

    expect(container.querySelector('.custom-row')).toBeInTheDocument()
    expect(container.querySelector('.custom-avatar')).toBeInTheDocument()
    expect(container.querySelector('.custom-bubble')).toBeInTheDocument()
    expect(container.querySelector('.custom-dot')).toBeInTheDocument()
  })

  it('renders within cbw-loading-row container', () => {
    const { container } = render(<LoadingIndicator />)

    expect(container.querySelector('.cbw-loading-row')).toBeInTheDocument()
  })

  it('includes avatar component', () => {
    const { container } = render(<LoadingIndicator />)

    expect(container.querySelector('.cbw-avatar')).toBeInTheDocument()
  })

  it('includes loading bubble component', () => {
    const { container } = render(<LoadingIndicator />)

    expect(container.querySelector('.cbw-loading-bubble')).toBeInTheDocument()
  })
})
