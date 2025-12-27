import { describe, it, expect } from 'vitest'
import { sanitizeMarkdown, sanitizeMessage, sanitizeUrl } from '../utils/sanitize'

describe('XSS Prevention', () => {
  describe('sanitizeMarkdown', () => {
    it('strips script tags with content', () => {
      const input = '<script>alert("xss")</script>Hello'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('<script>')
      expect(sanitized).not.toContain('</script>')
      expect(sanitized).not.toContain('alert')
      expect(sanitized).toContain('Hello')
    })

    it('strips script tags case-insensitively', () => {
      const input = '<SCRIPT>alert("xss")</SCRIPT>Hello'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('<SCRIPT>')
      expect(sanitized).not.toContain('</SCRIPT>')
      expect(sanitized).toContain('Hello')
    })

    it('strips standalone script tags', () => {
      const input = '<script src="evil.js"></script>Hello'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('<script')
      expect(sanitized).not.toContain('</script>')
      expect(sanitized).toContain('Hello')
    })

    it('neutralizes event handlers - onerror', () => {
      const input = '<img src="x" onerror="alert(1)">'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('onerror')
      expect(sanitized).not.toContain('alert')
    })

    it('neutralizes event handlers - onclick', () => {
      const input = '<button onclick="alert(1)">Click</button>'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('onclick')
      expect(sanitized).not.toContain('alert')
      expect(sanitized).toContain('Click')
    })

    it('neutralizes event handlers - onmouseover', () => {
      const input = '<div onmouseover="evil()">Hover</div>'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('onmouseover')
      expect(sanitized).toContain('Hover')
    })

    it('neutralizes event handlers - onload', () => {
      const input = '<body onload="steal()">Content</body>'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('onload')
      expect(sanitized).toContain('Content')
    })

    it('blocks javascript: URLs in href', () => {
      const input = '<a href="javascript:alert(1)">click</a>'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('javascript:')
      expect(sanitized).toContain('click')
    })

    it('blocks javascript: URLs in src', () => {
      const input = '<img src="javascript:evil()">'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('javascript:')
    })

    it('strips style tags', () => {
      const input = '<style>body { display: none }</style>Visible'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('<style>')
      expect(sanitized).not.toContain('</style>')
      expect(sanitized).not.toContain('display')
      expect(sanitized).toContain('Visible')
    })

    it('strips iframe tags', () => {
      const input = '<iframe src="evil.com"></iframe>Safe'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('<iframe')
      expect(sanitized).toContain('Safe')
    })

    it('strips object, embed, and form tags', () => {
      const input = '<object data="x"></object><embed src="y"><form action="z">Input</form>'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('<object')
      expect(sanitized).not.toContain('<embed')
      expect(sanitized).not.toContain('<form')
      expect(sanitized).toContain('Input')
    })

    it('allows safe markdown - bold and italic', () => {
      const input = '**bold** and _italic_'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).toContain('**bold**')
      expect(sanitized).toContain('_italic_')
    })

    it('allows safe HTML links', () => {
      const input = '<a href="https://example.com">Safe Link</a>'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).toContain('href="https://example.com"')
      expect(sanitized).toContain('Safe Link')
    })

    it('allows safe images', () => {
      const input = '<img src="https://example.com/image.png" alt="Image">'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).toContain('src="https://example.com/image.png"')
    })

    it('handles empty input', () => {
      expect(sanitizeMarkdown('')).toBe('')
    })

    it('handles nested malicious content', () => {
      const input = '<div onclick="x"><script>y</script><a href="javascript:z">link</a></div>'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized).not.toContain('onclick')
      expect(sanitized).not.toContain('<script>')
      expect(sanitized).not.toContain('javascript:')
      expect(sanitized).toContain('link')
    })

    it('handles mixed case event handlers', () => {
      const input = '<img OnErRoR="alert(1)" src="x">'
      const sanitized = sanitizeMarkdown(input)
      
      expect(sanitized.toLowerCase()).not.toContain('onerror')
    })
  })

  describe('sanitizeMessage', () => {
    it('escapes HTML special characters', () => {
      const input = '<script>alert("test")</script>'
      const sanitized = sanitizeMessage(input)
      
      expect(sanitized).not.toContain('<')
      expect(sanitized).not.toContain('>')
      expect(sanitized).toContain('&lt;')
      expect(sanitized).toContain('&gt;')
    })

    it('escapes ampersands', () => {
      const input = 'Tom & Jerry'
      const sanitized = sanitizeMessage(input)
      
      expect(sanitized).toBe('Tom &amp; Jerry')
    })

    it('escapes quotes', () => {
      const input = "He said \"hello\" and 'goodbye'"
      const sanitized = sanitizeMessage(input)
      
      expect(sanitized).toContain('&quot;')
      expect(sanitized).toContain('&#x27;')
    })

    it('handles empty input', () => {
      expect(sanitizeMessage('')).toBe('')
    })

    it('preserves safe content', () => {
      const input = 'Hello, this is a normal message!'
      const sanitized = sanitizeMessage(input)
      
      expect(sanitized).toBe(input)
    })
  })

  describe('sanitizeUrl', () => {
    it('blocks javascript: protocol', () => {
      const result = sanitizeUrl('javascript:alert(1)')
      expect(result).toBeUndefined()
    })

    it('blocks vbscript: protocol', () => {
      const result = sanitizeUrl('vbscript:msgbox(1)')
      expect(result).toBeUndefined()
    })

    it('allows https: protocol', () => {
      const result = sanitizeUrl('https://example.com')
      expect(result).toBe('https://example.com/')
    })

    it('allows http: protocol', () => {
      const result = sanitizeUrl('http://example.com')
      expect(result).toBe('http://example.com/')
    })

    it('allows safe data: URLs for images', () => {
      const dataUrl = 'data:image/png;base64,iVBOR='
      const result = sanitizeUrl(dataUrl)
      expect(result).toBe(dataUrl)
    })

    it('blocks unsafe data: URLs', () => {
      const result = sanitizeUrl('data:text/html,<script>alert(1)</script>')
      expect(result).toBeUndefined()
    })

    it('allows relative URLs starting with /', () => {
      const result = sanitizeUrl('/path/to/resource')
      expect(result).toBe('/path/to/resource')
    })

    it('allows relative URLs starting with ./', () => {
      const result = sanitizeUrl('./local-file.html')
      expect(result).toBe('./local-file.html')
    })

    it('handles undefined input', () => {
      expect(sanitizeUrl(undefined)).toBeUndefined()
    })

    it('handles empty string', () => {
      expect(sanitizeUrl('')).toBeUndefined()
    })

    it('strips dangerous characters', () => {
      const result = sanitizeUrl('https://example.com/<script>')
      // URL should be sanitized or return undefined
      expect(result).not.toContain('<script>')
    })
  })
})
