import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import OverviewPanel from '../OverviewPanel'

describe('OverviewPanel', () => {
  it('renders and allows editing of name and custom instruction', () => {
    const setName = vi.fn()
    const setCustomInstruction = vi.fn()
    const setModel = vi.fn()
    const setTemperature = vi.fn()
    const setMaxTokens = vi.fn()

    render(
      <OverviewPanel
        name="Bot"
        setName={setName}
        customInstruction="Merhaba"
        setCustomInstruction={setCustomInstruction}
        model="openai:gpt-4o"
        setModel={setModel}
        temperature={0.5}
        setTemperature={setTemperature}
        maxTokens={1024}
        setMaxTokens={setMaxTokens}
      />
    )

    expect(screen.getByText(/Kimlik/i)).toBeInTheDocument()
    const nameInput = screen.getByDisplayValue('Bot') as HTMLInputElement
    fireEvent.change(nameInput, { target: { value: 'Destek Botu' } })
    expect(setName).toHaveBeenCalledWith('Destek Botu')

    const promptTextarea = screen.getByPlaceholderText('Botunuza özel davranış kuralları ekleyin...') as HTMLTextAreaElement
    fireEvent.change(promptTextarea, { target: { value: 'Yeni talimat' } })
    expect(setCustomInstruction).toHaveBeenCalledWith('Yeni talimat')
  })
})
