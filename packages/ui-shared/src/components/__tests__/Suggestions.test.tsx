import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { Suggestions } from '../Suggestions'

describe('Suggestions Component', () => {
  it('renders all suggestions in carousel', () => {
    const items = ['Question 1', 'Question 2', 'Question 3']
    const onPick = vi.fn()

    render(<Suggestions items={items} onPick={onPick} />)

    // Should show first item initially
    expect(screen.getByText('Question 1')).toBeInTheDocument()
    expect(screen.getByText('1 / 3')).toBeInTheDocument()
  })

  it('navigates to next suggestion when next button is clicked', () => {
    const items = ['Q1', 'Q2', 'Q3']
    const onPick = vi.fn()

    render(<Suggestions items={items} onPick={onPick} />)

    const nextBtn = screen.getByLabelText('Sonraki soru')
    fireEvent.click(nextBtn)

    expect(screen.getByText('Q2')).toBeInTheDocument()
    expect(screen.getByText('2 / 3')).toBeInTheDocument()
  })

  it('navigates to previous suggestion when prev button is clicked', () => {
    const items = ['Q1', 'Q2', 'Q3']
    const onPick = vi.fn()

    render(<Suggestions items={items} onPick={onPick} />)

    const prevBtn = screen.getByLabelText('Önceki soru')
    fireEvent.click(prevBtn)

    // Should wrap around to last item
    expect(screen.getByText('Q3')).toBeInTheDocument()
    expect(screen.getByText('3 / 3')).toBeInTheDocument()
  })

  it('calls onPick when suggestion is clicked', () => {
    const items = ['Test question']
    const onPick = vi.fn()

    render(<Suggestions items={items} onPick={onPick} />)

    const suggestion = screen.getByText('Test question')
    fireEvent.click(suggestion)

    expect(onPick).toHaveBeenCalledWith('Test question')
  })

  it('does not show navigation buttons when only one suggestion', () => {
    const items = ['Single question']
    const onPick = vi.fn()

    render(<Suggestions items={items} onPick={onPick} />)

    expect(screen.queryByLabelText('Sonraki soru')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Önceki soru')).not.toBeInTheDocument()
    expect(screen.queryByText('1 / 1')).not.toBeInTheDocument()
  })

  it('disables buttons when disabled prop is true', () => {
    const items = ['Q1', 'Q2']
    const onPick = vi.fn()

    render(<Suggestions items={items} disabled={true} onPick={onPick} />)

    const suggestion = screen.getByText('Q1')
    const nextBtn = screen.getByLabelText('Sonraki soru')
    const prevBtn = screen.getByLabelText('Önceki soru')

    expect(suggestion).toBeDisabled()
    expect(nextBtn).toBeDisabled()
    expect(prevBtn).toBeDisabled()
  })

  it('returns null when items array is empty', () => {
    const onPick = vi.fn()

    const { container } = render(<Suggestions items={[]} onPick={onPick} />)

    expect(container.firstChild).toBeNull()
  })

  it('wraps around to first item after last item', () => {
    const items = ['Q1', 'Q2', 'Q3']
    const onPick = vi.fn()

    render(<Suggestions items={items} onPick={onPick} />)

    const nextBtn = screen.getByLabelText('Sonraki soru')
    
    // Navigate to last item
    fireEvent.click(nextBtn) // Q2
    fireEvent.click(nextBtn) // Q3
    fireEvent.click(nextBtn) // Should wrap to Q1

    expect(screen.getByText('Q1')).toBeInTheDocument()
    expect(screen.getByText('1 / 3')).toBeInTheDocument()
  })

  it('applies custom class names', () => {
    const items = ['Test']
    const onPick = vi.fn()

    const { container } = render(
      <Suggestions
        items={items}
        onPick={onPick}
        classNames={{
          container: 'custom-container',
          header: 'custom-header',
        }}
      />
    )

    expect(container.querySelector('.custom-container')).toBeInTheDocument()
    expect(container.querySelector('.custom-header')).toBeInTheDocument()
  })
})
