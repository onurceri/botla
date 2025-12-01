# Frontend Design & Implementation Plan

## 1. Design Philosophy
**"Premium, Modern, & Data-Driven"**
We will build a high-end SaaS dashboard that feels responsive, robust, and visually stunning.

### 1.1. Visual Identity
-   **Theme**: **Deep Dark Mode**. Not just pure black, but rich dark grays (`#0f172a`, `#1e293b`) to create depth.
-   **Accent Color**: **Botla Cherry** (Vibrant Pink/Red/Purple gradient). Used for primary buttons, active states, and key data points.
    -   Primary: `#ec4899` (Pink-500) to `#ef4444` (Red-500) gradients.
-   **Effects**:
    -   **Glassmorphism**: Subtle transparency on cards and sidebars (`backdrop-blur-md`, `bg-opacity-10`).
    -   **Glows**: Soft shadows and glows behind active elements to give a "neon" feel without being overwhelming.
    -   **Micro-interactions**: Smooth transitions on hover, loading states, and page navigation.

### 1.2. Typography
-   **Font**: `Inter` or `Plus Jakarta Sans` (Modern, geometric sans-serif).
-   **Hierarchy**: Bold, large headings for page titles; clean, legible text for data tables and chat logs.

---

## 2. Layout Structure
We will move away from a simple top-down layout to a professional **Sidebar Layout**.

### 2.1. Sidebar (Left)
-   **Brand**: Logo at the top with a "Beta" or "Pro" badge.
-   **Navigation**:
    -   Dashboard (Analytics)
    -   Chatbots (My Assistants)
    -   Knowledge Base (Global Sources - Optional)
    -   Settings (Account, Billing)
-   **User Profile**: Minimized profile card at the bottom with "Logout".

### 2.2. Top Bar (Header)
-   **Context**: Breadcrumbs (e.g., "Chatbots > Customer Support Bot").
-   **Actions**: Global search, Notifications bell, Theme toggle (optional).

### 2.3. Main Content Area
-   **Padding**: Generous whitespace.
-   **Grid System**: Responsive grid for cards and widgets.

---

## 3. Page Designs

### 3.1. Dashboard (Home)
**Goal**: At-a-glance performance overview.
-   **Hero Section**: "Welcome back, [Name]".
-   **Stats Cards**:
    -   Total Conversations (with % trend).
    -   Total Messages.
    -   Avg. User Satisfaction (Thumbs up/down ratio).
    -   Total Tokens Used (Cost estimation).
-   **Charts**:
    -   Activity over time (Line chart).
    -   Source usage distribution (Pie chart).

### 3.2. Chatbots List
**Goal**: Manage multiple bots easily.
-   **View**: Grid of cards.
-   **Card Content**: Bot Name, Description, Model (GPT-4/3.5), Status (Active/Inactive), Last Updated.
-   **Action**: "Create New Chatbot" button (Prominent).

### 3.3. Chatbot Detail (The Core)
**Goal**: All-in-one management for a specific bot.
**Layout**: Tabbed Interface.
1.  **Overview**: Quick stats specific to this bot.
2.  **Knowledge Base (Sources)**:
    -   **Upload Area**: Drag & drop zone for PDFs, Text files.
    -   **URL Input**: Add website links.
    -   **Source List**: Table showing status (Processing, Ready, Failed) with "Delete" and "Re-sync" options.
3.  **Playground (Test)**:
    -   Split screen: Settings on left (System Prompt, Temperature), Chat window on right.
    -   Real-time testing of the bot behavior.
4.  **Settings**:
    -   General: Name, Description.
    -   Model Config: Model selection, Max Tokens.
    -   Appearance: Widget color, Welcome message, Logo.
5.  **Integration**:
    -   "Copy Widget Code" snippet.
    -   API Key generation for this bot.

### 3.4. Authentication
-   **Login/Register**: Split screen design. Left side: Branding/Art. Right side: Clean form.

---

## 4. Technical Stack & Libraries
-   **Framework**: React (Vite) + TypeScript.
-   **Styling**: Tailwind CSS v4.
-   **Icons**: `lucide-react` (Clean, consistent SVG icons).
-   **Charts**: `recharts` (already installed).
-   **State Management**: React Query (`@tanstack/react-query`) for API data.
-   **UI Components**: We will build a custom "Design System" folder (`src/components/ui`) containing:
    -   `Button` (Variants: Primary, Secondary, Ghost, Destructive)
    -   `Input` / `Textarea`
    -   `Card`
    -   `Modal` / `Dialog`
    -   `Badge`
    -   `Toast` (Notifications)

## 5. Implementation Steps
1.  **Setup Design System**: Define colors in `index.css` / Tailwind config. Create base UI components.
2.  **Layout Implementation**: Create `DashboardLayout` with Sidebar.
3.  **Auth Pages**: Redesign Login/Register.
4.  **Dashboard Page**: Implement Analytics charts using real data.
5.  **Chatbot Flows**: CRUD pages and the complex "Detail" view.
6.  **Polish**: Animations, Loading skeletons, Error states.
