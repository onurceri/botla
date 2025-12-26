/**
 * Navigation buttons for the onboarding wizard (Back, Next, Finish).
 */

import { ArrowLeft, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { StepNumber } from '../types'

interface NavigationButtonsProps {
  currentStep: StepNumber
  isLoading: boolean
  canProceed: boolean
  onNext: () => void
  onBack: () => void
  onFinish: () => void
}

export function NavigationButtons({
  currentStep,
  isLoading,
  canProceed,
  onNext,
  onBack,
  onFinish,
}: NavigationButtonsProps) {
  const showBackButton = currentStep > 1 && currentStep < 4

  return (
    <div className="flex items-center justify-between mt-8 pt-6 border-t border-border/50">
      {/* Back Button */}
      {showBackButton ? (
        <Button variant="ghost" onClick={onBack} className="gap-2">
          <ArrowLeft className="w-4 h-4" />
          Geri
        </Button>
      ) : (
        <div />
      )}

      {/* Next / Finish Button */}
      {currentStep < 4 ? (
        <Button
          onClick={onNext}
          isLoading={isLoading}
          disabled={!canProceed}
          className="gap-2 bg-primary hover:bg-primary/90 shadow-lg shadow-primary/25"
        >
          {currentStep === 3 ? 'Botu Oluştur' : 'İleri'}
          {!isLoading && <ArrowRight className="w-4 h-4" />}
        </Button>
      ) : (
        <Button
          onClick={onFinish}
          className="w-full gap-2 bg-primary hover:bg-primary/90 shadow-lg shadow-primary/25"
        >
          Botu Görüntüle
          <ArrowRight className="w-4 h-4" />
        </Button>
      )}
    </div>
  )
}
