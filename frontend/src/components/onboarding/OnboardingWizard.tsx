/**
 * Onboarding Wizard - Main Orchestrator Component
 *
 * This component manages the multi-step onboarding flow for creating a new chatbot.
 * It uses a custom hook for state management and delegates rendering to step components.
 *
 * Architecture:
 * - State management: useOnboardingWizard hook
 * - Navigation UI: StepIndicator, NavigationButtons
 * - Step content: StepBotName, StepDataSource, StepPersonality, StepComplete
 */

import { useOnboardingWizard } from './hooks/useOnboardingWizard'
import { StepIndicator, NavigationButtons } from './components'
import {
  StepBotName,
  StepDataSource,
  StepPersonality,
  StepComplete,
} from './steps'

/**
 * Renders the current step's content based on wizard state.
 */
function StepContent({
  state,
  actions,
}: {
  state: ReturnType<typeof useOnboardingWizard>[0]
  actions: ReturnType<typeof useOnboardingWizard>[1]
}) {
  switch (state.currentStep) {
    case 1:
      return (
        <StepBotName
          botName={state.botName}
          onBotNameChange={actions.setBotName}
        />
      )
    case 2:
      return (
        <StepDataSource
          sourceType={state.sourceType}
          textContent={state.textContent}
          urlContent={state.urlContent}
          pdfFile={state.pdfFile}
          planData={actions.planFeatures}
          planCode={actions.planLimits?.code || 'free'}
          onSourceTypeChange={actions.setSourceType}
          onTextContentChange={actions.setTextContent}
          onUrlContentChange={actions.setUrlContent}
          onFileSelect={actions.handleFileSelect}
          onFileRemove={() => actions.setPdfFile(null)}
          onSkipStep={actions.skipStep2}
        />
      )
    case 3:
      return (
        <StepPersonality
          systemPrompt={state.systemPrompt}
          welcomeMessage={state.welcomeMessage}
          onSystemPromptChange={actions.setSystemPrompt}
          onWelcomeMessageChange={actions.setWelcomeMessage}
        />
      )
    case 4:
      return <StepComplete botName={state.botName} />
    default:
      return null
  }
}

/**
 * Main onboarding wizard component.
 * Guides users through creating their first chatbot.
 */
const OnboardingWizard = () => {
  const [state, actions] = useOnboardingWizard()

  return (
    <div className="min-h-screen bg-background relative overflow-hidden flex items-center justify-center p-6">
      {/* Animated Background */}
      <div className="absolute inset-0 gradient-mesh opacity-50" />
      <div className="absolute top-20 left-20 w-72 h-72 bg-primary/10 rounded-full blur-3xl animate-float" />
      <div
        className="absolute bottom-20 right-20 w-96 h-96 bg-accent/30 rounded-full blur-3xl animate-float"
        style={{ animationDelay: '-3s' }}
      />

      <div className="relative z-10 w-full max-w-xl">
        {/* Progress Header */}
        <StepIndicator
          currentStep={state.currentStep}
          onSkip={actions.skip}
        />

        {/* Form Card */}
        <div className="glass-card p-8 lg:p-10">
          <StepContent state={state} actions={actions} />

          {/* Navigation Buttons */}
          <NavigationButtons
            currentStep={state.currentStep}
            isLoading={state.isLoading}
            canProceed={actions.canProceed()}
            onNext={actions.goToNextStep}
            onBack={actions.goToPreviousStep}
            onFinish={actions.finish}
          />
        </div>
      </div>
    </div>
  )
}

export default OnboardingWizard
