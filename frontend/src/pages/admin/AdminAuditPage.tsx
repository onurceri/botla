import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import * as adminApi from '@/api/admin'
import { 
  Activity, 
  Calendar, 
  ChevronLeft, 
  ChevronRight, 
  Eye, 
  Loader2, 
  RefreshCw, 
  User 
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

/**
 * AdminAuditPage - Monitor administrative actions
 */
export function AdminAuditPage() {
  const [page, setPage] = useState(1)
  const [selectedLog, setSelectedLog] = useState<any | null>(null)
  const limit = 20

  const { data: qData, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['admin', 'audit-logs', page],
    queryFn: () => adminApi.listAuditLogs({ offset: (page - 1) * limit, limit }),
  })

  // Ensure data structure is safe
  const data = qData || { data: [], total: 0 }

  const handleRefresh = () => {
    refetch()
  }

  const getActionColor = (action: string) => {
    const a = action || ''
    if (a.includes('delete')) return 'text-destructive'
    if (a.includes('create')) return 'text-green-600 dark:text-green-400'
    if (a.includes('update')) return 'text-blue-600 dark:text-blue-400'
    return 'text-muted-foreground'
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Denetim Günlüğü</h1>
          <p className="text-muted-foreground">
            Admin işlemleri ve sistem değişikliklerini takip et.
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

      <Card>
        <CardHeader className="pb-3 border-b">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">İşlem Kayıtları</CardTitle>
            <div className="text-xs text-muted-foreground">
              Toplam: {data.total || 0} kayıt
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-20 gap-2">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Audit kayıtları yükleniyor...</p>
            </div>
          ) : !data.data?.length ? (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-3">
                <Activity className="w-6 h-6 text-muted-foreground" />
              </div>
              <h3 className="font-medium">Kayıt bulunamadı</h3>
              <p className="text-sm text-muted-foreground mt-1">
                Henüz herhangi bir admin işlemi kaydedilmemiş.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-left border-collapse">
                <thead className="bg-muted/50 text-muted-foreground sticky top-0">
                  <tr>
                    <th className="p-4 font-medium border-b w-[180px]">Admin</th>
                    <th className="p-4 font-medium border-b w-[150px]">İşlem</th>
                    <th className="p-4 font-medium border-b w-[120px]">Hedef</th>
                    <th className="p-4 font-medium border-b">Detaylar</th>
                    <th className="p-4 font-medium border-b w-[160px]">Tarih</th>
                    <th className="p-4 font-medium border-b text-right w-16">İncele</th>
                  </tr>
                </thead>
                <tbody>
                  {data.data.map((log: any) => (
                    <tr 
                      key={log.id} 
                      className="border-b last:border-0 hover:bg-muted/30 transition-colors"
                    >
                      <td className="p-4">
                        <div className="flex items-center gap-2">
                          <User className="w-4 h-4 text-muted-foreground" />
                          <span className="font-mono text-xs truncate max-w-[120px]" title={log.admin_user_id}>
                            {log.admin_user_id}
                          </span>
                        </div>
                      </td>
                      <td className="p-4">
                        <span className={cn("font-medium uppercase text-[10px] tracking-wider", getActionColor(log.action))}>
                          {(log.action || '').replace(/_/g, ' ')}
                        </span>
                      </td>
                      <td className="p-4">
                        <Badge variant="outline" className="capitalize">
                          {log.target_type}
                        </Badge>
                      </td>
                      <td className="p-4">
                        <p className="text-muted-foreground truncate max-w-[200px]">
                          {JSON.stringify(log.details)}
                        </p>
                      </td>
                      <td className="p-4 text-muted-foreground whitespace-nowrap">
                        <div className="flex items-center gap-1.5">
                          <Calendar className="w-3 h-3" />
                          {log.created_at ? new Date(log.created_at).toLocaleString('tr-TR') : 'N/A'}
                        </div>
                      </td>
                      <td className="p-4 text-right">
                        <Button 
                          size="icon" 
                          variant="ghost" 
                          className="h-8 w-8"
                          onClick={() => setSelectedLog(log)}
                          aria-label="Detayları İncele"
                          data-testid="audit-detail-button"
                        >
                          <Eye className="w-4 h-4" />
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

      {/* Audit Detail Dialog */}
      <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Audit Kaydı Detayları</DialogTitle>
          </DialogHeader>

          <div className="space-y-6 mt-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div className="space-y-1">
                <p className="text-muted-foreground">İşlem Yapan Admin</p>
                <p className="font-mono text-xs truncate" title={selectedLog?.admin_user_id}>{selectedLog?.admin_user_id}</p>
              </div>
              <div className="space-y-1">
                <p className="text-muted-foreground">İşlem Türü</p>
                <p className="font-medium uppercase text-xs">{selectedLog?.action}</p>
              </div>
              <div className="space-y-1">
                <p className="text-muted-foreground">Hedef Türü</p>
                <p className="font-medium capitalize">{selectedLog?.target_type}</p>
              </div>
              <div className="space-y-1">
                <p className="text-muted-foreground">Hedef ID</p>
                <p className="font-mono text-xs">{selectedLog?.target_id || 'N/A'}</p>
              </div>
              <div className="space-y-1">
                <p className="text-muted-foreground">IP Adresi</p>
                <p className="font-mono text-xs">{selectedLog?.ip_address}</p>
              </div>
              <div className="space-y-1">
                <p className="text-muted-foreground">Tarih</p>
                <p className="font-medium">
                  {selectedLog?.created_at ? new Date(selectedLog.created_at).toLocaleString('tr-TR') : 'N/A'}
                </p>
              </div>
            </div>

            <div className="space-y-2">
              <p className="text-sm font-medium">Değişiklik Detayları (JSON)</p>
              <div className="bg-muted p-4 rounded-md overflow-x-auto border">
                <pre className="text-xs font-mono whitespace-pre-wrap">
                  {JSON.stringify(selectedLog?.details, null, 2)}
                </pre>
              </div>
            </div>

            <div className="space-y-2">
              <p className="text-sm font-medium">User Agent</p>
              <p className="text-xs text-muted-foreground bg-muted p-2 rounded border font-mono italic break-words">
                {selectedLog?.user_agent}
              </p>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
