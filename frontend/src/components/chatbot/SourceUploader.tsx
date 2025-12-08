import { Upload, Link as LinkIcon, FileText, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { useSourceUpload } from '@/features/chatbot/hooks/useSourceUpload'

interface SourceUploaderProps {
  onUploadPDF: (file: File) => Promise<void>
  onUploadURL: (url: string) => Promise<void>
  onUploadText: (text: string) => Promise<void>
  extraUrlSettings?: React.ReactNode
}

const SourceUploader = ({ onUploadPDF, onUploadURL, onUploadText, extraUrlSettings }: SourceUploaderProps) => {
  const { 
    activeMode, setActiveMode,
    loading,
    inputValue, setInputValue,
    fileInputRef,
    handleFileChange,
    handleSubmit,
  } = useSourceUpload({ onUploadPDF, onUploadURL, onUploadText })

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-3 gap-4">
        <button
          onClick={() => setActiveMode(activeMode === 'pdf' ? null : 'pdf')}
          className={cn(
            "flex flex-col items-center justify-center gap-3 p-6 rounded-2xl border-2 border-dashed transition-all duration-200 hover:bg-white/60 backdrop-blur shadow-sm",
            activeMode === 'pdf' 
              ? "border-primary bg-primary/5 text-primary" 
              : "border-border text-muted-foreground hover:border-primary/50 hover:text-foreground"
          )}
        >
          <div className={cn("p-3 rounded-full bg-muted shadow-inner", activeMode === 'pdf' && "bg-primary/20")}>
            <Upload className="w-6 h-6" />
          </div>
          <span className="font-medium">PDF Yükle</span>
        </button>

        <button
          onClick={() => setActiveMode(activeMode === 'url' ? null : 'url')}
          className={cn(
            "flex flex-col items-center justify-center gap-3 p-6 rounded-2xl border-2 border-dashed transition-all duration-200 hover:bg-white/60 backdrop-blur shadow-sm",
            activeMode === 'url' 
              ? "border-blue-500 bg-blue-500/5 text-blue-400" 
              : "border-border text-muted-foreground hover:border-blue-500/50 hover:text-foreground"
          )}
        >
          <div className={cn("p-3 rounded-full bg-muted shadow-inner", activeMode === 'url' && "bg-blue-500/20")}> 
            <LinkIcon className="w-6 h-6" />
          </div>
          <span className="font-medium">Web Sitesi</span>
        </button>

        <button
          onClick={() => setActiveMode(activeMode === 'text' ? null : 'text')}
          className={cn(
            "flex flex-col items-center justify-center gap-3 p-6 rounded-2xl border-2 border-dashed transition-all duration-200 hover:bg-white/60 backdrop-blur shadow-sm",
            activeMode === 'text' 
              ? "border-emerald-500 bg-emerald-500/5 text-emerald-400" 
              : "border-border text-muted-foreground hover:border-emerald-500/50 hover:text-foreground"
          )}
        >
          <div className={cn("p-3 rounded-full bg-muted shadow-inner", activeMode === 'text' && "bg-emerald-500/20")}> 
            <FileText className="w-6 h-6" />
          </div>
          <span className="font-medium">Metin Gir</span>
        </button>
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
                <p className="mt-2 text-xs text-muted-foreground">Maksimum 50MB</p>
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
                  className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder="Metin içeriğini buraya yapıştırın..."
                />
                <div className="flex justify-end">
                  <Button onClick={handleSubmit} disabled={loading} className="rounded-full">
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
