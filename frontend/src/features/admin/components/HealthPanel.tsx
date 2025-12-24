import { useQuery } from '@tanstack/react-query'
import { getDetailedHealth } from '@/api/admin'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Activity, CheckCircle2, AlertCircle, XCircle } from 'lucide-react'

function StatusBadge({ status }: { status: string }) {
  switch (status.toLowerCase()) {
    case 'ok':
    case 'healthy':
      return (
        <Badge variant="outline" className="bg-green-500/10 text-green-500 border-green-500/20">
          <CheckCircle2 className="w-3 h-3 mr-1" />
          HEALTHY
        </Badge>
      )
    case 'degraded':
      return (
        <Badge variant="outline" className="bg-yellow-500/10 text-yellow-500 border-yellow-500/20">
          <AlertCircle className="w-3 h-3 mr-1" />
          DEGRADED
        </Badge>
      )
    case 'down':
    case 'unhealthy':
      return (
        <Badge variant="destructive">
          <XCircle className="w-3 h-3 mr-1" />
          UNHEALTHY
        </Badge>
      )
    default:
      return (
        <Badge variant="secondary">
          <Activity className="w-3 h-3 mr-1" />
          {status.toUpperCase()}
        </Badge>
      )
  }
}

export function HealthPanel() {
  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'health'],
    queryFn: () => getDetailedHealth(),
    refetchInterval: 30000, // Refresh every 30s
  })

  if (isLoading) {
    return (
      <Card className="animate-pulse">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-base font-semibold">System Health</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4 mt-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex justify-between items-center">
                <div className="h-4 bg-muted rounded w-1/4" />
                <div className="h-6 bg-muted rounded w-20" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card data-testid="health-panel">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-semibold">System Health</CardTitle>
        <StatusBadge status={data?.status || 'unknown'} />
      </CardHeader>
      <CardContent>
        <div className="space-y-4 mt-4">
          {data?.dependencies.map((dep) => (
            <div key={dep.name} className="flex items-center justify-between">
              <div className="flex flex-col">
                <span className="text-sm font-medium capitalize">{dep.name}</span>
                {dep.message && (
                  <span className="text-xs text-muted-foreground line-clamp-1">{dep.message}</span>
                )}
              </div>
              <div className="flex items-center gap-3">
                <span className="text-xs text-muted-foreground">{dep.latency_ms}ms</span>
                <StatusBadge status={dep.status} />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
