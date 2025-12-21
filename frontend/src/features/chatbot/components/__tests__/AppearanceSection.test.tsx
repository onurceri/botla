import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within } from '@testing-library/react'
import AppearanceSection from '../AppearanceSection'

describe('AppearanceSection', () => {
  it('updates position', () => {
    const onToggle = vi.fn()
    const setPosition = vi.fn()
    render(
      <AppearanceSection
        isExpanded={true}
        onToggle={onToggle}
        position="bottom-right"
        setPosition={setPosition}
      />
    )
    const posSelect = screen.getByLabelText('Konum') as HTMLSelectElement
    fireEvent.change(posSelect, { target: { value: 'bottom-left' } })
    expect(setPosition).toHaveBeenCalledWith('bottom-left')
  })
})
