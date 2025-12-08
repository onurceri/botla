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
    <div className="border border-border rounded-xl bg-card overflow-hidden">
      <button 
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 bg-white/50 backdrop-blur hover:bg-white/70 transition-colors"
      >
        <div className="flex items-center gap-2 font-medium">
          <Tag className="w-4 h-4 text-primary" />
          Branding Ayarları
        </div>
        {isExpanded ? <ChevronDown className="w-4 h-4 text-muted-foreground" /> : <ChevronRight className="w-4 h-4 text-muted-foreground" />}
      </button>
      {isExpanded && (
        <div className="p-4 space-y-4 border-t border-border animate-in slide-in-from-top-2 duration-200">
          {/* Hide Branding Toggle */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-xs font-medium text-muted-foreground uppercase">
                Botla Logosunu Gizle
              </label>
              {!canHideBranding && (
                <span className="flex items-center gap-1 text-[10px] text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded-full">
                  <Crown className="w-2.5 h-2.5" />
                  Pro+
                </span>
              )}
            </div>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => {
                  if (!canHideBranding) return
                  const next = !hideBranding
                  setHideBranding(next)
                  if (!next) setCustomBranding(null)
                }}
                disabled={!canHideBranding}
                className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
                  hideBranding ? 'bg-primary' : 'bg-gray-200'
                } ${!canHideBranding ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
              >
                <span
                  className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${
                    hideBranding ? 'translate-x-[18px]' : 'translate-x-1'
                  }`}
                />
              </button>
              <span className="text-xs text-muted-foreground">
                {hideBranding ? 'Gizli' : 'Görünür'}
              </span>
              {!canHideBranding && (
                <Lock className="w-3 h-3 text-muted-foreground" />
              )}
            </div>
          </div>

          {/* Custom Branding Section - only enabled when hideBranding is ON */}
          <div className={`space-y-2 pt-3 border-t border-border ${!hideBranding ? 'opacity-50' : ''}`}>
            <div className="flex items-center justify-between">
              <label className="text-xs font-medium text-muted-foreground uppercase">
                Özel Branding
              </label>
              {!canCustomBranding && (
                <span className="flex items-center gap-1 text-[10px] text-purple-600 bg-purple-50 px-1.5 py-0.5 rounded-full">
                  <Crown className="w-2.5 h-2.5" />
                  Enterprise
                </span>
              )}
            </div>
            
            {!hideBranding && (
              <p className="text-[10px] text-muted-foreground italic">
                Önce yukarıdaki "Botla Logosunu Gizle" seçeneğini aktifleştirin.
              </p>
            )}
            
            {hideBranding && canCustomBranding ? (
              <div className="space-y-2.5">
                <div className="space-y-1">
                  <label htmlFor="brand-logo" className="text-[10px] text-muted-foreground">Logo</label>
                  <Input 
                    id="brand-logo"
                    placeholder="https://site.com/logo.png"
                    value={customBranding?.logo_url || ''}
                    onChange={(e) => {
                      setLogoError(false)
                      setCustomBranding({ ...(customBranding || {}), logo_url: e.target.value })
                    }}
                    className="bg-background h-8 text-xs"
                  />
                  {logoError && (
                    <p className="text-[9px] text-rose-500 flex items-center gap-1">
                      <AlertCircle className="w-2.5 h-2.5 flex-shrink-0" />
                      Yüklenemedi
                    </p>
                  )}
                </div>
                <div className="space-y-1">
                  <label htmlFor="brand-text" className="text-[10px] text-muted-foreground">Metin</label>
                  <Input 
                    id="brand-text"
                    placeholder="Powered by Şirket"
                    value={customBranding?.text || ''}
                    onChange={(e) => setCustomBranding({ ...(customBranding || {}), text: e.target.value })}
                    className="bg-background h-8 text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <label htmlFor="brand-link" className="text-[10px] text-muted-foreground">Link</label>
                  <Input 
                    id="brand-link"
                    placeholder="https://sirket.com"
                    value={customBranding?.link || ''}
                    onChange={(e) => setCustomBranding({ ...(customBranding || {}), link: e.target.value })}
                    className="bg-background h-8 text-xs"
                  />
                </div>
                
              </div>
            ) : hideBranding && !canCustomBranding ? (
              <div className="p-3 bg-gray-50 rounded-lg text-center">
                <Lock className="w-5 h-5 text-muted-foreground mx-auto mb-1" />
                <p className="text-[10px] text-muted-foreground">
                  Enterprise plana yükseltin.
                </p>
              </div>
            ) : null}
          </div>

          
        </div>
      )}
    </div>
  )
}
