# Task 015: Frontend Domain Layer Extraction

**Priority:** 🟢 Low (Code Quality)  
**Phase:** 6 - Service Layer Refactoring  
**Estimated Time:** 4-5 hours  
**Dependencies:** None  

---

## Problem Statement

Business logic is embedded in UI components:
- Plan limits checked in components
- Error code mapping scattered
- Feature flags in UI layer
- Makes refactoring risky

---

## Objective

Extract frontend domain logic into a dedicated layer:
1. Plan/feature configuration
2. Error handling utilities
3. Business rule validation
4. Domain types

---

## Implementation

### Step 1: Create Domain Directory Structure

```
frontend/src/domain/
├── plans/
│   ├── index.ts        # Plan types and utilities
│   └── limits.ts       # Limit checking functions
├── errors/
│   ├── index.ts        # Error handling
│   └── codes.ts        # Error code definitions
├── chatbot/
│   └── validation.ts   # Chatbot validation rules
└── index.ts            # Re-exports
```

### Step 2: Extract Plan Logic

**File:** `frontend/src/domain/plans/index.ts` (NEW)

```typescript
export type PlanCode = 'free' | 'starter' | 'pro' | 'enterprise';

export interface PlanLimits {
  maxChatbots: number;
  maxSources: number;
  maxURLs: number;
  maxPDFs: number;
  maxFileSizeMB: number;
  maxMonthlyTokens: number;
  features: {
    customBranding: boolean;
    analytics: boolean;
    handoff: boolean;
    apiAccess: boolean;
  };
}

export const PLAN_LIMITS: Record<PlanCode, PlanLimits> = {
  free: {
    maxChatbots: 1,
    maxSources: 5,
    maxURLs: 3,
    maxPDFs: 2,
    maxFileSizeMB: 5,
    maxMonthlyTokens: 10000,
    features: {
      customBranding: false,
      analytics: false,
      handoff: false,
      apiAccess: false,
    },
  },
  // ... other plans
};

export function canCreateChatbot(plan: PlanCode, currentCount: number): boolean {
  return currentCount < PLAN_LIMITS[plan].maxChatbots;
}

export function canAddSource(plan: PlanCode, currentCount: number): boolean {
  return currentCount < PLAN_LIMITS[plan].maxSources;
}
```

### Step 3: Extract Error Handling

**File:** `frontend/src/domain/errors/index.ts` (NEW)

```typescript
import { ErrorCode, errorMessages } from './codes';

export interface AppError {
  code: ErrorCode;
  message: string;
  userMessage: string;
  recoverable: boolean;
  retryable: boolean;
}

export function parseError(error: unknown, lang: string = 'tr'): AppError {
  if (typeof error === 'object' && error !== null && 'error' in error) {
    const code = (error as { error: string }).error as ErrorCode;
    return {
      code,
      message: code,
      userMessage: getUserMessage(code, lang),
      recoverable: isRecoverable(code),
      retryable: isRetryable(code),
    };
  }
  
  return {
    code: 'ERR_UNKNOWN',
    message: String(error),
    userMessage: lang === 'tr' ? 'Bir hata oluştu' : 'An error occurred',
    recoverable: true,
    retryable: true,
  };
}

function isRecoverable(code: ErrorCode): boolean {
  const unrecoverable = ['ERR_UNAUTHORIZED', 'ERR_FORBIDDEN'];
  return !unrecoverable.includes(code);
}

function isRetryable(code: ErrorCode): boolean {
  const retryable = ['ERR_RATE_LIMITED', 'ERR_INTERNAL_SERVER'];
  return retryable.includes(code);
}
```

### Step 4: Update Components to Use Domain Layer

**Before:**
```tsx
// In component
if (chatbots.length >= 1 && plan === 'free') {
  showError('Limit reached');
}
```

**After:**
```tsx
import { canCreateChatbot } from '@/domain/plans';

if (!canCreateChatbot(plan, chatbots.length)) {
  showError(t('errors.chatbot_limit'));
}
```

---

## Acceptance Criteria

- [ ] Domain directory structure created
- [ ] Plan limits extracted
- [ ] Error handling centralized
- [ ] At least 3 components updated to use domain layer
- [ ] No duplicate business logic in components
- [ ] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `frontend/src/domain/plans/index.ts` | CREATE |
| `frontend/src/domain/plans/limits.ts` | CREATE |
| `frontend/src/domain/errors/index.ts` | CREATE |
| `frontend/src/domain/errors/codes.ts` | CREATE |
| `frontend/src/domain/chatbot/validation.ts` | CREATE |
| `frontend/src/domain/index.ts` | CREATE |
| Multiple components | MODIFY |
