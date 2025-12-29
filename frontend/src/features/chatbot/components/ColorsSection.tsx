import { Input } from '@/components/ui/input'
import { ColorPicker } from '@/components/ui/color-picker'
import { ChevronDown, Palette } from 'lucide-react'

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
  isExpanded,
  onToggle,
  chatBackgroundColor,
  setChatBackgroundColor,
  chatHeaderColor,
  setChatHeaderColor,
  chatHeaderTextColor,
  setChatHeaderTextColor,
  botMessageColor,
  setBotMessageColor,
  botMessageTextColor,
  setBotMessageTextColor,
  userMessageColor,
  setUserMessageColor,
  userMessageTextColor,
  setUserMessageTextColor,
  inputBackgroundColor,
  setInputBackgroundColor,
  inputTextColor,
  setInputTextColor,
  sendButtonColor,
  setSendButtonColor,
  chatFontFamily,
  setChatFontFamily,
  themeColor,
  setThemeColor,
  bubbleRadius,
  setBubbleRadius,
}: Props) {
  return (
    <div
      className={`transition-all duration-300 border border-slate-200/60 rounded-2xl overflow-hidden ${isExpanded ? 'bg-white shadow-[0_4px_20px_rgba(0,0,0,0.03)]' : 'bg-white/40 hover:bg-white/60'}`}
    >
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 transition-colors"
      >
        <div className="flex items-center gap-3">
          <div
            className={`p-2 rounded-xl transition-colors ${isExpanded ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}
          >
            <Palette className="w-4 h-4" />
          </div>
          <span
            className={`text-[13px] font-bold tracking-tight ${isExpanded ? 'text-slate-900' : 'text-slate-600'}`}
          >
            Yazı ve Renkler
          </span>
        </div>
        <div className={`transition-transform duration-300 ${isExpanded ? 'rotate-180' : ''}`}>
          <ChevronDown className="w-4 h-4 text-slate-300" />
        </div>
      </button>
      {isExpanded && (
        <div className="p-4 pt-0 space-y-6 animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="h-px bg-slate-100/80 mb-6" />

          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">
              Genel
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-5">
              <div className="space-y-2.5">
                <label
                  htmlFor="font-select"
                  className="text-[11px] font-bold text-slate-500 tracking-tight ml-1"
                >
                  Yazı Tipi
                </label>
                <select
                  id="font-select"
                  className="flex h-11 w-full rounded-2xl border-2 border-slate-100 bg-slate-50/30 px-4 py-1 text-[13px] font-medium transition-all hover:bg-slate-50/50 focus:bg-white focus:outline-none focus:ring-4 focus:ring-primary/5"
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

              <div className="space-y-2.5">
                <label
                  htmlFor="bubble-radius"
                  className="text-[11px] font-bold text-slate-500 tracking-tight ml-1"
                >
                  Kabarcık Ovalleşmesi
                </label>
                <Input
                  id="bubble-radius"
                  value={bubbleRadius}
                  onChange={(e) => setBubbleRadius(e.target.value)}
                  className="h-11 rounded-2xl border-2 border-slate-100 bg-slate-50/30 px-4 focus:bg-white transition-all font-medium text-[13px]"
                  placeholder="22px"
                />
              </div>

              <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                <label
                  htmlFor="theme-color"
                  className="text-[12px] font-bold text-slate-600 tracking-tight"
                >
                  Varsayılan İkon Rengi
                </label>
                <ColorPicker
                  id="theme-color"
                  value={themeColor}
                  onChange={setThemeColor}
                  label="Varsayılan İkon Rengi"
                />
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">
              Panel & Header
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                <div className="space-y-0.5">
                  <span className="text-[12px] font-bold text-slate-600 tracking-tight">Arka Plan</span>
                  <p className="text-[10px] text-slate-400 font-medium">Chat panel ana rengi</p>
                </div>
                <ColorPicker
                  id="chat-bg"
                  value={chatBackgroundColor}
                  onChange={setChatBackgroundColor}
                  label="Chat Arka Plan"
                />
              </div>
              <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                <div className="space-y-0.5">
                  <span className="text-[12px] font-bold text-slate-600 tracking-tight">Üst Bilgi</span>
                  <p className="text-[10px] text-slate-400 font-medium">Header arka planı</p>
                </div>
                <ColorPicker
                  id="header-color"
                  value={chatHeaderColor}
                  onChange={setChatHeaderColor}
                  label="Header"
                />
              </div>
              <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                <div className="space-y-0.5">
                  <span className="text-[12px] font-bold text-slate-600 tracking-tight">Başlık Yazısı</span>
                  <p className="text-[10px] text-slate-400 font-medium">Header yazı rengi</p>
                </div>
                <ColorPicker
                  id="header-text-color"
                  value={chatHeaderTextColor}
                  onChange={setChatHeaderTextColor}
                  label="Header Yazı"
                />
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">
                Bot Mesajları
              </h4>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                  <span className="text-[12px] font-bold text-slate-600 tracking-tight">Arka Plan</span>
                  <ColorPicker
                    id="bot-msg-color"
                    value={botMessageColor}
                    onChange={setBotMessageColor}
                    label="Bot Mesaj Arka Plan"
                  />
                </div>
                <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                  <span className="text-[12px] font-bold text-slate-600 tracking-tight">Yazı Rengi</span>
                  <ColorPicker
                    id="bot-text-color"
                    value={botMessageTextColor}
                    onChange={setBotMessageTextColor}
                    label="Bot Mesaj Yazı"
                  />
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">
                Kullanıcı Mesajları
              </h4>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                  <span className="text-[12px] font-bold text-slate-600 tracking-tight">Arka Plan</span>
                  <ColorPicker
                    id="user-msg-color"
                    value={userMessageColor}
                    onChange={setUserMessageColor}
                    label="Kullanıcı Mesaj Arka Plan"
                  />
                </div>
                <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                  <span className="text-[12px] font-bold text-slate-600 tracking-tight">Yazı Rengi</span>
                  <ColorPicker
                    id="user-text-color"
                    value={userMessageTextColor}
                    onChange={setUserMessageTextColor}
                    label="Kullanıcı Mesaj Yazı"
                  />
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-[0.15em] ml-1">
              Giriş Alanı
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                <span className="text-[12px] font-bold text-slate-600 tracking-tight">Arka Plan</span>
                <ColorPicker
                  id="input-bg"
                  value={inputBackgroundColor}
                  onChange={setInputBackgroundColor}
                  label="Giriş Alanı Arka Plan"
                />
              </div>
              <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                <span className="text-[12px] font-bold text-slate-600 tracking-tight">Yazı Rengi</span>
                <ColorPicker
                  id="input-text"
                  value={inputTextColor}
                  onChange={setInputTextColor}
                  label="Giriş Alanı Yazı"
                />
              </div>
              <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50/30 border-2 border-slate-100">
                <span className="text-[12px] font-bold text-slate-600 tracking-tight">Gönder Butonu</span>
                <ColorPicker
                  id="send-btn"
                  value={sendButtonColor}
                  onChange={setSendButtonColor}
                  label="Gönder Butonu"
                />
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
