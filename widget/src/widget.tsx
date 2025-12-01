import { render } from 'preact'
import { WidgetApp } from './widgetApp'
import styles from './styles.css?raw'

const defaultHostId = 'chatbot-widget-host'

function currentParams() {
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

  const host = ensureHost()
  const shadow = host.shadowRoot || host.attachShadow({ mode: 'open' })
  injectStyles(shadow)
  const root = document.createElement('div')
  shadow.appendChild(root)
  render(<WidgetApp chatbotId={chatbotId} apiBase={apiBase} themeColor={themeColor} welcome={welcome} />, root)
}

export function unmount() {
  const host = document.getElementById(defaultHostId)
  if (!host || !host.shadowRoot) return
  host.shadowRoot.innerHTML = ''
}

try { mount() } catch (e) {}

declare global { interface Window { ChatbotWidget: { mount: typeof mount; unmount: typeof unmount } } }
// @ts-ignore
window.ChatbotWidget = { mount, unmount }
