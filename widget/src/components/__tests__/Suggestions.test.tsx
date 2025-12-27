import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/preact'
import { Suggestions } from '../Suggestions'

describe('Suggestions', () => {
  afterEach(() => {
    cleanup()
  })

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

  it('navigates through items with carousel buttons', async () => {
    const onPick = vi.fn()
    const items = ["First Question", "Second Question", "Third Question"]
    render(<Suggestions items={items} disabled={false} onPick={onPick} />)

    // Initially shows the first item
    expect(screen.getByText("First Question")).toBeDefined()
    expect(screen.getByText("1 / 3")).toBeDefined()

    // Click next
    const nextBtn = screen.getByLabelText('Sonraki soru')
    fireEvent.click(nextBtn)

    // Should show the second item
    expect(screen.getByText("Second Question")).toBeDefined()
    expect(screen.getByText("2 / 3")).toBeDefined()

    // Click next again
    fireEvent.click(nextBtn)

    // Should show the third item
    expect(screen.getByText("Third Question")).toBeDefined()
    expect(screen.getByText("3 / 3")).toBeDefined()

    // Click next again (should wrap to first)
    fireEvent.click(nextBtn)
    expect(screen.getByText("First Question")).toBeDefined()
    expect(screen.getByText("1 / 3")).toBeDefined()

    // Click prev (should wrap to last)
    const prevBtn = screen.getByLabelText('Önceki soru')
    fireEvent.click(prevBtn)
    expect(screen.getByText("Third Question")).toBeDefined()
    expect(screen.getByText("3 / 3")).toBeDefined()
  })

  it('disables buttons when disabled prop is true', () => {
    const items = ["A", "B"]
    render(<Suggestions items={items} disabled={true} onPick={() => {}} />)

    expect((screen.getByLabelText('A') as HTMLButtonElement).disabled).toBe(true)
    expect((screen.getByLabelText('Sonraki soru') as HTMLButtonElement).disabled).toBe(true)
    expect((screen.getByLabelText('Önceki soru') as HTMLButtonElement).disabled).toBe(true)
  })

  it('does not show carousel buttons for single item', () => {
    render(<Suggestions items={["Single"]} disabled={false} onPick={() => {}} />)

    expect(screen.getByText("Single")).toBeDefined()
    expect(screen.queryByLabelText('Sonraki soru')).toBeNull()
    expect(screen.queryByLabelText('Önceki soru')).toBeNull()
    expect(screen.queryByText(/1 \/ 1/)).toBeNull()
  })
})
