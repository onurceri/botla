import { createContext, useContext, ReactNode, useState, useEffect } from 'react'
import { useChatbotForm } from '../hooks/useChatbotForm'
import { api } from '@/api/client'

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
  const [userPlan, setUserPlan] = useState('free')
  const [planConfig, setPlanConfig] = useState<ChatbotContextType['planConfig']>({})

  // Fetch user profile and plan config
  const fetchUserProfile = () => {
    api.get('/api/v1/me').then(({ data }) => {
      const plan = data.plan_code || data.subscription_plan || 'free'
      setUserPlan(plan)
      if (data.config) {
        setPlanConfig(data.config)
      }
    }).catch(() => {})
  }

  useEffect(() => {
    fetchUserProfile()

    // Refetch when user returns to tab (in case plan was upgraded)
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        fetchUserProfile()
      }
    }
    document.addEventListener('visibilitychange', handleVisibilityChange)

    if (!isNew && chatbotId) {
      api.get(`/api/v1/chatbots/${chatbotId}`).then(({ data }) => {
        form.setFromServer(data)
      }).catch(() => {})
    }

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [chatbotId, isNew])

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
