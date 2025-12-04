import { Input } from '@/components/ui/input'
import { ChevronDown, ChevronRight, Layout } from 'lucide-react'

type Props = {
  isExpanded: boolean
  onToggle: () => void
  position: string
  setPosition: (v: string) => void
  chatFontFamily: string
  setChatFontFamily: (v: string) => void
  themeColor: string
  setThemeColor: (v: string) => void
}

export default function AppearanceSection({ isExpanded, onToggle, position, setPosition, chatFontFamily, setChatFontFamily, themeColor, setThemeColor }: Props) {
  return (
    <div className="border border-border rounded-xl bg-card overflow-hidden">
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 bg-white/50 backdrop-blur hover:bg-white/70 transition-colors"
      >
        <div className="flex items-center gap-2 font-medium">
          <Layout className="w-4 h-4 text-primary" />
          Görünüm
        </div>
        {isExpanded ? <ChevronDown className="w-4 h-4 text-muted-foreground" /> : <ChevronRight className="w-4 h-4 text-muted-foreground" />}
      </button>
      {isExpanded && (
        <div className="p-4 space-y-4 border-t border-border animate-in slide-in-from-top-2 duration-200">
          <div className="space-y-2">
            <label htmlFor="position-select" className="text-xs font-medium text-muted-foreground uppercase">Konum</label>
            <select 
              id="position-select"
              className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              value={position}
              onChange={(e) => setPosition(e.target.value)}
            >
              <option value="bottom-right">Sağ Alt</option>
              <option value="bottom-left">Sol Alt</option>
            </select>
          </div>
          <div className="space-y-2">
            <label htmlFor="font-select" className="text-xs font-medium text-muted-foreground uppercase">Yazı Tipi</label>
            <select 
              id="font-select"
              className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              value={chatFontFamily}
              onChange={(e) => setChatFontFamily(e.target.value)}
            >
              <option value="Inter, sans-serif">Inter (Modern)</option>
              <option value="Roboto, sans-serif">Roboto</option>
              <option value="Open Sans, sans-serif">Open Sans</option>
              <option value="Lato, sans-serif">Lato</option>
              <option value="Montserrat, sans-serif">Montserrat</option>
            </select>
          </div>
          <div className="space-y-2">
            <label htmlFor="theme-color" className="text-xs font-medium text-muted-foreground uppercase">Ana Renk (Theme)</label>
            <div className="flex gap-2">
              <Input id="theme-color" type="color" value={themeColor} onChange={(e) => setThemeColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
              <Input value={themeColor} onChange={(e) => setThemeColor(e.target.value)} className="flex-1 bg-background font-mono" />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

