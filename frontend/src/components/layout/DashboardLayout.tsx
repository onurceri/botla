import { useState } from 'react'
import { Link, useLocation, useNavigate, Outlet } from 'react-router-dom'
import {
  LayoutDashboard,
  Bot,
  LogOut,
  Menu,
  X,
  ChevronRight,
  Pin,
  MousePointer,
  User,
  CreditCard,
  Shield,
  Sparkles,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { OrganizationSwitcher } from '@/features/organization/components/OrganizationSwitcher'
import { usePlan, useProfile } from '@/hooks/queries/useProfile'

/**
 * Premium sidebar navigation item with animated indicator
 */
const SidebarItem = ({
  icon: Icon,
  label,
  to,
  active,
  collapsed,
}: {
  icon: any
  label: string
  to: string
  active?: boolean
  collapsed?: boolean
}) => {
  return (
    <Link to={to}>
      <div
        className={cn(
          'sidebar-nav-item',
          active && 'active',
          collapsed && 'lg:justify-center lg:group-hover/sidebar:justify-start',
        )}
      >
        <Icon
          className={cn(
            'nav-icon w-5 h-5 flex-shrink-0',
            active
              ? 'text-primary'
              : 'text-muted-foreground group-hover:text-foreground',
          )}
          strokeWidth={active ? 2 : 1.5}
        />
        <span
          className={cn(
            'font-medium transition-colors duration-200',
            active ? 'text-foreground' : 'text-muted-foreground',
            collapsed ? 'lg:hidden lg:group-hover/sidebar:inline' : undefined,
          )}
        >
          {label}
        </span>
        {active && (
          <div
            className={cn(
              'nav-indicator-dot ml-auto',
              collapsed && 'lg:hidden lg:group-hover/sidebar:block',
            )}
          />
        )}
      </div>
    </Link>
  )
}

/**
 * Premium Dashboard Layout with glassmorphism sidebar
 */
const DashboardLayout = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [sidebarMode, setSidebarMode] = useState<'pinned' | 'hover'>(() => {
    const raw =
      typeof window !== 'undefined' ? window.localStorage.getItem('botla_sidebar_mode') : null
    return (raw as 'pinned' | 'hover') || 'hover'
  })
  const isCollapsed = sidebarMode === 'hover'
  const { data: profile } = useProfile()
  const { data: plan } = usePlan()

  const profileName = profile?.full_name || ''
  const profileEmail = profile?.email || ''
  const planName = plan?.name || ''
  const planCode = plan?.code || 'free'

  const toggleSidebarMode = () => {
    const next = sidebarMode === 'pinned' ? 'hover' : 'pinned'
    setSidebarMode(next)
    window.localStorage.setItem('botla_sidebar_mode', next)
  }

  const handleLogout = () => {
    window.localStorage.removeItem('botla_token')
    window.localStorage.removeItem('botla_refresh_token')
    navigate('/login')
  }

  const navItems = [
    { icon: LayoutDashboard, label: 'Panel', to: '/dashboard' },
    { icon: Bot, label: 'Chatbotlar', to: '/dashboard/chatbots' },
  ]

  if (profile?.is_platform_admin) {
    navItems.push({ icon: Shield, label: 'Yönetim', to: '/admin' })
  }

  const settingsItems = [
    { icon: User, label: 'Profil', to: '/dashboard/settings/profile' },
    { icon: CreditCard, label: 'Plan', to: '/dashboard/settings/plan' },
    { icon: Shield, label: 'Gizlilik', to: '/dashboard/settings/privacy' },
  ]

  // Get plan badge color
  const getPlanBadgeStyle = () => {
    switch (planCode.toLowerCase()) {
      case 'pro':
        return 'bg-gradient-to-r from-primary to-orange-500 text-white'
      case 'enterprise':
        return 'bg-gradient-to-r from-violet-500 to-purple-600 text-white'
      default:
        return 'bg-muted text-muted-foreground'
    }
  }

  return (
    <div className="h-screen bg-background text-foreground flex overflow-hidden">
      {/* Mobile Overlay */}
      {isMobileMenuOpen && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden animate-in fade-in duration-200"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* Premium Glassmorphism Sidebar */}
      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 flex flex-col sidebar-glass border-r border-white/20 transition-all duration-300 ease-out group/sidebar',
          // Mobile behavior
          'w-72',
          isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full',
          // Desktop behavior - always fixed, never static
          'lg:translate-x-0',
          isCollapsed ? 'lg:w-[72px] lg:hover:w-72' : 'lg:w-72',
          // Premium shadow
          'shadow-[4px_0_24px_rgba(0,0,0,0.06)]',
        )}
      >
        {/* Logo Area */}
        <div className={cn(
          "h-16 flex items-center border-b border-black/5 transition-all duration-300 ease-in-out",
          isCollapsed 
            ? "justify-center lg:group-hover/sidebar:justify-start lg:group-hover/sidebar:px-5" 
            : "px-5"
        )}>
          <Link to="/dashboard" className="flex items-center gap-3 logo-glow">
            <div className="relative flex-shrink-0">
              <img
                src="/logo-128.png"
                alt="Botla Logo"
                className="w-9 h-9 rounded-xl shadow-lg"
              />
              {/* Subtle glow behind logo */}
              <div className="absolute inset-0 w-9 h-9 rounded-xl bg-primary/20 blur-xl -z-10" />
            </div>
            <span
              className={cn(
                'font-bold text-lg tracking-tight bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text whitespace-nowrap',
                isCollapsed ? 'hidden lg:group-hover/sidebar:inline' : undefined,
              )}
            >
              botla.app
            </span>
          </Link>
          
          <div className={cn(
            "ml-auto flex items-center gap-2 overflow-hidden",
            isCollapsed && "hidden lg:group-hover/sidebar:flex"
          )}>
            {/* Pin/Hover toggle */}
            <Button
              variant="ghost"
              size="icon"
              className="hidden lg:inline-flex h-8 w-8 rounded-lg hover:bg-black/5 transition-all duration-200"
              onClick={toggleSidebarMode}
              title={sidebarMode === 'pinned' ? 'Sabit → Hover' : 'Hover → Sabit'}
            >
              {sidebarMode === 'pinned' ? (
                <Pin className="w-4 h-4 text-primary" />
              ) : (
                <MousePointer className="w-4 h-4 text-muted-foreground" />
              )}
            </Button>
            {/* Mobile close button */}
            <button
              className="lg:hidden text-muted-foreground hover:text-foreground transition-colors"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              <X className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* Navigation */}
        <div className="flex-1 py-6 px-3 space-y-6 overflow-y-auto overflow-x-hidden scrollbar-none">
          {/* Platform Section */}
          <div>
            <div
              className={cn(
                'sidebar-section-label',
                isCollapsed ? 'lg:hidden lg:group-hover/sidebar:block' : undefined,
              )}
            >
              Platform
            </div>
            <div className="space-y-1">
              {navItems.map((item) => (
                <SidebarItem
                  key={item.to}
                  icon={item.icon}
                  label={item.label}
                  to={item.to}
                  active={
                    item.to === '/dashboard'
                      ? location.pathname === '/dashboard'
                      : location.pathname.startsWith(item.to)
                  }
                  collapsed={isCollapsed}
                />
              ))}
            </div>
          </div>

          {/* Settings Section */}
          <div>
            <div
              className={cn(
                'sidebar-section-label',
                isCollapsed ? 'lg:hidden lg:group-hover/sidebar:block' : undefined,
              )}
            >
              Ayarlar
            </div>
            <div className="space-y-1">
              {settingsItems.map((item) => (
                <SidebarItem
                  key={item.to}
                  icon={item.icon}
                  label={item.label}
                  to={item.to}
                  active={location.pathname.startsWith(item.to)}
                  collapsed={isCollapsed}
                />
              ))}
            </div>
          </div>
        </div>

        {/* User Profile / Logout */}
        <div
          className={cn(
            'border-t border-black/5 p-3',
            isCollapsed ? 'lg:p-2 lg:group-hover/sidebar:p-3' : undefined,
          )}
        >
          {/* User Profile Card */}
          <div
            className={cn(
              'user-profile-card flex items-center gap-3 mb-3 cursor-pointer',
              isCollapsed ? 'lg:hidden lg:group-hover/sidebar:flex' : undefined,
            )}
            onClick={() => navigate('/dashboard/settings/profile')}
          >
            {/* Avatar with gradient ring */}
            <div className="avatar-ring">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary via-orange-400 to-amber-300 flex items-center justify-center text-white font-semibold text-sm shadow-lg">
                {(profileName || profileEmail || 'U').charAt(0).toUpperCase()}
              </div>
            </div>
            <div className="flex-1 min-w-0">
              <div className="text-sm font-semibold truncate text-foreground">
                {profileName || profileEmail || 'Kullanıcı'}
              </div>
              <div className="flex items-center gap-1.5 mt-0.5">
                <span
                  className={cn(
                    'text-[10px] font-medium px-1.5 py-0.5 rounded-full',
                    getPlanBadgeStyle(),
                  )}
                >
                  {planName || planCode.toUpperCase()}
                </span>
                {planCode !== 'free' && (
                  <Sparkles className="w-3 h-3 text-primary" />
                )}
              </div>
            </div>
          </div>

          {/* Collapsed avatar only */}
          <div
            className={cn(
              'hidden mb-3',
              isCollapsed ? 'lg:flex lg:justify-center lg:group-hover/sidebar:hidden' : undefined,
            )}
          >
            <div className="avatar-ring">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary via-orange-400 to-amber-300 flex items-center justify-center text-white font-semibold text-sm shadow-lg">
                {(profileName || profileEmail || 'U').charAt(0).toUpperCase()}
              </div>
            </div>
          </div>

          {/* Logout Button */}
          <Button
            variant="ghost"
            className={cn(
              'logout-btn w-full justify-start rounded-xl text-muted-foreground',
              isCollapsed
                ? 'lg:justify-center lg:group-hover/sidebar:justify-start lg:px-2 lg:group-hover/sidebar:px-3'
                : 'px-3',
            )}
            onClick={handleLogout}
          >
            <LogOut
              className={cn(
                'w-5 h-5 flex-shrink-0',
                isCollapsed ? 'lg:mr-0 lg:group-hover/sidebar:mr-3' : 'mr-3',
              )}
            />
            <span
              className={cn(
                'font-medium',
                isCollapsed ? 'lg:hidden lg:group-hover/sidebar:inline' : undefined,
              )}
            >
              Çıkış Yap
            </span>
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <main className={cn(
        "flex-1 flex flex-col min-w-0 overflow-hidden bg-background",
        // Add margin-left to account for fixed sidebar on desktop
        isCollapsed ? "lg:ml-[72px]" : "lg:ml-72"
      )}>
        {/* Top Header */}
        <header className="h-16 border-b border-border bg-background/80 backdrop-blur-md flex items-center justify-between px-4 lg:px-8 sticky top-0 z-30">
          <div className="flex items-center gap-4">
            <button
              className="lg:hidden text-muted-foreground hover:text-foreground transition-colors"
              onClick={() => setIsMobileMenuOpen(true)}
            >
              <Menu className="w-6 h-6" />
            </button>

            {/* Breadcrumbs */}
            <div className="hidden md:flex items-center text-sm text-muted-foreground">
              <span className="font-medium">Botla</span>
              <ChevronRight className="w-4 h-4 mx-1.5 text-border" />
              <span className="text-foreground font-semibold">
                {navItems.find((i) => location.pathname.startsWith(i.to) && i.to !== '/')?.label ||
                  settingsItems.find((i) => location.pathname.startsWith(i.to))?.label ||
                  'Panel'}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <OrganizationSwitcher />
          </div>
        </header>

        {/* Page Content */}
        <div className="flex-1 overflow-auto p-3 sm:p-4 lg:p-6">
          <div className="max-w-7xl mx-auto animate-in fade-in slide-in-from-bottom-4 duration-500">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  )
}

export default DashboardLayout
