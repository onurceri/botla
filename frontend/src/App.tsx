import { useEffect, useRef, useCallback } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import DashboardLayout from '@/components/layout/DashboardLayout'
import DashboardPage from '@/pages/DashboardPage'
import ChatbotsPage from '@/pages/ChatbotsPage'
import ChatbotDetailPage from '@/pages/ChatbotDetailPage'
import ProfilePage from '@/pages/ProfilePage'
import PlanPage from '@/pages/PlanPage'
import PrivacySettingsPage from '@/pages/PrivacySettingsPage'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import OnboardingPage from '@/pages/OnboardingPage'
import { ToastProvider, useToast } from '@/components/ui/toast'
import { OrganizationProvider } from '@/features/organization/context/OrganizationContext'
import { OrganizationSettingsPage } from '@/features/organization/pages/OrganizationSettingsPage'
import { WorkspaceSettingsPage } from '@/features/organization/pages/WorkspaceSettingsPage'
import LandingPage from '@/pages/LandingPage'

// Dashboard 7-tab structure
import SettingsTab from '@/features/chatbot/pages/tabs/SettingsTab'
import SecurityTab from '@/features/chatbot/pages/tabs/SecurityTab'
import SourcesTab from '@/features/chatbot/pages/tabs/SourcesTab'
import ActionsTab from '@/features/chatbot/pages/tabs/ActionsTab'
import PlaygroundTab from '@/features/chatbot/pages/tabs/PlaygroundTab'
import DeployTab from '@/features/chatbot/pages/tabs/DeployTab'
import InsightsTab from '@/features/chatbot/pages/tabs/InsightsTab'

// Admin routes
import { AdminRoute } from '@/features/admin/AdminRoute'
import {
  AdminLayout,
  AdminDashboardPage,
  AdminUsersPage,
  AdminOrganizationsPage,
  AdminChatbotsPage,
  AdminSourcesPage,
  AdminSystemPage,
  AdminQueuesPage,
  AdminErrorsPage,
  AdminAuditPage,
  AdminPrivacyPage,
} from '@/pages/admin'

// Helper to validate stored tokens - checks JWT format and expiry
export const isValidToken = (token: string | null): boolean => {
  // Basic null/empty checks
  if (token === null || token === 'undefined' || token === 'null' || token.length === 0) {
    return false
  }

  // JWT format check: must have exactly 3 parts separated by dots
  const parts = token.split('.')
  if (parts.length !== 3) {
    return false
  }

  // Each part must be non-empty and valid base64url
  const base64urlRegex = /^[A-Za-z0-9_-]+$/
  for (const part of parts) {
    if (part.length === 0 || !base64urlRegex.test(part)) {
      return false
    }
  }

  // Optional: Check expiry from payload
  try {
    // Decode the payload (second part) - replace base64url chars with base64
    const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const decoded = JSON.parse(atob(payload))
    
    // If exp claim exists, check if token is expired
    if (decoded.exp && typeof decoded.exp === 'number') {
      const now = Math.floor(Date.now() / 1000)
      if (decoded.exp < now) {
        return false
      }
    }
  } catch {
    // If decoding fails, the token structure is invalid
    return false
  }

  return true
}

const isAuthenticated = () => {
  // In E2E mode, bypass auth gating for visual tests
  // VITE_E2E is set in Playwright webServer env
  if (import.meta.env.VITE_E2E) return true
  const token = typeof window !== 'undefined' ? window.localStorage.getItem('botla_token') : null
  return isValidToken(token)
}

const PrivateRoute = ({ children }: { children: React.ReactNode }) => {
  return isAuthenticated() ? <>{children}</> : <Navigate to="/login" />
}

// Component to handle session expiry notifications
function SessionExpiryHandler() {
  const { toast } = useToast()
  
  // Use a ref to always have the latest toast function without causing effect re-runs
  const toastRef = useRef(toast)
  useEffect(() => {
    toastRef.current = toast
  }, [toast])
  
  // Stable event handler using useCallback with no dependencies
  const handleSessionExpired = useCallback(() => {
    toastRef.current('Oturumunuz sona erdi. Lütfen tekrar giriş yapın.', 'error')
  }, [])

  useEffect(() => {
    // Effect runs only once on mount, cleanup uses the same stable function reference
    window.addEventListener('session-expired', handleSessionExpired)
    return () => window.removeEventListener('session-expired', handleSessionExpired)
  }, [handleSessionExpired])

  return null
}

function App() {
  return (
    <ToastProvider>
      <SessionExpiryHandler />
      <Router>
        <Routes>
          {/* Public Routes */}
          <Route path="/" element={<LandingPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/onboarding"
            element={
              <PrivateRoute>
                <OrganizationProvider>
                  <OnboardingPage />
                </OrganizationProvider>
              </PrivateRoute>
            }
          />

          {/* Protected Routes */}
          <Route
            path="/dashboard"
            element={
              <PrivateRoute>
                <OrganizationProvider>
                  <DashboardLayout />
                </OrganizationProvider>
              </PrivateRoute>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="chatbots" element={<ChatbotsPage />} />

            <Route path="chatbots/:id" element={<ChatbotDetailPage />}>
              <Route index element={<Navigate to="settings" replace />} />
              <Route path="settings" element={<SettingsTab />} />
              <Route path="security" element={<SecurityTab />} />
              <Route path="sources" element={<SourcesTab />} />
              <Route path="actions" element={<ActionsTab />} />
              <Route path="playground" element={<PlaygroundTab />} />
              <Route path="deploy" element={<DeployTab />} />
              <Route path="insights" element={<InsightsTab />} />
            </Route>

            <Route
              path="settings"
              element={<Navigate to="/dashboard/settings/profile" replace />}
            />
            <Route path="settings/profile" element={<ProfilePage />} />
            <Route path="settings/organization" element={<OrganizationSettingsPage />} />
            <Route path="settings/workspace" element={<WorkspaceSettingsPage />} />
            <Route path="settings/plan" element={<PlanPage />} />
            <Route path="settings/privacy" element={<PrivacySettingsPage />} />
          </Route>

          {/* Admin Routes - Protected by AdminRoute */}
          <Route
            path="/admin"
            element={
              <PrivateRoute>
                <AdminRoute>
                  <AdminLayout />
                </AdminRoute>
              </PrivateRoute>
            }
          >
            <Route index element={<AdminDashboardPage />} />
            <Route path="users" element={<AdminUsersPage />} />
            <Route path="organizations" element={<AdminOrganizationsPage />} />
            <Route path="chatbots" element={<AdminChatbotsPage />} />
            <Route path="sources" element={<AdminSourcesPage />} />
            <Route path="system" element={<AdminSystemPage />} />
            <Route path="queues" element={<AdminQueuesPage />} />
            <Route path="errors" element={<AdminErrorsPage />} />
            <Route path="audit" element={<AdminAuditPage />} />
            <Route path="privacy" element={<AdminPrivacyPage />} />
          </Route>

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ToastProvider>
  )
}

export default App
