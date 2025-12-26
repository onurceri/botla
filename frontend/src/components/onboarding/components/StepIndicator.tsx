/**
 * Step progress indicator showing current step and navigation status.
 */

import { Bot, Upload, Palette, Rocket, CheckCircle2 } from 'lucide-react'
import type { StepNumber, StepDefinition } from '../types'

interface StepIndicatorProps {
  currentStep: StepNumber
  onSkip: () => void
}

const STEPS: StepDefinition[] = [
  {
    id: 1,
    title: 'Botunuzu Adlandırın',
    subtitle: 'Chatbotunuza benzersiz bir isim verin',
    icon: Bot,
  },
  {
    id: 2,
    title: 'Bilgi Kaynağı Ekleyin',
    subtitle: 'Botunuzun öğreneceği içeriği yükleyin',
    icon: Upload,
  },
  {
    id: 3,
    title: 'Kişiliğini Belirleyin',
    subtitle: 'Botunuzun nasıl konuşacağını ayarlayın',
    icon: Palette,
  },
  {
    id: 4,
    title: 'Hazır!',
    subtitle: 'Botunuz kullanıma hazır',
    icon: Rocket,
  },
]

export function StepIndicator({ currentStep, onSkip }: StepIndicatorProps) {
  return (
    <div className="mb-8">
      {/* Skip Button */}
      <div className="flex justify-end mb-4">
        <button
          onClick={onSkip}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          Atla
        </button>
      </div>

      {/* Progress Steps */}
      <div className="flex items-center justify-between mb-4">
        {STEPS.map((step, index) => (
          <div key={step.id} className="flex items-center">
            <div
              className={`w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300
                ${
                  currentStep >= step.id
                    ? 'bg-primary text-white'
                    : 'bg-muted text-muted-foreground'
                }
                ${currentStep === step.id ? 'ring-4 ring-primary/20' : ''}
              `}
            >
              {currentStep > step.id ? (
                <CheckCircle2 className="w-5 h-5" />
              ) : (
                <step.icon className="w-5 h-5" />
              )}
            </div>
            {index < STEPS.length - 1 && (
              <div
                className={`w-16 sm:w-24 h-1 mx-2 rounded-full transition-colors duration-300
                ${currentStep > step.id ? 'bg-primary' : 'bg-muted'}
              `}
              />
            )}
          </div>
        ))}
      </div>

      {/* Current Step Label */}
      <div className="text-center">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          Adım {currentStep} / {STEPS.length}
        </p>
      </div>
    </div>
  )
}
