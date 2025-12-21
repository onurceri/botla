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
  inputBackgroundColor: string
  setInputBackgroundColor: (v: string) => void
  inputTextColor: string
  setInputTextColor: (v: string) => void
  sendButtonColor: string
  setSendButtonColor: (v: string) => void
  chatFontFamily: string
  setChatFontFamily: (v: string) => void
  themeColor: string
  setThemeColor: (v: string) => void
  bubbleRadius: string
  setBubbleRadius: (v: string) => void
}

export default function ColorsSection({ 
  isExpanded, onToggle, 
  chatBackgroundColor, setChatBackgroundColor, 
  chatHeaderColor, setChatHeaderColor, 
  chatHeaderTextColor, setChatHeaderTextColor, 
  botMessageColor, setBotMessageColor, 
  botMessageTextColor, setBotMessageTextColor, 
  userMessageColor, setUserMessageColor, 
  userMessageTextColor, setUserMessageTextColor,
  inputBackgroundColor, setInputBackgroundColor,
  inputTextColor, setInputTextColor,
  sendButtonColor, setSendButtonColor,
  chatFontFamily, setChatFontFamily,
  themeColor, setThemeColor,
  bubbleRadius, setBubbleRadius
}: Props) {
  return (
    <div className={`transition-all duration-300 border border-slate-200/60 rounded-2xl overflow-hidden ${isExpanded ? 'bg-white shadow-[0_4px_20px_rgba(0,0,0,0.03)]' : 'bg-white/40 hover:bg-white/60'}`}>
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 transition-colors"
      >
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-xl transition-colors ${isExpanded ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}>
            <Palette className="w-4 h-4" />
          </div>
          <span className={`text-[13px] font-bold tracking-tight ${isExpanded ? 'text-slate-900' : 'text-slate-600'}`}>Yazı ve Renkler</span>
        </div>
        <div className={`transition-transform duration-300 ${isExpanded ? 'rotate-180' : ''}`}>
          <ChevronDown className="w-4 h-4 text-slate-300" />
        </div>
      </button>
      {isExpanded && (
        <div className="p-4 pt-0 space-y-6 animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="h-px bg-slate-100/80 mb-6" />
          
          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">Genel</h4>
            <div className="space-y-4">
              <div className="space-y-2.5">
                <label htmlFor="font-select" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Yazı Tipi</label>
                <select 
                  id="font-select"
                  className="flex h-11 w-full rounded-xl border border-slate-200/60 bg-slate-50/50 px-3 py-1 text-sm transition-all focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/5"
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

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2.5">
                  <label htmlFor="theme-color" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Varsayılan İkon Rengi</label>
                  <div className="flex gap-2.5">
                    <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                      <Input 
                        id="theme-color" 
                        type="color" 
                        value={themeColor} 
                        onChange={(e) => setThemeColor(e.target.value)} 
                        className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" 
                      />
                    </div>
                    <Input 
                      value={themeColor} 
                      onChange={(e) => setThemeColor(e.target.value)} 
                      className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" 
                    />
                  </div>
                </div>

                <div className="space-y-2.5">
                  <label htmlFor="bubble-radius" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Kabarcık Ovalleşmesi</label>
                  <div className="flex gap-2.5 items-center">
                    <Input 
                      id="bubble-radius"
                      value={bubbleRadius} 
                      onChange={(e) => setBubbleRadius(e.target.value)} 
                      className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-sm" 
                      placeholder="22px"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">Panel & Header</h4>
            <div className="grid grid-cols-1 gap-4">
              <div className="space-y-2.5">
                <label htmlFor="chat-bg" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Chat Arka Plan</label>
                <div className="flex gap-2.5">
                  <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                    <Input id="chat-bg" type="color" value={chatBackgroundColor} onChange={(e) => setChatBackgroundColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                  </div>
                  <Input value={chatBackgroundColor} onChange={(e) => setChatBackgroundColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2.5">
                  <label htmlFor="header-color" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Header</label>
                  <div className="flex gap-2.5">
                    <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                      <Input id="header-color" type="color" value={chatHeaderColor} onChange={(e) => setChatHeaderColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                    </div>
                    <Input value={chatHeaderColor} onChange={(e) => setChatHeaderColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                  </div>
                </div>
                <div className="space-y-2.5">
                  <label htmlFor="header-text-color" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Yazı</label>
                  <div className="flex gap-2.5">
                    <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                      <Input id="header-text-color" type="color" value={chatHeaderTextColor} onChange={(e) => setChatHeaderTextColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                    </div>
                    <Input value={chatHeaderTextColor} onChange={(e) => setChatHeaderTextColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">Bot Mesajları</h4>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2.5">
                <label htmlFor="bot-msg-color" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Arka Plan</label>
                <div className="flex gap-2.5">
                  <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                    <Input id="bot-msg-color" type="color" value={botMessageColor} onChange={(e) => setBotMessageColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                  </div>
                  <Input value={botMessageColor} onChange={(e) => setBotMessageColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                </div>
              </div>
              <div className="space-y-2.5">
                <label htmlFor="bot-text-color" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Yazı</label>
                <div className="flex gap-2.5">
                  <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                    <Input id="bot-text-color" type="color" value={botMessageTextColor} onChange={(e) => setBotMessageTextColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                  </div>
                  <Input value={botMessageTextColor} onChange={(e) => setBotMessageTextColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">Kullanıcı Mesajları</h4>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2.5">
                <label htmlFor="user-msg-color" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Arka Plan</label>
                <div className="flex gap-2.5">
                  <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                    <Input id="user-msg-color" type="color" value={userMessageColor} onChange={(e) => setUserMessageColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                  </div>
                  <Input value={userMessageColor} onChange={(e) => setUserMessageColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                </div>
              </div>
              <div className="space-y-2.5">
                <label htmlFor="user-text-color" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Yazı</label>
                <div className="flex gap-2.5">
                  <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                    <Input id="user-text-color" type="color" value={userMessageTextColor} onChange={(e) => setUserMessageTextColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                  </div>
                  <Input value={userMessageTextColor} onChange={(e) => setUserMessageTextColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">Giriş Alanı</h4>
            <div className="space-y-4">
              <div className="space-y-2.5">
                <label htmlFor="input-bg" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Arka Plan</label>
                <div className="flex gap-2.5">
                  <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                    <Input id="input-bg" type="color" value={inputBackgroundColor} onChange={(e) => setInputBackgroundColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                  </div>
                  <Input value={inputBackgroundColor} onChange={(e) => setInputBackgroundColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2.5">
                  <label htmlFor="input-text" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Yazı Rengi</label>
                  <div className="flex gap-2.5">
                    <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                      <Input id="input-text" type="color" value={inputTextColor} onChange={(e) => setInputTextColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                    </div>
                    <Input value={inputTextColor} onChange={(e) => setInputTextColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                  </div>
                </div>
                <div className="space-y-2.5">
                  <label htmlFor="send-btn" className="text-[11px] font-bold text-slate-500 tracking-tight ml-1">Gönder Butonu</label>
                  <div className="flex gap-2.5">
                    <div className="relative w-11 h-11 rounded-xl overflow-hidden border border-slate-200/60 bg-slate-50/50 shrink-0">
                      <Input id="send-btn" type="color" value={sendButtonColor} onChange={(e) => setSendButtonColor(e.target.value)} className="absolute inset-0 w-[150%] h-[150%] -top-[25%] -left-[25%] p-0 border-0 cursor-pointer" />
                    </div>
                    <Input value={sendButtonColor} onChange={(e) => setSendButtonColor(e.target.value)} className="h-11 rounded-xl bg-slate-50/50 border-slate-200/60 focus:bg-white transition-all font-mono text-xs flex-1" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
