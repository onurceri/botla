# Task 03: Frontend Component Consolidation - Dashboard and Widget Shared Library

## Priority
**Medium** - Reduces maintenance burden and ensures visual consistency

## Problem Statement

The project maintains two distinct frontend applications:
- **Dashboard**: React 19 + Vite (`frontend/`)
- **Widget**: Preact + Vite (`widget/`)

Both applications have significant overlap in UI components, particularly:
- Chat bubbles and message rendering
- Input fields and forms
- Markdown rendering
- Tailwind configurations

This duplication leads to:
- Visual and functional inconsistencies
- Bug fixes in one app being forgotten in the other
- Doubled QA surface area
- Slowed feature parity

## Evidence

### Overlapping Components

| Component | Dashboard Location | Widget Location |
|-----------|-------------------|-----------------|
| Chat Bubble | `frontend/src/components/chatbot/` | `widget/src/components/ChatBubble.tsx` |
| Message Rendering | `frontend/src/components/chatbot/` | `widget/src/components/Message.tsx` |
| Chat Drawer | - | `widget/src/components/ChatDrawer.tsx` |
| Suggestions | - | `widget/src/components/Suggestions.tsx` |

### Tailwind Configurations

Both applications have separate Tailwind configurations that may drift over time:
- `frontend/tailwind.config.js`
- `widget/tailwind.config.js`

## Refactoring Goals

1. **Shared Component Library**: Extract common UI components into a shared package
2. **Consistent Styling**: Single source of truth for design tokens
3. **Framework Agnostic**: Components work with both React and Preact
4. **Type Safety**: Shared TypeScript types for API responses and component props

## Implementation Plan

### Phase 1: Setup Monorepo Structure

**Option A: npm Workspaces (Recommended for simplicity)**

```json
// package.json (root)
{
  "name": "botla-co",
  "private": true,
  "workspaces": [
    "packages/*",
    "frontend",
    "widget"
  ]
}
```

**Option B: Turborepo (Recommended for scale)**

```json
// turbo.json
{
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**"]
    },
    "dev": {
      "cache": false
    }
  }
}
```

### Phase 2: Create Shared UI Package

**New Directory**: `packages/ui-shared/`

```
packages/
└── ui-shared/
    ├── package.json
    ├── tsconfig.json
    ├── src/
    │   ├── index.ts
    │   ├── components/
    │   │   ├── ChatBubble.tsx
    │   │   ├── Message.tsx
    │   │   ├── MessageInput.tsx
    │   │   └── MarkdownRenderer.tsx
    │   ├── hooks/
    │   │   ├── useChat.ts
    │   │   └── useTypingIndicator.ts
    │   └── styles/
    │       ├── tokens.css
    │       └── components.css
    └── tailwind.config.js
```

**File**: `packages/ui-shared/package.json`

```json
{
  "name": "@botla/ui-shared",
  "version": "0.1.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/index.js",
      "types": "./dist/index.d.ts"
    },
    "./styles": "./dist/styles/index.css"
  },
  "peerDependencies": {
    "react": ">=17.0.0 || >=18.0.0 || >=19.0.0",
    "preact": ">=10.0.0"
  },
  "peerDependenciesMeta": {
    "react": { "optional": true },
    "preact": { "optional": true }
  }
}
```

### Phase 3: Create Framework-Agnostic Components

**Strategy**: Use JSX pragma that works with both React and Preact

**File**: `packages/ui-shared/src/components/ChatBubble.tsx`

```tsx
/** @jsxImportSource react */
import type { ReactNode } from 'react';

export interface ChatBubbleProps {
  variant: 'user' | 'bot';
  children: ReactNode;
  className?: string;
}

export function ChatBubble({ variant, children, className = '' }: ChatBubbleProps) {
  const baseClasses = 'rounded-2xl px-4 py-2 max-w-[80%]';
  const variantClasses = variant === 'user' 
    ? 'bg-primary-600 text-white ml-auto' 
    : 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white';
  
  return (
    <div className={`${baseClasses} ${variantClasses} ${className}`}>
      {children}
    </div>
  );
}
```

**File**: `packages/ui-shared/src/components/Message.tsx`

```tsx
/** @jsxImportSource react */
import { ChatBubble } from './ChatBubble';
import { MarkdownRenderer } from './MarkdownRenderer';

export interface MessageProps {
  role: 'user' | 'assistant';
  content: string;
  timestamp?: Date;
  isTyping?: boolean;
}

export function Message({ role, content, timestamp, isTyping }: MessageProps) {
  const variant = role === 'user' ? 'user' : 'bot';
  
  return (
    <div className={`flex ${role === 'user' ? 'justify-end' : 'justify-start'} mb-3`}>
      <ChatBubble variant={variant}>
        {isTyping ? (
          <TypingIndicator />
        ) : (
          <MarkdownRenderer content={content} />
        )}
      </ChatBubble>
    </div>
  );
}
```

### Phase 4: Create Shared API Client

**File**: `packages/ui-shared/src/api/chat.ts`

```typescript
export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface ChatRequest {
  message: string;
  session_id: string;
}

export interface ChatResponse {
  response: string;
  tokens_used: number;
  sources_used: Array<{
    chunk_index: number;
    source_type: string;
  }>;
}

export function createChatClient(baseUrl: string) {
  return {
    async sendMessage(chatbotId: string, request: ChatRequest): Promise<ChatResponse> {
      const response = await fetch(`${baseUrl}/c/${chatbotId}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
      });
      
      if (!response.ok) {
        throw new Error(`Chat failed: ${response.status}`);
      }
      
      return response.json();
    },
  };
}
```

### Phase 5: Shared Design Tokens

**File**: `packages/ui-shared/src/styles/tokens.css`

```css
:root {
  /* Colors */
  --color-primary-50: #f0f9ff;
  --color-primary-100: #e0f2fe;
  --color-primary-500: #0ea5e9;
  --color-primary-600: #0284c7;
  --color-primary-700: #0369a1;
  
  /* Spacing */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  
  /* Border Radius */
  --radius-sm: 0.375rem;
  --radius-md: 0.5rem;
  --radius-lg: 1rem;
  --radius-full: 9999px;
  
  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
}

.dark {
  --color-bg-primary: #1f2937;
  --color-bg-secondary: #111827;
  --color-text-primary: #f9fafb;
}
```

### Phase 6: Update Consumers

**Dashboard**: `frontend/package.json`

```json
{
  "dependencies": {
    "@botla/ui-shared": "workspace:*"
  }
}
```

**Widget**: `widget/package.json`

```json
{
  "dependencies": {
    "@botla/ui-shared": "workspace:*"
  }
}
```

**Dashboard Usage**:

```tsx
// frontend/src/pages/ChatbotPreview.tsx
import { Message, MessageInput } from '@botla/ui-shared';
import '@botla/ui-shared/styles';

function ChatbotPreview() {
  return (
    <div className="flex flex-col h-full">
      {messages.map((msg, i) => (
        <Message key={i} role={msg.role} content={msg.content} />
      ))}
      <MessageInput onSend={handleSend} />
    </div>
  );
}
```

## Migration Strategy

1. **Setup workspace structure** without moving existing code
2. **Extract one component at a time** (start with ChatBubble)
3. **Update dashboard first** (easier to debug in full React)
4. **Update widget second** (verify Preact compatibility)
5. **Remove duplicate code** from both applications
6. **Consolidate Tailwind configs** to extend shared tokens

## Affected Files

| File | Action | Description |
|------|--------|-------------|
| `package.json` (root) | MODIFY | Add workspaces configuration |
| `packages/ui-shared/` | NEW | Entire shared package |
| `frontend/package.json` | MODIFY | Add workspace dependency |
| `widget/package.json` | MODIFY | Add workspace dependency |
| `frontend/src/components/chatbot/` | MODIFY | Replace with shared imports |
| `widget/src/components/*.tsx` | MODIFY | Replace with shared imports |

## Testing Strategy

### Component Tests

```tsx
// packages/ui-shared/src/components/__tests__/ChatBubble.test.tsx
import { render, screen } from '@testing-library/react';
import { ChatBubble } from '../ChatBubble';

describe('ChatBubble', () => {
  it('renders user variant with correct styling', () => {
    render(<ChatBubble variant="user">Hello</ChatBubble>);
    expect(screen.getByText('Hello')).toHaveClass('bg-primary-600');
  });
  
  it('renders bot variant with correct styling', () => {
    render(<ChatBubble variant="bot">Hi there</ChatBubble>);
    expect(screen.getByText('Hi there')).toHaveClass('bg-gray-100');
  });
});
```

### Visual Regression Tests

Consider adding Chromatic or Percy for visual regression testing.

## Acceptance Criteria

- [ ] Workspace structure configured and working
- [ ] At least 5 components extracted to shared package
- [ ] Dashboard uses shared components without regressions
- [ ] Widget uses shared components without regressions
- [ ] Shared Tailwind tokens in use
- [ ] Build and dev scripts work for all packages
- [ ] Documentation for adding new shared components

## Estimated Effort

**Size**: Large (5-7 days)
- Phase 1-2: 1 day (setup)
- Phase 3-4: 2 days (components)
- Phase 5: 0.5 day (tokens)
- Phase 6: 2-3 days (migration)
- Testing: 1 day

## Dependencies

- Node.js with npm workspaces support (v7+)
- Optional: Turborepo for build caching

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Preact compatibility issues | Test each component with both frameworks |
| Bundle size increase | Tree-shaking, only import what's needed |
| Build complexity | Clear documentation, CI validation |

## Future Improvements

- Storybook for component development and documentation
- Visual regression testing with Chromatic
- Component library versioning for external use
