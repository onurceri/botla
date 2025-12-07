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

const isAuthenticated = () => {
  // In E2E mode, bypass auth gating for visual tests
  // VITE_E2E is set in Playwright webServer env
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
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
              <DashboardLayout />
            </PrivateRoute>
          }>
            <Route index element={<DashboardPage />} />
            <Route path="chatbots" element={<ChatbotsPage />} />
            <Route path="chatbots/:id" element={<ChatbotDetailPage />} />
            <Route path="settings" element={<Navigate to="/settings/profile" replace />} />
            <Route path="settings/profile" element={<ProfilePage />} />
            <Route path="settings/plan" element={<PlanPage />} />
          </Route>
          
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ToastProvider>
  )
}

export default App
