import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import OverviewPanel from '../OverviewPanel'

describe('OverviewPanel', () => {
  it('renders and allows editing of name and system prompt', () => {
    const setName = vi.fn()
    const setSystemPrompt = vi.fn()

    render(
      <OverviewPanel
        name="Bot"
        setName={setName}
        systemPrompt="Merhaba"
        setSystemPrompt={setSystemPrompt}
      />
    )

    expect(screen.getByText(/Kimlik/i)).toBeInTheDocument()
    const nameInput = screen.getByDisplayValue('Bot') as HTMLInputElement
    fireEvent.change(nameInput, { target: { value: 'Destek Botu' } })
    expect(setName).toHaveBeenCalledWith('Destek Botu')

    const promptTextarea = screen.getByPlaceholderText('Sen yardımcı bir asistansın...') as HTMLTextAreaElement
    fireEvent.change(promptTextarea, { target: { value: 'Yeni sistem mesajı' } })
    expect(setSystemPrompt).toHaveBeenCalledWith('Yeni sistem mesajı')
  })
})
