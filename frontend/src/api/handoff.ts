import { api } from './client'

export interface HandoffRequest {
  id: string
  chatbot_id: string
  conversation_id: string
  status: 'pending' | 'assigned' | 'resolved'
  assigned_to?: string
  notes?: string
  user_email?: string
  created_at: string
  resolved_at?: string
}

export interface Message {
  id: string
  conversation_id: string
  role: 'user' | 'assistant'
  content: string
  tokens_used: number
  thumbs_up?: boolean
  created_at: string
  type?: string
}

export interface HandoffRequestDetail {
  request: HandoffRequest
  messages: Message[]
}

export async function getHandoffRequests(chatbotId: string): Promise<HandoffRequest[]> {
  const res = await api.get(`/api/v1/chatbots/${chatbotId}/handoff-requests`)
  return res.data.requests || []
}

export async function getHandoffRequestDetail(
  chatbotId: string,
  requestId: string,
): Promise<HandoffRequestDetail> {
  const res = await api.get(`/api/v1/chatbots/${chatbotId}/handoff-requests/${requestId}`)
  return res.data
}

export async function updateHandoffStatus(
  chatbotId: string,
  requestId: string,
  status: string,
  assignedTo?: string,
): Promise<void> {
  await api.patch(`/api/v1/chatbots/${chatbotId}/handoff-requests/${requestId}`, {
    status,
    assigned_to: assignedTo,
  })
}
