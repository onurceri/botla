import { useEffect, useState } from 'react'
import {
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  FileText,
  Globe,
  Type as TypeIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface ProcessingSource {
  id: string
  source_type: 'pdf' | 'url' | 'text'
  name: string
  status: 'queued' | 'processing' | 'completed' | 'failed'
  progress?: number // 0-100
  error_message?: string
}

interface IngestionProgressProps {
  sources: ProcessingSource[]
  className?: string
}

const getSourceIcon = (type: string) => {
  switch (type) {
    case 'pdf':
      return <FileText className="w-4 h-4" />
    case 'url':
      return <Globe className="w-4 h-4" />
    case 'text':
      return <TypeIcon className="w-4 h-4" />
    default:
      return <FileText className="w-4 h-4" />
  }
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'completed':
      return <CheckCircle2 className="w-4 h-4 text-emerald-500" />
    case 'failed':
      return <AlertCircle className="w-4 h-4 text-red-500" />
    case 'processing':
      return <RefreshCw className="w-4 h-4 text-blue-500 animate-spin" />
    default:
      return <RefreshCw className="w-4 h-4 text-amber-500 animate-pulse" />
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'completed':
      return 'Tamamlandı'
    case 'failed':
      return 'Başarısız'
    case 'processing':
      return 'İşleniyor'
    default:
      return 'Beklemede'
  }
}

function ProgressItem({ source }: { source: ProcessingSource }) {
  const [displayProgress, setDisplayProgress] = useState(0)

  // Smooth progress animation
  useEffect(() => {
    const target =
      source.progress ??
      (source.status === 'completed' ? 100 : source.status === 'processing' ? 50 : 0)
    const interval = setInterval(() => {
      setDisplayProgress((prev) => {
        if (prev < target) {
          return Math.min(prev + 2, target)
        }
        return prev
      })
    }, 20)
    return () => clearInterval(interval)
  }, [source.progress, source.status])

  return (
    <div
      className={cn(
        'flex items-center gap-3 p-3 rounded-xl',
        'bg-white/60 backdrop-blur-sm border border-border/40',
        'animate-in slide-in-from-top-2 fade-in duration-300',
        source.status === 'failed' && 'border-red-200 bg-red-50/50',
      )}
      data-testid="progress-item"
    >
      {/* Source Type Icon */}
      <div
        className={cn(
          'flex items-center justify-center w-8 h-8 rounded-lg shrink-0',
          source.status === 'completed'
            ? 'bg-emerald-100 text-emerald-600'
            : source.status === 'failed'
              ? 'bg-red-100 text-red-600'
              : source.status === 'processing'
                ? 'bg-blue-100 text-blue-600'
                : 'bg-amber-100 text-amber-600',
        )}
      >
        {getSourceIcon(source.source_type)}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2 mb-1">
          <span className="text-sm font-medium text-foreground truncate" title={source.name}>
            {source.name}
          </span>
          <div className="flex items-center gap-1.5 shrink-0">
            {getStatusIcon(source.status)}
            <span
              className={cn(
                'text-xs font-medium',
                source.status === 'completed'
                  ? 'text-emerald-600'
                  : source.status === 'failed'
                    ? 'text-red-600'
                    : source.status === 'processing'
                      ? 'text-blue-600'
                      : 'text-amber-600',
              )}
            >
              {getStatusText(source.status)}
            </span>
          </div>
        </div>

        {/* Progress Bar */}
        {(source.status === 'processing' || source.status === 'queued') && (
          <div className="h-1.5 bg-muted/50 rounded-full overflow-hidden">
            <div
              className={cn(
                'h-full rounded-full transition-all duration-300 ease-out',
                source.status === 'processing'
                  ? 'bg-gradient-to-r from-blue-400 to-blue-600'
                  : 'bg-amber-400',
              )}
              style={{ width: `${displayProgress}%` }}
              data-testid="progress-bar"
            />
          </div>
        )}

        {/* Error Message */}
        {source.status === 'failed' && source.error_message && (
          <p className="text-xs text-red-500 mt-1 truncate" title={source.error_message}>
            {source.error_message}
          </p>
        )}
      </div>
    </div>
  )
}

export default function IngestionProgress({ sources, className }: IngestionProgressProps) {
  const processingCount = sources.filter(
    (s) => s.status === 'processing' || s.status === 'queued',
  ).length

  if (sources.length === 0) return null

  return (
    <div className={cn('space-y-3', className)} data-testid="ingestion-progress">
      {/* Header */}
      {processingCount > 0 && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <RefreshCw className="w-4 h-4 animate-spin text-blue-500" />
          <span>
            <span className="font-medium text-foreground">{processingCount}</span> kaynak
            işleniyor...
          </span>
        </div>
      )}

      {/* Progress Items */}
      <div className="space-y-2">
        {sources.map((source) => (
          <ProgressItem key={source.id} source={source} />
        ))}
      </div>
    </div>
  )
}

export type { ProcessingSource, IngestionProgressProps }
