/**
 * Unit tests for NavigationButtons component
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { NavigationButtons } from '../NavigationButtons'

const defaultProps = {
  currentStep: 1 as const,
  isLoading: false,
  canProceed: true,
  onNext: vi.fn(),
  onBack: vi.fn(),
  onFinish: vi.fn(),
}

describe('NavigationButtons', () => {
  describe('step 1', () => {
    it('does not show back button on first step', () => {
      render(<NavigationButtons {...defaultProps} currentStep={1} />)

      expect(screen.queryByText('Geri')).not.toBeInTheDocument()
    })

    it('shows next button with "İleri" text', () => {
      render(<NavigationButtons {...defaultProps} currentStep={1} />)

      expect(screen.getByText('İleri')).toBeInTheDocument()
    })

    it('calls onNext when clicking next button', () => {
      const onNext = vi.fn()
      render(<NavigationButtons {...defaultProps} currentStep={1} onNext={onNext} />)

      fireEvent.click(screen.getByText('İleri'))

      expect(onNext).toHaveBeenCalled()
    })
  })

  describe('step 2', () => {
    it('shows back button on step 2', () => {
      render(<NavigationButtons {...defaultProps} currentStep={2} />)

      expect(screen.getByText('Geri')).toBeInTheDocument()
    })

    it('calls onBack when clicking back button', () => {
      const onBack = vi.fn()
      render(<NavigationButtons {...defaultProps} currentStep={2} onBack={onBack} />)

      fireEvent.click(screen.getByText('Geri'))

      expect(onBack).toHaveBeenCalled()
    })
  })

  describe('step 3', () => {
    it('shows "Botu Oluştur" text on step 3', () => {
      render(<NavigationButtons {...defaultProps} currentStep={3} />)

      expect(screen.getByText('Botu Oluştur')).toBeInTheDocument()
    })
  })

  describe('step 4 (complete)', () => {
    it('shows finish button on step 4', () => {
      render(<NavigationButtons {...defaultProps} currentStep={4} />)

      expect(screen.getByText('Botu Görüntüle')).toBeInTheDocument()
    })

    it('does not show back button on step 4', () => {
      render(<NavigationButtons {...defaultProps} currentStep={4} />)

      expect(screen.queryByText('Geri')).not.toBeInTheDocument()
    })

    it('calls onFinish when clicking finish button', () => {
      const onFinish = vi.fn()
      render(<NavigationButtons {...defaultProps} currentStep={4} onFinish={onFinish} />)

      fireEvent.click(screen.getByText('Botu Görüntüle'))

      expect(onFinish).toHaveBeenCalled()
    })
  })

  describe('disabled state', () => {
    it('disables next button when canProceed is false', () => {
      render(<NavigationButtons {...defaultProps} currentStep={1} canProceed={false} />)

      expect(screen.getByText('İleri').closest('button')).toBeDisabled()
    })

    it('enables next button when canProceed is true', () => {
      render(<NavigationButtons {...defaultProps} currentStep={1} canProceed={true} />)

      expect(screen.getByText('İleri').closest('button')).not.toBeDisabled()
    })
  })
})
