import { Input } from '@/components/ui/input'
import { ChevronDown, ChevronRight, Palette } from 'lucide-react'

type Props = {
  isExpanded: boolean
  onToggle: () => void
  chatBackgroundColor: string
  setChatBackgroundColor: (v: string) => void
  chatHeaderColor: string
  setChatHeaderColor: (v: string) => void
  chatHeaderTextColor: string
  setChatHeaderTextColor: (v: string) => void
  botMessageColor: string
  setBotMessageColor: (v: string) => void
  botMessageTextColor: string
  setBotMessageTextColor: (v: string) => void
  userMessageColor: string
  setUserMessageColor: (v: string) => void
  userMessageTextColor: string
  setUserMessageTextColor: (v: string) => void
}

export default function ColorsSection({ isExpanded, onToggle, chatBackgroundColor, setChatBackgroundColor, chatHeaderColor, setChatHeaderColor, chatHeaderTextColor, setChatHeaderTextColor, botMessageColor, setBotMessageColor, botMessageTextColor, setBotMessageTextColor, userMessageColor, setUserMessageColor, userMessageTextColor, setUserMessageTextColor }: Props) {
  return (
    <div className="border border-border rounded-xl bg-card overflow-hidden">
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 bg-white/50 backdrop-blur hover:bg-white/70 transition-colors"
      >
        <div className="flex items-center gap-2 font-medium">
          <Palette className="w-4 h-4 text-primary" />
          Renkler
        </div>
        {isExpanded ? <ChevronDown className="w-4 h-4 text-muted-foreground" /> : <ChevronRight className="w-4 h-4 text-muted-foreground" />}
      </button>
      {isExpanded && (
        <div className="p-4 space-y-4 border-t border-border animate-in slide-in-from-top-2 duration-200">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label htmlFor="chat-bg" className="text-xs font-medium text-muted-foreground uppercase">Chat Arka Plan</label>
              <div className="flex gap-2 items-center">
                <Input id="chat-bg" type="color" value={chatBackgroundColor} onChange={(e) => setChatBackgroundColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                <Input value={chatBackgroundColor} onChange={(e) => setChatBackgroundColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
              </div>
            </div>
            <div className="space-y-2">
              <label htmlFor="header-color" className="text-xs font-medium text-muted-foreground uppercase">Header</label>
              <div className="flex gap-2 items-center">
                <Input id="header-color" type="color" value={chatHeaderColor} onChange={(e) => setChatHeaderColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                <Input value={chatHeaderColor} onChange={(e) => setChatHeaderColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
              </div>
            </div>
            <div className="space-y-2">
              <label htmlFor="header-text-color" className="text-xs font-medium text-muted-foreground uppercase">Header Yazı</label>
              <div className="flex gap-2 items-center">
                <Input id="header-text-color" type="color" value={chatHeaderTextColor} onChange={(e) => setChatHeaderTextColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                <Input value={chatHeaderTextColor} onChange={(e) => setChatHeaderTextColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label htmlFor="bot-msg-color" className="text-xs font-medium text-muted-foreground uppercase">Bot Mesaj</label>
              <div className="flex gap-2 items-center">
                <Input id="bot-msg-color" type="color" value={botMessageColor} onChange={(e) => setBotMessageColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                <Input value={botMessageColor} onChange={(e) => setBotMessageColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
              </div>
            </div>
            <div className="space-y-2">
              <label htmlFor="bot-text-color" className="text-xs font-medium text-muted-foreground uppercase">Bot Yazı</label>
              <div className="flex gap-2 items-center">
                <Input id="bot-text-color" type="color" value={botMessageTextColor} onChange={(e) => setBotMessageTextColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                <Input value={botMessageTextColor} onChange={(e) => setBotMessageTextColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label htmlFor="user-msg-color" className="text-xs font-medium text-muted-foreground uppercase">Kullanıcı Mesaj</label>
              <div className="flex gap-2 items-center">
                <Input id="user-msg-color" type="color" value={userMessageColor} onChange={(e) => setUserMessageColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                <Input value={userMessageColor} onChange={(e) => setUserMessageColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
              </div>
            </div>
            <div className="space-y-2">
              <label htmlFor="user-text-color" className="text-xs font-medium text-muted-foreground uppercase">Kullanıcı Yazı</label>
              <div className="flex gap-2 items-center">
                <Input id="user-text-color" type="color" value={userMessageTextColor} onChange={(e) => setUserMessageTextColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                <Input value={userMessageTextColor} onChange={(e) => setUserMessageTextColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

