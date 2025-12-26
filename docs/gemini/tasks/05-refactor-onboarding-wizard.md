# Task 05: Refactor OnboardingWizard Component (Optional)

**Priority:** ⚪ Optional  
**Effort:** High (1-2 days)  
**Risk Level:** Medium (UI regression risk)

---

## Problem Statement

The `OnboardingWizard.tsx` component is 632 lines with complex state logic. While it's borderline for a "God Component," it may benefit from decomposition if the onboarding flow needs significant changes.

### Evidence

**File:** `frontend/src/components/onboarding/OnboardingWizard.tsx`
- **Size:** 22.83kb
- **Lines:** 632
- **Internal Functions:** 9+ handlers and render methods

### Current Assessment

**Not Urgent.** The component is reasonably well-structured internally with:
- Clear step definitions
- Separate handler functions (`handleNext`, `handleBack`, `handleFinish`)
- Dedicated render function (`renderStepContent`)

**Consider this task only if:**
- You're adding new onboarding steps
- You need to A/B test different wizard variants
- Unit testing individual steps becomes necessary

---

## Acceptance Criteria

- [x] Wizard state managed by a custom hook (`useOnboardingWizard`)
- [x] Each step is a separate component (`StepBotName`, `StepDataSource`, `StepPersonality`, `StepComplete`)
- [x] Step components are stateless (receive props only)
- [x] Main wizard component under 200 lines (now 113 lines, down from 632)
- [x] All existing functionality preserved (51 tests passing)
- [x] Visual regression tests pass

---

## Proposed Architecture

### Current Structure

```
OnboardingWizard.tsx (632 lines)
├── State: currentStep, botName, sources, personality...
├── Handlers: handleNext, handleBack, handleFinish...
└── Render: renderStepContent (switch statement)
```

### Target Structure

```
onboarding/
├── OnboardingWizard.tsx (150 lines) - Orchestrator
├── hooks/
│   └── useOnboardingState.ts - State machine
├── steps/
│   ├── StepBotName.tsx
│   ├── StepDataSource.tsx
│   ├── StepPersonality.tsx
│   └── StepComplete.tsx
├── types.ts - Shared types
└── index.ts - Exports
```

---

## Implementation Steps

### Step 1: Extract Types

**File:** `frontend/src/components/onboarding/types.ts`

```typescript
export type OnboardingStep = 1 | 2 | 3 | 4;

export type SourceType = 'url' | 'file' | 'text';

export interface OnboardingState {
  currentStep: OnboardingStep;
  botName: string;
  sourceType: SourceType;
  sourceUrl: string;
  sourceFile: File | null;
  sourceText: string;
  personality: string;
  welcomeMessage: string;
  isSubmitting: boolean;
  chatbotId: string | null;
}

export interface OnboardingActions {
  nextStep: () => void;
  prevStep: () => void;
  setBotName: (name: string) => void;
  setSourceType: (type: SourceType) => void;
  setSourceUrl: (url: string) => void;
  setSourceFile: (file: File | null) => void;
  setSourceText: (text: string) => void;
  setPersonality: (personality: string) => void;
  setWelcomeMessage: (message: string) => void;
  finish: () => Promise<void>;
  skip: () => void;
}
```

### Step 2: Create State Hook

**File:** `frontend/src/components/onboarding/hooks/useOnboardingState.ts`

```typescript
import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useToast } from '@/components/ui/toast';
import * as onboardingApi from '@/api/onboarding';
import type { OnboardingState, OnboardingActions } from '../types';

const initialState: OnboardingState = {
  currentStep: 1,
  botName: '',
  sourceType: 'url',
  sourceUrl: '',
  sourceFile: null,
  sourceText: '',
  personality: 'helpful',
  welcomeMessage: 'Merhaba! Size nasıl yardımcı olabilirim?',
  isSubmitting: false,
  chatbotId: null,
};

export function useOnboardingState(): [OnboardingState, OnboardingActions] {
  const [state, setState] = useState<OnboardingState>(initialState);
  const navigate = useNavigate();
  const { toast } = useToast();

  const nextStep = useCallback(() => {
    setState((prev) => ({
      ...prev,
      currentStep: Math.min(prev.currentStep + 1, 4) as OnboardingStep,
    }));
  }, []);

  const prevStep = useCallback(() => {
    setState((prev) => ({
      ...prev,
      currentStep: Math.max(prev.currentStep - 1, 1) as OnboardingStep,
    }));
  }, []);

  const setBotName = useCallback((name: string) => {
    setState((prev) => ({ ...prev, botName: name }));
  }, []);

  // ... other setters

  const finish = useCallback(async () => {
    setState((prev) => ({ ...prev, isSubmitting: true }));
    try {
      const chatbotId = await onboardingApi.completeOnboarding({
        name: state.botName,
        sourceType: state.sourceType,
        sourceUrl: state.sourceUrl,
        sourceFile: state.sourceFile,
        sourceText: state.sourceText,
        personality: state.personality,
        welcomeMessage: state.welcomeMessage,
      });
      
      setState((prev) => ({ ...prev, chatbotId, isSubmitting: false }));
      navigate(`/dashboard/chatbots/${chatbotId}`);
    } catch (error) {
      toast({ title: 'Error', description: 'Failed to create chatbot', variant: 'destructive' });
      setState((prev) => ({ ...prev, isSubmitting: false }));
    }
  }, [state, navigate, toast]);

  const skip = useCallback(() => {
    navigate('/dashboard');
  }, [navigate]);

  return [
    state,
    {
      nextStep,
      prevStep,
      setBotName,
      setSourceType,
      setSourceUrl,
      setSourceFile,
      setSourceText,
      setPersonality,
      setWelcomeMessage,
      finish,
      skip,
    },
  ];
}
```

### Step 3: Create Step Components

**File:** `frontend/src/components/onboarding/steps/StepBotName.tsx`

```typescript
import { Input } from '@/components/ui/input';
import { Bot } from 'lucide-react';

interface StepBotNameProps {
  botName: string;
  onBotNameChange: (name: string) => void;
}

export function StepBotName({ botName, onBotNameChange }: StepBotNameProps) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-center">
        <div className="p-4 bg-primary/10 rounded-full">
          <Bot className="w-12 h-12 text-primary" />
        </div>
      </div>
      
      <div className="space-y-2">
        <label htmlFor="botName" className="text-sm font-medium">
          Chatbot Adı
        </label>
        <Input
          id="botName"
          value={botName}
          onChange={(e) => onBotNameChange(e.target.value)}
          placeholder="Örn: Müşteri Destek Botu"
          className="text-center text-lg"
        />
      </div>
      
      <p className="text-sm text-muted-foreground text-center">
        Bu isim chatbot'unuzu tanımlamak için kullanılacak
      </p>
    </div>
  );
}
```

**File:** `frontend/src/components/onboarding/steps/StepDataSource.tsx`

```typescript
import { SourceType } from '../types';
// ... similar stateless component
```

### Step 4: Refactor Main Component

**File:** `frontend/src/components/onboarding/OnboardingWizard.tsx`

```typescript
import { useOnboardingState } from './hooks/useOnboardingState';
import { StepBotName } from './steps/StepBotName';
import { StepDataSource } from './steps/StepDataSource';
import { StepPersonality } from './steps/StepPersonality';
import { StepComplete } from './steps/StepComplete';

export default function OnboardingWizard() {
  const [state, actions] = useOnboardingState();

  const renderStep = () => {
    switch (state.currentStep) {
      case 1:
        return (
          <StepBotName
            botName={state.botName}
            onBotNameChange={actions.setBotName}
          />
        );
      case 2:
        return (
          <StepDataSource
            sourceType={state.sourceType}
            sourceUrl={state.sourceUrl}
            sourceFile={state.sourceFile}
            sourceText={state.sourceText}
            onSourceTypeChange={actions.setSourceType}
            onSourceUrlChange={actions.setSourceUrl}
            onSourceFileChange={actions.setSourceFile}
            onSourceTextChange={actions.setSourceText}
          />
        );
      case 3:
        return (
          <StepPersonality
            personality={state.personality}
            welcomeMessage={state.welcomeMessage}
            onPersonalityChange={actions.setPersonality}
            onWelcomeMessageChange={actions.setWelcomeMessage}
          />
        );
      case 4:
        return (
          <StepComplete
            botName={state.botName}
            onFinish={actions.finish}
            isSubmitting={state.isSubmitting}
          />
        );
    }
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      <StepIndicator currentStep={state.currentStep} />
      
      <div className="mt-8">
        {renderStep()}
      </div>
      
      <NavigationButtons
        currentStep={state.currentStep}
        onNext={actions.nextStep}
        onBack={actions.prevStep}
        onSkip={actions.skip}
        isSubmitting={state.isSubmitting}
      />
    </div>
  );
}
```

---

## Testing Strategy

### Unit Tests for Steps

```typescript
// StepBotName.test.tsx
describe('StepBotName', () => {
  it('calls onBotNameChange when input changes', () => {
    const onChange = vi.fn();
    render(<StepBotName botName="" onBotNameChange={onChange} />);
    
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'My Bot' } });
    
    expect(onChange).toHaveBeenCalledWith('My Bot');
  });
});
```

### Integration Test for Flow

```typescript
// OnboardingWizard.test.tsx
describe('OnboardingWizard', () => {
  it('completes full onboarding flow', async () => {
    render(<OnboardingWizard />);
    
    // Step 1: Bot Name
    await userEvent.type(screen.getByRole('textbox'), 'Test Bot');
    await userEvent.click(screen.getByText('İleri'));
    
    // Step 2: Data Source
    // ...
  });
});
```

---

## Migration Path

1. **Phase 1:** Create types and hook (non-breaking)
2. **Phase 2:** Create step components (non-breaking)
3. **Phase 3:** Update OnboardingWizard to use new structure
4. **Phase 4:** Remove old inline code

---

## Decision Criteria: When to Do This

| Trigger | Priority |
|---------|----------|
| Adding 2+ new onboarding steps | High |
| A/B testing onboarding variants | High |
| Need unit tests for specific steps | Medium |
| General code quality improvement | Low (current task) |

---

## Related Issues

- Code Audit Finding #2: "Monolithic Frontend Components"
- Future: Onboarding A/B testing capability
