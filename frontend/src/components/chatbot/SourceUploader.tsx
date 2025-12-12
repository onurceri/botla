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
  disabledModes = []
}: SourceUploaderProps) => {
  const { 
    activeMode, setActiveMode,
    loading,
    inputValue, setInputValue,
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
              <div className="space-y-2">
                <textarea 
                  className={cn(
                    "flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
                    inputValue.length > maxTextLength && "border-destructive focus-visible:ring-destructive"
                  )}
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder="Metin içeriğini buraya yapıştırın..."
                />
                <div className="flex justify-between items-center">
                  <span className={cn("text-xs", inputValue.length > maxTextLength ? "text-destructive font-medium" : "text-muted-foreground")}>
                    {inputValue.length > maxTextLength 
                      ? `${(inputValue.length - maxTextLength).toLocaleString()} karakter aşıldı` 
                      : `${(maxTextLength - inputValue.length).toLocaleString()} karakter kaldı`
                    }
                  </span>
                  <Button 
                    onClick={handleSubmit} 
                    disabled={loading || inputValue.length > maxTextLength || inputValue.length === 0} 
                    className="rounded-full"
                  >
                    {loading ? 'Ekleniyor...' : 'Ekle'}
                  </Button>
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
