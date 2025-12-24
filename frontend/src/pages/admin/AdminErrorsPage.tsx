import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getErrors, getErrorStats } from '@/api/admin'
import { 
  AlertTriangle, 
  AlertCircle, 
  Info, 
  Search, 
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Loader2
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

/**
 * AdminErrorsPage - Monitor system and application errors
 */
export function AdminErrorsPage() {
  const [severityFilter, setSeverityFilter] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [selectedError, setSelectedError] = useState<any | null>(null)
  const limit = 20

  const { data: statsData } = useQuery({
    queryKey: ['admin', 'errors', 'stats'],
    queryFn: () => getErrorStats(),
    refetchInterval: 30000,
  })
  const stats = statsData || { critical: 0, error: 0, warning: 0, info: 0 }
  const totalErrorCount = Number(stats.critical || 0) + Number(stats.error || 0)
  const warningCount = Number(stats.warning || 0)
  const infoCount = Number(stats.info || 0)

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['admin', 'errors', severityFilter, page],
    queryFn: () => getErrors(severityFilter === 'all' ? undefined : severityFilter, (page - 1) * limit, limit),
  })

  const handleRefresh = () => {
    refetch()
  }

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
      case 'error':
        return <AlertCircle className="w-4 h-4 text-destructive" />
      case 'warning':
        return <AlertTriangle className="w-4 h-4 text-amber-500" />
      default:
        return <Info className="w-4 h-4 text-blue-500" />
    }
  }

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <Badge variant="destructive" className="bg-red-800">Kritik</Badge>
      case 'error':
        return <Badge variant="destructive">Hata</Badge>
      case 'warning':
        return <Badge variant="outline" className="text-amber-600 border-amber-600">Uyarı</Badge>
      default:
        return <Badge variant="secondary">Bilgi</Badge>
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Hata Kayıtları</h1>
          <p className="text-muted-foreground">
            Sistem genelinde oluşan hata ve uyarıları izle.
          </p>
        </div>
        <Button 
          variant="outline" 
          size="sm" 
          onClick={handleRefresh} 
          disabled={isFetching}
        >
          <RefreshCw className={cn("w-4 h-4 mr-2", isFetching && "animate-spin")} />
          Yenile
        </Button>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-red-500/5 border-red-500/20">
          <CardContent className="pt-4">
            <div className="text-sm font-medium text-red-600 dark:text-red-400">Kritik/Hata (24s)</div>
            <div className="text-2xl font-bold mt-1">
              {totalErrorCount}
            </div>
          </CardContent>
        </Card>
        <Card className="bg-amber-500/5 border-amber-500/20">
          <CardContent className="pt-4">
            <div className="text-sm font-medium text-amber-600 dark:text-amber-400">Uyarı (24s)</div>
            <div className="text-2xl font-bold mt-1">{warningCount}</div>
          </CardContent>
        </Card>
        <Card className="bg-blue-500/5 border-blue-500/20">
          <CardContent className="pt-4">
            <div className="text-sm font-medium text-blue-600 dark:text-blue-400">Bilgi (24s)</div>
            <div className="text-2xl font-bold mt-1">{infoCount}</div>
          </CardContent>
        </Card>
      </div>

      {/* Filters & Table */}
      <Card>
        <CardHeader className="pb-3 border-b">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2 flex-1 max-w-sm">
              <Select value={severityFilter} onValueChange={(val) => { setSeverityFilter(val); setPage(1); }}>
                <SelectTrigger className="w-[140px]" data-testid="severity-select">
                  <SelectValue placeholder="Önem Derecesi" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Hepsi</SelectItem>
                  <SelectItem value="critical">Kritik</SelectItem>
                  <SelectItem value="error">Hata</SelectItem>
                  <SelectItem value="warning">Uyarı</SelectItem>
                  <SelectItem value="info">Bilgi</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="flex items-center gap-2 text-sm text-muted-foreground font-medium">
              Toplam: {data?.total || 0} kayıt
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-20 gap-2">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Loglar yükleniyor...</p>
            </div>
          ) : !data?.data?.length ? (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-3">
                <Search className="w-6 h-6 text-muted-foreground" />
              </div>
              <h3 className="font-medium">Kayıt bulunamadı</h3>
              <p className="text-sm text-muted-foreground px-6 mt-1">
                Seçilen filtrelere uygun herhangi bir hata kaydı bulunmuyor.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-left border-collapse">
                <thead className="bg-muted/50 text-muted-foreground sticky top-0">
                  <tr>
                    <th className="p-4 font-medium border-b w-10"></th>
                    <th className="p-4 font-medium border-b w-[100px]">Önem</th>
                    <th className="p-4 font-medium border-b w-[100px]">Tür</th>
                    <th className="p-4 font-medium border-b min-w-[300px]">Mesaj</th>
                    <th className="p-4 font-medium border-b w-[160px]">Tarih</th>
                    <th className="p-4 font-medium border-b text-right w-16">Detay</th>
                  </tr>
                </thead>
                <tbody>
                  {data.data.map((log: any) => (
                    <tr 
                      key={log.id} 
                      className="border-b last:border-0 hover:bg-muted/30 transition-colors cursor-pointer"
                      onClick={() => setSelectedError(log)}
                    >
                      <td className="p-4 text-center">
                        {getSeverityIcon(log.severity)}
                      </td>
                      <td className="p-4">
                        {getSeverityBadge(log.severity)}
                      </td>
                      <td className="p-4">
                        <span className="text-[10px] uppercase font-bold tracking-wider text-muted-foreground">
                          {log.error_type}
                        </span>
                      </td>
                      <td className="p-4">
                        <div className="flex flex-col gap-0.5">
                          <span className="font-medium truncate max-w-md">{log.message}</span>
                          <span className="text-[10px] text-muted-foreground font-mono">
                            {log.request_method || ''} {log.request_path || ''}
                          </span>
                        </div>
                      </td>
                      <td className="p-4 text-muted-foreground whitespace-nowrap">
                        {new Date(log.created_at).toLocaleString('tr-TR')}
                      </td>
                      <td className="p-4 text-right">
                        <Button size="icon" variant="ghost" className="h-8 w-8">
                          <ExternalLink className="w-4 h-4" />
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
        {data && data.total > limit && (
          <div className="p-4 border-t flex items-center justify-between">
            <p className="text-xs text-muted-foreground">
              {limit * (page - 1) + 1} - {Math.min(limit * page, data.total)} / {data.total} gösteriliyor
            </p>
            <div className="flex items-center gap-2">
              <Button 
                variant="outline" 
                size="icon" 
                className="h-8 w-8"
                disabled={page === 1}
                onClick={() => setPage(p => p - 1)}
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <span className="text-xs font-medium px-2">Sayfa {page}</span>
              <Button 
                variant="outline" 
                size="icon" 
                className="h-8 w-8"
                disabled={page * limit >= data.total}
                onClick={() => setPage(p => p + 1)}
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Error Detail Dialog */}
      <Dialog open={!!selectedError} onOpenChange={(open) => !open && setSelectedError(null)}>
        <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <div className="flex items-center gap-2 mb-2">
              {selectedError && getSeverityBadge(selectedError.severity)}
              <span className="text-xs text-muted-foreground font-mono">{selectedError?.id}</span>
            </div>
            <DialogTitle className="text-xl break-words pr-6">
              {selectedError?.message}
            </DialogTitle>
            <DialogDescription>
              {selectedError && new Date(selectedError.created_at).toLocaleString('tr-TR', { dateStyle: 'full', timeStyle: 'medium' })}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6 mt-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div className="space-y-1">
                <p className="text-muted-foreground">İstek (Request)</p>
                <p className="font-medium">{selectedError?.request_method || 'N/A'} {selectedError?.request_path || ''}</p>
              </div>
              <div className="space-y-1">
                <p className="text-muted-foreground">Tür (Type)</p>
                <p className="font-medium capitalize">{selectedError?.error_type?.replace(/_/g, ' ') || 'N/A'}</p>
              </div>
              {selectedError?.user_id && (
                <div className="space-y-1">
                  <p className="text-muted-foreground">Kullanıcı ID</p>
                  <p className="font-mono text-xs">{selectedError.user_id}</p>
                </div>
              )}
              {selectedError?.chatbot_id && (
                <div className="space-y-1">
                  <p className="text-muted-foreground">Chatbot ID</p>
                  <p className="font-mono text-xs">{selectedError.chatbot_id}</p>
                </div>
              )}
            </div>

            {selectedError?.stack_trace && (
              <div className="space-y-2">
                <p className="text-sm font-medium">Stack Trace</p>
                <div className="bg-muted p-4 rounded-md overflow-x-auto border">
                  <pre className="text-[10px] leading-relaxed font-mono">
                    {selectedError.stack_trace}
                  </pre>
                </div>
              </div>
            )}

            {selectedError?.context && (
              <div className="space-y-2">
                <p className="text-sm font-medium">Ek Bağlam (Context)</p>
                <div className="bg-muted p-4 rounded-md border">
                  <pre className="text-[10px] font-mono whitespace-pre-wrap">
                    {typeof selectedError.context === 'string' 
                      ? selectedError.context 
                      : JSON.stringify(selectedError.context, null, 2)}
                  </pre>
                </div>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
