import { cn } from '@/lib/utils'
import { CheckCircle2, Clock, Loader2, AlertCircle, XCircle } from 'lucide-react'

export type StatusType = 
  | 'completed' 
  | 'ready' 
  | 'processing' 
  | 'pending' 
  | 'failed' 
  | 'denied'
  | 'active'
  | 'inactive'
  | 'unknown'

interface StatusBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  status: string
  showIcon?: boolean
  size?: 'xs' | 'sm' | 'md'
}

const statusConfig: Record<StatusType, { 
  label: string
  icon: React.ComponentType<{ className?: string }>
  colors: string
}> = {
  completed: {
    label: 'Tamamlandı',
    icon: CheckCircle2,
    colors: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  },
  ready: {
    label: 'Hazır',
    icon: CheckCircle2,
    colors: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  },
  active: {
    label: 'Aktif',
    icon: CheckCircle2,
    colors: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  },
  processing: {
    label: 'İşleniyor',
    icon: Loader2,
    colors: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  },
  pending: {
    label: 'Bekliyor',
    icon: Clock,
    colors: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
  },
  failed: {
    label: 'Başarısız',
    icon: XCircle,
    colors: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  },
  denied: {
    label: 'Reddedildi',
    icon: XCircle,
    colors: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  },
  inactive: {
    label: 'Pasif',
    icon: AlertCircle,
    colors: 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20',
  },
  unknown: {
    label: 'Bilinmiyor',
    icon: AlertCircle,
    colors: 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20',
  },
}

/**
 * Normalizes a status string to a valid StatusType
 */
export function normalizeStatus(status: string | null | undefined): StatusType {
  if (!status) return 'unknown'
  const lower = status.toLowerCase()
  if (lower === 'completed' || lower === 'complete' || lower === 'done') return 'completed'
  if (lower === 'ready') return 'ready'
  if (lower === 'active' || lower === 'healthy' || lower === 'ok') return 'active'
  if (lower === 'processing' || lower === 'in_progress' || lower === 'running') return 'processing'
  if (lower === 'pending' || lower === 'waiting' || lower === 'queued' || lower === 'degraded') return 'pending'
  if (lower === 'failed' || lower === 'error' || lower === 'failure' || lower === 'unhealthy' || lower === 'down') return 'failed'
  if (lower === 'denied' || lower === 'rejected') return 'denied'
  if (lower === 'inactive' || lower === 'disabled') return 'inactive'
  return 'unknown'
}

/**
 * A reusable status badge component for consistent status display across admin pages
 */
export function StatusBadge({
  status,
  showIcon = true,
  size = 'sm',
  className,
  ...props
}: StatusBadgeProps) {
  const normalizedStatus = normalizeStatus(status)
  const config = statusConfig[normalizedStatus]
  const Icon = config.icon

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border font-medium',
        config.colors,
        size === 'xs' && 'px-2 py-0.5 text-[10px]',
        size === 'sm' && 'px-2.5 py-0.5 text-xs',
        size === 'md' && 'px-3 py-1 text-sm',
        className
      )}
      {...props}
    >
      {showIcon && (
        <Icon 
          className={cn(
            size === 'xs' && 'w-3 h-3',
            size === 'sm' && 'w-3.5 h-3.5',
            size === 'md' && 'w-4 h-4',
            normalizedStatus === 'processing' && 'animate-spin'
          )} 
        />
      )}
      <span>{config.label}</span>
    </span>
  )
}

export { statusConfig }
