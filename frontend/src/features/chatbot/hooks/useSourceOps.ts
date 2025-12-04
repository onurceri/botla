import { useCallback, useEffect, useState } from 'react'
import { listSources, getSourceStatus, deleteSource } from '@/api/source'
import { useToast } from '@/components/ui/toast'

type Source = { id: string; source_type: string; original_filename?: string; source_url?: string; status: string; chunk_count: number }

export function useSourceOps(id: string | undefined, isNew: boolean) {
  const [sources, setSources] = useState<Source[]>([])
  const { toast } = useToast()

  const refreshSources = useCallback(() => {
    if (!isNew && id) {
      listSources(id).then(setSources).catch(() => {})
    }
  }, [id, isNew])

  const pollStatus = useCallback((sid: string) => {
    let attempts = 0
    const interval = setInterval(async () => {
      attempts++
      try {
        const s = await getSourceStatus(sid)
        if (s.status !== 'pending' && s.status !== 'processing') {
          clearInterval(interval)
          refreshSources()
        }
      } catch { clearInterval(interval) }
      if (attempts > 60) clearInterval(interval)
    }, 1000)
  }, [refreshSources])

  const handleDeleteSource = useCallback(async (sourceId: string) => {
    if (!confirm('Bu kaynağı silmek istediğinize emin misiniz?')) return
    try {
      await deleteSource(sourceId)
      toast('Kaynak başarıyla silindi.', 'success')
      refreshSources()
    } catch (error) {
      toast('Kaynak silinirken bir hata oluştu.', 'error')
    }
  }, [refreshSources, toast])

  useEffect(() => { refreshSources() }, [refreshSources])

  return { sources, setSources, refreshSources, pollStatus, handleDeleteSource }
}

