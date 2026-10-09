/**
 * postMessage helpers for the widget's Playground / host-page integration.
 *
 * Trust model:
 * - Inbound `WIDGET_CONFIG` messages are accepted only from the Botla dashboard
 *   origins (exact origin match, fail closed when none are configured) and only
 *   from the parent window.
 * - Outbound `WIDGET_EVENT_*` messages are delivered only to the dashboard
 *   origins plus an optional host-configured `parent-origin`; they are never
 *   broadcast with a `'*'` target.
 */

/** Production dashboard origins used when VITE_DASHBOARD_URL is not set. */
export const DEFAULT_DASHBOARD_ORIGINS = ['https://botla.app', 'https://www.botla.app']

/** Local dashboard dev-server origins, only trusted in development builds. */
export const DEV_DASHBOARD_ORIGINS = ['http://localhost:5173', 'http://localhost:3000']

const LOCAL_HOSTNAMES = new Set(['localhost', '127.0.0.1', '[::1]'])

/**
 * Normalizes a URL or origin string to its serialized origin
 * (`scheme://host[:port]`). Returns null for unparsable values, opaque
 * origins, and non-http(s) schemes.
 */
export function normalizeOrigin(value: string | null | undefined): string | null {
  if (!value) return null
  const trimmed = value.trim()
  if (!trimmed || trimmed === '*') return null
  try {
    const url = new URL(trimmed)
    if (url.protocol !== 'https:' && url.protocol !== 'http:') return null
    if (url.origin === 'null') return null
    return url.origin
  } catch {
    return null
  }
}

/**
 * Parses one or more comma-separated origin lists into a de-duplicated list of
 * normalized origins. Invalid entries are dropped.
 */
export function parseOriginList(...values: Array<string | null | undefined>): string[] {
  const out: string[] = []
  for (const value of values) {
    if (!value) continue
    for (const part of value.split(',')) {
      const origin = normalizeOrigin(part)
      if (origin && !out.includes(origin)) out.push(origin)
    }
  }
  return out
}

/**
 * Exact origin comparison against an allowlist. An empty allowlist rejects
 * everything (fail closed).
 */
export function isAllowedOrigin(origin: string, allowed: readonly string[]): boolean {
  if (!origin || allowed.length === 0) return false
  return allowed.includes(origin)
}

/**
 * Resolves the dashboard origins that may configure the widget and receive its
 * events. `VITE_DASHBOARD_URL` (comma-separated) overrides the production
 * defaults; local dev-server origins are added only in development builds.
 */
export function getDashboardOrigins(
  configured: string | undefined = import.meta.env.VITE_DASHBOARD_URL,
  isDev: boolean = import.meta.env.DEV,
): string[] {
  const base = parseOriginList(configured)
  const origins = base.length > 0 ? base : [...DEFAULT_DASHBOARD_ORIGINS]
  if (isDev) {
    for (const o of DEV_DASHBOARD_ORIGINS) if (!origins.includes(o)) origins.push(o)
  }
  return origins
}

/**
 * Target origins for outbound widget events: the dashboard origins plus an
 * optional, explicitly configured parent origin (`parent-origin` parameter).
 */
export function getEventTargetOrigins(
  parentOrigin?: string | null,
  dashboardOrigins: readonly string[] = getDashboardOrigins(),
): string[] {
  const out = [...dashboardOrigins]
  const extra = normalizeOrigin(parentOrigin)
  if (extra && !out.includes(extra)) out.push(extra)
  return out
}

/**
 * Validates an API base URL supplied at runtime. Requires https, except for
 * localhost/loopback hosts where http is allowed for development. Credentials,
 * query strings and fragments are rejected. Returns the normalized base
 * (origin + path, no trailing slash) or null when invalid.
 */
export function validateApiBase(value: unknown): string | null {
  if (typeof value !== 'string' || !value.trim()) return null
  let url: URL
  try {
    url = new URL(value.trim())
  } catch {
    return null
  }
  const isLocal = LOCAL_HOSTNAMES.has(url.hostname)
  if (url.protocol !== 'https:' && !(url.protocol === 'http:' && isLocal)) return null
  if (url.username || url.password || url.search || url.hash) return null
  return url.origin + url.pathname.replace(/\/+$/, '')
}

/** Minimal window surface needed to post a message. */
export interface MessageTarget {
  postMessage(message: unknown, targetOrigin: string): void
}

/**
 * Posts a widget event to the parent window once per trusted target origin.
 * The browser drops deliveries whose target origin does not match the actual
 * parent, so untrusted framing pages never receive the payload. No-op when the
 * widget is not framed.
 */
export function postWidgetEvent(
  type: string,
  payload: unknown,
  targetOrigins: readonly string[],
  self: Window = window,
): void {
  const parent = self.parent as MessageTarget | null
  if (!parent || (parent as unknown) === self) return
  const message = { type: `WIDGET_EVENT_${type.toUpperCase()}`, payload }
  for (const origin of targetOrigins) {
    if (!origin || origin === '*') continue
    try {
      parent.postMessage(message, origin)
    } catch {
      // Ignore delivery errors (e.g. detached parent)
    }
  }
}

export interface ConfigMessageOptions {
  /** Origins allowed to send WIDGET_CONFIG (exact match). */
  allowedOrigins: readonly string[]
  /** Window expected to be the sender (the embedding parent). */
  expectedSource: MessageEventSource | null
  /** Called with the config object for accepted messages. */
  onConfig: (config: Record<string, unknown>, event: MessageEvent) => void
  /** Called for rejected WIDGET_CONFIG messages (for logging). */
  onReject?: (reason: string, event: MessageEvent) => void
}

/**
 * Builds a `message` event listener that applies WIDGET_CONFIG messages only
 * when they come from an allowed origin and the expected source window.
 */
export function createConfigMessageHandler(opts: ConfigMessageOptions): (event: MessageEvent) => void {
  return (event: MessageEvent) => {
    const data = event.data as { type?: unknown; config?: unknown } | null
    if (!data || typeof data !== 'object' || data.type !== 'WIDGET_CONFIG') return
    if (!isAllowedOrigin(event.origin, opts.allowedOrigins)) {
      opts.onReject?.('origin', event)
      return
    }
    if (opts.expectedSource && event.source !== opts.expectedSource) {
      opts.onReject?.('source', event)
      return
    }
    const config = data.config
    if (!config || typeof config !== 'object' || Array.isArray(config)) {
      opts.onReject?.('config', event)
      return
    }
    opts.onConfig(config as Record<string, unknown>, event)
  }
}
