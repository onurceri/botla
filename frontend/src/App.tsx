import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import DashboardLayout from '@/components/layout/DashboardLayout'
import DashboardPage from '@/pages/DashboardPage'
import ChatbotsPage from '@/pages/ChatbotsPage'
import ChatbotDetailPage from '@/pages/ChatbotDetailPage'
import ProfilePage from '@/pages/ProfilePage'
import PlanPage from '@/pages/PlanPage'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import { ToastProvider } from '@/components/ui/toast'
import { OrganizationProvider } from '@/features/organization/context/OrganizationContext'
import { OrganizationSettingsPage } from '@/features/organization/pages/OrganizationSettingsPage'
import { WorkspaceSettingsPage } from '@/features/organization/pages/WorkspaceSettingsPage'

import OverviewTab from '@/features/chatbot/pages/tabs/OverviewTab'
import GuardrailsTab from '@/features/chatbot/pages/tabs/GuardrailsTab'
import HandoffTab from '@/features/chatbot/pages/tabs/HandoffTab'
import SourcesTab from '@/features/chatbot/pages/tabs/SourcesTab'
import ActionsTab from '@/features/chatbot/pages/tabs/ActionsTab'
import SuggestionsTab from '@/features/chatbot/pages/tabs/SuggestionsTab'
import PlaygroundTab from '@/features/chatbot/pages/tabs/PlaygroundTab'
import ConnectTab from '@/features/chatbot/pages/tabs/ConnectTab'
import AnalyticsTab from '@/features/chatbot/pages/tabs/AnalyticsTab'

const isAuthenticated = () => {
  // In E2E mode, bypass auth gating for visual tests
  // VITE_E2E is set in Playwright webServer env
  // @ts-ignore
  if (import.meta.env && (import.meta.env as any).VITE_E2E) return true
  return !!localStorage.getItem('botla_token')
}

const PrivateRoute = ({ children }: { children: React.ReactNode }) => {
  return isAuthenticated() ? <>{children}</> : <Navigate to="/login" />
}

function App() {
  return (
    <ToastProvider>
      <Router>
        <Routes>
          {/* Public Routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* Protected Routes */}
          <Route path="/" element={
            <PrivateRoute>
              <OrganizationProvider>
                <DashboardLayout />
              </OrganizationProvider>
            </PrivateRoute>
          }>
            <Route index element={<DashboardPage />} />
            <Route path="chatbots" element={<ChatbotsPage />} />
            
            <Route path="chatbots/:id" element={<ChatbotDetailPage />}>
              <Route index element={<Navigate to="overview" replace />} />
              <Route path="overview" element={<OverviewTab />} />
              <Route path="guardrails" element={<GuardrailsTab />} />
              <Route path="handoff" element={<HandoffTab />} />
              <Route path="sources" element={<SourcesTab />} />
              <Route path="actions" element={<ActionsTab />} />
              <Route path="suggestions" element={<SuggestionsTab />} />
              <Route path="playground" element={<PlaygroundTab />} />
              <Route path="connect" element={<ConnectTab />} />
              <Route path="analytics" element={<AnalyticsTab />} />
            </Route>

            <Route path="settings" element={<Navigate to="/settings/profile" replace />} />
            <Route path="settings/profile" element={<ProfilePage />} />
            <Route path="settings/organization" element={<OrganizationSettingsPage />} />
            <Route path="settings/workspace" element={<WorkspaceSettingsPage />} />
            <Route path="settings/plan" element={<PlanPage />} />
          </Route>
          
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ToastProvider>
  )
}

export default App
