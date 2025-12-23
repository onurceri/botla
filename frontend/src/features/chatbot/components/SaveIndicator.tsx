import { Check, Loader2, AlertCircle, RefreshCw } from 'lucide-react'
import { AUTO_SAVE_RETRY_SUFFIX, SAVE_INDICATOR_MESSAGES } from '@/lib/errorMessages'

type Props = {
  isSaving: boolean
  lastSavedAt: Date | null
  error: string | null
}

export function SaveIndicator({ isSaving, lastSavedAt, error }: Props) {
  if (error) {
    const isRetrying = error.endsWith(AUTO_SAVE_RETRY_SUFFIX)
    return (
      <span className="flex items-center gap-1.5 text-sm text-destructive animate-in fade-in duration-200">
        {isRetrying ? (
          <RefreshCw className="w-4 h-4 animate-spin" />
        ) : (
          <AlertCircle className="w-4 h-4" />
        )}
        <span>
          {isRetrying ? SAVE_INDICATOR_MESSAGES.retrying : SAVE_INDICATOR_MESSAGES.failed}
        </span>
      </span>
    )
  }

  if (isSaving) {
    return (
      <span className="flex items-center gap-1.5 text-sm text-muted-foreground animate-pulse">
        <Loader2 className="w-4 h-4 animate-spin" />
        {SAVE_INDICATOR_MESSAGES.saving}
      </span>
    )
  }

  if (lastSavedAt) {
    return (
      <span className="flex items-center gap-1.5 text-sm text-green-600 animate-in fade-in duration-200">
        <Check className="w-4 h-4" />
        {SAVE_INDICATOR_MESSAGES.saved}
      </span>
    )
  }

  return null
}
