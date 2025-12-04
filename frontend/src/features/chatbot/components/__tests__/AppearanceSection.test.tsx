import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AppearanceSection from '../AppearanceSection'

describe('AppearanceSection', () => {
  it('updates position and theme color', () => {
    const onToggle = vi.fn()
    const setPosition = vi.fn()
    const setFont = vi.fn()
    const setTheme = vi.fn()
    render(
      <AppearanceSection
        isExpanded={true}
        onToggle={onToggle}
        position="bottom-right"
        setPosition={setPosition}
        chatFontFamily="Inter, sans-serif"
        setChatFontFamily={setFont}
        themeColor="#a78bfa"
        setThemeColor={setTheme}
      />
    )
    const posSelect = screen.getByLabelText('Konum') as HTMLSelectElement
    fireEvent.change(posSelect, { target: { value: 'bottom-left' } })
    expect(setPosition).toHaveBeenCalledWith('bottom-left')
    const colorInput = screen.getByLabelText('Ana Renk (Theme)') as HTMLInputElement
    fireEvent.change(colorInput, { target: { value: '#ffffff' } })
    expect(setTheme).toHaveBeenCalledWith('#ffffff')
  })

  it('updates font family', () => {
    const setFont = vi.fn()
    const utils = render(
      <AppearanceSection
        isExpanded={true}
        onToggle={() => {}}
        position="bottom-right"
        setPosition={() => {}}
        chatFontFamily="Inter, sans-serif"
        setChatFontFamily={setFont}
        themeColor="#a78bfa"
        setThemeColor={() => {}}
      />
    )
    const view = within(utils.container)
    const fontSelect = view.getAllByRole('combobox')[1] as HTMLSelectElement
    fireEvent.change(fontSelect, { target: { value: 'Roboto, sans-serif' } })
    expect(setFont).toHaveBeenCalledWith('Roboto, sans-serif')
  })
})
