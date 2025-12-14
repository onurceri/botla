# AGENTS.md - Botla Frontend

A React-based dashboard frontend for the Botla chatbot platform. Built with TypeScript, Vite, and Tailwind CSS.

## Project Structure

```
frontend/
├── src/
│   ├── api/              # API client and endpoint modules
│   │   └── __tests__/    # API unit tests
│   ├── components/       # Shared UI components (Button, Card, Input, etc.)
│   ├── features/         # Feature-based modules
│   │   ├── chatbot/      # Chatbot management components
│   │   ├── sources/      # Source management (URL, file upload)
│   │   └── analytics/    # Analytics charts and displays
│   ├── pages/            # Route page components
│   │   └── __tests__/    # Page component tests
│   ├── lib/              # Utility functions
│   ├── App.tsx           # Main app with routing
│   ├── main.tsx          # Entry point
│   ├── index.css         # Global styles and Tailwind imports
│   └── setupTests.ts     # Test configuration
├── e2e/                  # Playwright E2E tests
├── public/               # Static assets
└── dist/                 # Production build output
```

## Dev Environment Setup

### Prerequisites
- Node.js 18+
- npm 9+

### Install Dependencies
```bash
cd frontend
npm install
```

## Build & Run Commands

```bash
# Start development server (port 5173)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Testing Instructions

### Run Tests
```bash
# Run unit tests with Vitest
npm run test

# Run tests with coverage
npm run test:coverage

# Run E2E tests (requires running dev server)
npm run e2e

# Run E2E tests in headed mode (visible browser)
npm run e2e:headed

# Run E2E tests for CI
npm run e2e:ci
```

### Test Structure
- Unit tests: `**/__tests__/*.test.ts(x)` using Vitest + Testing Library
- E2E tests: `e2e/*.spec.ts` using Playwright
- Test mocking: Uses `vi.mock()` for API mocking
- Test utilities: `@testing-library/react`, `@testing-library/user-event`

### Running Specific Tests
```bash
# Run specific test file
npx vitest run src/pages/__tests__/LoginPage.test.tsx

# Run tests matching pattern
npx vitest run -t "should display"

# Watch mode for development
npx vitest
```

## Code Style Guidelines

### Linting & Formatting
```bash
# Run ESLint
npm run lint

# Type check without emitting
npm run typecheck

# Format with Prettier
npm run format

# Check formatting
npm run format:check

# Full CI check (lint + typecheck + test:coverage)
npm run ci
```

### TypeScript Configuration
- Strict mode enabled
- Path aliases: `@/*` → `src/*`, `@widget/*` → `../widget/src/*`
- Target: ES2020
- No unused locals/parameters

### Prettier Configuration
```json
{
  "printWidth": 100,
  "semi": false,
  "singleQuote": true,
  "trailingComma": "all",
  "arrowParens": "always"
}
```

### Conventions
- **Components**: Functional components with TypeScript interfaces for props
- **State management**: React Query (`@tanstack/react-query`) for server state
- **Routing**: React Router v7 (`react-router-dom`)
- **Styling**: Tailwind CSS v4 with utility-first approach
- **UI Components**: Radix UI primitives (`@radix-ui/*`) for accessible components
- **Icons**: Lucide React icons (`lucide-react`)
- **API calls**: Axios-based client in `src/api/client.ts`
- **Feature organization**: Group related components in `src/features/<feature>/`

### Component Guidelines
```typescript
// Prefer explicit interface definitions
interface ComponentProps {
  title: string
  onAction: () => void
}

// Use destructuring with defaults
export function Component({ title, onAction }: ComponentProps) {
  return (...)
}
```

### Naming Conventions
- Files: PascalCase for components (`ChatbotDetailPage.tsx`)
- Test files: `ComponentName.test.tsx` or `ComponentName.<scenario>.test.tsx`
- API modules: camelCase (`chatbot.ts`, `source.ts`)
- CSS classes: Tailwind utilities, no custom CSS unless necessary

## API Integration

### Client Setup
- Base URL configured via `VITE_API_URL` environment variable
- Axios instance with interceptors in `src/api/client.ts`
- JWT token handling with automatic refresh

### Environment Variables
```bash
# .env.development
VITE_API_URL=http://localhost:8080

# .env.production
VITE_API_URL=https://api.botla.app
```

## UI Component Library

Using Radix UI primitives with custom styling:
- `@radix-ui/react-slot` - Slot component for composition
- `@radix-ui/react-switch` - Toggle switches
- `@radix-ui/react-tabs` - Tab navigation
- `@radix-ui/react-tooltip` - Tooltips

Helper utilities:
- `class-variance-authority` (CVA) - Variant management
- `clsx` + `tailwind-merge` - Class name merging

## Charts & Analytics

Using Recharts for data visualization with responsive containers.

## Security Considerations

- Auth tokens stored in memory/localStorage
- Protected routes via `PrivateRoute` component
- API calls include auth headers automatically
- XSS prevention via React's built-in escaping

## PR Instructions

- Run `npm run ci` before committing
- All tests must pass
- No TypeScript errors
- No ESLint warnings
- Follow existing code patterns and naming conventions
