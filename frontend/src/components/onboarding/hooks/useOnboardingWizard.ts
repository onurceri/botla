/**
 * Custom hook for managing onboarding wizard state and logic.
 * Extracts all state management from the OnboardingWizard component.
 */

import { useState, useCallback, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'
import * as onboardingApi from '@/api/onboarding'
import type {
  WizardState,
  WizardActions,
  StepNumber,
  SourceType,
} from '../types'
import {
  DEFAULT_SYSTEM_PROMPT,
  DEFAULT_WELCOME_MESSAGE,
} from '../types'
import { usePlan } from '@/hooks/queries/useProfile'

const initialState: WizardState = {
  currentStep: 1,
  isLoading: false,
  createdBotId: null,
  skipDataSource: false,
  botName: '',
  sourceType: 'text',
  textContent: '',
  urlContent: '',
  pdfFile: null,
  systemPrompt: DEFAULT_SYSTEM_PROMPT,
  welcomeMessage: DEFAULT_WELCOME_MESSAGE,
}

/**
 * Manages all onboarding wizard state, navigation, and API interactions.
 * Returns a tuple of [state, actions] for use in wizard components.
 */
export function useOnboardingWizard(): [WizardState, WizardActions & { planLimits: any }] {
  const navigate = useNavigate()
  const { toast } = useToast()
  const { data: planData } = usePlan()
  
  const planLimits = planData?.limits || {
    max_chatbots: 1,
    max_monthly_ingestions: 100,
  }
  
  const planFeatures = planData?.features || {
    files: { max_size_mb: 5, max_text_length: 10000 }
  }

  const fileInputRef = useRef<HTMLInputElement>(null)
  const [state, setState] = useState<WizardState>(initialState)

  // --- Load saved state on mount ---
  useEffect(() => {
    const loadOnboardingState = async () => {
      try {
        const savedState = await onboardingApi.getOnboardingState()

        // Redirect if already completed or skipped
        if (savedState.completed || savedState.skipped) {
          navigate('/dashboard')
          return
        }

        // Restore saved state
        if (savedState.step > 0 && savedState.data) {
          setState((prev) => ({
            ...prev,
            currentStep: Math.min(savedState.step, 4) as StepNumber,
            botName: savedState.data?.bot_name ?? prev.botName,
            sourceType: (savedState.data?.source_type as SourceType) ?? prev.sourceType,
            textContent: savedState.data?.text_content ?? prev.textContent,
            urlContent: savedState.data?.url_content ?? prev.urlContent,
            systemPrompt: savedState.data?.system_prompt ?? prev.systemPrompt,
            welcomeMessage: savedState.data?.welcome_message ?? prev.welcomeMessage,
            createdBotId: savedState.data?.created_bot_id ?? prev.createdBotId,
          }))
        }
      } catch (error) {
        console.error('Failed to load onboarding state:', error)
      }
    }

    loadOnboardingState()
  }, [navigate])

  // --- Auto-save state (debounced) ---
  useEffect(() => {
    const saveState = async () => {
      // Don't save on final step (completion)
      if (state.currentStep === 4) return

      const data: onboardingApi.OnboardingData = {
        bot_name: state.botName,
        source_type: state.sourceType,
        text_content: state.textContent,
        url_content: state.urlContent,
        system_prompt: state.systemPrompt,
        welcome_message: state.welcomeMessage,
        created_bot_id: state.createdBotId || undefined,
      }

      try {
        await onboardingApi.updateOnboardingState(state.currentStep, data)
      } catch (error) {
        console.error('Failed to save onboarding state:', error)
      }
    }

    const timer = setTimeout(saveState, 500)
    return () => clearTimeout(timer)
  }, [state])

  // --- Validation ---
  const canProceed = useCallback((): boolean => {
    switch (state.currentStep) {
      case 1:
        return state.botName.trim().length >= 2
      case 2:
        // Step 2 is optional - can always proceed (skipStep2 handles skip case)
        if (state.skipDataSource) return true
        if (state.sourceType === 'text') {
          const maxTextLength = planFeatures.files?.max_text_length || 10000
          return state.textContent.trim().length >= 50 && state.textContent.trim().length <= maxTextLength
        }
        if (state.sourceType === 'url') return state.urlContent.trim().startsWith('http')
        if (state.sourceType === 'file') return state.pdfFile !== null
        return false
      case 3:
        return state.systemPrompt.trim().length >= 10
      case 4:
        return true
      default:
        return false
    }
  }, [state.currentStep, state.botName, state.sourceType, state.textContent, state.urlContent, state.pdfFile, state.systemPrompt, state.skipDataSource])

  // --- Navigation ---
  const goToNextStep = useCallback(async () => {
    if (!canProceed()) {
      toast('Lütfen gerekli alanları doldurun.', 'error')
      return
    }

    // If on step 3, create the bot
    if (state.currentStep === 3) {
      setState((prev) => ({ ...prev, isLoading: true }))
      try {
        // Create the chatbot
        const { data: chatbot } = await api.post('/api/v1/chatbots', {
          name: state.botName,
          system_prompt: state.systemPrompt,
          welcome_message: state.welcomeMessage,
        })

        const createdBotId = chatbot.id as string

        // Add source based on type (only if not skipped)
        if (!state.skipDataSource) {
          if (state.sourceType === 'text' && state.textContent.trim()) {
            const formData = new FormData()
            formData.append('source_type', 'text')
            formData.append('text', state.textContent)
            await api.post(`/api/v1/chatbots/${createdBotId}/sources`, formData, {
              headers: { 'Content-Type': 'multipart/form-data' },
            })
          } else if (state.sourceType === 'url' && state.urlContent.trim()) {
            const formData = new FormData()
            formData.append('source_type', 'url')
            formData.append('source_url', state.urlContent)
            await api.post(`/api/v1/chatbots/${createdBotId}/sources`, formData, {
              headers: { 'Content-Type': 'multipart/form-data' },
            })
          } else if (state.sourceType === 'file' && state.pdfFile) {
            const formData = new FormData()
            formData.append('source_type', 'pdf')
            formData.append('file', state.pdfFile)
            await api.post(`/api/v1/chatbots/${createdBotId}/sources`, formData, {
              headers: { 'Content-Type': 'multipart/form-data' },
            })
          }
        }

        // Mark onboarding as completed
        await onboardingApi.completeOnboarding(createdBotId)

        toast('Botunuz başarıyla oluşturuldu!', 'success')
        setState((prev) => ({
          ...prev,
          createdBotId,
          currentStep: 4,
          isLoading: false,
        }))
      } catch {
        toast('Bot oluşturulurken bir hata oluştu.', 'error')
        setState((prev) => ({ ...prev, isLoading: false }))
      }
      return
    }

    // Normal step advancement
    setState((prev) => ({
      ...prev,
      currentStep: Math.min(prev.currentStep + 1, 4) as StepNumber,
    }))
  }, [state, canProceed, toast])

  const goToPreviousStep = useCallback(() => {
    setState((prev) => ({
      ...prev,
      currentStep: Math.max(prev.currentStep - 1, 1) as StepNumber,
    }))
  }, [])

  const finish = useCallback(() => {
    if (state.createdBotId) {
      navigate(`/dashboard/chatbots/${state.createdBotId}/playground`)
    } else {
      navigate('/dashboard')
    }
  }, [state.createdBotId, navigate])

  const skip = useCallback(async () => {
    try {
      await onboardingApi.skipOnboarding()
      navigate('/dashboard')
    } catch (error) {
      console.error('Failed to skip onboarding:', error)
      navigate('/dashboard') // Navigate anyway
    }
  }, [navigate])

  const skipStep2 = useCallback(() => {
    setState((prev) => ({
      ...prev,
      skipDataSource: true,
      currentStep: 3 as StepNumber,
    }))
  }, [])

  // --- Form updates ---
  const setBotName = useCallback((name: string) => {
    setState((prev) => ({ ...prev, botName: name }))
  }, [])

  const setSourceType = useCallback((type: SourceType) => {
    setState((prev) => ({ ...prev, sourceType: type }))
  }, [])

  const setTextContent = useCallback((content: string) => {
    setState((prev) => ({ ...prev, textContent: content }))
  }, [])

  const setUrlContent = useCallback((url: string) => {
    setState((prev) => ({ ...prev, urlContent: url }))
  }, [])

  const setPdfFile = useCallback((file: File | null) => {
    setState((prev) => ({ ...prev, pdfFile: file }))
    if (file === null && fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }, [])

  const setSystemPrompt = useCallback((prompt: string) => {
    setState((prev) => ({ ...prev, systemPrompt: prompt }))
  }, [])

  const setWelcomeMessage = useCallback((message: string) => {
    setState((prev) => ({ ...prev, welcomeMessage: message }))
  }, [])

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.type !== 'application/pdf') {
      toast('Yalnızca PDF dosyaları desteklenir.', 'error')
      if (fileInputRef.current) fileInputRef.current.value = ''
      return
    }

    const maxFileSizeMB = planFeatures.files?.max_size_mb || 10
    const maxSize = maxFileSizeMB * 1024 * 1024
    if (file.size > maxSize) {
      toast(`Dosya boyutu ${maxFileSizeMB}MB'den büyük olamaz.`, 'error')
      if (fileInputRef.current) fileInputRef.current.value = ''
      return
    }

    setState((prev) => ({ ...prev, pdfFile: file }))
  }, [toast, planFeatures.files?.max_size_mb])

  return [
    state,
    {
      goToNextStep,
      goToPreviousStep,
      finish,
      skip,
      skipStep2,
      setBotName,
      setSourceType,
      setTextContent,
      setUrlContent,
      setPdfFile,
      setSystemPrompt,
      setWelcomeMessage,
      handleFileSelect,
      canProceed,
      planLimits,
      planFeatures,
    },
  ]
}
