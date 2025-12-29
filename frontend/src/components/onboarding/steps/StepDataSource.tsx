/**
 * Step 2: Data Source Selection
 * Allows the user to add content for the chatbot to learn from.
 * This step is optional - users can skip it and add sources later.
 */

import { useRef } from 'react'
import { Upload, Type, Globe, FileText, Info, SkipForward } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import type { SourceType } from '../types'
import { MIN_TEXT_LENGTH } from '../types'

interface StepDataSourceProps {
  sourceType: SourceType
  textContent: string
  urlContent: string
  pdfFile: File | null
  planData?: any
  planCode?: string
  onSourceTypeChange: (type: SourceType) => void
  onTextContentChange: (content: string) => void
  onUrlContentChange: (url: string) => void
  onFileSelect: (e: React.ChangeEvent<HTMLInputElement>) => void
  onFileRemove: () => void
  onSkipStep?: () => void
}

const SOURCE_OPTIONS = [
  { type: 'text' as const, icon: Type, label: 'Metin' },
  { type: 'url' as const, icon: Globe, label: 'URL' },
  { type: 'file' as const, icon: FileText, label: 'PDF' },
] as const

/**
 * Checks if a source type has content added
 */
function hasContent(
  type: SourceType,
  textContent: string,
  urlContent: string,
  pdfFile: File | null
): boolean {
  switch (type) {
    case 'text':
      return textContent.trim().length > 0
    case 'url':
      return urlContent.trim().length > 0
    case 'file':
      return pdfFile !== null
    default:
      return false
  }
}

/**
 * Formats large numbers with thousand separators
 */
function formatNumber(num: number): string {
  return num.toLocaleString('tr-TR')
}

export function StepDataSource({
  sourceType,
  textContent,
  urlContent,
  pdfFile,
  planData,
  planCode = 'free',
  onSourceTypeChange,
  onTextContentChange,
  onUrlContentChange,
  onFileSelect,
  onFileRemove,
  onSkipStep,
}: StepDataSourceProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  
  const maxTextLength = planData?.files?.max_text_length || 10000
  const maxFileSizeMB = planData?.files?.max_size_mb || 10
  const maxURLs = planData?.scraping?.max_urls_per_bot || 1
  const maxFiles = planData?.files?.max_files_per_bot || 1
  
  const isFreePlan = planCode === 'free'
  const planName = planCode.charAt(0).toUpperCase() + planCode.slice(1)

  // Determine if any source type has content (to disable others)
  const anyHasContent = hasContent('text', textContent, urlContent, pdfFile) ||
    hasContent('url', textContent, urlContent, pdfFile) ||
    hasContent('file', textContent, urlContent, pdfFile)

  // Determine if current source type has content
  const currentHasContent = hasContent(sourceType, textContent, urlContent, pdfFile)

  // Calculate if text is within valid range
  const textLength = textContent.length
  const isTextValid = textLength >= MIN_TEXT_LENGTH && textLength <= maxTextLength
  const isTextTooShort = textLength > 0 && textLength < MIN_TEXT_LENGTH
  const isTextTooLong = textLength > maxTextLength

  return (
    <div className="space-y-6 animate-fade-up">
      <div className="text-center mb-6">
        <div
          className="inline-flex items-center justify-center w-16 h-16 rounded-2xl 
                      bg-primary/10 mb-4"
        >
          <Upload className="w-8 h-8 text-primary" />
        </div>
        <h2 className="heading-md text-foreground mb-2">Bilgi Kaynağı Ekleyin</h2>
        <p className="body-sm">Botunuzun öğrenmesini istediğiniz içeriği seçin</p>
      </div>

      {/* Info Alert */}
      <div className="flex items-start gap-3 p-4 rounded-xl bg-blue-50 border border-blue-200 text-blue-800">
        <Info className="w-5 h-5 flex-shrink-0 mt-0.5" />
        <div className="text-sm">
          <p className="font-medium mb-1">Bu adım opsiyoneldir</p>
          <p className="text-blue-700">
            Kaynakları daha sonra Dashboard'dan ekleyebilirsiniz. 
            {isFreePlan ? (
              <>Ücretsiz planda <strong>tek bir kaynak</strong> ekleyebilirsiniz.</>
            ) : (
              <>{planName} planınızla birden fazla kaynak ekleyebilirsiniz.</>
            )}
          </p>
        </div>
      </div>

      {/* Source Type Selector */}
      <div className="grid grid-cols-3 gap-3">
        {SOURCE_OPTIONS.map(({ type, icon: Icon, label }) => {
          // Disable if another source type has content
          const isDisabled = anyHasContent && !hasContent(type, textContent, urlContent, pdfFile) && sourceType !== type
          const isSelected = sourceType === type

          return (
            <button
              key={type}
              onClick={() => !isDisabled && onSourceTypeChange(type)}
              disabled={isDisabled}
              className={`p-4 rounded-xl border-2 transition-all duration-200
                ${isSelected
                  ? 'border-primary bg-primary/5 text-foreground'
                  : isDisabled
                    ? 'border-border/30 bg-muted/50 text-muted-foreground/50 cursor-not-allowed'
                    : 'border-border/50 bg-white/50 text-muted-foreground hover:border-border cursor-pointer'
                }
              `}
            >
              <Icon className="w-6 h-6 mx-auto mb-2" />
              <span className="text-sm font-medium">{label}</span>
              {isDisabled && (
                <span className="block text-xs mt-1 text-muted-foreground/60">Başka kaynak seçili</span>
              )}
            </button>
          )
        })}
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
            className={`min-h-[160px] rounded-xl border-border/50 bg-white/50 
                     focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                     ${isTextTooLong ? 'border-red-500 focus:border-red-500 focus:ring-red-200' : ''}`}
          />
          <div className="flex items-center justify-between text-xs">
            <span className={`${isTextTooShort ? 'text-amber-600' : isTextTooLong ? 'text-red-600' : 'text-muted-foreground'}`}>
              {formatNumber(MIN_TEXT_LENGTH)} - {formatNumber(maxTextLength)} karakter
            </span>
            <span className={`font-medium ${isTextValid ? 'text-green-600' : isTextTooShort ? 'text-amber-600' : isTextTooLong ? 'text-red-600' : 'text-muted-foreground'}`}>
              {formatNumber(textLength)} karakter
            </span>
          </div>
          {/* Progress bar */}
          <div className="h-1.5 bg-muted rounded-full overflow-hidden">
            <div 
              className={`h-full transition-all duration-300 ${
                isTextTooLong ? 'bg-red-500' : isTextValid ? 'bg-green-500' : 'bg-amber-500'
              }`}
              style={{ width: `${Math.min((textLength / maxTextLength) * 100, 100)}%` }}
            />
          </div>
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
            Site içeriği otomatik olarak analiz edilecek (maksimum {maxURLs} sayfa)
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
                Maksimum {maxFileSizeMB}MB
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

      {/* Skip Step Button */}
      {onSkipStep && !currentHasContent && (
        <div className="pt-4 border-t border-border/50">
          <button
            onClick={onSkipStep}
            className="flex items-center justify-center gap-2 w-full py-3 text-sm text-muted-foreground 
                     hover:text-foreground transition-colors rounded-lg hover:bg-muted/50"
          >
            <SkipForward className="w-4 h-4" />
            Bu adımı atla, daha sonra ekle
          </button>
        </div>
      )}
    </div>
  )
}

