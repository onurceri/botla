import { cn } from '@/lib/utils'

export type PlanTier = 'free' | 'pro' | 'business' | 'ultra'

interface PlanBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  plan: PlanTier
  showIcon?: boolean
  size?: 'xs' | 'sm' | 'md' | 'lg'
  variant?: 'solid' | 'soft' | 'outline'
}

const planConfig = {
  ultra: {
    label: 'ULTRA',
    colors: {
      solid: 'bg-primary text-primary-foreground shadow-sm shadow-primary/30',
      soft: 'bg-primary/10 text-primary border border-primary/20',
      outline: 'bg-transparent text-primary border border-primary/30',
    },
  },
  business: {
    label: 'BUSINESS',
    colors: {
      solid: 'bg-violet-600 text-white shadow-sm shadow-violet-600/30',
      soft: 'bg-violet-500/10 text-violet-600 dark:text-violet-400 border border-violet-500/20',
      outline: 'bg-transparent text-violet-600 dark:text-violet-400 border border-violet-500/30',
    },
  },
  pro: {
    label: 'PRO',
    colors: {
      solid: 'bg-amber-500 text-white shadow-sm shadow-amber-500/30',
      soft: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20',
      outline: 'bg-transparent text-amber-600 dark:text-amber-400 border border-amber-500/30',
    },
  },
  free: {
    label: 'FREE',
    colors: {
      solid: 'bg-muted-foreground text-muted shadow-sm',
      soft: 'bg-muted text-muted-foreground border border-border',
      outline: 'bg-transparent text-muted-foreground border border-border',
    },
  },
}

/**
 * Normalizes a plan ID (which may be a UUID or a friendly name) to a valid PlanTier.
 * Useful for admin pages that receive raw database plan_ids.
 */
export function normalizePlanId(planId: string | null | undefined): PlanTier {
  if (!planId) return 'free'
  const lower = planId.toLowerCase()
  if (lower === 'ultra' || lower.includes('ultra')) return 'ultra'
  if (lower === 'business' || lower.includes('business')) return 'business'
  if (lower === 'pro' || lower.includes('pro')) return 'pro'
  if (lower === 'free' || lower.includes('free')) return 'free'
  // If it looks like a UUID or unknown, default to free
  return 'free'
}

/**
 * A reusable badge component for displaying plan tiers (PRO, ULTRA, FREE)
 * Now supports multiple sizes and variants (soft, solid, outline).
 */
function PlanBadge({
  plan,
  size = 'sm',
  variant = 'soft',
  className,
  ...props
}: PlanBadgeProps) {
  const config = planConfig[plan]
  const colorClass = config.colors[variant]

  return (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full font-black tracking-wider transition-all duration-300',
        colorClass,
        size === 'xs' && 'px-1.5 py-0 px-1 text-[9px]',
        size === 'sm' && 'px-2 py-0.5 text-[10px]',
        size === 'md' && 'px-3 py-1 text-[11px]',
        size === 'lg' && 'px-4 py-1.5 text-xs',
        className,
      )}
      {...props}
    >
      <span>{config.label}</span>
    </div>
  )
}

export { PlanBadge }
