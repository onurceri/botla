import * as React from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface StatsCardProps {
  title: string
  value: number | string
  subtitle?: string
  icon: React.ReactNode
  trend?: { value: number; isPositive: boolean }
  className?: string
  isLoading?: boolean
}

export function StatsCard({ title, value, subtitle, icon, trend, className, isLoading }: StatsCardProps) {
  if (isLoading) {
    return (
      <Card className={cn('overflow-hidden', className)} data-testid="stats-card-skeleton">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
              <div className="h-8 w-24 bg-muted animate-pulse rounded" />
              <div className="h-3 w-16 bg-muted animate-pulse rounded" />
            </div>
            <div className="h-10 w-10 bg-muted animate-pulse rounded-full" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={cn('overflow-hidden', className)} data-testid="stats-card">
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">{title}</p>
            <p className="text-3xl font-bold tracking-tight">{value}</p>
            {(subtitle || trend) && (
              <div className="flex items-center gap-2">
                {trend && (
                  <span
                    className={cn(
                      'text-xs font-semibold',
                      trend.isPositive ? 'text-green-500' : 'text-red-500',
                    )}
                  >
                    {trend.isPositive ? '+' : '-'}
                    {trend.value}%
                  </span>
                )}
                {subtitle && <p className="text-xs text-muted-foreground">{subtitle}</p>}
              </div>
            )}
          </div>
          <div className="rounded-full bg-primary/10 p-3 text-primary">{icon}</div>
        </div>
      </CardContent>
    </Card>
  )
}
