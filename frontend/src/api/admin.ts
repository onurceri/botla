import { api } from './client'

// Types
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page?: number
  per_page?: number
}

export interface UserListResponse {
  users: AdminUser[]
  total: number
}

export interface OrganizationListResponse {
  organizations: AdminOrganization[]
  total: number
}

export interface OverviewStats {
  total_users: number
  total_organizations: number
  total_chatbots: number
  total_messages: number
  // The plan had these, but backend doesn't yet. Added for future-proofing.
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
  oldest_pending?: string
}

export interface StuckJob {
  id: string
  queue_name: string
  source_id?: string
  status: string
  started_at: string
  stuck_duration: string
  error_message?: string
}

export interface ErrorLogEntry {
  id: string
  error_type: string
  message: string
  severity: string
  created_at: string
  stack_trace?: string
}

export interface AdminUser {
  id: string
  email: string
  full_name: string
  avatar_url?: string
  plan_id: string
  is_platform_admin: boolean
  is_suspended?: boolean
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

// API Functions

/**
 * Get high-level platform metrics for the admin dashboard.
 */
export const getOverviewStats = async () => {
  const { data } = await api.get<OverviewStats>('/api/v1/admin/stats/overview')
  return data
}

/**
 * Get comprehensive health status of all system dependencies.
 */
export const getDetailedHealth = async () => {
  const { data } = await api.get<DetailedHealth>('/api/v1/admin/health/detailed')
  return data
}

/**
 * List all users with pagination and filtering.
 */
export const listUsers = async (params?: {
  limit?: number
  offset?: number
  email?: string
  is_platform_admin?: boolean
  plan_id?: string
}) => {
  const { data } = await api.get<UserListResponse>('/api/v1/admin/users', { params })
  return data
}

/**
 * Get details for a single user.
 */
export const getUser = async (id: string) => {
  const { data } = await api.get<AdminUser>(`/api/v1/admin/users/${id}`)
  return data
}

/**
 * Update a user's details or admin status.
 */
export const updateUser = async (
  id: string,
  updates: { full_name?: string; plan_id?: string; is_platform_admin?: boolean; status?: string },
) => {
  const { data } = await api.patch<{ status: string }>(`/api/v1/admin/users/${id}`, updates)
  return data
}

/**
 * List all organizations with pagination and filtering.
 */
export const listOrganizations = async (params?: {
  limit?: number
  offset?: number
  name?: string
  plan_id?: string
}) => {
  const { data } = await api.get<OrganizationListResponse>('/api/v1/admin/organizations', { params })
  return data
}

/**
 * Get details for a single organization.
 */
export const getOrganization = async (id: string) => {
  const { data } = await api.get<AdminOrganization>(`/api/v1/admin/organizations/${id}`)
  return data
}

// Placeholder functions for future implementation (not yet in backend)

/**
 * Get status of background job queues.
 */
export const getQueues = async () => {
  const { data } = await api.get<QueueStats[]>('/api/v1/admin/queues')
  return data
}

/**
 * Get list of jobs that seem to be stuck.
 */
export const getStuckJobs = async (threshold?: string) => {
  const { data } = await api.get<StuckJob[]>('/api/v1/admin/queues/stuck', { params: { threshold } })
  return data
}

/**
 * Retry a failed or stuck job.
 */
export const retryJob = async (id: string) => {
  const { data } = await api.post(`/api/v1/admin/queues/${id}/retry`)
  return data
}

/**
 * Delete a job from the queue.
 */
export const deleteJob = async (id: string) => {
  const { data } = await api.delete(`/api/v1/admin/queues/${id}`)
  return data
}

/**
 * List system error logs.
 */
export const listErrors = async (params?: { page?: number; severity?: string; type?: string }) => {
  const { data } = await api.get<PaginatedResponse<ErrorLogEntry>>('/api/v1/admin/errors', { params })
  return data
}

/**
 * Get details of a specific error log entry.
 */
export const getError = async (id: string) => {
  const { data } = await api.get<ErrorLogEntry>(`/api/v1/admin/errors/${id}`)
  return data
}

/**
 * List all chatbots across all organizations.
 */
export const listChatbots = async (params?: { page?: number; search?: string; status?: string }) => {
  const { data } = await api.get<PaginatedResponse<any>>('/api/v1/admin/chatbots', { params })
  return data
}

/**
 * Force a refresh of a chatbot's content.
 */
export const forceRefreshChatbot = async (id: string) => {
  const { data } = await api.post(`/api/v1/admin/chatbots/${id}/force-refresh`)
  return data
}

/**
 * List admin audit logs.
 */
export const listAuditLogs = async (params?: { page?: number }) => {
  const { data } = await api.get<PaginatedResponse<any>>('/api/v1/admin/audit-logs', { params })
  return data
}
