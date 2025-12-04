import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/preact'
import { Suggestions } from '../Suggestions'

describe('Suggestions', () => {
  it('renders items and fires onPick', async () => {
    const onPick = vi.fn()
    render(<Suggestions items={["A", "B"]} disabled={false} onPick={onPick} />)
    const btn = await screen.findByRole('button', { name: 'A' })
    fireEvent.click(btn)
    expect(onPick).toHaveBeenCalledWith('A')
  })
  it('does not render when items empty', () => {
    const { container } = render(<Suggestions items={[]} disabled={false} onPick={() => {}} />)
    expect(container.textContent).toBe('')
  })
})
