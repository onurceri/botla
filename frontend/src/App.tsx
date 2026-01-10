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
import { AuthProvider, useAuth } from '@/contexts/AuthContext'

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

/**
 * PrivateRoute - Protects routes that require authentication
 * 
 * Uses API-based auth check via AuthContext (not localStorage).
 * The server is the single source of truth for authentication state.
 * 
 * HttpOnly cookies are used for actual API authentication, but the
 * client-side auth state is determined by a successful /api/v1/me call.
 */
function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth()

  // Show loading spinner while checking auth status
  // This prevents flash of login page for authenticated users
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="animate-pulse flex flex-col items-center gap-4">
          <div className="w-12 h-12 rounded-full bg-gradient-to-r from-cyan-500 to-teal-500 animate-spin" />
          <span className="text-slate-400">Yükleniyor...</span>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
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

// Component to handle account deleted notifications
function AccountDeletedHandler() {
  const { toast } = useToast()

  // Use a ref to always have the latest toast function without causing effect re-runs
  const toastRef = useRef(toast)
  useEffect(() => {
    toastRef.current = toast
  }, [toast])

  // Stable event handler using useCallback with no dependencies
  const handleAccountDeleted = useCallback(() => {
    toastRef.current('Hesabınız silindiği için oturumunuz sonlandırıldı.', 'error')
  }, [])

  useEffect(() => {
    // Effect runs only once on mount, cleanup uses the same stable function reference
    window.addEventListener('account-deleted', handleAccountDeleted)
    return () => window.removeEventListener('account-deleted', handleAccountDeleted)
  }, [handleAccountDeleted])

  return null
}

function AppRoutes() {
  return (
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
  )
}

/**
 * Main App Component
 * 
 * Authentication Architecture:
 * - Backend: Sets HttpOnly cookies (botla_token, botla_refresh_token)
 * - Frontend API: Uses axios with withCredentials: true
 * - Auth State: Determined by /api/v1/me API call in AuthContext
 * 
 * This is the secure best practice for cookie-based authentication:
 * - Tokens cannot be accessed by JavaScript (XSS protection)
 * - Server is the single source of truth for auth state
 */
function App() {
  return (
    <ToastProvider>
      <Router>
        <AuthProvider>
          <SessionExpiryHandler />
          <AccountDeletedHandler />
          <AppRoutes />
        </AuthProvider>
      </Router>
    </ToastProvider>
  )
}

export default App
