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

export const getDetailedHealth = async () => {
  const { data } = await api.get<DetailedHealth>(`${ADMIN_BASE}/health/detailed`)
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
