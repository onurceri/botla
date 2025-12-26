/**
 * Type definitions for the Onboarding Wizard
 */

import type { LucideIcon } from 'lucide-react'

/** Valid step numbers in the wizard */
export type StepNumber = 1 | 2 | 3 | 4

/** Source type for data input */
export type SourceType = 'text' | 'url' | 'file'

/** Step definition for display */
export interface StepDefinition {
  id: StepNumber
  title: string
  subtitle: string
  icon: LucideIcon
}

/** Core onboarding form state */
export interface OnboardingFormState {
  botName: string
  sourceType: SourceType
  textContent: string
  urlContent: string
  pdfFile: File | null
  systemPrompt: string
  welcomeMessage: string
}

/** Full wizard state including navigation */
export interface WizardState extends OnboardingFormState {
  currentStep: StepNumber
  isLoading: boolean
  createdBotId: string | null
}

/** Actions available for wizard navigation and form updates */
export interface WizardActions {
  // Navigation
  goToNextStep: () => Promise<void>
  goToPreviousStep: () => void
  finish: () => void
  skip: () => Promise<void>

  // Form updates
  setBotName: (name: string) => void
  setSourceType: (type: SourceType) => void
  setTextContent: (content: string) => void
  setUrlContent: (url: string) => void
  setPdfFile: (file: File | null) => void
  setSystemPrompt: (prompt: string) => void
  setWelcomeMessage: (message: string) => void
  handleFileSelect: (e: React.ChangeEvent<HTMLInputElement>) => void

  // Validation
  canProceed: () => boolean
}

/** Default values for the wizard */
export const DEFAULT_SYSTEM_PROMPT =
  'Sen yardımsever ve samimi bir müşteri destek asistanısın. Kısa ve öz cevaplar ver.'

export const DEFAULT_WELCOME_MESSAGE = 'Merhaba! Size nasıl yardımcı olabilirim?'

export const MAX_FILE_SIZE_MB = 10
