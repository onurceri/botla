import { useRef, useEffect, useCallback } from 'react'

type Props = {
  id: string
  themeColor: string
  chatHeaderColor: string
  chatHeaderTextColor: string
  botMessageColor: string
  botMessageTextColor: string
  userMessageColor: string
  userMessageTextColor: string
  inputBackgroundColor: string
  inputTextColor: string
  sendButtonColor: string
  chatFontFamily: string
  position: string
  botDisplayName: string
  botIcon: string
  chatBackgroundColor: string
  bubbleRadius: string
  welcomeMessage: string
  previewOpen: boolean
  sessionId: string
  suggestionsEnabled: boolean
  suggestedQuestions: string[]
  refreshKey?: number
  hideBranding?: boolean
  customBranding?: { logo_url?: string; text?: string; link?: string } | null
  autoOpen?: boolean
  panelHeight?: string
  panelWidth?: string
}

// Get widget URL from environment, fallback to localhost for development
const WIDGET_URL = import.meta.env.VITE_WIDGET_URL || 'http://localhost:5174'
const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

/**
 * PlaygroundPreview - Embeds the actual widget via iframe with postMessage config updates
 *
 * This component loads the widget in an isolated iframe and sends configuration
 * updates via postMessage. This ensures the preview matches the production widget exactly.
 */
export default function PlaygroundPreview(props: Props) {
  const {
    id,
    themeColor,
    chatHeaderColor,
    chatHeaderTextColor,
    botMessageColor,
    botMessageTextColor,
    userMessageColor,
    userMessageTextColor,
    inputBackgroundColor,
    inputTextColor,
    sendButtonColor,
    chatFontFamily,
    position,
    botDisplayName,
    botIcon,
    chatBackgroundColor,
    bubbleRadius,
    welcomeMessage,
    sessionId,
    suggestionsEnabled,
    suggestedQuestions,
    refreshKey,
    hideBranding,
    customBranding,
    autoOpen = false,
    panelHeight,
    panelWidth,
  } = props

  const iframeRef = useRef<HTMLIFrameElement>(null)
  const configSentRef = useRef(false)

  // Build widget config from props
  const buildConfig = useCallback(() => {
    const config: Record<string, string> = {
      'chatbot-id': id,
      'api-base': API_BASE,
      'session-id': sessionId,
      'auto-open': autoOpen ? '1' : '0',
    }

    if (panelHeight) config['panel-height'] = panelHeight
    if (panelWidth) config['panel-width'] = panelWidth

    // Theme & colors
    if (themeColor) config['color'] = themeColor
    if (chatHeaderColor) config['header-color'] = chatHeaderColor
    if (chatHeaderTextColor) config['header-text-color'] = chatHeaderTextColor
    if (botMessageColor) config['bot-message-color'] = botMessageColor
    if (botMessageTextColor) config['bot-message-text-color'] = botMessageTextColor
    if (userMessageColor) config['user-message-color'] = userMessageColor
    if (userMessageTextColor) config['user-message-text-color'] = userMessageTextColor
    if (inputBackgroundColor) config['input-bg-color'] = inputBackgroundColor
    if (inputTextColor) config['input-text-color'] = inputTextColor
    if (sendButtonColor) config['send-button-color'] = sendButtonColor
    if (chatBackgroundColor) config['chat-bg-color'] = chatBackgroundColor
    if (bubbleRadius) config['bubble-radius'] = bubbleRadius
    if (chatFontFamily) config['font-family'] = chatFontFamily

    // Position
    if (position) config['position'] = position

    // Bot identity
    if (botDisplayName) config['bot-name'] = botDisplayName
    if (botIcon) config['bot-icon'] = botIcon

    // Welcome message
    if (welcomeMessage) config['welcome'] = welcomeMessage

    // Suggestions - serialize as JSON in URL param
    if (suggestionsEnabled) {
      config['suggestions'] = JSON.stringify(suggestedQuestions || [])
    } else {
      config['suggestions'] = JSON.stringify([])
    }

    // Branding
    if (hideBranding !== undefined) {
      config['hide-branding'] = hideBranding ? '1' : '0'
    }
    if (customBranding) {
      config['custom-branding'] = JSON.stringify(customBranding)
    }

    return config
  }, [
    id,
    themeColor,
    chatHeaderColor,
    chatHeaderTextColor,
    botMessageColor,
    botMessageTextColor,
    userMessageColor,
    userMessageTextColor,
    inputBackgroundColor,
    inputTextColor,
    sendButtonColor,
    chatFontFamily,
    position,
    botDisplayName,
    botIcon,
    chatBackgroundColor,
    bubbleRadius,
    welcomeMessage,
    sessionId,
    suggestionsEnabled,
    suggestedQuestions,
    hideBranding,
    customBranding,
    panelHeight,
  ])

  // Send config to iframe
  const sendConfig = useCallback(() => {
    const iframe = iframeRef.current
    if (!iframe?.contentWindow) return

    const config = buildConfig()
    iframe.contentWindow.postMessage(
      {
        type: 'WIDGET_CONFIG',
        config,
      },
      '*',
    )
  }, [buildConfig])

  // Send config when iframe loads
  const handleIframeLoad = useCallback(() => {
    configSentRef.current = false
    // Small delay to ensure widget is ready
    setTimeout(() => {
      sendConfig()
      configSentRef.current = true
    }, 100)
  }, [sendConfig])

  // Send config updates when props change (after initial load)
  // Debounce to avoid excessive updates during typing
  useEffect(() => {
    if (!configSentRef.current) return

    const timeoutId = setTimeout(() => {
      sendConfig()
    }, 300) // 300ms debounce for live preview updates

    return () => clearTimeout(timeoutId)
  }, [sendConfig, refreshKey])

  // Listen for messages from iframe (optional - for debugging)
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type === 'WIDGET_CONFIG_APPLIED') {
        // Config applied successfully
      }
    }
    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [])

  return (
    <div className="w-full h-full relative">
      <iframe
        ref={iframeRef}
        src={`${WIDGET_URL}/preview.html`}
        onLoad={handleIframeLoad}
        title="Chatbot Önizlemesi"
        className="absolute inset-0 w-full h-full border-0 bg-transparent"
        sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
        loading="eager"
      />
    </div>
  )
}
