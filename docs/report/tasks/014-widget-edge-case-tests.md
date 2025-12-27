# Task 014: Widget Edge Case Tests

**Priority:** 🟡 Medium (Quality)  
**Phase:** 8 - Test Coverage  
**Estimated Time:** 3-4 hours  
**Dependencies:** None  

---

## Problem Statement

Widget lacks tests for edge cases:
- Network failures
- Invalid chatbot ID
- Rate limiting
- Large messages
- Special characters / XSS
- Mobile responsiveness

---

## Tests to Write

### File: `widget/src/__tests__/edge-cases.test.ts` (NEW)

```typescript
import { describe, it, expect, vi } from 'vitest';
import { sendMessage } from '../api/chat';

describe('Widget Edge Cases', () => {
  it('handles network failure gracefully', async () => {
    vi.spyOn(global, 'fetch').mockRejectedValue(new Error('Network error'));
    
    const result = await sendMessage('test', 'msg');
    
    expect(result.error).toBeDefined();
    expect(result.error).toContain('network');
  });

  it('handles 404 chatbot not found', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: 'ERR_NOT_FOUND' }),
    } as Response);
    
    const result = await sendMessage('invalid-id', 'msg');
    
    expect(result.error).toBeDefined();
  });

  it('handles rate limiting', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 429,
      json: () => Promise.resolve({ error: 'ERR_RATE_LIMITED' }),
    } as Response);
    
    const result = await sendMessage('bot-id', 'msg');
    
    expect(result.rateLimited).toBe(true);
  });

  it('truncates overly long messages', async () => {
    const longMessage = 'a'.repeat(10000);
    
    // Should truncate or reject
    const result = await sendMessage('bot-id', longMessage);
    
    // Either truncated or error
    expect(result.error || result.message.length <= 4000).toBeTruthy();
  });
});
```

### File: `widget/src/__tests__/xss-prevention.test.ts` (NEW)

```typescript
import { describe, it, expect } from 'vitest';
import { sanitizeMessage, sanitizeMarkdown } from '../utils/sanitize';

describe('XSS Prevention', () => {
  it('strips script tags', () => {
    const input = '<script>alert("xss")</script>Hello';
    const sanitized = sanitizeMarkdown(input);
    
    expect(sanitized).not.toContain('<script>');
    expect(sanitized).toContain('Hello');
  });

  it('neutralizes event handlers', () => {
    const input = '<img src="x" onerror="alert(1)">';
    const sanitized = sanitizeMarkdown(input);
    
    expect(sanitized).not.toContain('onerror');
  });

  it('blocks javascript: URLs', () => {
    const input = '<a href="javascript:alert(1)">click</a>';
    const sanitized = sanitizeMarkdown(input);
    
    expect(sanitized).not.toContain('javascript:');
  });

  it('allows safe markdown', () => {
    const input = '**bold** and _italic_';
    const sanitized = sanitizeMarkdown(input);
    
    expect(sanitized).toContain('bold');
    expect(sanitized).toContain('italic');
  });
});
```

### File: `widget/e2e/mobile.spec.ts` (NEW)

```typescript
import { test, expect, devices } from '@playwright/test';

test.describe('Mobile Responsiveness', () => {
  test.use(devices['iPhone 13']);

  test('widget opens on mobile', async ({ page }) => {
    await page.goto('/widget-test.html');
    
    await page.click('.botla-widget-button');
    
    await expect(page.locator('.botla-widget-container')).toBeVisible();
  });

  test('keyboard does not break layout', async ({ page }) => {
    await page.goto('/widget-test.html');
    await page.click('.botla-widget-button');
    
    await page.click('input[type="text"]');
    
    // Container should still be visible
    await expect(page.locator('.botla-widget-container')).toBeVisible();
  });
});
```

---

## Acceptance Criteria

- [ ] Network failure handling test
- [ ] 404 chatbot test
- [ ] Rate limiting test
- [ ] Long message test
- [ ] XSS prevention tests
- [ ] Mobile responsive tests
- [ ] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `widget/src/__tests__/edge-cases.test.ts` | CREATE |
| `widget/src/__tests__/xss-prevention.test.ts` | CREATE |
| `widget/e2e/mobile.spec.ts` | CREATE |
