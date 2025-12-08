import { useCallback, useEffect, useState } from 'react'
import { listSources, getSourceStatus, deleteSource, refreshSource } from '@/api/source'
import { useToast } from '@/components/ui/toast'

type Source = { id: string; source_type: string; original_filename?: string; source_url?: string; status: string; chunk_count: number }

export function useSourceOps(id: string | undefined, isNew: boolean) {
  const [sources, setSources] = useState<Source[]>([])
  const [refreshingId, setRefreshingId] = useState<string | undefined>()
  const { toast } = useToast()

  const refreshSources = useCallback(() => {
    if (!isNew && id) {
      listSources(id).then(setSources).catch(() => { })
    }
  }, [id, isNew])

  const pollStatus = useCallback((sid: string) => {
    let attempts = 0
    let delay = 1000
    let etag: string | undefined
    const tick = async () => {
      attempts++
      try {
        const res = await getSourceStatus(sid, etag)
        if (!res.notModified) {
          etag = res.etag || etag
          const s = res.data as Source
          if (s && s.status !== 'pending' && s.status !== 'processing') {
            refreshSources()
            setRefreshingId(undefined)
            return
          }
        }
      } catch {
        setRefreshingId(undefined)
        return
      }
      if (attempts > 60) {
        setRefreshingId(undefined)
        return
      }
      setTimeout(tick, delay)
      delay = Math.min(delay * 2, 32000)
    }
    // Trigger first poll immediately to avoid test flakiness and improve responsiveness
    tick()
  }, [refreshSources])

  const handleDeleteSource = useCallback(async (sourceId: string) => {
    if (!confirm('Bu kaynağı silmek istediğinize emin misiniz?')) return
    try {
      await deleteSource(sourceId)
      toast('Kaynak başarıyla silindi.', 'success')
      refreshSources()
    } catch {
      toast('Kaynak silinirken bir hata oluştu.', 'error')
    }
  }, [refreshSources, toast])

  const handleRefreshSource = useCallback(async (sourceId: string) => {
    try {
      setRefreshingId(sourceId)
      await refreshSource(sourceId)
      toast('Kaynak yenileniyor...', 'success')
      refreshSources()
      pollStatus(sourceId)
    } catch (error: any) {
      setRefreshingId(undefined)
      const status = error?.response?.status
      if (status === 403) {
        toast('Yenileme özelliği planınızda aktif değil.', 'error')
      } else if (status === 402) {
        toast('Aylık yenileme limitinize ulaştınız.', 'error')
      } else if (status === 429) {
        toast('Yenileme için bekleme süresi aktif.', 'error')
      } else {
        toast('Yenileme başlatılamadı.', 'error')
      }
    }
  }, [refreshSources, pollStatus, toast])

  useEffect(() => { refreshSources() }, [refreshSources])

  return { sources, setSources, refreshSources, pollStatus, handleDeleteSource, handleRefreshSource, refreshingId }
}
