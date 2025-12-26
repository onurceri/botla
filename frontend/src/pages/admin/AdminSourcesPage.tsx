/**
 * AdminSourcesPage - Data sources management page
 * Lists all data sources and their processing status
 */
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { FileText, Globe, RefreshCw, MoreHorizontal, CheckCircle2, Clock, Loader2, AlertCircle } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { tr } from 'date-fns/locale'
import * as adminApi from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/DropdownMenu'
import { useToast } from '@/components/ui/toast'
import { StatusBadge } from '@/components/ui/status-badge'

export function AdminSourcesPage() {
  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [offset, setOffset] = useState(0)
  const limit = 20

  const queryClient = useQueryClient()
  const { toast } = useToast()

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin', 'sources', { statusFilter, typeFilter, offset, limit }],
    queryFn: () =>
      adminApi.listSources({
        status: statusFilter || undefined,
        source_type: typeFilter || undefined,
        limit,
        offset,
      }),
  })

  const { data: stats } = useQuery({
    queryKey: ['admin', 'sources', 'stats'],
    queryFn: adminApi.getSourceStats,
  })

  const reprocessMutation = useMutation({
    mutationFn: adminApi.reprocessSource,
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'sources'] })
      toast(
        result.queued
          ? 'Kaynak işlenmek üzere kuyruğa eklendi.'
          : 'Kaynak sıfırlandı ancak kuyruk kullanılamıyor.',
        'success'
      )
    },
    onError: () => {
      toast('Kaynak yeniden işlenirken bir hata oluştu.', 'error')
    },
  })

  const sources = data?.sources ?? []
  const total = data?.total ?? 0
  const hasNextPage = offset + limit < total
  const hasPrevPage = offset > 0

  const getSourceIcon = (type: string) => {
    switch (type) {
      case 'url':
        return <Globe className="w-4 h-4 text-primary" />
      default:
        return <FileText className="w-4 h-4 text-primary" />
    }
  }

  const getTypeBadgeClass = (type: string) => {
    switch (type) {
      case 'url':
        return 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20'
      case 'pdf':
        return 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20'
      case 'file':
        return 'bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-500/20'
      default:
        return 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20'
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Kaynaklar</h1>
        <p className="text-muted-foreground">
          Veri kaynaklarını ve işleme durumlarını görüntüle. Toplam: {total}
        </p>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Card className="bg-green-500/5 border-green-500/20">
            <CardContent className="pt-4">
              <div className="flex items-center gap-2 mb-2">
                <CheckCircle2 className="w-4 h-4 text-green-500" />
                <span className="text-sm text-muted-foreground">Tamamlandı</span>
              </div>
              <p className="text-2xl font-bold">{stats.completed || stats.ready || 0}</p>
            </CardContent>
          </Card>
          <Card className="bg-blue-500/5 border-blue-500/20">
            <CardContent className="pt-4">
              <div className="flex items-center gap-2 mb-2">
                <Loader2 className="w-4 h-4 text-blue-500" />
                <span className="text-sm text-muted-foreground">İşleniyor</span>
              </div>
              <p className="text-2xl font-bold">{stats.processing || 0}</p>
            </CardContent>
          </Card>
          <Card className="bg-amber-500/5 border-amber-500/20">
            <CardContent className="pt-4">
              <div className="flex items-center gap-2 mb-2">
                <Clock className="w-4 h-4 text-amber-500" />
                <span className="text-sm text-muted-foreground">Bekliyor</span>
              </div>
              <p className="text-2xl font-bold">{stats.pending || 0}</p>
            </CardContent>
          </Card>
          <Card className="bg-red-500/5 border-red-500/20">
            <CardContent className="pt-4">
              <div className="flex items-center gap-2 mb-2">
                <AlertCircle className="w-4 h-4 text-red-500" />
                <span className="text-sm text-muted-foreground">Başarısız</span>
              </div>
              <p className="text-2xl font-bold">{stats.failed || 0}</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <select
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value)
            setOffset(0)
          }}
          className="px-4 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
        >
          <option value="">Tüm Durumlar</option>
          <option value="completed">Tamamlandı</option>
          <option value="ready">Hazır</option>
          <option value="processing">İşleniyor</option>
          <option value="pending">Bekliyor</option>
          <option value="failed">Başarısız</option>
        </select>
        <select
          value={typeFilter}
          onChange={(e) => {
            setTypeFilter(e.target.value)
            setOffset(0)
          }}
          className="px-4 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
        >
          <option value="">Tüm Tipler</option>
          <option value="url">URL</option>
          <option value="file">Dosya</option>
          <option value="pdf">PDF</option>
          <option value="text">Metin</option>
        </select>
      </div>

      {/* Sources Table */}
      <Card>
        <CardHeader className="pb-3 border-b">
          <CardTitle className="text-sm font-medium">Kaynak Listesi</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-8 text-center text-muted-foreground">Yükleniyor...</div>
          ) : error ? (
            <div className="p-8 text-center text-destructive">Hata: {(error as Error).message}</div>
          ) : sources.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground">Kaynak bulunamadı.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-muted/50 text-muted-foreground">
                  <tr className="text-left">
                    <th className="px-4 py-3 font-medium">Kaynak</th>
                    <th className="px-4 py-3 font-medium">Chatbot</th>
                    <th className="px-4 py-3 font-medium">Tip</th>
                    <th className="px-4 py-3 font-medium">Durum</th>
                    <th className="px-4 py-3 font-medium">Chunk</th>
                    <th className="px-4 py-3 font-medium">Oluşturulma</th>
                    <th className="px-4 py-3 font-medium text-right">İşlemler</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {sources.map((source) => (
                    <tr key={source.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
                            {getSourceIcon(source.source_type)}
                          </div>
                          <div className="max-w-xs truncate">
                            <span className="font-medium text-sm">
                              {source.original_filename || source.source_url || source.id}
                            </span>
                            {source.error_message && (
                              <p className="text-xs text-destructive truncate">{source.error_message}</p>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {source.chatbot_name}
                      </td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-1 text-[10px] rounded-full border font-semibold uppercase ${getTypeBadgeClass(source.source_type)}`}>
                          {source.source_type}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <StatusBadge status={source.status} size="sm" />
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {source.chunk_count}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {formatDistanceToNow(new Date(source.created_at), {
                          addSuffix: true,
                          locale: tr,
                        })}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreHorizontal className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => reprocessMutation.mutate(source.id)}
                              disabled={reprocessMutation.isPending}
                            >
                              <RefreshCw
                                className={`w-4 h-4 mr-2 ${
                                  reprocessMutation.isPending ? 'animate-spin' : ''
                                }`}
                              />
                              Yeniden İşle
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
        {/* Pagination */}
        {total > limit && (
          <div className="p-4 border-t flex items-center justify-between">
            <span className="text-xs text-muted-foreground">
              {offset + 1} - {Math.min(offset + limit, total)} / {total} kaynak
            </span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setOffset(Math.max(0, offset - limit))}
                disabled={!hasPrevPage}
              >
                Önceki
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setOffset(offset + limit)}
                disabled={!hasNextPage}
              >
                Sonraki
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}
