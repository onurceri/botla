import { useState } from 'react'

type CustomBranding = {
  logo_url?: string
  text?: string
  link?: string
}

export function useChatbotForm() {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('')
  const [temperature, setTemperature] = useState(0.7)
  const [maxTokens, setMaxTokens] = useState(512)
  const [themeColor, setThemeColor] = useState('#a78bfa')
  const [welcomeMessage, setWelcomeMessage] = useState('Merhaba! Size nasıl yardımcı olabilirim?')
  const [position, setPosition] = useState('bottom-right')
  const [botMessageColor, setBotMessageColor] = useState('#fcfcfd')
  const [userMessageColor, setUserMessageColor] = useState('#2e408a')
  const [botMessageTextColor, setBotMessageTextColor] = useState('#030303')
  const [userMessageTextColor, setUserMessageTextColor] = useState('#ffffff')
  const [chatFontFamily, setChatFontFamily] = useState('Inter, sans-serif')
  const [chatHeaderColor, setChatHeaderColor] = useState('#3b82f6')
  const [chatHeaderTextColor, setChatHeaderTextColor] = useState('#ffffff')
  const [chatBackgroundColor, setChatBackgroundColor] = useState('#FFF5E6')
  const [botIcon, setBotIcon] = useState('')
  const [botDisplayName, setBotDisplayName] = useState('')
  const [secureEmbedEnabled, setSecureEmbedEnabled] = useState(false)
  const [allowedDomains, setAllowedDomains] = useState('')
  const [embedSecret, setEmbedSecret] = useState('')
  const [suggestionsEnabled, setSuggestionsEnabled] = useState(false)
  const [suggestedQuestions, setSuggestedQuestions] = useState<string[]>([])
  const [includePaths, setIncludePaths] = useState<string[]>([])
  const [excludePaths, setExcludePaths] = useState<string[]>([])
  const [selectorWhitelist, setSelectorWhitelist] = useState<string[]>([])
  const [discoveryMode, setDiscoveryMode] = useState<'auto' | 'pending' | 'disabled'>('auto')
  const [refreshPolicy, setRefreshPolicy] = useState<'manual' | 'auto'>('manual')
  const [refreshFrequency, setRefreshFrequency] = useState<'daily' | 'weekly' | 'monthly' | null>(null)
  const [nextRefreshAt, setNextRefreshAt] = useState<string | null>(null)
  const [lastRefreshAt, setLastRefreshAt] = useState<string | null>(null)
  const [hideBranding, setHideBranding] = useState(false)
  const [customBranding, setCustomBranding] = useState<CustomBranding | null>(null)

  function setFromServer(data: any) {
    setName(data.name || '')
    setDescription(data.description || '')
    setSystemPrompt(data.system_prompt || '')
    setTemperature(data.temperature ?? 0.7)
    setMaxTokens(data.max_tokens ?? 512)
    setThemeColor(data.theme_color || '#a78bfa')
    setWelcomeMessage(data.welcome_message || '')
    setPosition(data.position || 'bottom-right')
    setBotMessageColor(data.bot_message_color || '#fcfcfd')
    setUserMessageColor(data.user_message_color || '#2e408a')
    setBotMessageTextColor(data.bot_message_text_color || '#030303')
    setUserMessageTextColor(data.user_message_text_color || '#ffffff')
    setChatFontFamily(data.chat_font_family || 'Inter, sans-serif')
    setChatHeaderColor(data.chat_header_color || '#3b82f6')
    setChatHeaderTextColor(data.chat_header_text_color || '#ffffff')
    setChatBackgroundColor(data.chat_background_color || '#FFF5E6')
    setBotIcon(data.bot_icon || '')
    setBotDisplayName(data.bot_display_name || '')
    setAllowedDomains(data.allowed_domains || '')
    setEmbedSecret(data.embed_secret || '')
    setSecureEmbedEnabled(!!data.secure_embed_enabled)
    setSuggestionsEnabled(!!data.suggestions_enabled)
    setSuggestedQuestions(Array.isArray(data.suggested_questions) ? data.suggested_questions : [])
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
  }

  function validate() {
    return !!name.trim()
  }

  function buildPayload() {
    return {
      name,
      description,
      system_prompt: systemPrompt,
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
      custom_branding: customBranding,
    }
  }

  return {
    name, setName,
    description, setDescription,
    systemPrompt, setSystemPrompt,
    temperature, setTemperature,
    maxTokens, setMaxTokens,
    themeColor, setThemeColor,
    welcomeMessage, setWelcomeMessage,
    position, setPosition,
    botMessageColor, setBotMessageColor,
    userMessageColor, setUserMessageColor,
    botMessageTextColor, setBotMessageTextColor,
    userMessageTextColor, setUserMessageTextColor,
    chatFontFamily, setChatFontFamily,
    chatHeaderColor, setChatHeaderColor,
    chatHeaderTextColor, setChatHeaderTextColor,
    chatBackgroundColor, setChatBackgroundColor,
    botIcon, setBotIcon,
    botDisplayName, setBotDisplayName,
    secureEmbedEnabled, setSecureEmbedEnabled,
    allowedDomains, setAllowedDomains,
    embedSecret, setEmbedSecret,
    suggestionsEnabled, setSuggestionsEnabled,
    suggestedQuestions, setSuggestedQuestions,
    includePaths, setIncludePaths,
    excludePaths, setExcludePaths,
    selectorWhitelist, setSelectorWhitelist,
    discoveryMode, setDiscoveryMode,
    refreshPolicy, setRefreshPolicy,
    refreshFrequency, setRefreshFrequency,
    nextRefreshAt,
    lastRefreshAt,
    hideBranding, setHideBranding,
    customBranding, setCustomBranding,
    setFromServer,
    validate,
    buildPayload,
  }
}
