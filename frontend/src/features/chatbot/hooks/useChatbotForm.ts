import { useState } from 'react'

export type CustomBranding = {
  logo_url?: string
  text?: string
  link?: string
}

export type FallbackMessages = {
  no_info_found?: string
  error_message?: string
  handoff_message?: string
}

export type TopicConfig = {
  allowed_topics?: string[]
  blocked_topics?: string[]
  blocked_message?: string
}

export type HandoffConfig = {
  email_to?: string
  email_subject?: string
}

export type ThresholdConfig = {
  high_threshold: number
  medium_threshold: number
  fallback_mode: 'smart' | 'static' | 'escalate'
  show_confidence_warning: boolean
}

export const DEFAULT_THRESHOLD_CONFIG: ThresholdConfig = {
  high_threshold: 0.5,
  medium_threshold: 0.3,
  fallback_mode: 'smart',
  show_confidence_warning: true,
}

export function useChatbotForm() {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [customInstruction, setCustomInstruction] = useState('')
  const [model, setModel] = useState('gpt-4o-mini')
  const [temperature, setTemperature] = useState(0.7)
  const [maxTokens, setMaxTokens] = useState(512)
  const [themeColor, setThemeColor] = useState('rgba(255, 174, 0, 1)')
  const [welcomeMessage, setWelcomeMessage] = useState('Merhaba! Size nasıl yardımcı olabilirim?')
  const [position, setPosition] = useState('bottom-right')
  const [botMessageColor, setBotMessageColor] = useState('rgba(252, 252, 253, 1)')
  const [userMessageColor, setUserMessageColor] = useState('rgba(250, 171, 0, 0.91)')
  const [botMessageTextColor, setBotMessageTextColor] = useState('rgba(0, 0, 0, 1)')
  const [userMessageTextColor, setUserMessageTextColor] = useState('rgba(255, 255, 255, 1)')
  const [chatFontFamily, setChatFontFamily] = useState('Inter, sans-serif')
  const [chatHeaderColor, setChatHeaderColor] = useState('rgba(242, 167, 36, 1)')
  const [chatHeaderTextColor, setChatHeaderTextColor] = useState('rgba(247, 241, 241, 1)')
  const [chatBackgroundColor, setChatBackgroundColor] = useState('rgba(255, 245, 230, 1)')
  const [bubbleRadius, setBubbleRadius] = useState('22px')
  const [inputBackgroundColor, setInputBackgroundColor] = useState('rgba(255, 255, 255, 0.5)')
  const [inputTextColor, setInputTextColor] = useState('rgba(28, 28, 30, 1)')
  const [sendButtonColor, setSendButtonColor] = useState('rgba(246, 140, 0, 1)')
  const [botIcon, setBotIcon] = useState('')
  const [botDisplayName, setBotDisplayName] = useState('')
  const [secureEmbedEnabled, setSecureEmbedEnabled] = useState(false)
  const [allowedDomains, setAllowedDomains] = useState('')
  const [embedSecret, setEmbedSecret] = useState('')
  const [suggestionsEnabled, setSuggestionsEnabled] = useState(false)
  const [suggestedQuestions, setSuggestedQuestions] = useState<string[]>([])
  const [manualQuestions, setManualQuestions] = useState<string[]>([])
  const [includePaths, setIncludePaths] = useState<string[]>([])
  const [excludePaths, setExcludePaths] = useState<string[]>([])
  const [selectorWhitelist, setSelectorWhitelist] = useState<string[]>([])
  const [discoveryMode, setDiscoveryMode] = useState<'auto' | 'pending' | 'disabled'>('auto')
  const [refreshPolicy, setRefreshPolicy] = useState<'manual' | 'auto'>('manual')
  const [refreshFrequency, setRefreshFrequency] = useState<'daily' | 'weekly' | 'monthly' | null>(
    null,
  )
  const [nextRefreshAt, setNextRefreshAt] = useState<string | null>(null)
  const [lastRefreshAt, setLastRefreshAt] = useState<string | null>(null)
  const [hideBranding, setHideBranding] = useState(false)
  const [customBranding, setCustomBranding] = useState<CustomBranding | null>(null)
  const [confidenceThreshold, setConfidenceThreshold] = useState(0.7)
  const [thresholdConfig, setThresholdConfig] = useState<ThresholdConfig>(DEFAULT_THRESHOLD_CONFIG)
  const [fallbackMessages, setFallbackMessages] = useState<FallbackMessages | null>(null)
  const [topicRestrictions, setTopicRestrictions] = useState<TopicConfig | null>(null)
  const [handoffEnabled, setHandoffEnabled] = useState(false)
  const [handoffType, setHandoffType] = useState<'email'>('email')
  const [handoffConfig, setHandoffConfig] = useState<HandoffConfig | null>(null)

  function setFromServer(data: any) {
    setName(data.name || '')
    setDescription(data.description || '')
    setCustomInstruction(data.custom_instruction || '')
    setModel(data.model || 'gpt-4o-mini')
    setTemperature(data.temperature ?? 0.7)
    setMaxTokens(data.max_tokens ?? 512)
    setThemeColor(data.theme_color || 'rgba(167, 139, 250, 1)')
    setWelcomeMessage(data.welcome_message || '')
    setPosition(data.position || 'bottom-right')
    setBotMessageColor(data.bot_message_color || 'rgba(252, 252, 253, 1)')
    setUserMessageColor(data.user_message_color || 'rgba(46, 64, 138, 1)')
    setBotMessageTextColor(data.bot_message_text_color || 'rgba(3, 3, 3, 1)')
    setUserMessageTextColor(data.user_message_text_color || 'rgba(255, 255, 255, 1)')
    setChatFontFamily(data.chat_font_family || 'Inter, sans-serif')
    setChatHeaderColor(data.chat_header_color || 'rgba(59, 130, 246, 1)')
    setChatHeaderTextColor(data.chat_header_text_color || 'rgba(255, 255, 255, 1)')
    setChatBackgroundColor(data.chat_background_color || 'rgba(255, 245, 230, 1)')
    setBubbleRadius(data.bubble_radius || '22px')
    setInputBackgroundColor(data.input_background_color || 'rgba(237, 237, 237, 1)')
    setInputTextColor(data.input_text_color || 'rgba(0, 0, 0, 1)')
    setSendButtonColor(data.send_button_color || 'rgba(235, 184, 0, 1)')
    setBotIcon(data.bot_icon || '')
    setBotDisplayName(data.bot_display_name || '')
    // Sanitize allowed_domains: remove quotes from each domain
    const rawDomains = data.allowed_domains || ''
    const sanitizedDomains = rawDomains
      .split(',')
      .map((d: string) => d.trim().replace(/^["']|["']$/g, ''))
      .filter((d: string) => d)
      .join(', ')
    setAllowedDomains(sanitizedDomains)
    setEmbedSecret(data.embed_secret || '')
    setSecureEmbedEnabled(!!data.secure_embed_enabled)
    setSuggestionsEnabled(!!data.suggestions_enabled)
    setSuggestedQuestions(Array.isArray(data.suggested_questions) ? data.suggested_questions : [])
    setManualQuestions(Array.isArray(data.manual_questions) ? data.manual_questions : [])
    setIncludePaths(Array.isArray(data.include_paths) ? data.include_paths : [])
    setExcludePaths(Array.isArray(data.exclude_paths) ? data.exclude_paths : [])
    setSelectorWhitelist(Array.isArray(data.selector_whitelist) ? data.selector_whitelist : [])
    setDiscoveryMode(data.discovery_mode || 'auto')
    setRefreshPolicy(data.refresh_policy || 'manual')
    setRefreshFrequency(data.refresh_frequency || null)
    setNextRefreshAt(data.next_refresh_at || null)
    setLastRefreshAt(data.last_refresh_at || null)
    setHideBranding(!!data.hide_branding)
    setCustomBranding(data.custom_branding || null)
    setConfidenceThreshold(data.confidence_threshold ?? 0.7)
    setThresholdConfig(data.threshold_config || DEFAULT_THRESHOLD_CONFIG)
    setFallbackMessages(data.fallback_messages || null)
    setTopicRestrictions(data.topic_restrictions || null)
    setHandoffEnabled(!!data.handoff_enabled)
    setHandoffType(data.handoff_type || 'email')
    setHandoffConfig(data.handoff_config || null)
  }

  function validate() {
    return !!name.trim()
  }

  function buildPayload() {
    return {
      name,
      description,
      custom_instruction: customInstruction,
      model,
      temperature,
      max_tokens: maxTokens,
      theme_color: themeColor,
      welcome_message: welcomeMessage,
      position,
      bot_message_color: botMessageColor,
      user_message_color: userMessageColor,
      bot_message_text_color: botMessageTextColor,
      user_message_text_color: userMessageTextColor,
      chat_font_family: chatFontFamily,
      chat_header_color: chatHeaderColor,
      chat_header_text_color: chatHeaderTextColor,
      chat_background_color: chatBackgroundColor,
      bubble_radius: bubbleRadius,
      input_background_color: inputBackgroundColor,
      input_text_color: inputTextColor,
      send_button_color: sendButtonColor,
      bot_icon: botIcon,
      bot_display_name: botDisplayName,
      secure_embed_enabled: secureEmbedEnabled,
      allowed_domains: secureEmbedEnabled ? allowedDomains : undefined,
      embed_secret: secureEmbedEnabled ? embedSecret : undefined,
      suggestions_enabled: suggestionsEnabled,
      suggested_questions: suggestedQuestions,
      include_paths: includePaths,
      exclude_paths: excludePaths,
      selector_whitelist: selectorWhitelist,
      discovery_mode: discoveryMode,
      refresh_policy: refreshPolicy,
      refresh_frequency: refreshFrequency,
      hide_branding: hideBranding,
      custom_branding: hideBranding ? customBranding : null,
      confidence_threshold: confidenceThreshold,
      threshold_config: thresholdConfig,
      fallback_messages: fallbackMessages,
      topic_restrictions: topicRestrictions,
      handoff_enabled: handoffEnabled,
      handoff_type: handoffType,
      handoff_config: handoffEnabled ? handoffConfig : null,
    }
  }

  function buildOverviewPayload() {
    return {
      name,
      custom_instruction: customInstruction,
      model,
      temperature,
      max_tokens: maxTokens,
    }
  }

  function buildGuardrailsPayload() {
    return {
      threshold_config: thresholdConfig,
      fallback_messages: fallbackMessages,
      topic_restrictions: topicRestrictions,
    }
  }

  function buildHandoffPayload() {
    return {
      handoff_enabled: handoffEnabled,
      handoff_type: handoffType,
      handoff_config: handoffEnabled ? handoffConfig : null,
    }
  }

  function buildSuggestionsPayload() {
    return {
      suggestions_enabled: suggestionsEnabled,
      suggested_questions: suggestedQuestions,
    }
  }

  function buildAppearancePayload() {
    return {
      bot_display_name: botDisplayName,
      bot_icon: botIcon,
      welcome_message: welcomeMessage,
      position,
      chat_font_family: chatFontFamily,
      theme_color: themeColor,
      chat_background_color: chatBackgroundColor,
      chat_header_color: chatHeaderColor,
      chat_header_text_color: chatHeaderTextColor,
      bot_message_color: botMessageColor,
      bot_message_text_color: botMessageTextColor,
      user_message_color: userMessageColor,
      user_message_text_color: userMessageTextColor,
      bubble_radius: bubbleRadius,
      input_background_color: inputBackgroundColor,
      input_text_color: inputTextColor,
      send_button_color: sendButtonColor,
      hide_branding: hideBranding,
      custom_branding: hideBranding ? customBranding : null,
    }
  }

  function buildConnectPayload() {
    return {
      secure_embed_enabled: secureEmbedEnabled,
      allowed_domains: secureEmbedEnabled ? allowedDomains : undefined,
      embed_secret: secureEmbedEnabled ? embedSecret : undefined,
    }
  }

  function buildSourceSettingsPayload() {
    return {
      discovery_mode: discoveryMode,
      refresh_policy: refreshPolicy,
      refresh_frequency: refreshFrequency,
      include_paths: includePaths,
      exclude_paths: excludePaths,
      selector_whitelist: selectorWhitelist,
    }
  }

  return {
    name,
    setName,
    description,
    setDescription,
    customInstruction,
    setCustomInstruction,
    temperature,
    setTemperature,
    maxTokens,
    setMaxTokens,
    themeColor,
    setThemeColor,
    welcomeMessage,
    setWelcomeMessage,
    position,
    setPosition,
    botMessageColor,
    setBotMessageColor,
    userMessageColor,
    setUserMessageColor,
    botMessageTextColor,
    setBotMessageTextColor,
    userMessageTextColor,
    setUserMessageTextColor,
    chatFontFamily,
    setChatFontFamily,
    chatHeaderColor,
    setChatHeaderColor,
    chatHeaderTextColor,
    setChatHeaderTextColor,
    chatBackgroundColor,
    setChatBackgroundColor,
    bubbleRadius,
    setBubbleRadius,
    inputBackgroundColor,
    setInputBackgroundColor,
    inputTextColor,
    setInputTextColor,
    sendButtonColor,
    setSendButtonColor,
    botIcon,
    setBotIcon,
    botDisplayName,
    setBotDisplayName,
    secureEmbedEnabled,
    setSecureEmbedEnabled,
    allowedDomains,
    setAllowedDomains,
    embedSecret,
    setEmbedSecret,
    suggestionsEnabled,
    setSuggestionsEnabled,
    suggestedQuestions,
    setSuggestedQuestions,
    manualQuestions,
    setManualQuestions,
    includePaths,
    setIncludePaths,
    excludePaths,
    setExcludePaths,
    selectorWhitelist,
    setSelectorWhitelist,
    discoveryMode,
    setDiscoveryMode,
    refreshPolicy,
    setRefreshPolicy,
    refreshFrequency,
    setRefreshFrequency,
    nextRefreshAt,
    lastRefreshAt,
    hideBranding,
    setHideBranding,
    customBranding,
    setCustomBranding,
    buildOverviewPayload,
    buildGuardrailsPayload,
    buildHandoffPayload,
    buildSuggestionsPayload,
    buildAppearancePayload,
    buildConnectPayload,
    buildSourceSettingsPayload,
    setFromServer,
    validate,
    buildPayload,
    model,
    setModel,
    confidenceThreshold,
    setConfidenceThreshold,
    thresholdConfig,
    setThresholdConfig,
    fallbackMessages,
    setFallbackMessages,
    topicRestrictions,
    setTopicRestrictions,
    handoffEnabled,
    setHandoffEnabled,
    handoffType,
    setHandoffType,
    handoffConfig,
    setHandoffConfig,
  }
}
