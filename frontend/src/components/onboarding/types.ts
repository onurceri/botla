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
  skipDataSource: boolean
}

/** Actions available for wizard navigation and form updates */
export interface WizardActions {
  // Navigation
  goToNextStep: () => Promise<void>
  goToPreviousStep: () => void
  finish: () => void
  skip: () => Promise<void>
  skipStep2: () => void

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

/** Plan limits the wizard reads (subset of GET /api/v1/me/plan `limits`) */
export interface WizardPlanLimits {
  max_chatbots?: number
  max_monthly_ingestions?: number
}

/** Plan features the wizard reads (subset of GET /api/v1/me/plan `features`) */
export interface WizardPlanFeatures {
  files?: { max_size_mb?: number; max_text_length?: number; max_files_per_bot?: number }
  scraping?: { max_urls_per_bot?: number }
}

/** Plan information exposed by useOnboardingWizard alongside the actions */
export interface WizardPlanInfo {
  planCode: string
  planLimits: WizardPlanLimits
  planFeatures: WizardPlanFeatures
}

/** Default values for the wizard */
export const DEFAULT_SYSTEM_PROMPT =
  'Sen yardımsever ve samimi bir müşteri destek asistanısın. Kısa ve öz cevaplar ver.'

export const DEFAULT_WELCOME_MESSAGE = 'Merhaba! Size nasıl yardımcı olabilirim?'

/** Text source limits per plan (baseline minimum) */
export const MIN_TEXT_LENGTH = 50

