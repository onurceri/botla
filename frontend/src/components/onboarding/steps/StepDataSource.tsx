/**
 * Step 2: Data Source Selection
 * Allows the user to add content for the chatbot to learn from.
 */

import { useRef } from 'react'
import { Upload, Type, Globe, FileText } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import type { SourceType } from '../types'
import { MAX_FILE_SIZE_MB } from '../types'

interface StepDataSourceProps {
  sourceType: SourceType
  textContent: string
  urlContent: string
  pdfFile: File | null
  onSourceTypeChange: (type: SourceType) => void
  onTextContentChange: (content: string) => void
  onUrlContentChange: (url: string) => void
  onFileSelect: (e: React.ChangeEvent<HTMLInputElement>) => void
  onFileRemove: () => void
}

const SOURCE_OPTIONS = [
  { type: 'text' as const, icon: Type, label: 'Metin' },
  { type: 'url' as const, icon: Globe, label: 'URL' },
  { type: 'file' as const, icon: FileText, label: 'PDF' },
] as const

export function StepDataSource({
  sourceType,
  textContent,
  urlContent,
  pdfFile,
  onSourceTypeChange,
  onTextContentChange,
  onUrlContentChange,
  onFileSelect,
  onFileRemove,
}: StepDataSourceProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  return (
    <div className="space-y-6 animate-fade-up">
      <div className="text-center mb-8">
        <div
          className="inline-flex items-center justify-center w-16 h-16 rounded-2xl 
                      bg-primary/10 mb-4"
        >
          <Upload className="w-8 h-8 text-primary" />
        </div>
        <h2 className="heading-md text-foreground mb-2">Bilgi Kaynağı Ekleyin</h2>
        <p className="body-sm">Botunuzun öğrenmesini istediğiniz içeriği seçin</p>
      </div>

      {/* Source Type Selector */}
      <div className="grid grid-cols-3 gap-3 mb-6">
        {SOURCE_OPTIONS.map(({ type, icon: Icon, label }) => (
          <button
            key={type}
            onClick={() => onSourceTypeChange(type)}
            className={`p-4 rounded-xl border-2 transition-all duration-200 cursor-pointer
              ${
                sourceType === type
                  ? 'border-primary bg-primary/5 text-foreground'
                  : 'border-border/50 bg-white/50 text-muted-foreground hover:border-border'
              }
            `}
          >
            <Icon className="w-6 h-6 mx-auto mb-2" />
            <span className="text-sm font-medium">{label}</span>
          </button>
        ))}
      </div>

      {/* Text Content Input */}
      {sourceType === 'text' && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground" htmlFor="textContent">
            İçerik
          </label>
          <Textarea
            id="textContent"
            placeholder="Botunuzun öğrenmesini istediğiniz bilgileri buraya yapıştırın..."
            value={textContent}
            onChange={(e) => onTextContentChange(e.target.value)}
            className="min-h-[160px] rounded-xl border-border/50 bg-white/50 
                     focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
          />
          <p className="text-xs text-muted-foreground">
            Minimum 50 karakter ({textContent.length}/50)
          </p>
        </div>
      )}

      {/* URL Input */}
      {sourceType === 'url' && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground" htmlFor="urlContent">
            Web Sitesi URL'si
          </label>
          <Input
            id="urlContent"
            type="url"
            placeholder="https://example.com"
            value={urlContent}
            onChange={(e) => onUrlContentChange(e.target.value)}
            className="h-12 rounded-xl border-border/50 bg-white/50 
                     focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
          />
          <p className="text-xs text-muted-foreground">
            Site içeriği otomatik olarak analiz edilecek
          </p>
        </div>
      )}

      {/* File Upload */}
      {sourceType === 'file' && (
        <div className="space-y-4">
          <input
            ref={fileInputRef}
            type="file"
            accept=".pdf,application/pdf"
            onChange={onFileSelect}
            className="hidden"
            id="pdf-upload"
          />

          {!pdfFile ? (
            <label
              htmlFor="pdf-upload"
              className="flex flex-col items-center justify-center p-8 rounded-xl border-2 border-dashed 
                        border-border/50 bg-white/50 hover:border-primary/50 hover:bg-primary/5
                        cursor-pointer transition-all duration-200"
            >
              <Upload className="w-10 h-10 text-muted-foreground mb-3" />
              <span className="text-sm font-medium text-foreground">PDF Dosyası Seçin</span>
              <span className="text-xs text-muted-foreground mt-1">
                Maksimum {MAX_FILE_SIZE_MB}MB
              </span>
            </label>
          ) : (
            <div className="flex items-center gap-4 p-4 rounded-xl border border-border bg-white/80">
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
                <FileText className="w-6 h-6 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-foreground truncate">{pdfFile.name}</p>
                <p className="text-xs text-muted-foreground">
                  {(pdfFile.size / 1024 / 1024).toFixed(2)} MB
                </p>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  onFileRemove()
                  if (fileInputRef.current) fileInputRef.current.value = ''
                }}
                className="text-muted-foreground hover:text-destructive"
              >
                Değiştir
              </Button>
            </div>
          )}

          <p className="text-xs text-muted-foreground">
            PDF dosyası yükleyerek botunuzun bu içeriği öğrenmesini sağlayın
          </p>
        </div>
      )}
    </div>
  )
}
