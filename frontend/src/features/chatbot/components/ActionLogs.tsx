import { useState, useEffect } from 'react'
import { ActionLog, getActionLogs } from '@/api/action'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

interface Props {
  chatbotId: string
}

export default function ActionLogs({ chatbotId }: Props) {
  const [logs, setLogs] = useState<ActionLog[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedLog, setSelectedLog] = useState<ActionLog | null>(null)

  useEffect(() => {
    fetchLogs()
  }, [chatbotId])

  const fetchLogs = async () => {
    setLoading(true)
    try {
      const data = await getActionLogs(chatbotId)
      setLogs(data.logs || [])
    } catch (error) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Intl.DateTimeFormat('tr-TR', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    }).format(new Date(dateString))
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium">Aksiyon Geçmişi</h3>
        <Button variant="outline" size="sm" onClick={fetchLogs} disabled={loading}>
          Yenile
        </Button>
      </div>

      <div className="border rounded-md overflow-hidden">
        <table className="w-full text-sm text-left">
          <thead className="bg-muted/50 border-b">
            <tr>
              <th className="p-3 font-medium">Tarih</th>
              <th className="p-3 font-medium">Durum</th>
              <th className="p-3 font-medium">Süre</th>
              <th className="p-3 font-medium text-right">Detay</th>
            </tr>
          </thead>
          <tbody>
            {logs.length === 0 && !loading && (
              <tr>
                <td colSpan={4} className="text-center py-8 text-muted-foreground">
                  Henüz bir kayıt yok.
                </td>
              </tr>
            )}
            {logs.map((log) => (
              <tr key={log.id} className="border-b last:border-0 hover:bg-muted/30">
                <td className="p-3 whitespace-nowrap">{formatDate(log.created_at)}</td>
                <td className="p-3">
                  <Badge
                    variant={log.status === 'success' ? 'default' : 'destructive'}
                    className={log.status === 'success' ? 'bg-green-600 hover:bg-green-700' : ''}
                  >
                    {log.status === 'success' ? 'Başarılı' : 'Hata'}
                  </Badge>
                </td>
                <td className="p-3">{log.duration_ms} ms</td>
                <td className="p-3 text-right">
                  <Button variant="ghost" size="sm" onClick={() => setSelectedLog(log)}>
                    İncele
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>Aksiyon Detayı</DialogTitle>
          </DialogHeader>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <h4 className="font-medium text-sm text-muted-foreground">İstek (Request)</h4>
              <div className="h-[300px] w-full rounded-md border p-4 bg-muted/50 overflow-auto">
                <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                  {JSON.stringify(selectedLog?.request_payload, null, 2)}
                </pre>
              </div>
            </div>
            <div className="space-y-2">
              <h4 className="font-medium text-sm text-muted-foreground">Yanıt (Response)</h4>
              <div className="h-[300px] w-full rounded-md border p-4 bg-muted/50 overflow-auto">
                <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                  {JSON.stringify(selectedLog?.response_payload, null, 2)}
                </pre>
              </div>
            </div>
          </div>
          {selectedLog?.error_message && (
            <div className="mt-4 p-4 bg-red-50 text-red-900 rounded-md border border-red-200">
              <p className="font-medium text-sm">Hata Mesajı:</p>
              <p className="text-sm mt-1">{selectedLog.error_message}</p>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
