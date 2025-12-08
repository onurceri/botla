import { useState, useEffect, useCallback } from 'react'
import { CheckSquare, Square, ExternalLink, Trash2, CheckCircle, XCircle, RefreshCw, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import {
  listPendingURLs,
  approvePendingURLs,
  rejectPendingURLs,
  clearPendingURLs,
  PendingURL,
} from '@/api/source'
import { useToast } from '@/components/ui/toast'

interface PendingURLsPanelProps {
  chatbotId: string
  onSourcesCreated: () => void
}

export default function PendingURLsPanel({ chatbotId, onSourcesCreated }: PendingURLsPanelProps) {
  const [urls, setUrls] = useState<PendingURL[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [perPage] = useState(20)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)
  const { toast } = useToast()

  const fetchPendingURLs = useCallback(async () => {
    setLoading(true)
    try {
      const response = await listPendingURLs(chatbotId, page, perPage)
      setUrls(response.urls || [])
      setTotal(response.total || 0)
    } catch (error) {
      console.error('Failed to fetch pending URLs:', error)
      toast('URL listesi yüklenemedi.', 'error')
    } finally {
      setLoading(false)
    }
  }, [chatbotId, page, perPage, toast])

  useEffect(() => {
    fetchPendingURLs()
  }, [fetchPendingURLs])

  const toggleSelect = (id: string) => {
    setSelectedIds(prev => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const selectAll = () => {
    if (selectedIds.size === urls.length) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(urls.map(u => u.id)))
    }
  }

  const handleApprove = async () => {
    if (selectedIds.size === 0) return
    setActionLoading(true)
    try {
      const result = await approvePendingURLs(chatbotId, Array.from(selectedIds))
      toast(`${result.sources_created} kaynak oluşturuldu.`, 'success')
      setSelectedIds(new Set())
      await fetchPendingURLs()
      onSourcesCreated()
    } catch (error) {
      console.error('Failed to approve URLs:', error)
      toast('URL onaylama başarısız.', 'error')
    } finally {
      setActionLoading(false)
    }
  }

  const handleReject = async () => {
    if (selectedIds.size === 0) return
    setActionLoading(true)
    try {
      const result = await rejectPendingURLs(chatbotId, Array.from(selectedIds))
      toast(`${result.rejected_count} URL reddedildi.`, 'success')
      setSelectedIds(new Set())
      await fetchPendingURLs()
    } catch (error) {
      console.error('Failed to reject URLs:', error)
      toast('URL reddetme başarısız.', 'error')
    } finally {
      setActionLoading(false)
    }
  }

  const handleClearAll = async () => {
    if (!confirm('Tüm bekleyen URL\'leri silmek istediğinize emin misiniz?')) return
    setActionLoading(true)
    try {
      const result = await clearPendingURLs(chatbotId)
      toast(`${result.cleared_count} URL temizlendi.`, 'success')
      setSelectedIds(new Set())
      await fetchPendingURLs()
    } catch (error) {
      console.error('Failed to clear URLs:', error)
      toast('URL temizleme başarısız.', 'error')
    } finally {
      setActionLoading(false)
    }
  }

  const totalPages = Math.ceil(total / perPage)

  if (total === 0 && !loading) {
    return null // Don't show the panel if there are no pending URLs
  }

  return (
    <Card className="border-amber-200 bg-amber-50/50">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-amber-600" />
            <CardTitle className="text-lg text-amber-900">Keşfedilen URL'ler</CardTitle>
            <span className="text-sm text-amber-700 bg-amber-100 px-2 py-0.5 rounded-full font-medium">
              {total} adet
            </span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={fetchPendingURLs}
            disabled={loading}
            className="text-amber-700 hover:text-amber-900 hover:bg-amber-100"
          >
            <RefreshCw className={`w-4 h-4 mr-1 ${loading ? 'animate-spin' : ''}`} />
            Yenile
          </Button>
        </div>
        <CardDescription className="text-amber-700">
          Aşağıdaki URL'ler otomatik olarak keşfedildi. Hangilerini kaynak olarak eklemek istediğinizi seçin.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Toolbar */}
        <div className="flex flex-wrap items-center gap-2 pb-2 border-b border-amber-200">
          <Button
            variant="outline"
            size="sm"
            onClick={selectAll}
            className="border-amber-300 hover:bg-amber-100"
          >
            {selectedIds.size === urls.length && urls.length > 0 ? (
              <>
                <CheckSquare className="w-4 h-4 mr-1" />
                Seçimi Kaldır
              </>
            ) : (
              <>
                <Square className="w-4 h-4 mr-1" />
                Tümünü Seç
              </>
            )}
          </Button>
          
          <span className="text-sm text-amber-700 px-2">
            {selectedIds.size} seçili
          </span>

          <div className="flex-1" />

          <Button
            variant="outline"
            size="sm"
            onClick={handleReject}
            disabled={selectedIds.size === 0 || actionLoading}
            className="border-red-300 text-red-600 hover:bg-red-50 hover:text-red-700"
          >
            <XCircle className="w-4 h-4 mr-1" />
            Reddet
          </Button>
          
          <Button
            size="sm"
            onClick={handleApprove}
            disabled={selectedIds.size === 0 || actionLoading}
            className="bg-green-600 hover:bg-green-700 text-white"
          >
            <CheckCircle className="w-4 h-4 mr-1" />
            Onayla
          </Button>

          <Button
            variant="ghost"
            size="sm"
            onClick={handleClearAll}
            disabled={actionLoading}
            className="text-red-600 hover:text-red-700 hover:bg-red-50"
          >
            <Trash2 className="w-4 h-4 mr-1" />
            Tümünü Temizle
          </Button>
        </div>

        {/* URL List */}
        <div className="max-h-[300px] overflow-y-auto space-y-1 rounded-lg border border-amber-200 bg-white p-2">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="w-6 h-6 animate-spin text-amber-600" />
            </div>
          ) : urls.length === 0 ? (
            <div className="text-center py-8 text-amber-700">
              Bekleyen URL bulunmuyor.
            </div>
          ) : (
            urls.map((url) => (
              <div
                key={url.id}
                onClick={() => toggleSelect(url.id)}
                className={`
                  flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-colors
                  ${selectedIds.has(url.id) 
                    ? 'bg-green-50 border border-green-200' 
                    : 'hover:bg-amber-50 border border-transparent'
                  }
                `}
              >
                <div className="flex-shrink-0">
                  {selectedIds.has(url.id) ? (
                    <CheckSquare className="w-5 h-5 text-green-600" />
                  ) : (
                    <Square className="w-5 h-5 text-gray-400" />
                  )}
                </div>
                
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-gray-900 truncate">
                      {url.url}
                    </span>
                    <a
                      href={url.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      onClick={(e) => e.stopPropagation()}
                      className="text-blue-600 hover:text-blue-800 flex-shrink-0"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </a>
                  </div>
                  <div className="text-xs text-gray-500 mt-0.5">
                    Keşfedildi: {new Date(url.discovered_at).toLocaleDateString('tr-TR')}
                  </div>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 pt-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1 || loading}
            >
              ←
            </Button>
            <span className="text-sm text-amber-700">
              {page} / {totalPages}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages || loading}
            >
              →
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
