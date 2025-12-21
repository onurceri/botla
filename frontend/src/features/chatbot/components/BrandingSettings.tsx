import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Tag, ChevronDown, ChevronRight, Lock, Crown, AlertCircle } from 'lucide-react'

type CustomBranding = {
  logo_url?: string
  text?: string
  link?: string
}

type Props = {
  isExpanded: boolean
  onToggle: () => void
  hideBranding: boolean
  setHideBranding: (v: boolean) => void
  customBranding: CustomBranding | null
  setCustomBranding: (v: CustomBranding | null) => void
  canHideBranding: boolean
  canCustomBranding: boolean
}

export default function BrandingSettings({
  isExpanded, 
  onToggle, 
  hideBranding, 
  setHideBranding, 
  customBranding, 
  setCustomBranding,
  canHideBranding,
  canCustomBranding
}: Props) {
  const [logoError, setLogoError] = useState(false)
  
  return (
    <div className={`transition-all duration-300 border border-slate-200/60 rounded-2xl overflow-hidden ${isExpanded ? 'bg-white shadow-[0_4px_20px_rgba(0,0,0,0.03)]' : 'bg-white/40 hover:bg-white/60'}`}>
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 transition-colors"
      >
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-xl transition-colors ${isExpanded ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}>
            <Tag className="w-4 h-4" />
          </div>
          <span className={`text-[13px] font-bold tracking-tight ${isExpanded ? 'text-slate-900' : 'text-slate-600'}`}>Branding Ayarları</span>
        </div>
        <div className={`transition-transform duration-300 ${isExpanded ? 'rotate-180' : ''}`}>
          <ChevronDown className="w-4 h-4 text-slate-300" />
        </div>
      </button>

      {isExpanded && (
        <div className="px-4 pb-5 space-y-5 animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="h-px bg-slate-100 -mx-4 mb-4" />
          
          {/* Hide Branding Toggle */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400">
                  Botla Logosunu Gizle
                </label>
                {!canHideBranding && (
                  <span className="flex items-center gap-1 text-[9px] font-bold text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full border border-amber-100">
                    <Crown className="w-2.5 h-2.5" />
                    PRO+
                  </span>
                )}
              </div>
              <div className="flex items-center gap-3">
                {!canHideBranding && (
                  <Lock className="w-3.5 h-3.5 text-slate-300" />
                )}
                <button
                  type="button"
                  onClick={() => {
                    if (!canHideBranding) return
                    const next = !hideBranding
                    setHideBranding(next)
                    if (!next) setCustomBranding(null)
                  }}
                  disabled={!canHideBranding}
                  className={`relative inline-flex h-5 w-10 items-center rounded-full transition-all duration-300 ${
                    hideBranding ? 'bg-primary shadow-[0_2px_8px_rgba(var(--primary-rgb),0.3)]' : 'bg-slate-200'
                  } ${!canHideBranding ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer hover:scale-105 active:scale-95'}`}
                >
                  <span
                    className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow-sm transition-transform duration-300 ${
                      hideBranding ? 'translate-x-[22px]' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>
            </div>
            <p className="text-[11px] text-slate-500 leading-relaxed">
              Sohbet penceresinin altındaki "Powered by Botla" yazısını kaldırın.
            </p>
          </div>

          {/* Custom Branding Section */}
          <div className={`space-y-4 pt-4 border-t border-slate-100 transition-opacity duration-300 ${!hideBranding ? 'opacity-40 grayscale pointer-events-none' : ''}`}>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400">
                  Özel Branding
                </label>
                {!canCustomBranding && (
                  <span className="flex items-center gap-1 text-[9px] font-bold text-purple-600 bg-purple-50 px-2 py-0.5 rounded-full border border-purple-100">
                    <Crown className="w-2.5 h-2.5" />
                    ENTERPRISE
                  </span>
                )}
              </div>
            </div>
            
            {!hideBranding && (
              <div className="flex items-center gap-2 p-2.5 bg-slate-50 rounded-xl border border-slate-100">
                <AlertCircle className="w-3.5 h-3.5 text-slate-400" />
                <p className="text-[10px] text-slate-500 font-medium">
                  Önce yukarıdaki logoyu gizleme seçeneğini aktifleştirin.
                </p>
              </div>
            )}
            
            {hideBranding && canCustomBranding ? (
              <div className="space-y-4">
                <div className="space-y-1.5">
                  <label htmlFor="brand-logo" className="text-[11px] font-semibold text-slate-700 ml-1">Logo URL</label>
                  <Input 
                    id="brand-logo"
                    placeholder="https://site.com/logo.png"
                    value={customBranding?.logo_url || ''}
                    onChange={(e) => {
                      setLogoError(false)
                      setCustomBranding({ ...(customBranding || {}), logo_url: e.target.value })
                    }}
                    className="bg-slate-50/50 border-slate-200/60 rounded-xl h-9 text-xs focus:bg-white transition-all"
                  />
                  {logoError && (
                    <p className="text-[10px] text-rose-500 flex items-center gap-1 ml-1 mt-1">
                      <AlertCircle className="w-3 h-3 flex-shrink-0" />
                      Logo yüklenemedi, lütfen URL'yi kontrol edin.
                    </p>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <label htmlFor="brand-text" className="text-[11px] font-semibold text-slate-700 ml-1">Metin</label>
                    <Input 
                      id="brand-text"
                      placeholder="Powered by Şirket"
                      value={customBranding?.text || ''}
                      onChange={(e) => setCustomBranding({ ...(customBranding || {}), text: e.target.value })}
                      className="bg-slate-50/50 border-slate-200/60 rounded-xl h-9 text-xs focus:bg-white transition-all"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label htmlFor="brand-link" className="text-[11px] font-semibold text-slate-700 ml-1">Link</label>
                    <Input 
                      id="brand-link"
                      placeholder="https://sirket.com"
                      value={customBranding?.link || ''}
                      onChange={(e) => setCustomBranding({ ...(customBranding || {}), link: e.target.value })}
                      className="bg-slate-50/50 border-slate-200/60 rounded-xl h-9 text-xs focus:bg-white transition-all"
                    />
                  </div>
                </div>
              </div>
            ) : hideBranding && !canCustomBranding ? (
              <div className="p-5 bg-slate-50/50 rounded-2xl border border-slate-100 flex flex-col items-center text-center gap-2">
                <div className="p-2.5 rounded-full bg-white shadow-sm border border-slate-100">
                  <Lock className="w-5 h-5 text-slate-300" />
                </div>
                <div>
                  <p className="text-[12px] font-bold text-slate-900">Enterprise Özelliği</p>
                  <p className="text-[11px] text-slate-500 mt-0.5">Kendi logonuzu ve linkinizi eklemek için Enterprise plana yükseltin.</p>
                </div>
              </div>
            ) : null}
          </div>
        </div>
      )}
    </div>
  )
}
