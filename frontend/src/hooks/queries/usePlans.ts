import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'

/**
 * Public plan API types - matches backend PlanResponse
 */
export interface PublicPlanLimits {
  max_chatbots: number
  max_monthly_ingestions: number
  max_monthly_embedding_tokens: number
}

export interface PublicPlanFeatures {
  scraping: {
    dynamic_enabled: boolean
    max_urls_per_bot: number
    max_pages_per_crawl: number
  }
  files: {
    max_size_mb: number
    max_files_per_bot: number
    max_files_total: number
    total_storage_mb: number
    max_text_length: number
  }
  chat: {
    default_model?: string
    allowed_models: string[]
    max_monthly_tokens: number
    rag: {
      top_k: number
      max_context_tokens: number
    }
    max_suggested_questions: number
    max_manual_questions: number
    min_response_token_limit: number
    max_response_token_limit: number
  }
  refresh: {
    enabled: boolean
    max_monthly: number
  }
  security: {
    secure_embed_enabled: boolean
  }
  guardrails: {
    can_customize_thresholds: boolean
    can_use_smart_fallback: boolean
    can_use_escalate_fallback: boolean
    can_manage_topics: boolean
    can_customize_messages: boolean
  }
  branding: {
    can_hide_branding: boolean
    can_custom_branding: boolean
  }
  rate_limits: {
    requests_per_minute: number
    window_seconds: number
  }
}

export interface PublicPlan {
  code: string
  name?: string
  price: number
  currency: string
  limits: PublicPlanLimits
  features: PublicPlanFeatures
}

export const PUBLIC_PLANS_QUERY_KEY = ['publicPlans'] as const

/**
 * Fetch all plans from public endpoint (no auth required).
 * Use for landing page pricing, plan comparison, etc.
 */
export function usePlans() {
  return useQuery<PublicPlan[]>({
    queryKey: PUBLIC_PLANS_QUERY_KEY,
    queryFn: async () => {
      const { data } = await api.get<PublicPlan[]>('/api/v1/plans')
      return data
    },
    staleTime: 1000 * 60 * 15, // 15 minutes (plans rarely change)
    refetchOnWindowFocus: false,
  })
}

/**
 * Fetch a single plan by code from public endpoint (no auth required).
 * Use when you need specific plan details.
 */
export function usePlanByCode(code: string) {
  return useQuery<PublicPlan>({
    queryKey: [...PUBLIC_PLANS_QUERY_KEY, code],
    queryFn: async () => {
      const { data } = await api.get<PublicPlan>(`/api/v1/plans/${code}`)
      return data
    },
    staleTime: 1000 * 60 * 15, // 15 minutes
    refetchOnWindowFocus: false,
    enabled: !!code, // Only fetch if code is provided
  })
}

/**
 * Helper to find a plan by code from the plans array
 */
export function findPlanByCode(plans: PublicPlan[] | undefined, code: string): PublicPlan | undefined {
  return plans?.find(p => p.code === code)
}
