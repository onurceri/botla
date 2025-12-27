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

/**
 * Dangerous HTML event handler attributes that should be stripped
 */
const EVENT_HANDLERS = [
  'onabort', 'onanimationend', 'onanimationiteration', 'onanimationstart',
  'onauxclick', 'onbeforecopy', 'onbeforecut', 'onbeforeinput', 'onbeforepaste',
  'onblur', 'oncancel', 'oncanplay', 'oncanplaythrough', 'onchange', 'onclick',
  'onclose', 'oncontextmenu', 'oncopy', 'oncuechange', 'oncut', 'ondblclick',
  'ondrag', 'ondragend', 'ondragenter', 'ondragleave', 'ondragover', 'ondragstart',
  'ondrop', 'ondurationchange', 'onemptied', 'onended', 'onerror', 'onfocus',
  'onfocusin', 'onfocusout', 'onformdata', 'ongotpointercapture', 'oninput',
  'oninvalid', 'onkeydown', 'onkeypress', 'onkeyup', 'onload', 'onloadeddata',
  'onloadedmetadata', 'onloadstart', 'onlostpointercapture', 'onmousedown',
  'onmouseenter', 'onmouseleave', 'onmousemove', 'onmouseout', 'onmouseover',
  'onmouseup', 'onmousewheel', 'onpaste', 'onpause', 'onplay', 'onplaying',
  'onpointercancel', 'onpointerdown', 'onpointerenter', 'onpointerleave',
  'onpointermove', 'onpointerout', 'onpointerover', 'onpointerup', 'onprogress',
  'onratechange', 'onreset', 'onresize', 'onscroll', 'onsearch', 'onseeked',
  'onseeking', 'onselect', 'onselectionchange', 'onselectstart', 'onshow',
  'onstalled', 'onsubmit', 'onsuspend', 'ontimeupdate', 'ontoggle',
  'ontransitionend', 'onvolumechange', 'onwaiting', 'onwheel'
]

/**
 * Sanitizes a message string by escaping HTML special characters.
 * 
 * @param message - The message to sanitize
 * @returns The sanitized message with HTML entities escaped
 */
export function sanitizeMessage(message: string): string {
  if (!message) return ''
  
  return message
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
}

/**
 * Sanitizes markdown content by removing dangerous HTML elements and attributes.
 * This function strips script tags, event handlers, and javascript: URLs while
 * preserving safe markdown content.
 * 
 * @param markdown - The markdown content to sanitize
 * @returns The sanitized markdown content
 */
export function sanitizeMarkdown(markdown: string): string {
  if (!markdown) return ''
  
  let result = markdown
  
  // Remove script tags and their content (case-insensitive)
  result = result.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
  
  // Remove standalone script tags (opening or closing)
  result = result.replace(/<\/?script[^>]*>/gi, '')
  
  // Remove style tags and their content
  result = result.replace(/<style\b[^<]*(?:(?!<\/style>)<[^<]*)*<\/style>/gi, '')
  
  // Remove event handlers from any HTML tag
  for (const handler of EVENT_HANDLERS) {
    // Match both quoted and unquoted attribute values
    const pattern = new RegExp(`\\s*${handler}\\s*=\\s*(?:"[^"]*"|'[^']*'|[^\\s>]*)`, 'gi')
    result = result.replace(pattern, '')
  }
  
  // Remove javascript: URLs from href, src, action attributes
  result = result.replace(/\s*(href|src|action)\s*=\s*["']?\s*javascript:[^"'>\s]*/gi, '')
  
  // Also handle javascript: in plain anchor tags more thoroughly
  result = result.replace(/href\s*=\s*["']javascript:[^"']*["']/gi, 'href=""')
  result = result.replace(/href\s*=\s*javascript:[^\s>]*/gi, 'href=""')
  
  // Remove data: URLs that aren't images
  result = result.replace(/\s*(src|href)\s*=\s*["']?data:(?!image\/)[^"'>\s]*/gi, '')
  
  // Remove iframe, object, embed, form, and other dangerous elements
  const dangerousTags = ['iframe', 'object', 'embed', 'form', 'base', 'meta', 'link']
  for (const tag of dangerousTags) {
    result = result.replace(new RegExp(`<${tag}\\b[^>]*>`, 'gi'), '')
    result = result.replace(new RegExp(`</${tag}>`, 'gi'), '')
  }
  
  return result
}
