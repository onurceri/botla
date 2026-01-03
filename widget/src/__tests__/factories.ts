/**
 * Test data factories for widget tests.
 * Provides consistent, customizable test data generation.
 */

import type { ChatMessage, ChatbotConfig, CustomBranding } from '../types'

/**
 * Factory options interface
 */
interface FactoryOptions<T> {
  overrides?: Partial<T>
  count?: number
}

/**
 * Creates a single ChatMessage with optional overrides
 */
export function createMessage(overrides: Partial<ChatMessage> = {}): ChatMessage {
  const timestamp = Date.now()
  const isUser = overrides.role === 'user'
  
  return {
    id: `msg-${timestamp}-${Math.random().toString(36).substr(2, 9)}`,
    role: isUser ? 'user' : 'assistant',
    content: isUser 
      ? 'Bu bir test kullanıcı mesajıdır.' 
      : 'Bu bir test asistan yanıtıdır.',
    ts: timestamp,
    feedback: undefined,
    type: 'normal',
    handoffRequestId: undefined,
    emailSubmitted: false,
    ...overrides,
  }
}

/**
 * Creates a welcome message
 */
export function createWelcomeMessage(welcomeText: string = 'Merhaba! Size nasıl yardımcı olabilirim?'): ChatMessage {
  return createMessage({
    role: 'assistant',
    content: welcomeText,
    type: 'welcome',
  })
}

/**
 * Creates a handoff message
 */
export function createHandoffMessage(requestId: string = 'handoff-123'): ChatMessage {
  return createMessage({
    role: 'assistant',
    content: 'Destek talebinizi aldık.',
    type: 'handoff',
    handoffRequestId: requestId,
  })
}

/**
 * Creates a list of chat messages for conversation testing
 */
export function createMessageList(count: number = 5, options: FactoryOptions<ChatMessage> = {}): ChatMessage[] {
  const messages: ChatMessage[] = []
  let role: 'user' | 'assistant' = 'user'
  let baseTimestamp = Date.now()
  
  for (let i = 0; i < count; i++) {
    messages.push(createMessage({
      role,
      content: role === 'user' 
        ? `Kullanıcı mesajı ${i + 1}` 
        : `Asistan yanıtı ${i + 1}`,
      ts: baseTimestamp + (i * 1000),
      ...options.overrides,
    }))
    role = role === 'user' ? 'assistant' : 'user'
  }
  
  return messages
}

/**
 * Creates a ChatbotConfig with optional overrides
 */
export function createChatbotConfig(overrides: Partial<ChatbotConfig> = {}): ChatbotConfig {
  return {
    theme_color: '#6366f1',
    position: 'bottom-right',
    welcome_message: 'Merhaba! Size nasıl yardımcı olabilirim?',
    suggested_questions: [
      'Nasıl yardımcı olabilirim?',
      'Hızlı başlangıç',
      'Özellikler',
    ],
    bot_display_name: 'Test Bot',
    bot_icon: undefined,
    hide_branding: false,
    custom_branding: undefined,
    max_chars: 1000,
    bot_message_color: undefined,
    bot_message_text_color: undefined,
    user_message_color: undefined,
    user_message_text_color: undefined,
    chat_header_color: undefined,
    chat_header_text_color: undefined,
    chat_font_family: undefined,
    chat_panel_bg_color: undefined,
    chat_background_color: undefined,
    input_background_color: undefined,
    input_text_color: undefined,
    bubble_radius: undefined,
    send_button_color: undefined,
    chat_panel_height: undefined,
    chat_panel_width: undefined,
    ...overrides,
  }
}

/**
 * Creates a custom branding configuration
 */
export function createCustomBranding(overrides: Partial<CustomBranding> = {}): CustomBranding {
  return {
    logo_url: 'https://example.com/logo.png',
    text: 'Powered by',
    link: 'https://example.com',
    ...overrides,
  }
}

/**
 * Creates a mock chat API response
 */
export interface ChatAPIResponse {
  response: string
  message_id: string
  handoff_request_id?: string
}

export function createChatResponse(overrides: Partial<ChatAPIResponse> = {}): ChatAPIResponse {
  return {
    response: 'Bu bir test yanıtıdır.',
    message_id: `msg-${Date.now()}`,
    handoff_request_id: undefined,
    ...overrides,
  }
}

/**
 * Creates a mock feedback API response
 */
export function createFeedbackResponse(success: boolean = true) {
  return { success }
}

/**
 * Creates a mock session data
 */
export interface SessionData {
  sessionId: string
  messages: ChatMessage[]
}

export function createSessionData(overrides: Partial<SessionData> = {}): SessionData {
  return {
    sessionId: `session-${Date.now()}`,
    messages: [createWelcomeMessage()],
    ...overrides,
  }
}

/**
 * Creates widget mount parameters for testing
 */
export interface WidgetParams {
  chatbotId: string
  apiBase?: string
  themeColor?: string
  welcome?: string
  embedTokenUrl?: string
  captchaSiteKey?: string
  autoOpen?: boolean
  useOverrides?: boolean
  headerColor?: string
  headerTextColor?: string
  botMessageColor?: string
  botMessageTextColor?: string
  userMessageColor?: string
  userMessageTextColor?: string
  fontFamily?: string
  position?: 'bottom-left' | 'bottom-right'
  botName?: string
  botIcon?: string
  panelHeight?: string
  panelWidth?: string
  panelBg?: string
  inputBg?: string
  inputText?: string
  chatBg?: string
  bubbleRadius?: string
  sendButtonColor?: string
  resetSession?: boolean
  sessionId?: string
}

export function createWidgetParams(overrides: Partial<WidgetParams> = {}): WidgetParams {
  return {
    chatbotId: 'test-chatbot',
    apiBase: '',
    themeColor: undefined,
    welcome: undefined,
    embedTokenUrl: undefined,
    captchaSiteKey: undefined,
    autoOpen: false,
    useOverrides: false,
    headerColor: undefined,
    headerTextColor: undefined,
    botMessageColor: undefined,
    botMessageTextColor: undefined,
    userMessageColor: undefined,
    userMessageTextColor: undefined,
    fontFamily: undefined,
    position: 'bottom-right',
    botName: undefined,
    botIcon: undefined,
    panelHeight: undefined,
    panelWidth: undefined,
    panelBg: undefined,
    inputBg: undefined,
    inputText: undefined,
    chatBg: undefined,
    bubbleRadius: undefined,
    sendButtonColor: undefined,
    resetSession: false,
    sessionId: undefined,
    ...overrides,
  }
}

/**
 * Converts widget params to URL search params
 */
export function paramsToSearchParams(params: WidgetParams): URLSearchParams {
  const sp = new URLSearchParams()
  
  sp.set('chatbot-id', params.chatbotId)
  if (params.apiBase) sp.set('api-base', params.apiBase)
  if (params.themeColor) sp.set('color', params.themeColor)
  if (params.welcome) sp.set('welcome', params.welcome)
  if (params.embedTokenUrl) sp.set('embed-token-url', params.embedTokenUrl)
  if (params.captchaSiteKey) sp.set('captcha-site-key', params.captchaSiteKey)
  if (params.autoOpen) sp.set('auto-open', '1')
  if (params.useOverrides) sp.set('use-url-overrides', '1')
  if (params.headerColor) sp.set('header-color', params.headerColor)
  if (params.headerTextColor) sp.set('header-text-color', params.headerTextColor)
  if (params.botMessageColor) sp.set('bot-message-color', params.botMessageColor)
  if (params.botMessageTextColor) sp.set('bot-message-text-color', params.botMessageTextColor)
  if (params.userMessageColor) sp.set('user-message-color', params.userMessageColor)
  if (params.userMessageTextColor) sp.set('user-message-text-color', params.userMessageTextColor)
  if (params.fontFamily) sp.set('font-family', params.fontFamily)
  if (params.position) sp.set('position', params.position)
  if (params.botName) sp.set('bot-name', params.botName)
  if (params.botIcon) sp.set('bot-icon', params.botIcon)
  if (params.panelHeight) sp.set('panel-height', params.panelHeight)
  if (params.panelWidth) sp.set('panel-width', params.panelWidth)
  if (params.panelBg) sp.set('panel-bg-color', params.panelBg)
  if (params.inputBg) sp.set('input-bg-color', params.inputBg)
  if (params.inputText) sp.set('input-text-color', params.inputText)
  if (params.chatBg) sp.set('chat-bg-color', params.chatBg)
  if (params.bubbleRadius) sp.set('bubble-radius', params.bubbleRadius)
  if (params.sendButtonColor) sp.set('send-button-color', params.sendButtonColor)
  if (params.resetSession) sp.set('reset-session', '1')
  if (params.sessionId) sp.set('session-id', params.sessionId)
  
  return sp
}
