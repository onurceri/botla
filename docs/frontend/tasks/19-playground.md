# Task: Implement Playground Tests

> **Task ID**: 19-playground  
> **Source**: TEST_PATHS.md Section 6.1  
> **Priority**: Medium-High (Chat & Actions)  
> **Estimated Effort**: 10-12 hours  
> **Prerequisite**: 11-chatbots-detail.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Chat Playground. This task covers chat interface, message sending, loading states, suggestions, and feedback.

### Context

The Playground is the chat testing interface for chatbots. Testing this functionality ensures:
- Chat interface works correctly
- Messages can be sent and received
- Loading states show properly
- Suggestions work as intended
- Feedback can be submitted

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 6.1:

#### 6.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `chat-container` | container | Chat messages area |
| `message-user` | component | User message bubble |
| `message-bot` | component | Bot message bubble |
| `message-loading` | component | Loading indicator |
| `message-feedback` | component | Thumbs up/down |
| `input-message` | textarea | Message input |
| `btn-send` | button | Send button |
| `suggestions-carousel` | component | Suggested questions |
| `btn-clear-chat` | button | Clear conversation |
| `btn-download-chat` | button | Download chat history |

#### 6.1.2 Chat Interaction Flow

```
Chat Flow
├── Load playground
│   ├── Assert: Chat container empty
│   ├── Assert: Welcome message shown
│   ├── Assert: Suggestions visible (if enabled)
│   └── Assert: Input enabled
│
├── Send message
│   ├── Type: "Hello, how are you?"
│   ├── Assert: Message appears (user)
│   ├── Assert: Loading indicator
│   ├── Wait: Bot response
│   ├── Assert: Message appears (bot)
│   ├── Assert: Sources cited (if any)
│   └── Assert: Feedback buttons visible
│
├── Send empty message
│   ├── Type: ""
│   ├── Click: btn-send
│   └── Assert: No message sent
│
├── Send long message
│   ├── Type: "A" x 4000
│   ├── Assert: Character count = 4000/4000
│   ├── Type: 1 more char
│   └── Assert: Error "Max 4000 characters"
│
├── Typing indicator
│   ├── Send: Message
│   ├── Assert: Bot shows typing
│   ├── Show: Animated dots
│   └── Hide: After response
│
├── Suggestions
│   ├── Click: Suggestion chip
│   │   ├── Copy: Text to input
│   │   └── Auto-send: After delay
│   │
│   └── Hover: Suggestion chip
│       └── Highlight background
│
└── Clear chat
    ├── Click: btn-clear-chat
    ├── Assert: `modal-confirm-clear` opens
    ├── Click: btn-confirm
    ├── Assert: Chat cleared
    └── Assert: Welcome message shown
```

### Implementation Requirements

1. **Create Playground Test File** (`frontend/e2e/playground.spec.ts`)
2. **Create Playground Page Object** (`frontend/e2e/pages/playground.page.ts`)
3. **Create Playground Mocks** (`frontend/e2e/mocks/playground.mocks.ts`)

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/playground.page.ts` with chat interface locators

### Phase 2: Page Load Tests

- [ ] Test: Welcome message visible
- [ ] Test: Suggestions visible (if enabled)
- [ ] Test: Input enabled
- [ ] Test: Send button visible

### Phase 3: Message Sending Tests

- [ ] Test: Type message
- [ ] Test: User message appears
- [ ] Test: Loading indicator
- [ ] Test: Bot response appears
- [ ] Test: Sources cited
- [ ] Test: Feedback buttons visible

### Phase 4: Validation Tests

- [ ] Test: Empty message not sent
- [ ] Test: Max character limit
- [ ] Test: Character count display

### Phase 5: Suggestions Tests

- [ ] Test: Click suggestion
- [ ] Test: Auto-send after delay
- [ ] Test: Hover effect

### Phase 6: Feedback Tests

- [ ] Test: Thumbs up works
- [ ] Test: Thumbs down works
- [ ] Test: Feedback form appears
- [ ] Test: Submit feedback

### Phase 7: Clear Chat Tests

- [ ] Test: Clear button visible
- [ ] Test: Confirmation modal
- [ ] Test: Chat cleared
- [ ] Test: Welcome message shown

---

## Technical Notes

```typescript
// frontend/e2e/pages/playground.page.ts
export class PlaygroundPage {
  readonly page: Page;
  readonly chatContainer: Locator;
  readonly messageInput: Locator;
  readonly sendButton: Locator;
  readonly clearButton: Locator;
  readonly suggestions: Locator;

  async sendMessage(message: string) {
    await this.messageInput.fill(message);
    await this.sendButton.click();
  }

  async expectBotResponse() {
    // Wait for bot message
  }

  async expectWelcomeMessage() {
    // Check welcome message visible
  }
}
```

---

## Dependencies

- **Prerequisites**: 11-chatbots-detail.md (Playground tab)
- **Environment**: Backend API with chat endpoint

---

## Related Tasks

- 11-chatbots-detail.md - Navigate to playground
- 20-chat-history.md - Conversation history
- 21-actions-list.md - Smart actions

---

*Task created from: docs/frontend/TEST_PATHS.md Section 6.1*
