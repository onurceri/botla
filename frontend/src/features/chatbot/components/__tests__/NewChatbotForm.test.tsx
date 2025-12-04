import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import NewChatbotForm from '../NewChatbotForm'

describe('NewChatbotForm', () => {
  it('renders fields and triggers change handlers', () => {
    const onName = vi.fn()
    const onDesc = vi.fn()
    render(<NewChatbotForm name="" description="" onNameChange={onName} onDescriptionChange={onDesc} />)
    const nameInput = screen.getByPlaceholderText('Örn: Müşteri Temsilcisi') as HTMLInputElement
    const descInput = screen.getByPlaceholderText('Botun amacı nedir?') as HTMLInputElement
    fireEvent.change(nameInput, { target: { value: 'Destek' } })
    fireEvent.change(descInput, { target: { value: 'Müşteri destek' } })
    expect(onName).toHaveBeenCalledWith('Destek')
    expect(onDesc).toHaveBeenCalledWith('Müşteri destek')
  })
})

