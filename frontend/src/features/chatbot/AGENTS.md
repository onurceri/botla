# AGENTS.md - Chatbot Feature

Chatbot management UI with tabs, components, and hooks for configuration.

## WHERE TO LOOK

```
frontend/src/features/chatbot/
├── components/           # Reusable UI components (Card, Form, Settings panels)
│   └── __tests__/        # Component unit tests
├── pages/tabs/           # Tab page components (OverviewTab, PlaygroundTab, etc.)
│   ├── sections/         # Section components for tabs
│   └── __tests__/        # Tab unit tests
├── hooks/                # Custom hooks for state/logic (useChatbotForm, useAutoSave)
│   ├── __tests__/        # Hook unit tests
│   └── use*.ts           # Hook files (camelCase)
└── context/              # React Context for global state
```

## CONVENTIONS

- **State management**: `ChatbotContext` wraps form state + plan config + React Query data
- **Form handling**: `useChatbotForm` hook manages 50+ form fields with setters
- **Auto-save**: `useAutoSave` hook debounces and retries saves with progress indicator
- **Server state**: React Query hooks in `@/hooks/queries/` and mutations in `@/hooks/mutations/`
- **Component naming**: PascalCase (`ChatbotSidebar.tsx`), test files `Component.test.tsx`
- **Hook naming**: `use*` prefix, camelCase (`useSourceOps.ts`)
- **Payload builders**: `buildPayload()`, `build*Payload()` functions convert form state to API format
- **Path aliases**: Relative paths for local imports (`../../components/`, `../hooks/`)
- **Plan restrictions**: `ChatbotContext` enforces feature limits from user plan

## ANTI-PATTERNS

- **Manual API calls**: Use React Query hooks (`useChatbot`, `useUpdateBasicInfo`) instead of `api.get`/`api.put`
- **Skipping auto-save**: Configure `useAutoSave` with proper `saveFn` for mutation integration
- **Hardcoded thresholds**: Pull plan limits from `planConfig` context, don't inline limits
- **Prop drilling**: Use `useChatbotContext()` instead of passing chatbot state through multiple levels
- **Inline types**: Define proper interfaces in hook files, avoid `any` for payload types
