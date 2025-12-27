// Z-index for maximum layer priority
export const WIDGET_Z_INDEX = 2147483647

// Default values
export const DEFAULT_MAX_CHARS = 1000
export const DEFAULT_THEME_COLOR = '#3b82f6'
export const DEFAULT_POSITION: 'bottom-right' = 'bottom-right'

// Timeouts
export const SCROLL_DELAY_MS = 10
export const ERROR_DISPLAY_DELAY_MS = 300

// Storage
export const STORAGE_PREFIX = 'chatbot_session_'
export const DEBUG_STORAGE_KEY = 'botla_debug'

// API
export const DEFAULT_API_ENDPOINTS = {
  config: (chatbotId: string) => `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}`,
  chat: (chatbotId: string) => `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/chat`,
  feedback: (chatbotId: string) => `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/feedback`,
  handoff: (chatbotId: string, requestId: string) => 
    `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/handoff/${encodeURIComponent(requestId)}/contact`,
} as const

// i18n defaults
export const DEFAULT_ERROR_MESSAGE = {
  tr: 'Şu an bir hata oluştu, lütfen tekrar deneyin.',
  en: 'An error occurred, please try again.'
}
