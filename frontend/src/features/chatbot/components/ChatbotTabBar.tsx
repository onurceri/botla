import { NavLink, useLocation } from 'react-router-dom'
import {
  Settings,
  Shield,
  Database,
  Zap,
  Rocket,
  BarChart3,
  Terminal,
  MoreHorizontal,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useState } from 'react'

interface TabItem {
  id: string
  label: string
  icon: React.ComponentType<{ className?: string }>
  badge?: number
}

export function ChatbotTabBar() {
  const { pathname } = useLocation()
  const [isMoreMenuOpen, setIsMoreMenuOpen] = useState(false)

  const tabs: TabItem[] = [
    { id: 'settings', label: 'Ayarlar', icon: Settings },
    { id: 'security', label: 'Güvenlik', icon: Shield },
    { id: 'sources', label: 'Kaynaklar', icon: Database },
    { id: 'actions', label: 'Aksiyonlar', icon: Zap },
    { id: 'playground', label: 'Görünüm ve Test', icon: Terminal },
    { id: 'deploy', label: 'Yayınla', icon: Rocket },
    { id: 'insights', label: 'Raporlar', icon: BarChart3 },
  ]

  const priorityTabs = tabs.slice(0, 4)
  const moreTabs = tabs.slice(4)

  const isAnyMoreTabActive = moreTabs.some((tab) => pathname.endsWith(tab.id))

  return (
    <>
      {/* Desktop Tab Bar */}
      <nav className="hidden lg:flex items-center gap-1 p-1 bg-muted/50 rounded-xl border border-border/50">
        {tabs.map((tab) => (
          <NavLink
            key={tab.id}
            to={tab.id}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200',
                isActive
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground hover:bg-background/50',
              )
            }
          >
            <tab.icon className="w-4 h-4" />
            <span>{tab.label}</span>
            {tab.badge !== undefined && tab.badge > 0 && (
              <span className="ml-1 px-1.5 py-0.5 text-xs font-semibold bg-primary text-primary-foreground rounded-full">
                {tab.badge}
              </span>
            )}
          </NavLink>
        ))}
      </nav>

      {/* Mobile Bottom Tab Bar */}
      <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-background/95 backdrop-blur-lg border-t border-border safe-area-pb">
        <div className="flex items-center justify-around h-16 px-2">
          {priorityTabs.map((tab) => (
            <NavLink
              key={tab.id}
              to={tab.id}
              className={({ isActive }) =>
                cn(
                  'flex flex-col items-center justify-center gap-1 flex-1 py-2 rounded-lg transition-all duration-200 relative',
                  isActive ? 'text-primary' : 'text-muted-foreground active:text-foreground',
                )
              }
            >
              {({ isActive }) => (
                <>
                  <div className="relative">
                    <tab.icon className={cn('w-5 h-5', isActive && 'scale-110')} />
                    {tab.badge !== undefined && tab.badge > 0 && (
                      <span className="absolute -top-1 -right-1 w-4 h-4 text-[10px] font-bold bg-destructive text-destructive-foreground rounded-full flex items-center justify-center">
                        {tab.badge > 9 ? '9+' : tab.badge}
                      </span>
                    )}
                  </div>
                  <span className={cn('text-[10px] font-medium', isActive && 'font-semibold')}>
                    {tab.label}
                  </span>
                  {isActive && (
                    <div className="absolute bottom-1 w-1 h-1 rounded-full bg-primary" />
                  )}
                </>
              )}
            </NavLink>
          ))}

          {/* More Button */}
          <button
            onClick={() => setIsMoreMenuOpen(!isMoreMenuOpen)}
            className={cn(
              'flex flex-col items-center justify-center gap-1 flex-1 py-2 rounded-lg transition-all duration-200 relative',
              isMoreMenuOpen || isAnyMoreTabActive ? 'text-primary' : 'text-muted-foreground',
            )}
          >
            <div className="relative">
              <MoreHorizontal
                className={cn('w-5 h-5', (isMoreMenuOpen || isAnyMoreTabActive) && 'scale-110')}
              />
            </div>
            <span
              className={cn(
                'text-[10px] font-medium',
                (isMoreMenuOpen || isAnyMoreTabActive) && 'font-semibold',
              )}
            >
              Daha Fazla
            </span>
            {isAnyMoreTabActive && !isMoreMenuOpen && (
              <div className="absolute bottom-1 w-1 h-1 rounded-full bg-primary" />
            )}
          </button>
        </div>

        {/* More Menu Overlay */}
        {isMoreMenuOpen && (
          <>
            <div
              className="fixed inset-0 bg-background/40 backdrop-blur-sm z-[-1]"
              onClick={() => setIsMoreMenuOpen(false)}
            />
            <div className="absolute bottom-full left-0 right-0 mb-2 px-4 pb-4 animate-in fade-in slide-in-from-bottom-4 duration-200">
              <div className="bg-background/95 backdrop-blur-lg border border-border rounded-2xl shadow-xl overflow-hidden">
                <div className="grid grid-cols-3 gap-2 p-4">
                  {moreTabs.map((tab) => (
                    <NavLink
                      key={tab.id}
                      to={tab.id}
                      onClick={() => setIsMoreMenuOpen(false)}
                      className={({ isActive }) =>
                        cn(
                          'flex flex-col items-center justify-center gap-2 p-4 rounded-xl transition-all duration-200',
                          isActive
                            ? 'bg-primary/10 text-primary'
                            : 'text-muted-foreground hover:bg-muted',
                        )
                      }
                    >
                      <div className="relative">
                        <tab.icon className="w-6 h-6" />
                        {tab.badge !== undefined && tab.badge > 0 && (
                          <span className="absolute -top-1 -right-1 w-4 h-4 text-[10px] font-bold bg-destructive text-destructive-foreground rounded-full flex items-center justify-center">
                            {tab.badge > 9 ? '9+' : tab.badge}
                          </span>
                        )}
                      </div>
                      <span className="text-xs font-medium">{tab.label}</span>
                    </NavLink>
                  ))}
                </div>
              </div>
            </div>
          </>
        )}
      </nav>
    </>
  )
}
