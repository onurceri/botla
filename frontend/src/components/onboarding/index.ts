/**
 * Onboarding module exports
 */

// Main wizard component
export { default as OnboardingWizard } from './OnboardingWizard'

// Types for external use
export type { StepNumber, SourceType, OnboardingFormState, WizardState } from './types'

// Hook for advanced customization
export { useOnboardingWizard } from './hooks/useOnboardingWizard'

// Step components (for unit testing or custom wizards)
export * from './steps'

// UI components (for composition)
export * from './components'
