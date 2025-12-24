import { Outlet, NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  Users,
  Building2,
  Bot,
  Database,
  AlertTriangle,
  Activity,
  Clock,
  FileText,
  ArrowLeft,
} from 'lucide-react'

const navItems = [
  { to: '/admin', icon: LayoutDashboard, label: 'Genel Bakış', end: true },
  { to: '/admin/users', icon: Users, label: 'Kullanıcılar' },
  { to: '/admin/organizations', icon: Building2, label: 'Organizasyonlar' },
  { to: '/admin/chatbots', icon: Bot, label: 'Chatbotlar' },
  { to: '/admin/sources', icon: Database, label: 'Kaynaklar' },
  { to: '/admin/system', icon: Activity, label: 'Sistem Durumu' },
  { to: '/admin/queues', icon: Clock, label: 'Kuyruklar' },
  { to: '/admin/errors', icon: AlertTriangle, label: 'Hatalar' },
  { to: '/admin/audit', icon: FileText, label: 'Denetim Günlüğü' },
]

/**
 * AdminLayout - Shell layout for admin dashboard with sidebar navigation
 */
export function AdminLayout() {
  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <aside className="w-64 bg-card border-r border-border flex flex-col">
        <div className="p-4 border-b border-border">
          <h1 className="text-xl font-bold text-destructive">Admin Panel</h1>
          <p className="text-xs text-muted-foreground mt-1">Platform Yönetimi</p>
        </div>

        <nav className="p-4 space-y-1 flex-1 overflow-y-auto">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                }`
              }
            >
              <item.icon className="w-5 h-5" />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="p-4 border-t border-border">
          <NavLink
            to="/dashboard"
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
            Dashboard'a Dön
          </NavLink>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
