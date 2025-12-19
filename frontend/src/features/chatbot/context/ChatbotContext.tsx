import { createContext, useContext, ReactNode, useState, useEffect } from 'react'
import { useChatbotForm } from '../hooks/useChatbotForm'
import { useChatbot } from '@/hooks/queries/useChatbot'
import { usePlan } from '@/hooks/queries/useProfile'

type ChatbotFormReturn = ReturnType<typeof useChatbotForm>

interface ChatbotContextType extends ChatbotFormReturn {
  planConfig: {
    branding?: { can_hide_branding?: boolean; can_custom_branding?: boolean }
    scraping?: { max_pages_per_crawl?: number; max_urls_per_bot?: number; dynamic_enabled?: boolean }
    security?: { secure_embed_enabled?: boolean }
    guardrails?: {
      can_customize_thresholds?: boolean
      can_use_smart_fallback?: boolean
      can_use_escalate_fallback?: boolean
      can_manage_topics?: boolean
      can_customize_messages?: boolean
    }
    chat?: {
      allowed_models: string[]
      max_monthly_tokens: number
      rag: { top_k: number; max_context_tokens: number }
    }
    files?: {
      ocr_enabled: boolean
      max_size_mb: number
      max_text_length?: number
      max_files_per_bot: number
      max_files_total: number
      total_storage_mb: number
    }
    refresh?: {
      enabled: boolean
      max_monthly: number
    }
  }
  userPlan: string
}

export const ChatbotContext = createContext<ChatbotContextType | undefined>(undefined)

export function ChatbotProvider({ 
  children, 
  chatbotId, 
  isNew 
}: { 
  children: ReactNode
  chatbotId?: string
  isNew: boolean 
}) {
  const form = useChatbotForm()
  const [planConfig, setPlanConfig] = useState<ChatbotContextType['planConfig']>({})

  // Use React Query for plan data
  const { data: planData } = usePlan()
  const userPlan = planData?.code || 'free'

  // Use React Query for chatbot data (replaces manual api.get)
  const { data: chatbotData } = useChatbot(chatbotId, !isNew)

  // Sync form state when chatbot data changes
  useEffect(() => {
    if (chatbotData) {
      form.setFromServer(chatbotData)
    }
  }, [chatbotData])

  // Update plan config when plan data changes
  useEffect(() => {
    if (planData?.features) {
      setPlanConfig(planData.features)
    }
  }, [planData])

  // Enforce plan restrictions
  useEffect(() => {
    if (planConfig?.guardrails && !planConfig.guardrails.can_use_escalate_fallback && form.handoffEnabled) {
      form.setHandoffEnabled(false)
    }
  }, [planConfig, form.handoffEnabled, form.setHandoffEnabled])

  return (
    <ChatbotContext.Provider value={{ ...form, planConfig, userPlan }}>
      {children}
    </ChatbotContext.Provider>
  )
}

export function useChatbotContext() {
  const context = useContext(ChatbotContext)
  if (context === undefined) {
    throw new Error('useChatbotContext must be used within a ChatbotProvider')
  }
  return context
}
