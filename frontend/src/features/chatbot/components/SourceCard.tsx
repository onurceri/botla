import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  CheckCircle2,
  RefreshCw,
  AlertCircle,
  Trash2,
  FileText,
  Globe,
  Type as TypeIcon,
  Database,
  Clock,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import ChunkInspector from './ChunkInspector'

export interface Source {
  id: string
  source_type: 'pdf' | 'url' | 'text'
  original_filename?: string
  source_url?: string
  status: 'queued' | 'processing' | 'completed' | 'failed' | 'pending'
  chunk_count: number
  capability_summary?: string
  error_message?: string
}

interface SourceCardProps {
  source: Source
  userPlan: string
  onDelete: (id: string) => void
  onRefresh: (id: string) => void
  isRefreshing?: boolean
}

const getSourceIcon = (type: string) => {
  switch (type) {
    case 'pdf':
      return <FileText className="w-5 h-5" />
    case 'url':
      return <Globe className="w-5 h-5" />
    case 'text':
      return <TypeIcon className="w-5 h-5" />
    default:
      return <FileText className="w-5 h-5" />
  }
}

const getSourceTypeLabel = (type: string) => {
  switch (type) {
    case 'pdf':
      return 'PDF Doküman'
    case 'url':
      return 'Web Sayfası'
    case 'text':
      return 'Metin'
    default:
      return type.toUpperCase()
  }
}

const getStatusConfig = (status: string) => {
  switch (status) {
    case 'completed':
      return {
        icon: <CheckCircle2 className="w-4 h-4" />,
        label: 'Tamamlandı',
        className: 'bg-emerald-100 text-emerald-700 border-emerald-200',
        iconBg: 'bg-emerald-500/10 text-emerald-600',
      }
    case 'processing':
      return {
        icon: <RefreshCw className="w-4 h-4 animate-spin" />,
        label: 'İşleniyor',
        className: 'bg-blue-100 text-blue-700 border-blue-200',
        iconBg: 'bg-blue-500/10 text-blue-600',
      }
    case 'failed':
      return {
        icon: <AlertCircle className="w-4 h-4" />,
        label: 'Başarısız',
        className: 'bg-red-100 text-red-700 border-red-200',
        iconBg: 'bg-red-500/10 text-red-600',
      }
    case 'queued':
    case 'pending':
    default:
      return {
        icon: <Clock className="w-4 h-4 animate-pulse" />,
        label: 'Beklemede',
        className: 'bg-amber-100 text-amber-700 border-amber-200',
        iconBg: 'bg-amber-500/10 text-amber-600',
      }
  }
}

export default function SourceCard({
  source,
  userPlan,
  onDelete,
  onRefresh,
  isRefreshing,
}: SourceCardProps) {
  const [inspectorOpen, setInspectorOpen] = useState(false)
  const canRefresh = userPlan !== 'free'
  const statusConfig = getStatusConfig(source.status)
  const isProcessing = source.status === 'pending' || source.status === 'processing'
  const sourceName = source.original_filename || source.source_url || 'İsimsiz Kaynak'

  return (
    <>
      <ChunkInspector sourceId={source.id} open={inspectorOpen} onOpenChange={setInspectorOpen} />

      <div
        className={cn(
          'group relative rounded-2xl border border-border/60 p-5',
          'bg-white/80 backdrop-blur-xl shadow-sm',
          'md:hover:shadow-lg md:hover:shadow-primary/5 md:hover:border-primary/20',
          'md:hover:scale-[1.02]',
          'transition-all duration-300 ease-out',
          isProcessing && 'ring-2 ring-blue-500/20 ring-offset-2',
        )}
        data-testid="source-card"
      >
        {/* Status Indicator - Top Right */}
        <div className="absolute -top-2 -right-2">
          <TooltipProvider delayDuration={0}>
            <Tooltip>
              <TooltipTrigger asChild>
                <span
                  className={cn(
                    'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
                    'border shadow-sm',
                    statusConfig.className,
                    source.status === 'failed' && source.error_message && 'cursor-help',
                  )}
                >
                  {statusConfig.icon}
                  {statusConfig.label}
                </span>
              </TooltipTrigger>
              {source.status === 'failed' && source.error_message && (
                <TooltipContent side="top" className="max-w-xs">
                  <p className="break-words">{source.error_message}</p>
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        </div>

        {/* Header: Icon + Type */}
        <div className="flex items-start gap-4">
          <div
            className={cn(
              'flex items-center justify-center w-12 h-12 rounded-xl shrink-0',
              'transition-colors duration-300',
              statusConfig.iconBg,
            )}
          >
            {getSourceIcon(source.source_type)}
          </div>

          <div className="flex-1 min-w-0 pt-1">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground/70 mb-1">
              {getSourceTypeLabel(source.source_type)}
            </p>
            <h3 className="font-semibold text-foreground truncate pr-4" title={sourceName}>
              {sourceName}
            </h3>
          </div>
        </div>

        {/* Stats Row */}
        <div className="mt-4 flex items-center gap-4 text-sm text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <Database className="w-4 h-4" />
            <span className="font-medium text-foreground">{source.chunk_count}</span>
            <span>parça</span>
          </div>

          {source.capability_summary && (
            <Dialog>
              <DialogTrigger asChild>
                <button className="flex items-center gap-1.5 hover:text-primary transition-colors">
                  <FileText className="w-4 h-4" />
                  <span className="text-xs underline underline-offset-2">Özet</span>
                </button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Kaynak Özeti</DialogTitle>
                </DialogHeader>
                <div className="mt-4 text-sm text-foreground whitespace-pre-wrap">
                  {source.capability_summary}
                </div>
              </DialogContent>
            </Dialog>
          )}
        </div>

        {/* Action Buttons */}
        <div
          className={cn(
            'mt-4 pt-4 border-t border-border/50 flex items-center justify-between gap-2',
            'md:opacity-0 md:group-hover:opacity-100 transition-opacity duration-200',
          )}
        >
          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground hover:text-foreground"
            onClick={() => setInspectorOpen(true)}
            aria-label="Parçaları İncele"
          >
            <Database className="w-4 h-4 mr-1.5" />
            Parçalar
          </Button>

          <div className="flex items-center gap-1">
            {source.source_type === 'url' && (
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-primary"
                        aria-label="Kaynağı Yenile"
                        disabled={!canRefresh || isProcessing || isRefreshing}
                        onClick={() => onRefresh(source.id)}
                      >
                        <RefreshCw className={cn('w-4 h-4', isRefreshing && 'animate-spin')} />
                      </Button>
                    </span>
                  </TooltipTrigger>
                  {!canRefresh && (
                    <TooltipContent>
                      <p>Yenileme özelliği ücretli planlarda aktiftir</p>
                    </TooltipContent>
                  )}
                </Tooltip>
              </TooltipProvider>
            )}

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 text-muted-foreground hover:text-destructive"
                  aria-label="Kaynağı Sil"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    Bu kaynağı silmek istediğinizden emin misiniz?
                  </AlertDialogTitle>
                  <AlertDialogDescription>
                    "{sourceName}" kaynağı ve tüm ilişkili veri parçaları silinecektir. Bu işlem
                    geri alınamaz.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>İptal</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={() => onDelete(source.id)}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    Sil
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>
    </>
  )
}
