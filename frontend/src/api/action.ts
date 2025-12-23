import { api } from './client'

export interface Action {
  id: string
  chatbot_id: string
  name: string
  description?: string
  action_type: 'builtin' | 'http' | 'zapier'
  config?: any
  parameters?: any
  enabled: boolean
  created_at: string
  updated_at?: string
}

export interface CreateActionRequest {
  name: string
  description?: string
  action_type: 'builtin' | 'http' | 'zapier'
  config?: any
  parameters?: any
  enabled: boolean
}

export const getActions = async (chatbotId: string) => {
  const { data } = await api.get<{ actions: Action[] }>(`/api/v1/chatbots/${chatbotId}/actions`)
  return data.actions
}

export const getAction = async (chatbotId: string, actionId: string) => {
  const { data } = await api.get<Action>(`/api/v1/chatbots/${chatbotId}/actions/${actionId}`)
  return data
}

export const createAction = async (chatbotId: string, action: CreateActionRequest) => {
  const { data } = await api.post<Action>(`/api/v1/chatbots/${chatbotId}/actions`, action)
  return data
}

export const updateAction = async (
  chatbotId: string,
  actionId: string,
  action: Partial<CreateActionRequest>,
) => {
  const { data } = await api.put<Action>(
    `/api/v1/chatbots/${chatbotId}/actions/${actionId}`,
    action,
  )
  return data
}

export const deleteAction = async (chatbotId: string, actionId: string) => {
  await api.delete(`/api/v1/chatbots/${chatbotId}/actions/${actionId}`)
}

export interface ActionLog {
  id: string
  chatbot_id: string
  action_id: string
  conversation_id?: string
  status: string
  request_payload: any
  response_payload: any
  error_message?: string
  duration_ms: number
  created_at: string
}

export const getActionLogs = async (chatbotId: string, page = 1, limit = 20) => {
  const { data } = await api.get<{ logs: ActionLog[]; page: number; limit: number }>(
    `/api/v1/chatbots/${chatbotId}/actions/logs`,
    {
      params: { page, limit },
    },
  )
  return data
}
