import { useQuery } from '@tanstack/react-query'
import { getDetailedHealth } from '@/api/admin'
import { RefreshCw, Activity, Server, Clock, Code2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

/**
 * AdminSystemPage - Dedicated page for monitoring all system dependencies
 */
export function AdminSystemPage() {
  const { data: health, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['admin', 'health'],
    queryFn: () => getDetailedHealth(),
    refetchInterval: 30000, // Auto-refresh every 30s
  })

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'healthy':
      case 'ok':
        return 'bg-green-500/10 text-green-500 border-green-500/20'
      case 'degraded':
        return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20'
      case 'unhealthy':
      case 'down':
        return 'bg-red-500/10 text-red-500 border-red-500/20'
      default:
        return 'bg-gray-500/10 text-gray-500 border-gray-500/20'
    }
  }

  const getStatusDotColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'healthy':
      case 'ok':
        return 'bg-green-500'
      case 'degraded':
        return 'bg-yellow-500'
      case 'unhealthy':
      case 'down':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Sistem Durumu</h1>
          <p className="text-muted-foreground">
            Sistem bileşenlerinin sağlık durumunu ve performansını izle.
          </p>
        </div>
        <Button 
          variant="outline" 
          size="sm" 
          onClick={() => refetch()} 
          disabled={isFetching}
        >
          <RefreshCw className={cn("w-4 h-4 mr-2", isFetching && "animate-spin")} />
          Yenile
        </Button>
      </div>

      {/* Overall Status Card */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
            <div className="flex items-center gap-4">
              <div className={cn("w-12 h-12 rounded-full flex items-center justify-center", 
                health?.status === 'healthy' ? 'bg-green-500/10' : 
                health?.status === 'degraded' ? 'bg-yellow-500/10' : 'bg-red-500/10'
              )}>
                <Activity className={cn("w-6 h-6", 
                  health?.status === 'healthy' ? 'text-green-500' : 
                  health?.status === 'degraded' ? 'text-yellow-500' : 'text-red-500'
                )} />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="text-lg font-semibold capitalize">
                    {health?.status || (isLoading ? 'Yükleniyor...' : 'Bilinmiyor')}
                  </h3>
                  {health?.status && (
                    <div className={cn("w-2 h-2 rounded-full animate-pulse", getStatusDotColor(health.status))} />
                  )}
                </div>
                <p className="text-sm text-muted-foreground">Genel sistem durumu</p>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-8">
              <div className="space-y-1">
                <div className="flex items-center text-muted-foreground text-xs uppercase tracking-wider font-medium">
                  <Code2 className="w-3 h-3 mr-1" /> Versiyon
                </div>
                <p className="text-sm font-mono">{health?.version || '-'}</p>
              </div>
              <div className="space-y-1">
                <div className="flex items-center text-muted-foreground text-xs uppercase tracking-wider font-medium">
                  <Clock className="w-3 h-3 mr-1" /> Çalışma Süresi
                </div>
                <p className="text-sm">{health?.uptime || '-'}</p>
              </div>
              <div className="space-y-1">
                <div className="flex items-center text-muted-foreground text-xs uppercase tracking-wider font-medium">
                  <Server className="w-3 h-3 mr-1" /> Ortam
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
          <Card key={dep.name} className="overflow-hidden">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base capitalize">{dep.name}</CardTitle>
                <Badge variant="outline" className={cn("font-medium", getStatusColor(dep.status))}>
                  {dep.status.toUpperCase()}
                </Badge>
              </div>
              <CardDescription className="text-xs">
                Son kontrol: {dep.checked_at ? new Date(dep.checked_at).toLocaleTimeString() : '-'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between mt-2">
                <span className="text-sm text-muted-foreground">Gecikme</span>
                <span className={cn("text-sm font-medium", 
                  dep.latency_ms > 500 ? 'text-yellow-500' : 'text-green-500'
                )}>
                  {dep.latency_ms}ms
                </span>
              </div>
              {dep.message && (
                <div className="mt-4 p-2 rounded bg-destructive/10 text-destructive text-xs break-words border border-destructive/20">
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
