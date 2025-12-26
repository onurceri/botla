/**
 * Unit tests for StepComplete component
 */

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StepComplete } from '../StepComplete'

describe('StepComplete', () => {
  it('renders the congratulations message', () => {
    render(<StepComplete botName="Test Bot" />)

    expect(screen.getByRole('heading', { name: /Tebrikler/i })).toBeInTheDocument()
  })

  it('displays the bot name in success message', () => {
    render(<StepComplete botName="My Awesome Bot" />)

    expect(screen.getByText('My Awesome Bot')).toBeInTheDocument()
    expect(screen.getByText(/kullanıma hazır/i)).toBeInTheDocument()
  })

  it('renders the next steps section', () => {
    render(<StepComplete botName="Test Bot" />)

    expect(screen.getByText('Şimdi Yapabilecekleriniz')).toBeInTheDocument()
  })

  it('renders all three next step items', () => {
    render(<StepComplete botName="Test Bot" />)

    expect(screen.getByText(/Test Alanı'nda botunuzu test edin/)).toBeInTheDocument()
    expect(screen.getByText(/bilgi tabanını genişletin/)).toBeInTheDocument()
    expect(screen.getByText(/embed kodu ile entegre edin/)).toBeInTheDocument()
  })

  it('displays step numbers 1, 2, 3', () => {
    render(<StepComplete botName="Test Bot" />)

    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })
})
