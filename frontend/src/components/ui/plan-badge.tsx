import * as React from 'react'
import { cn } from '@/lib/utils'
import { Crown, Sparkles, User } from 'lucide-react'

export type PlanTier = 'free' | 'pro' | 'ultra'

interface PlanBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  plan: PlanTier
  showIcon?: boolean
  size?: 'xs' | 'sm' | 'md' | 'lg'
  variant?: 'solid' | 'soft' | 'outline'
}

const planConfig = {
  ultra: {
    label: 'ULTRA',
    icon: Crown,
    colors: {
      solid: 'bg-primary text-primary-foreground shadow-sm shadow-primary/30',
      soft: 'bg-primary/10 text-primary border border-primary/20',
      outline: 'bg-transparent text-primary border border-primary/30',
    },
  },
  pro: {
    label: 'PRO',
    icon: Sparkles,
    colors: {
      solid: 'bg-amber-500 text-white shadow-sm shadow-amber-500/30',
      soft: 'bg-amber-500/10 text-amber-600 border border-amber-500/20',
      outline: 'bg-transparent text-amber-600 border border-amber-500/30',
    },
  },
  free: {
    label: 'FREE',
    icon: User,
    colors: {
      solid: 'bg-muted-foreground text-muted shadow-sm',
      soft: 'bg-muted text-muted-foreground border border-border',
      outline: 'bg-transparent text-muted-foreground border border-border',
    },
  },
}

/**
 * A reusable badge component for displaying plan tiers (PRO, ULTRA, FREE)
 * Now supports multiple sizes and variants (soft, solid, outline).
 */
function PlanBadge({
  plan,
  showIcon = true,
  size = 'sm',
  variant = 'soft',
  className,
  ...props
}: PlanBadgeProps) {
  const config = planConfig[plan]
  const Icon = config.icon

  const sizeClasses = {
    xs: 'px-1.5 py-px text-[9px] gap-1',
    sm: 'px-2 py-0.5 text-[10px] gap-1.5',
    md: 'px-2.5 py-0.5 text-xs gap-1.5',
    lg: 'px-3 py-1 text-sm gap-2',
  }

  const iconSizes = {
    xs: 'h-2 w-2',
    sm: 'h-2.5 w-2.5',
    md: 'h-3 w-3',
    lg: 'h-3.5 w-3.5',
  }

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full font-bold tracking-wider transition-all duration-200 select-none',
        sizeClasses[size],
        config.colors[variant],
        className,
      )}
      {...props}
    >
      {showIcon && Icon && <Icon className={cn('shrink-0', iconSizes[size])} />}
      {config.label}
    </span>
  )
}

export { PlanBadge }
