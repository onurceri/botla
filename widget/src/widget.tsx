import { render } from 'preact'
import { WidgetApp } from './widgetApp'
import styles from './styles.css?raw'

const defaultHostId = 'chatbot-widget-host'

function currentParams() {
  const p = (window as any).__CBW_PARAMS
  if (p && typeof p === 'object') return new URLSearchParams(p)
  
  // document.currentScript is null for ES modules, so find script by src or data-bot
  let script = document.currentScript as HTMLScriptElement | null
  if (!script) {
    // Find widget script by looking for data-bot attribute or widget.js/widget.tsx in src
    script = document.querySelector('script[data-bot]') as HTMLScriptElement | null
    if (!script) {
      script = document.querySelector('script[src*="widget.js"], script[src*="widget.tsx"]') as HTMLScriptElement | null
    }
  }
  
  const src = script?.src || window.location.href
  const u = new URL(src)
  const params = u.searchParams
  if (script && script.dataset.bot) {
    params.set('chatbot-id', script.dataset.bot)
  }
  return params
}

function ensureHost(): HTMLElement {
  const id = defaultHostId
  let host = document.getElementById(id)
  if (!host) {
    host = document.createElement('div')
    host.id = id
    document.body.appendChild(host)
  }
  return host
}

function injectStyles(shadow: ShadowRoot) {
  const style = document.createElement('style')
  style.textContent = styles
  shadow.appendChild(style)
}

export function mount() {
  const params = currentParams()
  const chatbotId = params.get('chatbot-id')
  if (!chatbotId) {
    console.error('ChatbotWidget: chatbot-id is required')
    return
  }
  // Use VITE_API_BASE_URL env var as default if available, otherwise undefined (which defaults to relative)
  const defaultApiBase = import.meta.env.VITE_API_BASE_URL || undefined
  const apiBase = params.get('api-base') || defaultApiBase
  const themeColor = params.get('color') || undefined
  const welcome = params.get('welcome') || undefined
  const embedTokenUrl = params.get('embed-token-url') || undefined
  const captchaSiteKey = params.get('captcha-site-key') || undefined
  const autoOpen = params.get('auto-open') === '1' || params.get('auto-open') === 'true'
  const useOverrides = params.get('use-url-overrides') === '1' || params.get('use-url-overrides') === 'true'
  const headerColor = params.get('header-color') || undefined
  const headerTextColor = params.get('header-text-color') || undefined
  const botMessageColor = params.get('bot-message-color') || undefined
  const botMessageTextColor = params.get('bot-message-text-color') || undefined
  const userMessageColor = params.get('user-message-color') || undefined
  const userMessageTextColor = params.get('user-message-text-color') || undefined
  const fontFamily = params.get('font-family') || undefined
  const positionRaw = params.get('position') || undefined
  const position = positionRaw === 'bottom-left' ? 'bottom-left' : positionRaw === 'bottom-right' ? 'bottom-right' : undefined
  const botName = params.get('bot-name') || undefined
  const botIcon = params.get('bot-icon') || undefined
  const panelHeight = params.get('panel-height') || undefined
  const panelWidth = params.get('panel-width') || undefined
  const panelBg = params.get('panel-bg-color') || undefined
  const inputBg = params.get('input-bg-color') || undefined
  const inputText = params.get('input-text-color') || undefined
  const chatBg = params.get('chat-bg-color') || undefined
  const bubbleRadius = params.get('bubble-radius') || undefined
  const sendButtonColor = params.get('send-button-color') || undefined
  const resetSession = params.get('reset-session') === '1' || params.get('reset-session') === 'true'
  const sessionIdOverride = params.get('session-id') || undefined
  
  // Playground preview params
  const positionStrategyRaw = params.get('position-strategy') || undefined
  const positionStrategy = positionStrategyRaw === 'absolute' ? 'absolute' : 'fixed'
  const previewMode = params.get('preview-mode') === '1' || params.get('preview-mode') === 'true'
  
  // Parse suggestions from JSON if provided
  let suggestionsOverride: string[] | undefined
  const suggestionsRaw = params.get('suggestions')
  if (suggestionsRaw) {
    try { suggestionsOverride = JSON.parse(suggestionsRaw) } catch {}
  }
  
  // Parse branding options
  const hideBrandingOverride = params.get('hide-branding') === '1' ? true : params.get('hide-branding') === '0' ? false : undefined
  let customBrandingOverride: { logo_url?: string; text?: string; link?: string } | undefined
  const customBrandingRaw = params.get('custom-branding')
  if (customBrandingRaw) {
    try { customBrandingOverride = JSON.parse(customBrandingRaw) } catch {}
  }

  const host = ensureHost()
  const shadow = host.shadowRoot || host.attachShadow({ mode: 'open' })
  injectStyles(shadow)
  const root = document.createElement('div')
  shadow.appendChild(root)
  
  // Global state tracker for preserving open state during config updates
  let currentOpenState: boolean | undefined = undefined
  const setOpenState = (isOpen: boolean) => { currentOpenState = isOpen }
  
  const props = { chatbotId, apiBase, themeColor, headerColor, headerTextColor, botMessageColor, botMessageTextColor, userMessageColor, userMessageTextColor, fontFamily, position: position as "bottom-left" | "bottom-right" | undefined, botNameOverride: botName, botIconOverride: botIcon, panelHeight, panelWidth, panelBg, inputBg, inputText, chatBg, bubbleRadius, sendButtonColor, useOverrides, welcome, embedTokenUrl, captchaSiteKey, autoOpen, resetSession, sessionIdOverride, positionStrategy: positionStrategy as "fixed" | "absolute", suggestions: suggestionsOverride, hideBrandingOverride, customBrandingOverride, previewMode, onOpenChange: setOpenState }
  
  render(<WidgetApp {...props} />, root)

  // Allowed origins for postMessage (Playground support)
  const ALLOWED_ORIGINS = [
    import.meta.env.VITE_DASHBOARD_URL,
    import.meta.env.VITE_API_BASE_URL,
  ].filter(Boolean) as string[]

  // Listen for live config updates (Playground support)
  window.addEventListener('message', (event) => {
    // Origin validation - block unauthorized origins
    if (ALLOWED_ORIGINS.length > 0 && !ALLOWED_ORIGINS.some(origin => event.origin.startsWith(origin))) {
      console.warn('[Widget] Unauthorized postMessage origin:', event.origin)
      return
    }
    
    if (event.data?.type === 'WIDGET_CONFIG') {
      const newConfig = event.data.config
      
      // Preserve the current open state if widget is already mounted
      const preserveOpen = currentOpenState !== undefined ? currentOpenState : (newConfig['auto-open'] === '1')
      
      const updatedProps = {
        ...props,
        chatbotId: newConfig['chatbot-id'] || chatbotId,
        apiBase: newConfig['api-base'] || apiBase,
        themeColor: newConfig['color'] || themeColor,
        headerColor: newConfig['header-color'] || headerColor,
        headerTextColor: newConfig['header-text-color'] || headerTextColor,
        botMessageColor: newConfig['bot-message-color'] || botMessageColor,
        botMessageTextColor: newConfig['bot-message-text-color'] || botMessageTextColor,
        userMessageColor: newConfig['user-message-color'] || userMessageColor,
        userMessageTextColor: newConfig['user-message-text-color'] || userMessageTextColor,
        fontFamily: newConfig['font-family'] || fontFamily,
        position: (newConfig['position'] === 'bottom-left' ? 'bottom-left' : newConfig['position'] === 'bottom-right' ? 'bottom-right' : position) as "bottom-left" | "bottom-right" | undefined,
        botNameOverride: newConfig['bot-name'] || botName,
        botIconOverride: newConfig['bot-icon'] || botIcon,
        welcome: newConfig['welcome'] || welcome,
        chatBg: newConfig['chat-bg-color'] || chatBg,
        panelBg: newConfig['panel-bg-color'] || panelBg,
        inputBg: newConfig['input-bg-color'] || inputBg,
        inputText: newConfig['input-text-color'] || inputText,
        bubbleRadius: newConfig['bubble-radius'] || bubbleRadius,
        sendButtonColor: newConfig['send-button-color'] || sendButtonColor,
        panelHeight: newConfig['panel-height'] || panelHeight,
        panelWidth: newConfig['panel-width'] || panelWidth,
        useOverrides: true, // Always override in playground
        autoOpen: preserveOpen, // Use preserved open state instead of auto-open config
        sessionIdOverride: newConfig['session-id'] || sessionIdOverride,
        hideBrandingOverride: newConfig['hide-branding'] === '1',
        positionStrategy: (newConfig['position-strategy'] || positionStrategy) as "fixed" | "absolute",
        previewMode: newConfig['preview-mode'] === '1',
        onOpenChange: setOpenState, // Keep the callback
      }
      
      if (newConfig['suggestions']) {
        try { updatedProps.suggestions = JSON.parse(newConfig['suggestions']) } catch {}
      }
      
      if (newConfig['custom-branding']) {
        try { updatedProps.customBrandingOverride = JSON.parse(newConfig['custom-branding']) } catch {}
      }

      render(<WidgetApp {...updatedProps} />, root)
      
      // Notify parent that config was applied
      window.parent.postMessage({ type: 'WIDGET_CONFIG_APPLIED' }, '*')
    }
  })
}

export function unmount() {
  const host = document.getElementById(defaultHostId)
  if (!host || !host.shadowRoot) return
  host.shadowRoot.innerHTML = ''
}

// Skip auto-mount if skip-auto-mount flag is set (preview mode waits for postMessage config)
const skipMount = currentParams().get('skip-auto-mount') === '1'
if (!skipMount) {
  try { mount() } catch (e) {
    const params = currentParams()
    const debug = params.get('debug') === '1' || params.get('debug') === 'true'
    if (debug) {
      console.error('ChatbotWidget mount error:', e)
      try {
        const m = document.createElement('div')
        m.style.position = 'fixed'
        m.style.bottom = '8px'
        m.style.right = '8px'
        m.style.background = '#fee2e2'
        m.style.color = '#b91c1c'
        m.style.padding = '8px 10px'
        m.style.borderRadius = '8px'
        m.style.fontSize = '12px'
        m.style.zIndex = '2147483647'
        m.textContent = 'Widget failed to mount. See console.'
        document.body.appendChild(m)
      } catch {}
    }
  }
}

declare global { interface Window { ChatbotWidget: { mount: typeof mount; unmount: typeof unmount } } }
// @ts-ignore
window.ChatbotWidget = { mount, unmount }
