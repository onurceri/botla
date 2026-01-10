import { api } from './client'

const ADMIN_BASE = '/api/v1/admin' as const

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page?: number
  per_page?: number
}

export interface OverviewStats {
  total_users: number
  total_organizations: number
  total_chatbots: number
  total_messages: number
  users_today?: number
  conversations_today?: number
  active_plans?: Record<string, number>
}

export interface DependencyStatus {
  name: string
  status: 'ok' | 'degraded' | 'down'
  latency_ms: number
  message?: string
  checked_at: string
}

export interface DetailedHealth {
  status: 'healthy' | 'degraded' | 'unhealthy'
  version: string
  uptime: string
  environment: string
  dependencies: DependencyStatus[]
  last_updated: string
}

export interface QueueStats {
  queue_name: string
  pending_count: number
  processing_count: number
  failed_count: number
  oldest_pending?: string | null
}

export interface StuckJob {
  id: string
  queue_name: string
  source_id?: string
  chatbot_id?: string
  status: string
  started_at: string
  stuck_duration: string
  error_message?: string
}

export interface ErrorLogEntry {
  id: string
  error_type: string
  message: string
  request_path?: string
  request_method?: string
  user_id?: string
  chatbot_id?: string
  organization_id?: string
  severity: string
  created_at: string
  stack_trace?: string
  context?: string
}

export interface AdminUser {
  id: string
  email: string
  full_name?: string
  avatar_url?: string
  plan_id: string
  preferred_language_id?: string
  onboarding_completed?: boolean
  onboarding_step?: string
  onboarding_skipped?: boolean
  onboarding_data?: unknown
  is_platform_admin: boolean
  created_at: string
}

export interface AdminOrganization {
  id: string
  name: string
  slug: string
  owner_id: string
  plan_id: string
  user_count?: number
  chatbot_count?: number
  created_at: string
  updated_at: string
}

export interface UserListResponse {
  users: AdminUser[]
  total: number
}

export interface ListUsersParams {
  limit?: number
  offset?: number
  email?: string
  is_platform_admin?: boolean
  plan_id?: string
}

export interface OrganizationListResponse {
  organizations: AdminOrganization[]
  total: number
}

export interface ListOrganizationsParams {
  limit?: number
  offset?: number
  name?: string
  plan_id?: string
}

export interface ListErrorsParams {
  severity?: string
  offset?: number
  limit?: number
}

export interface AuditLogEntry {
  id: string
  admin_user_id: string
  action: string
  target_type: string
  target_id?: string | null
  details: Record<string, unknown>
  ip_address: string
  user_agent: string
  created_at: string
}

export interface ListAuditLogsParams {
  limit?: number
  offset?: number
  admin_user_id?: string
  action?: string
  target_type?: string
}

export const getOverviewStats = async () => {
  const { data } = await api.get<OverviewStats>(`${ADMIN_BASE}/stats/overview`)
  return data
}

export const getDetailedHealth = async (refresh = false) => {
  const { data } = await api.get<DetailedHealth>(`${ADMIN_BASE}/health/detailed`, {
    params: refresh ? { refresh: 'true' } : undefined,
  })
  return data
}

export const listUsers = async (params?: ListUsersParams) => {
  const { data } = await api.get<UserListResponse>(`${ADMIN_BASE}/users`, { params })
  return data
}

export const getUser = async (id: string) => {
  const { data } = await api.get<AdminUser>(`${ADMIN_BASE}/users/${id}`)
  return data
}

export const updateUser = async (
  id: string,
  updates: { full_name?: string; plan_id?: string; is_platform_admin?: boolean; status?: string },
) => {
  const { data } = await api.patch<{ status: string }>(`${ADMIN_BASE}/users/${id}`, updates)
  return data
}

export const listOrganizations = async (params?: ListOrganizationsParams) => {
  const { data } = await api.get<OrganizationListResponse>(`${ADMIN_BASE}/organizations`, { params })
  return data
}

export const getOrganization = async (id: string) => {
  const { data } = await api.get<AdminOrganization>(`${ADMIN_BASE}/organizations/${id}`)
  return data
}

export const getQueues = async () => {
  const { data } = await api.get<QueueStats[]>(`${ADMIN_BASE}/queues`)
  return data
}

export const getStuckJobs = async (threshold?: string) => {
  const { data } = await api.get<StuckJob[]>(`${ADMIN_BASE}/queues/stuck`, { params: { threshold } })
  return data
}

export const retryJob = async (id: string) => {
  const { data } = await api.post(`${ADMIN_BASE}/queues/${id}/retry`)
  return data
}

export const deleteJob = async (id: string) => {
  const { data } = await api.delete(`${ADMIN_BASE}/queues/${id}`)
  return data
}

export const getErrors = async (severity?: string, offset?: number, limit?: number) => {
  const { data } = await api.get<PaginatedResponse<ErrorLogEntry>>(`${ADMIN_BASE}/errors`, {
    params: { severity, offset, limit },
  })
  return data
}

export const getPrivacyDownloadURL = async (id: string) => {
  const { data } = await api.get<{ url: string }>(`${ADMIN_BASE}/privacy/requests/${id}/download-url`)
  return data
}

export const listErrors = async (params?: ListErrorsParams) => {
  const { data } = await api.get<PaginatedResponse<ErrorLogEntry>>(`${ADMIN_BASE}/errors`, { params })
  return data
}

export const getErrorStats = async () => {
  const { data } = await api.get<Record<string, number>>(`${ADMIN_BASE}/errors/stats`)
  return data
}

export const getError = async (id: string) => {
  const { data } = await api.get<ErrorLogEntry>(`${ADMIN_BASE}/errors/${id}`)
  return data
}

export const listAuditLogs = async (params?: ListAuditLogsParams) => {
  const { data } = await api.get<PaginatedResponse<AuditLogEntry>>(`${ADMIN_BASE}/audit-logs`, { params })
  return data
}

export interface PrivacyRequest {
  id: string
  user_id: string
  user_email: string
  request_type: 'export' | 'deletion' | 'correction'
  status: 'pending' | 'processing' | 'completed' | 'denied'
  reason?: string
  denial_reason?: string
  processed_at?: string
  processed_by?: string
  completed_at?: string
  export_url?: string
  export_expires_at?: string
  created_at: string
}

export interface ListPrivacyRequestsParams {
  status?: string
  limit?: number
  offset?: number
}

export const listPrivacyRequests = async (params?: ListPrivacyRequestsParams) => {
  const { data } = await api.get<PaginatedResponse<PrivacyRequest>>(`${ADMIN_BASE}/privacy/requests`, { params })
  return data
}

export const processPrivacyRequest = async (
  id: string,
  action: 'approve' | 'deny',
  denial_reason?: string
) => {
  const { data } = await api.patch(`${ADMIN_BASE}/privacy/requests/${id}`, {
    action,
    denial_reason,
  })
  return data
}

// === Chatbots ===

export interface AdminChatbot {
  id: string
  name: string
  owner_id: string
  workspace_id: string
  organization_id?: string
  organization_name?: string
  owner_email: string
  source_count: number
  message_count: number
  custom_branding?: unknown
  created_at: string
}

export interface ChatbotListResponse {
  chatbots: AdminChatbot[]
  total: number
}

export interface ListChatbotsParams {
  limit?: number
  offset?: number
  name?: string
  organization_id?: string
  owner_id?: string
}

export const listChatbots = async (params?: ListChatbotsParams) => {
  const { data } = await api.get<ChatbotListResponse>(`${ADMIN_BASE}/chatbots`, { params })
  return data
}

export const getChatbot = async (id: string) => {
  const { data } = await api.get<AdminChatbot>(`${ADMIN_BASE}/chatbots/${id}`)
  return data
}

export const forceRefreshChatbot = async (id: string) => {
  const { data } = await api.post<{ status: string; sources_reset: number; sources_queued: number }>(
    `${ADMIN_BASE}/chatbots/${id}/force-refresh`
  )
  return data
}



// === Data Sources ===

export interface AdminSource {
  id: string
  chatbot_id: string
  chatbot_name: string
  organization_name?: string
  owner_email: string
  source_type: string
  source_url?: string
  original_filename?: string
  status: string
  error_message?: string
  chunk_count: number
  size_bytes?: number
  processed_at?: string
  created_at: string
}

export interface SourceListResponse {
  sources: AdminSource[]
  total: number
}

export interface ListSourcesParams {
  limit?: number
  offset?: number
  chatbot_id?: string
  source_type?: string
  status?: string
  owner_id?: string
}

export const listSources = async (params?: ListSourcesParams) => {
  const { data } = await api.get<SourceListResponse>(`${ADMIN_BASE}/sources`, { params })
  return data
}

export const getSource = async (id: string) => {
  const { data } = await api.get<AdminSource>(`${ADMIN_BASE}/sources/${id}`)
  return data
}

export const getSourceStats = async () => {
  const { data } = await api.get<Record<string, number>>(`${ADMIN_BASE}/sources/stats`)
  return data
}

export const getSourceTypes = async () => {
  const { data } = await api.get<{ types: string[]; statuses: string[] }>(`${ADMIN_BASE}/sources/types`)
  return data
}

export const reprocessSource = async (id: string) => {
  const { data } = await api.post<{ status: string; queued: boolean }>(
    `${ADMIN_BASE}/sources/${id}/reprocess`
  )
  return data
}


// === Plans Management ===

export interface AdminPlanSummary {
  id: string
  code: string
  name: string
  status: string
  billing_cycle: string
  price: number
  currency: string
  trial_days: number
  max_chatbots: number
  max_monthly_ingestions: number
  files_max_size_mb: number
  chat_allowed_models_count: number
}

export interface PlanLimitsDetail {
  // Core Limits
  max_chatbots: number
  max_monthly_ingestions: number
  max_monthly_embedding_tokens: number
  min_readd_cooldown_minutes: number

  // Scraping
  scraping_dynamic_enabled: boolean
  scraping_max_urls_per_bot: number
  scraping_max_pages_per_crawl: number

  // Files
  files_max_size_mb: number
  files_max_files_per_bot: number
  files_max_files_total: number
  files_total_storage_mb: number
  files_max_text_length: number

  // Chat
  chat_default_model: string
  chat_allowed_models: string[]
  chat_max_monthly_tokens: number
  chat_rag_top_k: number
  chat_rag_max_context_tokens: number
  chat_max_suggested_questions: number
  chat_max_manual_questions: number
  chat_min_response_token_limit: number
  chat_max_response_token_limit: number

  // Refresh
  refresh_enabled: boolean
  refresh_max_monthly: number

  // Security
  security_secure_embed_enabled: boolean

  // Guardrails
  guardrails_can_customize_thresholds: boolean
  guardrails_can_use_smart_fallback: boolean
  guardrails_can_use_escalate_fallback: boolean
  guardrails_can_manage_topics: boolean
  guardrails_can_customize_messages: boolean

  // Branding
  branding_can_hide_branding: boolean
  branding_can_custom_branding: boolean

  // Rate Limits
  rate_limits_requests_per_minute: number
  rate_limits_window_seconds: number
  rate_limits_chat_rpm: number
  rate_limits_chat_window: number
  rate_limits_sources_rpm: number
  rate_limits_sources_window: number
}

export interface AdminPlanDetail {
  plan: {
    id: string
    code: string
    name: string
    status: string
    billing_cycle: string
    price: number
    currency: string
    trial_days: number
    created_at: string
    updated_at?: string
  }
  limits: PlanLimitsDetail
  translations: Array<{
    language_id: string
    language_code: string
    name: string
    description: string
  }>
}

export interface PlanListResponse {
  plans: AdminPlanSummary[]
  total: number
}

export interface ListPlansParams {
  limit?: number
  offset?: number
}

export const listPlans = async (params?: ListPlansParams) => {
  const { data } = await api.get<PlanListResponse>(`${ADMIN_BASE}/plans`, { params })
  return data
}

export const getPlan = async (id: string) => {
  const { data } = await api.get<AdminPlanDetail>(`${ADMIN_BASE}/plans/${id}`)
  return data
}

export interface UpdatePlanLimitsRequest {
  // Core Limits
  max_chatbots?: number
  max_monthly_ingestions?: number
  max_monthly_embedding_tokens?: number
  min_readd_cooldown_minutes?: number

  // Scraping
  scraping_dynamic_enabled?: boolean
  scraping_max_urls_per_bot?: number
  scraping_max_pages_per_crawl?: number

  // Files
  files_max_size_mb?: number
  files_max_files_per_bot?: number
  files_max_files_total?: number
  files_total_storage_mb?: number
  files_max_text_length?: number

  // Chat
  chat_default_model?: string
  chat_allowed_models?: string[]
  chat_max_monthly_tokens?: number
  chat_rag_top_k?: number
  chat_rag_max_context_tokens?: number
  chat_max_suggested_questions?: number
  chat_max_manual_questions?: number
  chat_min_response_token_limit?: number
  chat_max_response_token_limit?: number

  // Refresh
  refresh_enabled?: boolean
  refresh_max_monthly?: number

  // Security
  security_secure_embed_enabled?: boolean

  // Guardrails
  guardrails_can_customize_thresholds?: boolean
  guardrails_can_use_smart_fallback?: boolean
  guardrails_can_use_escalate_fallback?: boolean
  guardrails_can_manage_topics?: boolean
  guardrails_can_customize_messages?: boolean

  // Branding
  branding_can_hide_branding?: boolean
  branding_can_custom_branding?: boolean

  // Rate Limits
  rate_limits_requests_per_minute?: number
  rate_limits_window_seconds?: number
  rate_limits_chat_rpm?: number
  rate_limits_chat_window?: number
  rate_limits_sources_rpm?: number
  rate_limits_sources_window?: number
}

export const updatePlanLimits = async (planId: string, updates: UpdatePlanLimitsRequest) => {
  const { data } = await api.patch<{ status: string }>(`${ADMIN_BASE}/plans/${planId}/limits`, updates)
  return data
}

export const invalidatePlanCache = async (planId: string) => {
  const { data } = await api.post<{ status: string }>(`${ADMIN_BASE}/plans/${planId}/cache-invalidate`)
  return data
}

