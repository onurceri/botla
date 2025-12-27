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
    panelWidth,
    autoOpen,
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
  }, [sendConfig, refreshKey, buildConfig])

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
    <div className="w-full h-full relative overflow-hidden">
      {/* Decorative Website Mockup Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-slate-50 to-slate-100 pointer-events-none">
        {/* Browser Chrome */}
        <div className="h-8 bg-slate-200/80 border-b border-slate-300/50 flex items-center px-3 gap-2">
          <div className="flex gap-1.5">
            <div className="w-2.5 h-2.5 rounded-full bg-red-400/60" />
            <div className="w-2.5 h-2.5 rounded-full bg-yellow-400/60" />
            <div className="w-2.5 h-2.5 rounded-full bg-green-400/60" />
          </div>
          <div className="flex-1 mx-4">
            <div className="h-4 bg-white/80 rounded-md max-w-[200px] mx-auto shadow-inner" />
          </div>
        </div>

        {/* Website Header */}
        <div className="h-12 bg-white/60 border-b border-slate-200/60 flex items-center justify-between px-4">
          <div className="flex items-center gap-3">
            <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-primary/30 to-orange-400/30" />
            <div className="w-20 h-3 bg-slate-300/60 rounded-full" />
          </div>
          <div className="flex gap-4">
            <div className="w-12 h-2.5 bg-slate-200/80 rounded-full" />
            <div className="w-14 h-2.5 bg-slate-200/80 rounded-full" />
            <div className="w-10 h-2.5 bg-slate-200/80 rounded-full" />
          </div>
        </div>

        {/* Hero Section */}
        <div className="px-6 py-8">
          <div className="max-w-[200px] mx-auto text-center space-y-3">
            <div className="w-3/4 h-4 bg-slate-300/50 rounded-full mx-auto" />
            <div className="w-full h-3 bg-slate-200/50 rounded-full" />
            <div className="w-2/3 h-3 bg-slate-200/50 rounded-full mx-auto" />
            <div className="w-20 h-6 bg-primary/20 rounded-lg mx-auto mt-4" />
          </div>
        </div>

        {/* Content Grid */}
        <div className="px-4 grid grid-cols-3 gap-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-white/40 rounded-xl p-3 space-y-2 border border-slate-200/30">
              <div className="w-full h-12 bg-slate-200/40 rounded-lg" />
              <div className="w-3/4 h-2 bg-slate-200/50 rounded-full" />
              <div className="w-1/2 h-2 bg-slate-200/40 rounded-full" />
            </div>
          ))}
        </div>

        {/* Additional Content */}
        <div className="px-4 mt-4 space-y-2">
          <div className="w-2/3 h-3 bg-slate-200/40 rounded-full" />
          <div className="w-full h-2 bg-slate-200/30 rounded-full" />
          <div className="w-5/6 h-2 bg-slate-200/30 rounded-full" />
        </div>
      </div>

      {/* Widget iframe - on top of the mockup */}
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
