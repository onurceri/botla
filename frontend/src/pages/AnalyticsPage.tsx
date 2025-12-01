import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAnalytics } from '@/api/analytics'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import Card from '@/components/shared/Card'

const AnalyticsPage = () => {
  const [start, setStart] = useState('')
  const [end, setEnd] = useState('')
  const { data, isLoading, isError } = useQuery({
    queryKey: ['analytics'],
    queryFn: getAnalytics,
  })

  const filtered = useMemo(() => {
    const items = Array.isArray(data) ? data : []
    return items.filter((d) => {
      const t = new Date(d.date).getTime()
      const s = start ? new Date(start).getTime() : -Infinity
      const e = end ? new Date(end).getTime() : Infinity
      return t >= s && t <= e
    })
  }, [start, end, data])

  const totalMessages = filtered.reduce((acc, d) => acc + d.messages, 0)
  const totalConversations = filtered.reduce((acc, d) => acc + d.conversations, 0)
  const unansweredPercent = 7
  const avgRating = 4.6

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <input type="date" className="rounded border bg-input px-3 py-2" value={start} onChange={(e) => setStart(e.target.value)} />
        <input type="date" className="rounded border bg-input px-3 py-2" value={end} onChange={(e) => setEnd(e.target.value)} />
      </div>

      {isLoading && (
        <Card title="Analitikler Yükleniyor">
          <div className="p-4 text-sm text-muted-foreground">Veriler getiriliyor...</div>
        </Card>
      )}

      {isError && (
        <Card title="Analitikler Hatası">
          <div className="p-4 text-sm text-red-600">Veriler alınırken bir hata oluştu.</div>
        </Card>
      )}

      {!isLoading && !isError && filtered.length === 0 && (
        <Card title="Veri Yok">
          <div className="p-4 text-sm text-muted-foreground">Görüntülenecek veri bulunmuyor.</div>
        </Card>
      )}

      {!isLoading && !isError && filtered.length > 0 && (
      <Card title="Günlük Mesaj Sayısı">
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={filtered}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Line type="monotone" dataKey="messages" stroke="#8884d8" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Card>
      )}

      {!isLoading && !isError && filtered.length > 0 && (
      <Card title="Konuşma Sayısı">
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={filtered}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Line type="monotone" dataKey="conversations" stroke="#82ca9d" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Card>
      )}

      <Card title="Metrikler">
        <div className="grid gap-3 md:grid-cols-4">
          <div className="rounded border bg-card p-3">
            <div className="text-xs text-muted-foreground">Toplam Mesaj</div>
            <div className="text-xl font-semibold">{totalMessages}</div>
          </div>
          <div className="rounded border bg-card p-3">
            <div className="text-xs text-muted-foreground">Konuşma</div>
            <div className="text-xl font-semibold">{totalConversations}</div>
          </div>
          <div className="rounded border bg-card p-3">
            <div className="text-xs text-muted-foreground">Yanıtlanamayan %</div>
            <div className="text-xl font-semibold">{unansweredPercent}%</div>
          </div>
          <div className="rounded border bg-card p-3">
            <div className="text-xs text-muted-foreground">Ortalama Puan</div>
            <div className="text-xl font-semibold">{avgRating}</div>
          </div>
        </div>
      </Card>
    </div>
  )
}

export default AnalyticsPage
