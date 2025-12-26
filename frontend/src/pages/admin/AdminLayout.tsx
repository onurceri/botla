import { Outlet, NavLink, useNavigate } from 'react-router-dom'
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
  Shield,
  Sparkles,
} from 'lucide-react'
import { cn } from '@/lib/utils'

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
  { to: '/admin/privacy', icon: Shield, label: 'Gizlilik Talepleri' },
]

/**
 * AdminLayout - Premium shell layout for admin dashboard with glassmorphism sidebar
 */
export function AdminLayout() {
  const navigate = useNavigate()

  return (
    <div className="flex h-screen bg-background">
      {/* Premium Glassmorphism Sidebar */}
      <aside className="w-72 sidebar-glass border-r border-white/20 flex flex-col shadow-[4px_0_24px_rgba(0,0,0,0.06)]">
        {/* Header */}
        <div className="p-5 border-b border-black/5">
          <div className="flex items-center gap-3">
            <div className="relative">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500 to-red-600 flex items-center justify-center shadow-lg">
                <Shield className="w-5 h-5 text-white" />
              </div>
              {/* Glow effect */}
              <div className="absolute inset-0 w-10 h-10 rounded-xl bg-red-500/30 blur-xl -z-10" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-foreground flex items-center gap-2">
                Admin Panel
                <Sparkles className="w-4 h-4 text-red-500" />
              </h1>
              <p className="text-xs text-muted-foreground">Platform Yönetimi</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="p-3 space-y-1 flex-1 overflow-y-auto scrollbar-none">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                cn(
                  'sidebar-nav-item group',
                  isActive && 'active',
                )
              }
            >
              {({ isActive }) => (
                <>
                  <item.icon
                    className={cn(
                      'nav-icon w-5 h-5 flex-shrink-0 transition-colors',
                      isActive
                        ? 'text-red-500'
                        : 'text-muted-foreground group-hover:text-foreground',
                    )}
                    strokeWidth={isActive ? 2 : 1.5}
                  />
                  <span
                    className={cn(
                      'font-medium transition-colors',
                      isActive ? 'text-foreground' : 'text-muted-foreground',
                    )}
                  >
                    {item.label}
                  </span>
                  {isActive && (
                    <div className="ml-auto w-2 h-2 rounded-full bg-gradient-to-r from-red-500 to-red-600 shadow-[0_0_8px_rgba(239,68,68,0.6)] animate-pulse" />
                  )}
                </>
              )}
            </NavLink>
          ))}
        </nav>

        {/* Back to Dashboard */}
        <div className="p-3 border-t border-black/5">
          <button
            onClick={() => navigate('/dashboard')}
            className="w-full sidebar-nav-item group text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="nav-icon w-5 h-5 flex-shrink-0" strokeWidth={1.5} />
            <span className="font-medium">Dashboard'a Dön</span>
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto bg-background">
        <div className="p-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
