/**
 * Unit tests for StepPersonality component
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { StepPersonality } from '../StepPersonality'

const defaultProps = {
  systemPrompt: '',
  welcomeMessage: '',
  onSystemPromptChange: vi.fn(),
  onWelcomeMessageChange: vi.fn(),
}

describe('StepPersonality', () => {
  it('renders the step title and description', () => {
    render(<StepPersonality {...defaultProps} />)

    expect(screen.getByText('Kişiliğini Belirleyin')).toBeInTheDocument()
    expect(
      screen.getByText('Botunuzun nasıl davranacağını ve konuşacağını ayarlayın'),
    ).toBeInTheDocument()
  })

  it('renders system prompt field with label', () => {
    render(<StepPersonality {...defaultProps} />)

    expect(screen.getByLabelText('Sistem Talimatı')).toBeInTheDocument()
  })

  it('renders welcome message field with label', () => {
    render(<StepPersonality {...defaultProps} />)

    expect(screen.getByLabelText('Karşılama Mesajı')).toBeInTheDocument()
  })

  it('displays current systemPrompt value', () => {
    render(<StepPersonality {...defaultProps} systemPrompt="Test prompt" />)

    expect(screen.getByDisplayValue('Test prompt')).toBeInTheDocument()
  })

  it('displays current welcomeMessage value', () => {
    render(<StepPersonality {...defaultProps} welcomeMessage="Hello!" />)

    expect(screen.getByDisplayValue('Hello!')).toBeInTheDocument()
  })

  it('calls onSystemPromptChange when typing in system prompt', () => {
    const onChange = vi.fn()
    render(<StepPersonality {...defaultProps} onSystemPromptChange={onChange} />)

    fireEvent.change(screen.getByLabelText('Sistem Talimatı'), {
      target: { value: 'New prompt' },
    })

    expect(onChange).toHaveBeenCalledWith('New prompt')
  })

  it('calls onWelcomeMessageChange when typing in welcome message', () => {
    const onChange = vi.fn()
    render(<StepPersonality {...defaultProps} onWelcomeMessageChange={onChange} />)

    fireEvent.change(screen.getByLabelText('Karşılama Mesajı'), {
      target: { value: 'New message' },
    })

    expect(onChange).toHaveBeenCalledWith('New message')
  })
})
