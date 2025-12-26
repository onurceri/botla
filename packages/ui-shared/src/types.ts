/**
 * Shared types for chat messages and components
 */

export interface ChatMessage {
  id?: string
  role: 'user' | 'assistant'
  content: string
  ts?: number
  feedback?: boolean
  type?: 'welcome' | 'handoff' | 'normal'
  handoffRequestId?: string
  emailSubmitted?: boolean
}

export interface ChatConfig {
  botName?: string
  botIcon?: string
  maxChars?: number
  sessionId: string
  chatbotId: string
}

export interface CustomBranding {
  logo_url?: string
  text?: string
  link?: string
}

export interface ChatResponse {
  response: string
  tokens_used?: number
  sources_used?: Array<{
    chunk_index: number
    source_type: string
  }>
}

export interface ChatRequest {
  message: string
  session_id: string
}
