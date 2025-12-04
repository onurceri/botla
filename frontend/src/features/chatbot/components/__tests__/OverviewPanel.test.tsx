import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import OverviewPanel from '../OverviewPanel'

describe('OverviewPanel', () => {
  it('renders and allows editing of fields', () => {
    const setName = vi.fn()
    const setModel = vi.fn()
    const setSystemPrompt = vi.fn()
    const setTemperature = vi.fn()
    const setMaxTokens = vi.fn()

    render(
      <OverviewPanel
        name="Bot"
        setName={setName}
        model="gpt-3.5-turbo"
        setModel={setModel}
        systemPrompt="Merhaba"
        setSystemPrompt={setSystemPrompt}
        temperature={0.7}
        setTemperature={setTemperature}
        maxTokens={512}
        setMaxTokens={setMaxTokens}
      />
    )

    expect(screen.getByText(/Kimlik & Model/i)).toBeInTheDocument()
    expect(screen.getByText(/Model Ayarları/i)).toBeInTheDocument()
    const nameInput = screen.getByDisplayValue('Bot') as HTMLInputElement
    fireEvent.change(nameInput, { target: { value: 'Destek Botu' } })
    expect(setName).toHaveBeenCalledWith('Destek Botu')

    const modelSelect = screen.getByRole('combobox') as HTMLSelectElement
    fireEvent.change(modelSelect, { target: { value: 'gpt-4' } })
    expect(setModel).toHaveBeenCalledWith('gpt-4')

    const promptTextarea = screen.getByPlaceholderText('Sen yardımcı bir asistansın...') as HTMLTextAreaElement
    fireEvent.change(promptTextarea, { target: { value: 'Yeni sistem mesajı' } })
    expect(setSystemPrompt).toHaveBeenCalledWith('Yeni sistem mesajı')

    const tempRange = screen.getByRole('slider') as HTMLInputElement
    fireEvent.change(tempRange, { target: { value: '0.9' } })
    expect(setTemperature).toHaveBeenCalledWith(0.9)
    const tokenInput = screen.getByLabelText(/Maksimum Token/i) as HTMLInputElement
    expect(tokenInput.value).toBe('512')
    fireEvent.change(tokenInput, { target: { value: '1024' } })
    expect(setMaxTokens).toHaveBeenCalledWith(1024)
  })
})
