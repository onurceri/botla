import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getQueues, getStuckJobs, retryJob, deleteJob } from '@/api/admin'
import { RefreshCw, Trash2, Clock, AlertCircle, PlayCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/components/ui/toast'
import { cn } from '@/lib/utils'

/**
 * AdminQueuesPage - Monitor and manage background processing queues
 */
export function AdminQueuesPage() {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  // Stats for all queues
  const { data: queues, isLoading: isLoadingQueues, refetch: refetchQueues } = useQuery({
    queryKey: ['admin', 'queues'],
    queryFn: () => getQueues(),
    refetchInterval: 10000, // Frequent refresh for queues
  })

  // Stuck jobs list
  const { data: stuckJobs, isLoading: isLoadingStuck, refetch: refetchStuck } = useQuery({
    queryKey: ['admin', 'queues', 'stuck'],
    queryFn: () => getStuckJobs(),
    refetchInterval: 10000,
  })

  const retryMutation = useMutation({
    mutationFn: (id: string) => retryJob(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'queues'] })
      toast("Görev başarıyla bekleme kuyruğuna alındı.", "success")
    },
    onError: (error: any) => {
      toast("Görev yeniden başlatılamadı: " + (error.response?.data?.error || error.message), "error")
    }
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteJob(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'queues'] })
      toast("Görev kuyruktan başarıyla kaldırıldı.", "success")
    },
    onError: (error: any) => {
      toast("Görev silinemedi: " + (error.response?.data?.error || error.message), "error")
    }
  })

  const handleRefreshAll = () => {
    refetchQueues()
    refetchStuck()
  }

  const isRefreshing = isLoadingQueues || isLoadingStuck

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Kuyruk Yönetimi</h1>
          <p className="text-muted-foreground">
            Bileşenlerin işleme kuyruklarını ve takılmış görevleri izle.
          </p>
        </div>
        <Button 
          variant="outline" 
          size="sm" 
          onClick={handleRefreshAll} 
          disabled={isRefreshing}
        >
          <RefreshCw className={cn("w-4 h-4 mr-2", isRefreshing && "animate-spin")} />
          Yenile
        </Button>
      </div>

      {/* Queue Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {isLoadingQueues && !queues && Array.from({ length: 3 }).map((_, i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="pb-2">
              <div className="h-5 bg-muted rounded w-1/2 mb-1" />
              <div className="h-4 bg-muted rounded w-1/3" />
            </CardHeader>
            <CardContent>
              <div className="space-y-2 mt-4">
                <div className="h-4 bg-muted rounded w-full" />
                <div className="h-4 bg-muted rounded w-full" />
                <div className="h-4 bg-muted rounded w-full" />
              </div>
            </CardContent>
          </Card>
        ))}
        {queues?.map(q => (
          <Card key={q.queue_name}>
            <CardHeader className="pb-2">
              <CardTitle className="text-base capitalize">{q.queue_name.replace(/_/g, ' ')}</CardTitle>
              <CardDescription className="text-xs">Aktif iş kuyruğu durumu</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="mt-4 space-y-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <Clock className="w-3.5 h-3.5" /> Bekleyen
                  </span>
                  <span className="font-semibold">{q.pending_count}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <Loader2 className="w-3.5 h-3.5" /> İşlenen
                  </span>
                  <span className="font-semibold text-blue-500">{q.processing_count}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <AlertCircle className="w-3.5 h-3.5" /> Hatalı
                  </span>
                  <span className={cn("font-semibold", 
                    q.failed_count > 0 ? "text-destructive" : "text-muted-foreground"
                  )}>
                    {q.failed_count}
                  </span>
                </div>
                {q.oldest_pending && (
                  <div className="mt-4 pt-3 border-t text-[10px] text-muted-foreground flex justify-between">
                    <span>En Eski Bekleyen:</span>
                    <span>{new Date(q.oldest_pending).toLocaleString()}</span>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Stuck Jobs Section */}
      <Card className="overflow-hidden">
        <CardHeader className="border-b bg-muted/30">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Takılmış Görevler</CardTitle>
              <CardDescription>30 dakikadan uzun süredir işlenen görevler</CardDescription>
            </div>
            <Badge variant={(stuckJobs?.length ?? 0) > 0 ? "destructive" : "outline"}>
              {stuckJobs?.length || 0} Görev
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {isLoadingStuck && !stuckJobs ? (
            <div className="flex flex-col items-center justify-center py-12 gap-2">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Görevler yükleniyor...</p>
            </div>
          ) : !stuckJobs?.length ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="w-12 h-12 rounded-full bg-green-500/10 flex items-center justify-center mb-3">
                <PlayCircle className="w-6 h-6 text-green-500" />
              </div>
              <h3 className="font-medium">Takılmış görev yok</h3>
              <p className="text-sm text-muted-foreground px-6 mt-1">
                Şu anda sistemde normal süresini aşmış herhangi bir işlem bulunmuyor.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-left border-collapse">
                <thead className="bg-muted/50 text-muted-foreground sticky top-0">
                  <tr>
                    <th className="p-4 font-medium border-b w-1/4">Kuyruk / ID</th>
                    <th className="p-4 font-medium border-b w-1/6">Durum</th>
                    <th className="p-4 font-medium border-b w-1/6">Süre</th>
                    <th className="p-4 font-medium border-b">Son Hata</th>
                    <th className="p-4 font-medium border-b text-right w-24">Aksiyonlar</th>
                  </tr>
                </thead>
                <tbody>
                  {stuckJobs.map(job => (
                    <tr key={job.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="p-4">
                        <div className="flex flex-col gap-0.5">
                          <span className="font-medium capitalize">{job.queue_name.replace(/_/g, ' ')}</span>
                          <span className="text-[10px] font-mono text-muted-foreground truncate max-w-[150px]">
                            {job.id}
                          </span>
                        </div>
                      </td>
                      <td className="p-4">
                        <Badge variant="outline" className="capitalize text-[10px] font-normal">
                          {job.status}
                        </Badge>
                      </td>
                      <td className="p-4">
                        <span className="font-mono text-destructive">{job.stuck_duration}</span>
                      </td>
                      <td className="p-4">
                        <p className="text-xs text-destructive line-clamp-2 max-w-sm" title={job.error_message}>
                          {job.error_message || "Hata detayı yok"}
                        </p>
                      </td>
                      <td className="p-4 text-right">
                        <div className="flex justify-end gap-1">
                          <Button
                            size="icon"
                            variant="ghost"
                            className="h-8 w-8 text-blue-500 hover:text-blue-600 hover:bg-blue-500/10"
                            onClick={() => retryMutation.mutate(job.id)}
                            disabled={retryMutation.isPending}
                            title="Tekrar Dene"
                          >
                            <RefreshCw className={cn("w-4 h-4", retryMutation.isPending && "animate-spin")} />
                          </Button>
                          <Button
                            size="icon"
                            variant="ghost"
                            className="h-8 w-8 text-destructive hover:bg-destructive/10"
                            onClick={() => deleteMutation.mutate(job.id)}
                            disabled={deleteMutation.isPending}
                            title="Sil"
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
