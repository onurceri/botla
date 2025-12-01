import { useState } from 'react'
import { Link, useLocation, useNavigate, Outlet } from 'react-router-dom'
import { 
  LayoutDashboard, 
  Bot, 
  Settings, 
  LogOut, 
  Menu, 
  X,
  ChevronRight
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

const SidebarItem = ({ 
  icon: Icon, 
  label, 
  to, 
  active 
}: { 
  icon: any, 
  label: string, 
  to: string, 
  active?: boolean 
}) => {
  return (
    <Link to={to}>
      <div className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 group",
        active 
          ? "bg-primary/10 text-primary font-medium" 
          : "text-muted-foreground hover:bg-white/5 hover:text-foreground"
      )}>
        <Icon className={cn("w-5 h-5", active ? "text-primary" : "text-muted-foreground group-hover:text-foreground")} />
        <span>{label}</span>
        {active && <div className="ml-auto w-1.5 h-1.5 rounded-full bg-primary shadow-[0_0_8px_rgba(167,139,250,0.8)]" />}
      </div>
    </Link>
  )
}

const DashboardLayout = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  const handleLogout = () => {
    localStorage.removeItem('botla_token')
    localStorage.removeItem('botla_refresh_token')
    navigate('/login')
  }

  const navItems = [
    { icon: LayoutDashboard, label: 'Dashboard', to: '/' },
    { icon: Bot, label: 'Chatbots', to: '/chatbots' },
    { icon: Settings, label: 'Settings', to: '/settings' },
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
        "fixed lg:static inset-y-0 left-0 z-50 w-64 bg-card border-r border-border flex flex-col transition-transform duration-300 ease-in-out",
        isMobileMenuOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
      )}>
        {/* Logo Area */}
        <div className="h-16 flex items-center px-6 border-b border-border">
          <div className="flex items-center gap-2 font-bold text-xl tracking-tight">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-primary-foreground shadow-lg shadow-primary/20">
              B
            </div>
            <span className="text-foreground">
              Botla.co
            </span>
          </div>
          <button 
            className="ml-auto lg:hidden text-muted-foreground"
            onClick={() => setIsMobileMenuOpen(false)}
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Navigation */}
        <div className="flex-1 py-6 px-3 space-y-1">
          <div className="px-3 mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Platform
          </div>
          {navItems.map((item) => (
            <SidebarItem 
              key={item.to}
              icon={item.icon}
              label={item.label}
              to={item.to}
              active={location.pathname === item.to || (item.to !== '/' && location.pathname.startsWith(item.to))}
            />
          ))}
        </div>

        {/* User Profile / Logout */}
        <div className="p-4 border-t border-border">
          <div className="bg-muted/50 rounded-xl p-3 flex items-center gap-3 mb-3">
            <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-primary to-accent" />
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate text-foreground">User Account</div>
              <div className="text-xs text-muted-foreground truncate">Pro Plan</div>
            </div>
          </div>
          <Button 
            variant="ghost" 
            className="w-full justify-start text-muted-foreground hover:text-destructive hover:bg-destructive/10"
            onClick={handleLogout}
          >
            <LogOut className="w-4 h-4 mr-2" />
            Logout
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
                {navItems.find(i => location.pathname.startsWith(i.to) && i.to !== '/')?.label || 'Dashboard'}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {/* Add global actions here later (Search, Notifications) */}
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
