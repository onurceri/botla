# @botla/ui-shared

Shared UI components for Botla chat interfaces.

## Overview

This package contains reusable React/Preact components for building chat interfaces across both the Dashboard and Widget applications.

## Components

### `Message`
Displays a single chat message with support for:
- User and assistant messages
- Markdown rendering
- Feedback buttons (thumbs up/down)
- Handoff email submission forms

### `Suggestions`
Shows a carousel of suggested questions users can click to quickly ask common queries.

### `ChatBubble`
Floating button that opens the chat interface. Supports custom icons and unread message badges.

### `LoadingIndicator`
Animated typing indicator shown when the bot is processing a response.

### `ChatDrawer`
Complete chat interface including:
- Message history
- Input field with auto-resize
- Suggestions display
- Branding controls
- Character limits

## Usage

### In React (Dashboard)

```tsx
import { ChatDrawer, Message } from '@botla/ui-shared'
import type { ChatMessage } from '@botla/ui-shared'

function MyChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  
  return (
    <ChatDrawer
      messages={messages}
      loading={false}
      input={input}
      setInput={setInput}
      onSend={() => {/* send logic */}}
      onClose={() => {/* close logic */}}
      botName="My Bot"
    />
  )
}
```

### In Preact (Widget)

The components work seamlessly with Preact due to Preact's React compatibility layer:

```tsx
import { h } from 'preact'
import { ChatDrawer } from '@botla/ui-shared'

export function Widget() {
  // Same API as React
  return (
    <ChatDrawer
      messages={messages}
      // ... same props
    />
  )
}
```

## Styling

All components use CSS class names with the `cbw-` prefix (ChatBot Widget). They accept a `classNames` prop for customization:

```tsx
<Message
  message={msg}
  classNames={{
    row: 'custom-row',
    bubble: 'custom-bubble',
    content: 'custom-content',
  }}
/>
```

## Development

```bash
# Install dependencies
npm install

# Build the package
npm run build

# Run tests
npm test

# Watch mode for development
npm run dev
```

## Testing

All components have comprehensive test coverage. Run tests with:

```bash
npm test
```

## TypeScript

The package is written in TypeScript and exports all necessary types:

```tsx
import type { ChatMessage, ChatConfig, CustomBranding } from '@botla/ui-shared'
```
