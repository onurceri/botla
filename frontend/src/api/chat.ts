import { api } from './client'

export interface ChatMessage {
  message: string
  session_id: string
}

export interface ChatResponse {
  response: string
  tokens_used: number
  sources_used: Array<{ chunk_index: number; source_type: string }>
  message_id?: string
}

export const sendChatMessage = async (chatbotId: string, message: ChatMessage): Promise<ChatResponse> => {
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/chat`, message)
  return data
}
