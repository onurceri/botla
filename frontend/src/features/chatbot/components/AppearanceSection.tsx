import { Input } from '@/components/ui/input'
import { ChevronDown, ChevronRight, Layout } from 'lucide-react'

type Props = {
  isExpanded: boolean
  onToggle: () => void
  position: string
  setPosition: (v: string) => void
}

export default function AppearanceSection({ isExpanded, onToggle, position, setPosition }: Props) {
  return (
    <div className={`transition-all duration-300 border border-slate-200/60 rounded-2xl overflow-hidden ${isExpanded ? 'bg-white shadow-[0_4px_20px_rgba(0,0,0,0.03)]' : 'bg-white/40 hover:bg-white/60'}`}>
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 transition-colors"
      >
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-xl transition-colors ${isExpanded ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}>
            <Layout className="w-4 h-4" />
          </div>
          <span className={`text-[13px] font-bold tracking-tight ${isExpanded ? 'text-slate-900' : 'text-slate-600'}`}>Konum</span>
        </div>
        <div className={`transition-transform duration-300 ${isExpanded ? 'rotate-180' : ''}`}>
          <ChevronDown className="w-4 h-4 text-slate-300" />
        </div>
      </button>
      {isExpanded && (
        <div className="p-4 pt-0 space-y-5 animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="h-px bg-slate-100/80 mb-5" />
          <div className="space-y-2.5">
            <label htmlFor="position-select" className="text-[11px] font-bold text-slate-400 uppercase tracking-widest ml-1">Konum</label>
            <select 
              id="position-select"
              className="flex h-11 w-full rounded-xl border border-slate-200/60 bg-slate-50/50 px-3 py-1 text-sm transition-all focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/5"
              value={position}
              onChange={(e) => setPosition(e.target.value)}
            >
              <option value="bottom-right">Sağ Alt</option>
              <option value="bottom-left">Sol Alt</option>
            </select>
          </div>
        </div>
      )}
    </div>
  )
}

