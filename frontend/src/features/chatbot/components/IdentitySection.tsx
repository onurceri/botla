import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ChevronDown, ChevronRight, User } from 'lucide-react'

const WELCOME_MESSAGE_MAX_LENGTH = 200

type Props = {
  isExpanded: boolean
  onToggle: () => void
  botDisplayName: string
  setBotDisplayName: (v: string) => void
  botIcon: string
  setBotIcon: (v: string) => void
  welcomeMessage: string
  setWelcomeMessage: (v: string) => void
}

export default function IdentitySection({ isExpanded, onToggle, botDisplayName, setBotDisplayName, botIcon, setBotIcon, welcomeMessage, setWelcomeMessage }: Props) {
  return (
    <div className={`transition-all duration-300 border border-slate-200/60 rounded-2xl overflow-hidden ${isExpanded ? 'bg-white shadow-[0_4px_20px_rgba(0,0,0,0.03)]' : 'bg-white/40 hover:bg-white/60'}`}>
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 transition-colors"
      >
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-xl transition-colors ${isExpanded ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}>
            <User className="w-4 h-4" />
          </div>
          <span className={`text-[13px] font-bold tracking-tight ${isExpanded ? 'text-slate-900' : 'text-slate-600'}`}>Kimlik</span>
        </div>
        <div className={`transition-transform duration-300 ${isExpanded ? 'rotate-180' : ''}`}>
          <ChevronDown className="w-4 h-4 text-slate-300" />
        </div>
      </button>
      {isExpanded && (
        <div className="p-4 pt-0 space-y-5 animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="h-px bg-slate-100/80 mb-5" />
          <div className="space-y-2.5">
            <label htmlFor="bot-display-name" className="text-[11px] font-bold text-slate-400 uppercase tracking-widest ml-1">Bot Görünen Adı</label>
            <Input id="bot-display-name" value={botDisplayName} onChange={(e) => setBotDisplayName(e.target.value)} placeholder="Örn: Asistan" className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all" />
          </div>
          <div className="space-y-2.5">
            <label htmlFor="bot-icon-url" className="text-[11px] font-bold text-slate-400 uppercase tracking-widest ml-1">Bot İkon URL</label>
            <Input id="bot-icon-url" value={botIcon} onChange={(e) => setBotIcon(e.target.value)} placeholder="https://..." className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all" />
          </div>
          <div className="space-y-2.5">
            <div className="flex items-center justify-between ml-1">
              <label htmlFor="welcome-message" className="text-[11px] font-bold text-slate-400 uppercase tracking-widest">Karşılama Mesajı</label>
              <span className="text-[10px] font-bold text-slate-300 tabular-nums">
                {welcomeMessage.length}/{WELCOME_MESSAGE_MAX_LENGTH}
              </span>
            </div>
            <Textarea
              id="welcome-message"
              value={welcomeMessage}
              onChange={(e) => setWelcomeMessage(e.target.value)}
              className="bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all rounded-xl resize-none p-3 text-sm leading-relaxed"
              maxLength={WELCOME_MESSAGE_MAX_LENGTH}
              rows={4}
            />
          </div>
        </div>
      )}
    </div>
  )
}
