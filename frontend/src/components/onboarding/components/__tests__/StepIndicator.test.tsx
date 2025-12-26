/**
 * Unit tests for StepIndicator component
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { StepIndicator } from '../StepIndicator'

describe('StepIndicator', () => {
  it('renders skip button', () => {
    render(<StepIndicator currentStep={1} onSkip={() => {}} />)

    expect(screen.getByText('Atla')).toBeInTheDocument()
  })

  it('calls onSkip when clicking skip button', () => {
    const onSkip = vi.fn()
    render(<StepIndicator currentStep={1} onSkip={onSkip} />)

    fireEvent.click(screen.getByText('Atla'))

    expect(onSkip).toHaveBeenCalled()
  })

  it('shows current step counter', () => {
    render(<StepIndicator currentStep={2} onSkip={() => {}} />)

    expect(screen.getByText('Adım 2 / 4')).toBeInTheDocument()
  })

  it('renders all 4 step indicators', () => {
    render(<StepIndicator currentStep={1} onSkip={() => {}} />)

    // Each step should have an icon container
    const stepCircles = document.querySelectorAll('.rounded-full.w-10.h-10')
    expect(stepCircles.length).toBe(4)
  })

  it('highlights current and completed steps', () => {
    render(<StepIndicator currentStep={2} onSkip={() => {}} />)

    // Current step should have a ring
    const steps = document.querySelectorAll('.rounded-full.w-10.h-10')
    
    // Step 2 (index 1) should have the ring class
    expect(steps[1].className).toContain('ring-4')
  })
})
