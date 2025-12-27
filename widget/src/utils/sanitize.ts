/**
 * Security utilities for widget
 */

const ALLOWED_PROTOCOLS = ['http:', 'https:', 'data:']
const DATA_MIME_WHITELIST = ['image/png', 'image/jpeg', 'image/gif', 'image/svg+xml', 'image/webp']

/**
 * Sanitizes URLs to prevent XSS attacks via javascript: or malicious data: URLs.
 * 
 * @param u - The URL to sanitize
 * @returns The sanitized URL or undefined if the URL is unsafe
 */
export function sanitizeUrl(u?: string): string | undefined {
  if (!u) return undefined

  const trimmed = u.replace(/[`'"<>]/g, '').trim()

  try {
    const url = new URL(trimmed)

    // Protocol check
    if (!ALLOWED_PROTOCOLS.includes(url.protocol)) {
      console.warn('[Widget] Blocked unsafe URL protocol:', url.protocol)
      return undefined
    }

    // Data URL MIME type check
    if (url.protocol === 'data:') {
      const mimeMatch = trimmed.match(/^data:([^;,]+)/)
      if (mimeMatch && !DATA_MIME_WHITELIST.includes(mimeMatch[1])) {
        console.warn('[Widget] Blocked unsafe data URL MIME:', mimeMatch[1])
        return undefined
      }
    }

    return url.toString()
  } catch {
    // Relative URLs
    if (trimmed.startsWith('/') || trimmed.startsWith('./')) {
      return trimmed
    }
    return undefined
  }
}
