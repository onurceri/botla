# Plan: Comprehensive UI/UX Overhaul (Apple Design Style)

## Phase 1: Authentication & Onboarding Overhaul
- [x] Task: Re-design Login and Sign-up UI with "Apple" aesthetics (minimalism, glassmorphism)
    - [x] Write unit tests for the new Auth UI components
    - [x] Implement the new Auth UI using Tailwind and Radix UI
    - [x] Ensure fluid transitions between Login and Sign-up states
- [x] Task: Restructure the "First-to-Launch" Onboarding Flow
    - [x] Write unit tests for the multi-step onboarding wizard
    - [x] Implement a guided, visual journey from registration to bot deployment
    - [x] Add progress indicators and "magical" reveal animations for each step
- [x] Task: Conductor - User Manual Verification 'Authentication & Onboarding Overhaul' (Protocol in workflow.md)

## Phase 2: Bot Configuration & Branding (The "Visual Builder")
- [x] Task: Restructure Dashboard Navigation (7-Tab Structure)
    - [x] Replace 10-tab sidebar with 7-tab horizontal navigation
    - [x] Implement mobile-friendly bottom tab bar (iOS-style)
    - [x] New tabs: Ayarlar, Güvenlik, Kaynaklar, Aksiyonlar, Tasarım, Yayınla, Raporlar
    - [x] Add route redirects for legacy URLs
- [x] Task: Create the Unified Visual Builder Interface
    - [x] Implement the DesignTab with live preview and appearance settings
    - [x] Add glassmorphism effects to the configuration panels
- [x] Task: Implement Real-time Widget Preview
    - [x] Create a "Live Preview" frame that updates instantly as settings change
    - [x] Add responsive preview sizing for mobile and desktop
- [x] Task: Conductor - User Manual Verification 'Bot Configuration & Branding' (Protocol in workflow.md)

## Phase 3: Knowledge Base & RAG Management Streamlining
- [x] Task: Redesign Source Management with "Tactile" Cards
    - [x] Write tests for the new source card components and interactions
    - [x] Implement the visual card-based layout for PDF, URL, and Text sources
    - [x] Add smooth animations for adding/removing sources
- [x] Task: Enhance Ingestion Progress Visualization
    - [x] Write tests for progress tracking UI
    - [x] Implement beautiful, animated progress bars and status indicators for scraping/processing
- [x] Task: Conductor - User Manual Verification 'Knowledge Base & RAG Management' (Protocol in workflow.md)

## Phase 4: Chat Widget & Playground Transformation
- [x] Task: Overhaul Chat Widget UI (iOS-inspired)
    - [x] Write tests for the new Preact widget components
    - [x] Implement glassmorphism, background blurs, and rounded message bubbles
    - [x] Add fluid entrance/exit animations for messages and the widget itself
- [x] Task: Refine the Dashboard Playground
    - [x] Write tests for Playground-specific debug features
    - [x] Redesign the Playground to feel like a premium dev tool
- [x] Task: Conductor - User Manual Verification 'Chat Widget & Playground Transformation' (Protocol in workflow.md)

## Phase 5: Analytics & Insights Reimagining
- [ ] Task: Transform Analytics into "Glanceable" Storytelling Cards
    - [ ] Write tests for the new analytics data visualization components
    - [ ] Replace data tables with beautiful cards and simplified charts
    - [ ] Ensure data is digestible at a glance with high-impact typography
- [ ] Task: Conductor - User Manual Verification 'Analytics & Insights Reimagining' (Protocol in workflow.md)

## Phase 6: Global Responsiveness & Final Polish
- [x] Task: Mobile Optimization Pass (Dashboard)
    - [x] Implement iOS-style bottom navigation for mobile
    - [x] Fix preview area sizing and positioning issues
    - [x] Add bottom padding throughout for navigation clearance
- [ ] Task: Mobile Optimization Pass (Remaining)
    - [ ] Write E2E tests for mobile responsiveness across all new flows
    - [ ] Audit and fix any "non-native" feeling interactions on mobile devices
- [ ] Task: Global Transition and Loading State Audit
    - [ ] Implement unified skeleton screens and loading spinners
    - [ ] Perform a final "polish pass" on all micro-interactions and spacing
- [ ] Task: Conductor - User Manual Verification 'Global Responsiveness & Final Polish' (Protocol in workflow.md)

