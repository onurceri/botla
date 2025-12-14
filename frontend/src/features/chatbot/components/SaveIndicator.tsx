import { Check, Loader2, AlertCircle, RefreshCw } from 'lucide-react'

type Props = {
  isSaving: boolean
  lastSavedAt: Date | null
  error: string | null
}

export function SaveIndicator({ isSaving, lastSavedAt, error }: Props) {
  if (error) {
    const isRetrying = error.includes('Tekrar deneniyor')
    return (
      <span className="flex items-center gap-1.5 text-sm text-destructive animate-in fade-in duration-200">
        {isRetrying ? (
          <RefreshCw className="w-4 h-4 animate-spin" />
        ) : (
          <AlertCircle className="w-4 h-4" />
        )}
        <span className="max-w-[200px] truncate">{error}</span>
      </span>
    )
  }

  if (isSaving) {
    return (
      <span className="flex items-center gap-1.5 text-sm text-muted-foreground animate-pulse">
        <Loader2 className="w-4 h-4 animate-spin" />
        Kaydediliyor...
      </span>
    )
  }

  if (lastSavedAt) {
    return (
      <span className="flex items-center gap-1.5 text-sm text-green-600 animate-in fade-in duration-200">
        <Check className="w-4 h-4" />
        Kaydedildi
      </span>
    )
  }

  return null
}

