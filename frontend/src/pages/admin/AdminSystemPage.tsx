/**
 * AdminSystemPage - Dedicated page for monitoring all system dependencies
 */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getDetailedHealth } from '@/api/admin'
import { RefreshCw, Activity, Server, Clock, Code2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { cn } from '@/lib/utils'

export function AdminSystemPage() {
  const queryClient = useQueryClient()
  
  const { data: health, isLoading, isFetching } = useQuery({
    queryKey: ['admin', 'health'],
    queryFn: () => getDetailedHealth(false),
    refetchInterval: 30 * 60 * 1000, // Refresh every 30 minutes to match cache TTL, or keep shorter if we want to catch expiration sooner. Let's stick to 30s as before for auto-updates if cache expires? keeping 30s is fine, backend handles caching.
    // Actually, if backend caches for 30m, polling every 30s just hits Redis. That's fine.
  })

  const refreshMutation = useMutation({
    mutationFn: () => getDetailedHealth(true),
    onSuccess: (data) => {
      queryClient.setQueryData(['admin', 'health'], data)
    },
  })

  // Helper to force the display label for status badge if needed, 
  // though StatusBadge usually displays pre-defined labels.
  // We can just rely on StatusBadge's internal logic which maps healthy -> Active, etc.
  // Or we might want to keep exact labels (Healthy, Degraded, Down).
  // StatusBadge is opinionated about labels (Aktif, Bekliyor, Başarısız).
  // This is probably fine for "modern consistent" look.

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Sistem Durumu</h1>
          <p className="text-muted-foreground">
            Sistem bileşenlerinin sağlık durumunu ve performansını izle.
          </p>
        </div>
        <div className="flex items-center gap-4">
          {health?.last_updated && (
            <span className="text-sm text-muted-foreground">
              Son güncelleme: {new Date(health.last_updated).toLocaleTimeString()}
            </span>
          )}
          <Button 
            variant="outline" 
            size="sm" 
            onClick={() => refreshMutation.mutate()} 
            disabled={refreshMutation.isPending || isFetching}
          >
            <RefreshCw className={cn("w-4 h-4 mr-2", (refreshMutation.isPending || isFetching) && "animate-spin")} />
            Yenile
          </Button>
        </div>
      </div>

      {/* Overall Status Card */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
            <div className="flex items-center gap-4">
              <div className={cn("w-12 h-12 rounded-full flex items-center justify-center", 
                health?.status === 'healthy' ? 'bg-green-500/10' : 
                health?.status === 'degraded' ? 'bg-amber-500/10' : 'bg-red-500/10'
              )}>
                <Activity className={cn("w-6 h-6", 
                  health?.status === 'healthy' ? 'text-green-600' : 
                  health?.status === 'degraded' ? 'text-amber-600' : 'text-red-600'
                )} />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="text-lg font-semibold capitalize">
                    {health?.status || (isLoading ? 'Yükleniyor...' : 'Bilinmiyor')}
                  </h3>
                  {health?.status && (
                    <div className={cn("w-2.5 h-2.5 rounded-full animate-pulse", 
                      health.status === 'healthy' ? 'bg-green-500' : 
                      health.status === 'degraded' ? 'bg-amber-500' : 'bg-red-500'
                    )} />
                  )}
                </div>
                <p className="text-sm text-muted-foreground">Genel sistem durumu</p>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-8">
              <div className="space-y-1">
                <div className="flex items-center text-muted-foreground text-xs uppercase tracking-wider font-medium">
                  <Code2 className="w-3.5 h-3.5 mr-1" /> Versiyon
                </div>
                <p className="text-sm font-mono">{health?.version || '-'}</p>
              </div>
              <div className="space-y-1">
                <div className="flex items-center text-muted-foreground text-xs uppercase tracking-wider font-medium">
                  <Clock className="w-3.5 h-3.5 mr-1" /> Çalışma Süresi
                </div>
                <p className="text-sm">{health?.uptime || '-'}</p>
              </div>
              <div className="space-y-1">
                <div className="flex items-center text-muted-foreground text-xs uppercase tracking-wider font-medium">
                  <Server className="w-3.5 h-3.5 mr-1" /> Ortam
                </div>
                <p className="text-sm capitalize">{health?.environment || '-'}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Dependencies Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {health?.dependencies.map(dep => (
          <Card key={dep.name} className="overflow-hidden hover:border-primary/50 transition-colors">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base capitalize font-semibold">{dep.name}</CardTitle>
                <StatusBadge status={dep.status} size="xs" />
              </div>
              <CardDescription className="text-xs">
                Son kontrol: {dep.checked_at ? new Date(dep.checked_at).toLocaleTimeString() : '-'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between mt-2">
                <span className="text-sm text-muted-foreground">Gecikme</span>
                <span className={cn("text-sm font-medium font-mono", 
                  dep.latency_ms > 500 ? 'text-amber-600' : 'text-green-600'
                )}>
                  {dep.latency_ms}ms
                </span>
              </div>
              {dep.message && (
                <div className="mt-4 p-2 rounded bg-destructive/10 text-destructive text-xs break-words border border-destructive/20 font-medium">
                  {dep.message}
                </div>
              )}
            </CardContent>
          </Card>
        ))}
        {isLoading && Array.from({ length: 3 }).map((_, i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader>
              <div className="h-5 bg-muted rounded w-1/3 mb-2" />
              <div className="h-4 bg-muted rounded w-1/2" />
            </CardHeader>
            <CardContent>
              <div className="h-4 bg-muted rounded w-full mt-2" />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
