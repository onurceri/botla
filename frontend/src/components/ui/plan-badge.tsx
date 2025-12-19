import * as React from "react"
import { cn } from "@/lib/utils"
import { Crown, Sparkles } from "lucide-react"

export type PlanTier = 'free' | 'pro' | 'ultra'

interface PlanBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  plan: PlanTier
  showIcon?: boolean
  size?: 'sm' | 'md'
}

const planConfig = {
  ultra: {
    label: 'ULTRA',
    className: 'bg-primary text-primary-foreground shadow-sm shadow-primary/30',
    icon: Crown,
  },
  pro: {
    label: 'PRO',
    className: 'bg-secondary text-secondary-foreground border border-primary/20',
    icon: Sparkles,
  },
  free: {
    label: 'FREE',
    className: 'bg-muted text-muted-foreground border border-border',
    icon: null,
  },
}

/**
 * A reusable badge component for displaying plan tiers (PRO, ULTRA, FREE)
 * Uses the project's Amber color palette for consistency across the app.
 */
function PlanBadge({ 
  plan, 
  showIcon = true, 
  size = 'sm',
  className, 
  ...props 
}: PlanBadgeProps) {
  const config = planConfig[plan]
  const Icon = config.icon
  
  const sizeClasses = {
    sm: 'px-1.5 py-0.5 text-[10px] gap-1',
    md: 'px-2.5 py-0.5 text-xs gap-1.5',
  }
  
  const iconSizes = {
    sm: 'h-2.5 w-2.5',
    md: 'h-3 w-3',
  }

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full font-bold tracking-wide transition-colors",
        sizeClasses[size],
        config.className,
        className
      )}
      {...props}
    >
      {showIcon && Icon && <Icon className={iconSizes[size]} />}
      {config.label}
    </span>
  )
}

export { PlanBadge }
