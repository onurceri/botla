import { useEffect } from 'react'
import { Upload, Link as LinkIcon, X, Type } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { useSourceUpload } from '@/features/chatbot/hooks/useSourceUpload'

interface SourceUploaderProps {
  onUploadPDF: (file: File) => Promise<void>
  onUploadURL: (url: string) => Promise<void>
  onUploadText: (text: string) => Promise<void>
  extraUrlSettings?: React.ReactNode
  maxFileSizeMB?: number
  maxTextLength?: number
  disabled?: boolean
  disabledModes?: ('pdf' | 'url' | 'text')[]
}

const SourceUploader = ({
  onUploadPDF,
  onUploadURL,
  onUploadText,
  extraUrlSettings,
  maxFileSizeMB = 50,
  maxTextLength = 400000, // ~400k characters default
  disabled = false,
  disabledModes = [],
}: SourceUploaderProps) => {
  const {
    activeMode,
    setActiveMode,
    loading,
    inputValue,
    setInputValue,
    fileInputRef,
    handleFileChange,
    handleSubmit,
  } = useSourceUpload({ onUploadPDF, onUploadURL, onUploadText, maxFileSizeMB })

  useEffect(() => {
    if (activeMode && disabledModes.includes(activeMode)) {
      setActiveMode(null)
    }
  }, [activeMode, disabledModes, setActiveMode])

  return (
    <div className="space-y-4">
      {/* Upload Type Selection */}
      <div className="flex flex-wrap gap-2">
        <Button
          variant={activeMode === 'pdf' ? 'default' : 'outline'}
          onClick={() => setActiveMode(activeMode === 'pdf' ? null : 'pdf')}
          className="gap-2"
          disabled={disabled || (disabledModes.includes('pdf') && activeMode !== 'pdf')}
        >
          <Upload className="w-4 h-4" />
          PDF Yükle
        </Button>
        <Button
          variant={activeMode === 'url' ? 'default' : 'outline'}
          onClick={() => setActiveMode(activeMode === 'url' ? null : 'url')}
          className="gap-2"
          disabled={disabled || (disabledModes.includes('url') && activeMode !== 'url')}
        >
          <LinkIcon className="w-4 h-4" />
          Web Sitesi
        </Button>
        <Button
          variant={activeMode === 'text' ? 'default' : 'outline'}
          onClick={() => setActiveMode(activeMode === 'text' ? null : 'text')}
          className="gap-2"
          disabled={disabled || (disabledModes.includes('text') && activeMode !== 'text')}
        >
          <Type className="w-4 h-4" />
          Metin
        </Button>
      </div>

      {/* Input Area */}
      {activeMode && (
        <div className="animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="relative p-6 rounded-2xl bg-white/60 backdrop-blur border border-border shadow-sm">
            <button
              onClick={() => setActiveMode(null)}
              className="absolute top-2 right-2 p-1 text-muted-foreground hover:text-foreground"
            >
              <X className="w-4 h-4" />
            </button>

            {activeMode === 'pdf' && (
              <div className="text-center">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf"
                  onChange={handleFileChange}
                  className="hidden"
                  id="pdf-upload"
                />
                <label
                  htmlFor="pdf-upload"
                  className="inline-flex items-center justify-center px-4 py-2 rounded-md bg-primary text-primary-foreground font-medium cursor-pointer hover:bg-primary/90 transition-colors disabled:opacity-50"
                >
                  {loading ? 'Yükleniyor...' : 'Dosya Seç'}
                </label>
                <p className="mt-2 text-xs text-muted-foreground">Maksimum {maxFileSizeMB}MB</p>
              </div>
            )}

            {activeMode === 'url' && (
              <div className="space-y-4">
                <div className="flex gap-2">
                  <Input
                    value={inputValue}
                    onChange={(e) => setInputValue(e.target.value)}
                    placeholder="https://example.com"
                    className="flex-1"
                  />
                  <Button onClick={handleSubmit} disabled={loading} className="rounded-full">
                    {loading ? 'Ekleniyor...' : 'Ekle'}
                  </Button>
                </div>
                {extraUrlSettings}
              </div>
            )}

            {activeMode === 'text' && (
              <div className="space-y-4">
                {/* Header with info */}
                <div className="flex items-center justify-between text-xs text-muted-foreground">
                  <span className="font-medium">Metin İçeriği</span>
                  <span>Maks. {maxTextLength.toLocaleString()} karakter</span>
                </div>

                {/* Large resizable textarea */}
                <textarea
                  className={cn(
                    'flex min-h-[300px] max-h-[500px] w-full rounded-xl border border-input bg-white/80 backdrop-blur-sm px-4 py-3 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary/40 disabled:cursor-not-allowed disabled:opacity-50 resize-y font-mono leading-relaxed',
                    inputValue.length > maxTextLength &&
                      'border-destructive focus-visible:ring-destructive/20 focus-visible:border-destructive',
                  )}
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder="Metin içeriğinizi buraya yapıştırın...

Örnekler:
• Ürün açıklamaları
• SSS (Sıkça Sorulan Sorular) 
• Kullanım kılavuzları
• Şirket bilgileri
• Blog yazıları"
                />

                {/* Progress bar */}
                <div className="space-y-2">
                  <div className="h-1.5 bg-slate-100 rounded-full overflow-hidden">
                    <div 
                      className={cn(
                        "h-full rounded-full transition-all duration-300",
                        inputValue.length > maxTextLength 
                          ? "bg-red-500" 
                          : inputValue.length > maxTextLength * 0.9
                            ? "bg-amber-500"
                            : "bg-primary"
                      )}
                      style={{ width: `${Math.min((inputValue.length / maxTextLength) * 100, 100)}%` }}
                    />
                  </div>
                  
                  <div className="flex justify-between items-center">
                    <div className="flex items-center gap-3">
                      <span
                        className={cn(
                          'text-xs font-medium',
                          inputValue.length > maxTextLength
                            ? 'text-destructive'
                            : inputValue.length > maxTextLength * 0.9
                              ? 'text-amber-600'
                              : 'text-muted-foreground',
                        )}
                      >
                        {inputValue.length.toLocaleString()} / {maxTextLength.toLocaleString()} karakter
                      </span>
                      {inputValue.length > 0 && inputValue.length <= maxTextLength && (
                        <span className="text-xs text-emerald-600 font-medium">
                          ✓ {(maxTextLength - inputValue.length).toLocaleString()} kaldı
                        </span>
                      )}
                      {inputValue.length > maxTextLength && (
                        <span className="text-xs text-destructive font-medium">
                          ⚠ {(inputValue.length - maxTextLength).toLocaleString()} fazla
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      {inputValue.length > 0 && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setInputValue('')}
                          className="h-8 px-3 text-xs text-muted-foreground hover:text-foreground"
                        >
                          Temizle
                        </Button>
                      )}
                      <Button
                        onClick={handleSubmit}
                        disabled={
                          loading || inputValue.length > maxTextLength || inputValue.length === 0
                        }
                        className="rounded-full h-9 px-5"
                      >
                        {loading ? 'Ekleniyor...' : 'Ekle'}
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default SourceUploader
