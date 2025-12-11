import { Settings, Database, Play, Code, MessageSquare, Zap, Shield, Headphones, BarChart3 } from 'lucide-react'
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn } from '@/lib/utils'

interface ChatbotSidebarProps {
  activeTab: string
  onTabChange: (tab: string) => void
}

export function ChatbotSidebar({ activeTab, onTabChange }: ChatbotSidebarProps) {
  const groups = [
    {
      label: 'Genel',
      items: [
        { id: 'overview', label: 'Genel Ayarlar', icon: Settings },
        { id: 'guardrails', label: 'Güvenlik & Sınırlar', icon: Shield },
        { id: 'handoff', label: 'İnsan Desteği', icon: Headphones },
      ]
    },
    {
      label: 'Yetenekler',
      items: [
        { id: 'sources', label: 'Veri Kaynakları', icon: Database },
        { id: 'actions', label: 'Aksiyonlar', icon: Zap },
        { id: 'suggestions', label: 'Örnek Sorular', icon: MessageSquare },
      ]
    },
    {
      label: 'Test & Yayın',
      items: [
        { id: 'playground', label: 'Playground', icon: Play },
        { id: 'connect', label: 'Entegrasyon', icon: Code },
      ]
    },
    {
      label: 'Raporlar',
      items: [
        { id: 'analytics', label: 'Analizler', icon: BarChart3 },
      ]
    }
  ]

  return (
    <>
      {/* Mobile View - Horizontal Scroll */}
      {/* Mobile View - Dropdown Menu */}
      <div className="lg:hidden w-full pb-6">
        <Select value={activeTab} onValueChange={onTabChange}>
          <SelectTrigger className="w-full bg-background h-12">
             <SelectValue placeholder="Menü Seçin" />
          </SelectTrigger>
          <SelectContent>
            {groups.map((group) => (
                <SelectGroup key={group.label}>
                    <SelectLabel className="pl-4 text-xs font-semibold text-muted-foreground uppercase tracking-wider opacity-70 mt-2">
                        {group.label}
                    </SelectLabel>
                    {group.items.map(item => (
                        <SelectItem key={item.id} value={item.id} className="pl-4 py-3">
                            <div className="flex items-center gap-2">
                                <item.icon className="w-4 h-4" />
                                {item.label}
                            </div>
                        </SelectItem>
                    ))}
                </SelectGroup>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Desktop View - Vertical Sidebar */}
      <nav className="hidden lg:flex flex-col gap-8 w-64 flex-shrink-0 sticky top-0">
        {groups.map((group) => (
          <div key={group.label} className="space-y-2">
            <h3 className="px-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {group.label}
            </h3>
            <div className="space-y-1">
              {group.items.map((item) => (
                <button
                  key={item.id}
                  role="tab"
                  aria-selected={activeTab === item.id}
                  onClick={() => onTabChange(item.id)}
                  className={cn(
                    "w-full flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-md transition-colors text-left",
                    activeTab === item.id
                      ? "bg-primary/10 text-primary"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  )}
                >
                  <item.icon className="w-4 h-4" />
                  {item.label}
                </button>
              ))}
            </div>
          </div>
        ))}
      </nav>
    </>
  )
}
