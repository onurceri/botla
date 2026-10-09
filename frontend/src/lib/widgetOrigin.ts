/** Default widget dev-server URL used when VITE_WIDGET_URL is not set. */
export const DEFAULT_WIDGET_URL = 'http://localhost:5174'

/** Base URL the Playground loads the widget preview from. */
export function getWidgetUrl(): string {
  return import.meta.env.VITE_WIDGET_URL || DEFAULT_WIDGET_URL
}

/** Serialized origin of the widget preview, or null if the URL is invalid. */
export function getWidgetOrigin(widgetUrl: string = getWidgetUrl()): string | null {
  try {
    const origin = new URL(widgetUrl).origin
    return origin === 'null' ? null : origin
  } catch {
    return null
  }
}

/**
 * Returns true when a postMessage event comes from the widget preview: the
 * origin must exactly match the widget origin and, when an expected source
 * window is given, the event must come from that window (the preview iframe).
 */
export function isWidgetMessage(
  event: MessageEvent,
  expectedSource?: MessageEventSource | null,
): boolean {
  const origin = getWidgetOrigin()
  if (!origin || event.origin !== origin) return false
  if (
    expectedSource !== undefined &&
    (expectedSource === null || event.source !== expectedSource)
  ) {
    return false
  }
  return true
}
