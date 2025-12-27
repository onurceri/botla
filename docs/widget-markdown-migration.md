# markdown-to-jsx to react-markdown Migration

**Date:** 2025-12-27  
**Status:** ✅ Completed Successfully

## Problem

Build warnings about `eval` usage in `markdown-to-jsx`:
```
Use of eval in "../node_modules/markdown-to-jsx/dist/index.js" is strongly discouraged 
as it poses security risks and may cause issues with minification.
```

This contradicted our recent security hardening efforts for the widget.

## Solution

Migrated from `markdown-to-jsx` to `react-markdown`, which:
- ✅ Does not use `eval()`
- ✅ Does not use `dangerouslySetInnerHTML`
- ✅ Is the industry standard for secure markdown rendering
- ✅ Is actively maintained with better long-term support
- ✅ Works seamlessly with Preact via compatibility aliases

## Changes Made

### 1. Package Dependencies

**Removed:**
- `markdown-to-jsx@^9.3.5`

**Added:**
- `react-markdown@^10.1.0` - Secure markdown rendering
- `remark-gfm@^4.0.1` - GitHub Flavored Markdown support
- `rehype-sanitize@^6.0.0` - HTML sanitization

### 2. Component Updates

**File:** `src/components/Message.tsx`

- Replaced `Markdown` component with `ReactMarkdown`
- Removed standalone `SafeLink` component
- Integrated URL sanitization into markdown components configuration
- Used `rehype-sanitize` for HTML sanitization
- Maintained all existing security features:
  - URL sanitization via `sanitizeUrl()`
  - Automatic `target="_blank"` and `rel="noopener noreferrer"` for links
  - XSS prevention through sanitization

**Before:**
```typescript
import Markdown from 'markdown-to-jsx'

const secureMarkdownOptions = {
  disableParsingRawHTML: true,
  forceBlock: true,
  overrides: {
    script: () => null,
    iframe: () => null,
    object: () => null,
    embed: () => null,
    a: SafeLink as React.ComponentType,
  },
}

<Markdown options={secureMarkdownOptions}>{m.content}</Markdown>
```

**After:**
```typescript
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeSanitize from 'rehype-sanitize'

const markdownComponents = {
  a: ({ href, children }) => {
    const safeHref = sanitizeUrl(href)
    if (!safeHref) return <>{children}</>
    return (
      <a href={safeHref} target="_blank" rel="noopener noreferrer">
        {children}
      </a>
    )
  },
}

<ReactMarkdown
  remarkPlugins={[remarkGfm]}
  rehypePlugins={[rehypeSanitize]}
  components={markdownComponents as any}
>
  {m.content}
</ReactMarkdown>
```

### 3. Test Updates

**Files:**
- `src/components/__tests__/Message.test.tsx`
- `src/components/__tests__/ChatDrawer.test.tsx`

Updated mocks from `markdown-to-jsx` to `react-markdown` while maintaining the same testing behavior.

**Before:**
```typescript
vi.mock('markdown-to-jsx', () => ({
  default: ({ children }: { children: string }) => children
}))
```

**After:**
```typescript
vi.mock('react-markdown', () => ({
  default: ({ children }: { children: string }) => children
}))
```

## Verification

### ✅ All Tests Pass
```bash
npm run test
# Test Files  7 passed (7)
# Tests  52 passed (52)
```

### ✅ Build Succeeds Without Warnings
```bash
npm run build
# dist/widget.js  223.40 kB │ gzip: 67.65 kB
# ✓ built in 1.18s
# ✓ No eval warnings found
```

### ✅ Security Features Maintained
- URL sanitization still active
- XSS prevention through `rehype-sanitize`
- All unsafe HTML elements blocked
- Safe link handling with `noopener noreferrer`

## Benefits

1. **Enhanced Security**: Eliminated `eval()` usage completely
2. **Better Minification**: No eval-related minification issues
3. **Industry Standard**: Using the recommended React/Preact markdown library
4. **GitHub Flavored Markdown**: Added support for GFM features (tables, strikethrough, etc.)
5. **Robust HTML Sanitization**: Using `rehype-sanitize` for comprehensive protection
6. **Future-Proof**: Better long-term maintenance and security updates

## Bundle Size Impact

- Previous build: ~220 kB (with markdown-to-jsx)
- Current build: 223.40 kB (with react-markdown)
- Difference: +3.4 kB (~1.5% increase)

This small increase is acceptable given the security and maintainability improvements.

## Compatibility

- ✅ Fully compatible with Preact via existing React compatibility aliases
- ✅ No changes required to styling or CSS
- ✅ Maintains visual appearance
- ✅ All existing functionality preserved
