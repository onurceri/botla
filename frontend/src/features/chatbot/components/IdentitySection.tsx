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
    <div className="border border-border rounded-xl bg-card overflow-hidden">
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 bg-white/50 backdrop-blur hover:bg-white/70 transition-colors"
      >
        <div className="flex items-center gap-2 font-medium">
          <User className="w-4 h-4 text-primary" />
          Kimlik
        </div>
        {isExpanded ? <ChevronDown className="w-4 h-4 text-muted-foreground" /> : <ChevronRight className="w-4 h-4 text-muted-foreground" />}
      </button>
      {isExpanded && (
        <div className="p-4 space-y-4 border-t border-border animate-in slide-in-from-top-2 duration-200">
          <div className="space-y-2">
            <label htmlFor="bot-display-name" className="text-xs font-medium text-muted-foreground uppercase">Bot Görünen Adı</label>
            <Input id="bot-display-name" value={botDisplayName} onChange={(e) => setBotDisplayName(e.target.value)} placeholder="Örn: Asistan" className="bg-background" />
          </div>
          <div className="space-y-2">
            <label htmlFor="bot-icon-url" className="text-xs font-medium text-muted-foreground uppercase">Bot İkon URL</label>
            <Input id="bot-icon-url" value={botIcon} onChange={(e) => setBotIcon(e.target.value)} placeholder="https://..." className="bg-background" />
          </div>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label htmlFor="welcome-message" className="text-xs font-medium text-muted-foreground uppercase">Karşılama Mesajı (İlk Bot Mesajı)</label>
              <span className="text-[11px] text-muted-foreground">
                {welcomeMessage.length}/{WELCOME_MESSAGE_MAX_LENGTH}
              </span>
            </div>
            <Textarea
              id="welcome-message"
              value={welcomeMessage}
              onChange={(e) => setWelcomeMessage(e.target.value)}
              className="bg-background resize-none"
              maxLength={WELCOME_MESSAGE_MAX_LENGTH}
              rows={3}
            />
            <div className="flex justify-end">
              <span className="text-[11px] text-muted-foreground">
                {Math.max(0, WELCOME_MESSAGE_MAX_LENGTH - welcomeMessage.length).toLocaleString('tr-TR')} karakter kaldı
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
