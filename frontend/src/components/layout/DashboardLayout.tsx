import { useEffect, useState } from 'react'
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
  CreditCard
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { api } from '@/api/client'
import { OrganizationSwitcher } from '@/features/organization/components/OrganizationSwitcher'

const SidebarItem = ({ 
  icon: Icon, 
  label, 
  to, 
  active,
  collapsed
}: { 
  icon: any, 
  label: string, 
  to: string, 
  active?: boolean,
  collapsed?: boolean
}) => {
  return (
    <Link to={to}>
      <div className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 group",
        active 
          ? "bg-primary/10 text-foreground font-medium" 
          : "text-muted-foreground hover:bg-white/5 hover:text-foreground",
        collapsed && "lg:justify-center lg:group-hover/sidebar:justify-start"
      )}>
        <Icon className={cn("w-5 h-5", active ? "text-primary" : "text-muted-foreground group-hover:text-foreground")} strokeWidth={1.5} />
        <span className={cn(collapsed ? "hidden lg:group-hover/sidebar:inline" : undefined)}>{label}</span>
        {active && (
          <div className={cn(
            "ml-auto w-1.5 h-1.5 rounded-full bg-primary shadow-[0_0_8px_rgba(167,139,250,0.8)]",
            collapsed && "hidden lg:group-hover/sidebar:block"
          )} />
        )}
      </div>
    </Link>
  )
}

const DashboardLayout = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [sidebarMode, setSidebarMode] = useState<'pinned' | 'hover'>(() => (localStorage.getItem('botla_sidebar_mode') as 'pinned' | 'hover') || 'hover')
  const isCollapsed = sidebarMode === 'hover'
  const [profileName, setProfileName] = useState<string>('')
  const [profileEmail, setProfileEmail] = useState<string>('')
  const [planName, setPlanName] = useState<string>('')
  const [planCode, setPlanCode] = useState<string>('free')

  const toggleSidebarMode = () => {
    const next = sidebarMode === 'pinned' ? 'hover' : 'pinned'
    setSidebarMode(next)
    localStorage.setItem('botla_sidebar_mode', next)
  }

  const handleLogout = () => {
    localStorage.removeItem('botla_token')
    localStorage.removeItem('botla_refresh_token')
    navigate('/login')
  }

  useEffect(() => {
    api.get('/api/v1/me')
      .then((res) => {
        setProfileName(res.data?.full_name ?? '')
        setProfileEmail(res.data?.email ?? '')
        const code = res.data?.subscription_plan ?? res.data?.plan_code ?? 'free'
        setPlanCode(code)
        setPlanName(res.data?.plan_name ?? '')
      })
      .catch(() => {
        setProfileName('')
        setProfileEmail('')
        setPlanCode('free')
        setPlanName('')
      })
  }, [])

  const navItems = [
    { icon: LayoutDashboard, label: 'Panel', to: '/dashboard' },
    { icon: Bot, label: 'Chatbotlar', to: '/dashboard/chatbots' },
  ]

  const settingsItems = [
    { icon: User, label: 'Profil', to: '/dashboard/settings/profile' },
    { icon: CreditCard, label: 'Plan & Fatura', to: '/dashboard/settings/plan' },
  ]

  return (
    <div className="min-h-screen bg-background text-foreground flex">
      {/* Mobile Overlay */}
      {isMobileMenuOpen && (
        <div 
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside className={cn(
        "fixed inset-y-0 left-0 z-50 flex flex-col bg-card border-r border-border transition-transform duration-300 ease-in-out group/sidebar",
        // Mobile behavior: fixed width, transform based on state
        "w-64", 
        isMobileMenuOpen ? "translate-x-0" : "-translate-x-full",
        // Desktop behavior: static position, always visible (reset transform), width controlled by collapse state
        "lg:static lg:translate-x-0",
        isCollapsed ? "lg:w-16 lg:hover:w-64" : "lg:w-64"
      )}>
        {/* Logo Area */}
        <div className="h-16 flex items-center px-6 border-b border-border">
          <div className="flex items-center gap-2 font-bold text-xl tracking-tight">
            <img src="/logo-128.png" alt="Botla Logo" className="w-8 h-8 rounded-lg shadow-lg shadow-primary/20" />
            <span className={cn("text-foreground", isCollapsed ? "hidden lg:group-hover/sidebar:inline" : undefined)}>
              Botla.co
            </span>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <Button 
              variant="ghost"
              size="icon"
              className="hidden lg:inline-flex"
              onClick={toggleSidebarMode}
              title={sidebarMode === 'pinned' ? 'Sabit → Hover' : 'Hover → Sabit'}
            >
              {sidebarMode === 'pinned' ? <Pin className="w-4 h-4" /> : <MousePointer className="w-4 h-4" />}
            </Button>
            <button 
              className="lg:hidden text-muted-foreground"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              <X className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* Navigation */}
        <div className="flex-1 py-6 px-3 space-y-1 overflow-y-auto overflow-x-hidden">
          <div className={cn("px-3 mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider", isCollapsed ? "hidden lg:group-hover/sidebar:block" : undefined)}>
            Platform
          </div>
          {navItems.map((item) => (
            <SidebarItem 
              key={item.to}
              icon={item.icon}
              label={item.label}
              to={item.to}
              active={item.to === '/dashboard' ? location.pathname === '/dashboard' : location.pathname.startsWith(item.to)}
              collapsed={isCollapsed}
            />
          ))}

          <div className={cn("px-3 mt-6 mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider", isCollapsed ? "hidden lg:group-hover/sidebar:block" : undefined)}>
            Ayarlar
          </div>
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

        {/* User Profile / Logout */}
        <div className={cn("border-t border-border", isCollapsed ? "p-2 lg:group-hover/sidebar:p-4" : "p-4")}>
          <div className={cn("bg-muted/50 rounded-xl p-3 flex items-center gap-3 mb-3", isCollapsed ? "hidden lg:group-hover/sidebar:flex" : undefined)}>
            <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-primary to-accent" />
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate text-foreground">{profileName || profileEmail}</div>
              <div className="text-xs text-muted-foreground truncate">{planName || planCode}</div>
            </div>
          </div>
          <Button 
            variant="ghost" 
            className={cn(
              "w-full justify-start hover:text-destructive hover:bg-destructive/10",
              isCollapsed ? "text-muted-foreground lg:justify-center lg:group-hover/sidebar:justify-start lg:group-hover/sidebar:text-foreground px-2 lg:group-hover/sidebar:px-4" : "text-foreground"
            )}
            onClick={handleLogout}
          >
            <LogOut className={cn("w-4 h-4", isCollapsed ? "lg:group-hover/sidebar:mr-2" : "mr-2")} />
            <span className={cn(isCollapsed ? "hidden lg:group-hover/sidebar:inline" : undefined)}>Çıkış Yap</span>
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden bg-background">
        {/* Top Header */}
        <header className="h-16 border-b border-border bg-background/80 backdrop-blur-md flex items-center justify-between px-4 lg:px-8 sticky top-0 z-30">
          <div className="flex items-center gap-4">
            <button 
              className="lg:hidden text-muted-foreground hover:text-foreground"
              onClick={() => setIsMobileMenuOpen(true)}
            >
              <Menu className="w-6 h-6" />
            </button>
            
            {/* Breadcrumbs (Simple) */}
            <div className="hidden md:flex items-center text-sm text-muted-foreground">
              <span>Botla</span>
              <ChevronRight className="w-4 h-4 mx-1" />
              <span className="text-foreground font-medium">
                {navItems.find(i => location.pathname.startsWith(i.to) && i.to !== '/')?.label || 'Panel'}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <OrganizationSwitcher />
          </div>
        </header>

        {/* Page Content */}
        <div className="flex-1 overflow-auto p-4 lg:p-8">
          <div className="max-w-7xl mx-auto animate-in fade-in slide-in-from-bottom-4 duration-500">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  )
}

export default DashboardLayout
