/**
 * Unit tests for StepBotName component
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { StepBotName } from '../StepBotName'

describe('StepBotName', () => {
  it('renders the step title and description', () => {
    render(<StepBotName botName="" onBotNameChange={() => {}} />)

    expect(screen.getByText('Botunuza İsim Verin')).toBeInTheDocument()
    expect(screen.getByText("Bu isim dashboard'da ve widget'ta görünecektir")).toBeInTheDocument()
  })

  it('renders input with correct label', () => {
    render(<StepBotName botName="" onBotNameChange={() => {}} />)

    expect(screen.getByLabelText('Bot Adı')).toBeInTheDocument()
  })

  it('displays current botName value', () => {
    render(<StepBotName botName="Test Bot" onBotNameChange={() => {}} />)

    expect(screen.getByDisplayValue('Test Bot')).toBeInTheDocument()
  })

  it('calls onBotNameChange when input changes', () => {
    const onChange = vi.fn()
    render(<StepBotName botName="" onBotNameChange={onChange} />)

    fireEvent.change(screen.getByLabelText('Bot Adı'), { target: { value: 'New Bot' } })

    expect(onChange).toHaveBeenCalledWith('New Bot')
  })

  it('shows minimum character hint', () => {
    render(<StepBotName botName="" onBotNameChange={() => {}} />)

    expect(screen.getByText('Minimum 2 karakter')).toBeInTheDocument()
  })
})
