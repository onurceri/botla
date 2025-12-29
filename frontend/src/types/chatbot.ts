/**
 * Chatbot request types for mutations.
 * These types mirror the backend Go structs from internal/services/chatbot_service.go
 */

// Re-export shared types from useChatbotForm to avoid duplication
export type {
  CustomBranding,
  FallbackMessages,
  TopicConfig,
  HandoffConfig,
  ThresholdConfig,
} from '@/features/chatbot/hooks/useChatbotForm'

// Request types for update mutations
export interface ChatbotUpdateRequest {
  name?: string;
  description?: string | null;
  custom_instruction?: string | null;
  language?: string | null;
  model?: string | null;
  temperature?: number | null;
  max_tokens?: number | null;
  theme_color?: string | null;
  welcome_message?: string | null;
  position?: string | null;
  bot_message_color?: string | null;
  user_message_color?: string | null;
  bot_message_text_color?: string | null;
  user_message_text_color?: string | null;
  chat_font_family?: string | null;
  chat_header_color?: string | null;
  chat_header_text_color?: string | null;
  chat_background_color?: string | null;
  bubble_radius?: string | null;
  input_background_color?: string | null;
  input_text_color?: string | null;
  send_button_color?: string | null;
  bot_icon?: string | null;
  bot_display_name?: string | null;
  secure_embed_enabled?: boolean | null;
  allowed_domains?: string[];
  embed_secret?: string | null;
  suggested_questions?: string[] | null;
  manual_questions?: string[] | null;
  suggestions_enabled?: boolean | null;
  include_paths?: string[] | null;
  exclude_paths?: string[] | null;
  selector_whitelist?: string[] | null;
  discovery_mode?: string | null;
  refresh_policy?: string | null;
  refresh_frequency?: string | null;
  hide_branding?: boolean | null;
  custom_branding?: import('@/features/chatbot/hooks/useChatbotForm').CustomBranding | null;
  confidence_threshold?: number | null;
  fallback_messages?: import('@/features/chatbot/hooks/useChatbotForm').FallbackMessages | null;
  topic_restrictions?: import('@/features/chatbot/hooks/useChatbotForm').TopicConfig | null;
  threshold_config?: import('@/features/chatbot/hooks/useChatbotForm').ThresholdConfig | null;
  handoff_enabled?: boolean | null;
  handoff_type?: string | null;
  handoff_config?: import('@/features/chatbot/hooks/useChatbotForm').HandoffConfig | null;
}

export interface BasicInfoRequest {
  name: string;
  description?: string | null;
  language?: string | null;
  custom_instruction?: string | null;
}

export interface AppearanceRequest {
  theme_color?: string | null;
  welcome_message?: string | null;
  position?: string | null;
  bot_message_color?: string | null;
  user_message_color?: string | null;
  bot_message_text_color?: string | null;
  user_message_text_color?: string | null;
  chat_font_family?: string | null;
  chat_header_color?: string | null;
  chat_header_text_color?: string | null;
  chat_background_color?: string | null;
  bubble_radius?: string | null;
  input_background_color?: string | null;
  input_text_color?: string | null;
  send_button_color?: string | null;
  bot_icon?: string | null;
  bot_display_name?: string | null;
  hide_branding?: boolean | null;
  custom_branding?: import('@/features/chatbot/hooks/useChatbotForm').CustomBranding | null;
  suggested_questions?: string[] | null;
  manual_questions?: string[] | null;
  suggestions_enabled?: boolean | null;
}

export interface ModelSettingsRequest {
  model?: string | null;
  temperature?: number | null;
  max_tokens?: number | null;
  custom_instruction?: string | null;
}

export interface SecuritySettingsRequest {
  secure_embed_enabled?: boolean | null;
  allowed_domains?: string[];
  embed_secret?: string | null;
}

export interface GuardrailsRequest {
  confidence_threshold?: number | null;
  fallback_messages?: import('@/features/chatbot/hooks/useChatbotForm').FallbackMessages | null;
  topic_restrictions?: import('@/features/chatbot/hooks/useChatbotForm').TopicConfig | null;
  threshold_config?: import('@/features/chatbot/hooks/useChatbotForm').ThresholdConfig | null;
}

export interface HandoffRequest {
  handoff_enabled?: boolean | null;
  handoff_type?: string | null;
  handoff_config?: import('@/features/chatbot/hooks/useChatbotForm').HandoffConfig | null;
}

export interface RefreshRequest {
  refresh_policy?: string | null;
  refresh_frequency?: string | null;
}

export interface ScrapingConfigRequest {
  include_paths?: string[] | null;
  exclude_paths?: string[] | null;
  selector_whitelist?: string[] | null;
  discovery_mode?: string | null;
}

export interface CreateChatbotRequest {
  name: string;
  description?: string | null;
  custom_instruction?: string | null;
  language?: string | null;
  model?: string | null;
  temperature?: number | null;
  max_tokens?: number | null;
  theme_color?: string | null;
  welcome_message?: string | null;
  position?: string | null;
  bot_message_color?: string | null;
  user_message_color?: string | null;
  bot_message_text_color?: string | null;
  user_message_text_color?: string | null;
  chat_font_family?: string | null;
  chat_header_color?: string | null;
  chat_header_text_color?: string | null;
  chat_background_color?: string | null;
  bubble_radius?: string | null;
  input_background_color?: string | null;
  input_text_color?: string | null;
  send_button_color?: string | null;
  bot_icon?: string | null;
  bot_display_name?: string | null;
  secure_embed_enabled?: boolean | null;
  allowed_domains?: string;
  embed_secret?: string | null;
  suggested_questions?: string[] | null;
  suggestions_enabled?: boolean | null;
  include_paths?: string[] | null;
  exclude_paths?: string[] | null;
  selector_whitelist?: string[] | null;
  discovery_mode?: string | null;
  refresh_policy?: string | null;
  refresh_frequency?: string | null;
  hide_branding?: boolean | null;
  custom_branding?: import('@/features/chatbot/hooks/useChatbotForm').CustomBranding | null;
  confidence_threshold?: number | null;
  fallback_messages?: import('@/features/chatbot/hooks/useChatbotForm').FallbackMessages | null;
  topic_restrictions?: import('@/features/chatbot/hooks/useChatbotForm').TopicConfig | null;
  handoff_enabled?: boolean | null;
  handoff_type?: string | null;
  handoff_config?: import('@/features/chatbot/hooks/useChatbotForm').HandoffConfig | null;
  threshold_config?: import('@/features/chatbot/hooks/useChatbotForm').ThresholdConfig | null;
}
