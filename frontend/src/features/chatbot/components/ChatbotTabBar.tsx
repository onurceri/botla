import { NavLink } from 'react-router-dom'
import { Settings, Shield, Database, Zap, Palette, Rocket, BarChart3, Terminal } from 'lucide-react'
import { cn } from '@/lib/utils'

interface TabItem {
  id: string
  label: string
  icon: React.ComponentType<{ className?: string }>
  badge?: number
}

export function ChatbotTabBar() {
  const tabs: TabItem[] = [
    { id: 'settings', label: 'Ayarlar', icon: Settings },
    { id: 'security', label: 'Güvenlik', icon: Shield },
    { id: 'sources', label: 'Kaynaklar', icon: Database },
    { id: 'actions', label: 'Aksiyonlar', icon: Zap },
    { id: 'playground', label: 'Playground', icon: Terminal },
    { id: 'deploy', label: 'Yayınla', icon: Rocket },
    { id: 'insights', label: 'Raporlar', icon: BarChart3 },
  ]

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
                "flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
                isActive
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground hover:bg-background/50"
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
          {tabs.map((tab) => (
            <NavLink
              key={tab.id}
              to={tab.id}
              className={({ isActive }) =>
                cn(
                  "flex flex-col items-center justify-center gap-1 flex-1 py-2 rounded-lg transition-all duration-200",
                  isActive
                    ? "text-primary"
                    : "text-muted-foreground active:text-foreground"
                )
              }
            >
              {({ isActive }) => (
                <>
                  <div className="relative">
                    <tab.icon className={cn("w-5 h-5", isActive && "scale-110")} />
                    {tab.badge !== undefined && tab.badge > 0 && (
                      <span className="absolute -top-1 -right-1 w-4 h-4 text-[10px] font-bold bg-destructive text-destructive-foreground rounded-full flex items-center justify-center">
                        {tab.badge > 9 ? '9+' : tab.badge}
                      </span>
                    )}
                  </div>
                  <span className={cn("text-[10px] font-medium", isActive && "font-semibold")}>
                    {tab.label}
                  </span>
                  {isActive && (
                    <div className="absolute bottom-1 w-1 h-1 rounded-full bg-primary" />
                  )}
                </>
              )}
            </NavLink>
          ))}
        </div>
      </nav>
    </>
  )
}
