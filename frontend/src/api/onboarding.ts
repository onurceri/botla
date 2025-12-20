import { api } from './client'

export interface OnboardingState {
  completed: boolean
  skipped: boolean
  step: number
  data?: OnboardingData
}

export interface OnboardingData {
  bot_name?: string
  source_type?: 'text' | 'url' | 'file'
  text_content?: string
  url_content?: string
  system_prompt?: string
  welcome_message?: string
  created_bot_id?: string
}

export const getOnboardingState = async (): Promise<OnboardingState> => {
  const { data } = await api.get<OnboardingState>('/api/v1/me/onboarding')
  return data
}

export const updateOnboardingState = async (step: number, onboardingData: OnboardingData): Promise<void> => {
  await api.put('/api/v1/me/onboarding', { step, data: onboardingData })
}

export const skipOnboarding = async (): Promise<void> => {
  await api.post('/api/v1/me/onboarding/skip')
}

export const completeOnboarding = async (botId: string): Promise<void> => {
  await api.post('/api/v1/me/onboarding/complete', { bot_id: botId })
}
