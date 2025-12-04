import { render } from 'preact'
import { WidgetApp } from './widgetApp'
import styles from './styles.css?raw'

const defaultHostId = 'chatbot-widget-host'

function currentParams() {
  const p = (window as any).__CBW_PARAMS
  if (p && typeof p === 'object') return new URLSearchParams(p)
  const src = (document.currentScript as HTMLScriptElement | null)?.src || window.location.href
  const u = new URL(src)
  return u.searchParams
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
  const chatbotId = params.get('chatbot-id') || 'demo'
  const apiBase = params.get('api-base') || undefined
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
  const panelBg = params.get('panel-bg-color') || undefined
  const inputBg = params.get('input-bg-color') || undefined
  const inputText = params.get('input-text-color') || undefined
  const chatBg = params.get('chat-bg-color') || undefined
  const bubbleRadius = params.get('bubble-radius') || undefined
  const sendButtonColor = params.get('send-button-color') || undefined
  const resetSession = params.get('reset-session') === '1' || params.get('reset-session') === 'true'
  const sessionIdOverride = params.get('session-id') || undefined

  const host = ensureHost()
  const shadow = host.shadowRoot || host.attachShadow({ mode: 'open' })
  injectStyles(shadow)
  const root = document.createElement('div')
  shadow.appendChild(root)
  render(<WidgetApp chatbotId={chatbotId} apiBase={apiBase} themeColor={themeColor} headerColor={headerColor} headerTextColor={headerTextColor} botMessageColor={botMessageColor} botMessageTextColor={botMessageTextColor} userMessageColor={userMessageColor} userMessageTextColor={userMessageTextColor} fontFamily={fontFamily} position={position} botNameOverride={botName} botIconOverride={botIcon} panelHeight={panelHeight} panelBg={panelBg} inputBg={inputBg} inputText={inputText} chatBg={chatBg} bubbleRadius={bubbleRadius} sendButtonColor={sendButtonColor} useOverrides={useOverrides} welcome={welcome} embedTokenUrl={embedTokenUrl} captchaSiteKey={captchaSiteKey} autoOpen={autoOpen} resetSession={resetSession} sessionIdOverride={sessionIdOverride} />, root)
}

export function unmount() {
  const host = document.getElementById(defaultHostId)
  if (!host || !host.shadowRoot) return
  host.shadowRoot.innerHTML = ''
}

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

declare global { interface Window { ChatbotWidget: { mount: typeof mount; unmount: typeof unmount } } }
// @ts-ignore
window.ChatbotWidget = { mount, unmount }
