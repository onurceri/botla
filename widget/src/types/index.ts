/**
 * Widget type definitions
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

export interface CustomBranding {
  logo_url?: string
  text?: string
  link?: string
}

export interface ChatbotConfig {
  theme_color?: string
  position?: 'bottom-right' | 'bottom-left' | string
  welcome_message?: string
  suggested_questions?: string[]
  bot_display_name?: string
  bot_icon?: string
  hide_branding?: boolean
  custom_branding?: CustomBranding
  max_chars?: number
  // Styling
  bot_message_color?: string
  bot_message_text_color?: string
  user_message_color?: string
  user_message_text_color?: string
  chat_header_color?: string
  chat_header_text_color?: string
  chat_font_family?: string
  chat_panel_bg_color?: string
  chat_background_color?: string
  input_background_color?: string
  input_text_color?: string
  bubble_radius?: string
  send_button_color?: string
  chat_panel_height?: string
  chat_panel_width?: string
}

export interface SessionData {
  sessionId: string
  messages: ChatMessage[]
}

export type WidgetPosition = 'bottom-right' | 'bottom-left'
export type PositionStrategy = 'fixed' | 'absolute'

// Alias for backward compatibility during refactor if needed, 
// strictly following the plan implies we should use ChatMessage, 
// but code uses Message. I'll export Message as alias.
export type Message = ChatMessage
