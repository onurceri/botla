import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { getSourceUsageStats } from '@/api/analytics'

interface SourceStat {
  source_id: string
  source_name: string
  source_type: string
  times_used: number
  avg_relevance: number
  positive_feedback: number
  negative_feedback: number
  last_used: string
}

export function SourceUsageStats({ chatbotId }: { chatbotId: string }) {
  const [stats, setStats] = useState<SourceStat[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchSourceStats()
  }, [chatbotId])

  const fetchSourceStats = async () => {
    try {
      const data = await getSourceUsageStats(chatbotId, 30)
      setStats(data || [])
    } catch (error) {
      console.error('Failed to fetch source stats:', error)
    } finally {
      setLoading(false)
    }
  }

  const getFeedbackRate = (pos: number, neg: number) => {
    const total = pos + neg
    return total > 0 ? ((pos / total) * 100).toFixed(0) : '-'
  }

  if (loading) {
      return <div className="p-8 text-center text-muted-foreground">İstatistikler yükleniyor...</div>
  }

  if (stats.length === 0) {
      return (
        <Card>
            <CardHeader>
                <CardTitle>Kaynak Kullanım İstatistikleri</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="text-center text-muted-foreground py-4">Henüz veri yok.</div>
            </CardContent>
        </Card>
      )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kaynak Kullanım İstatistikleri</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="rounded-2xl border border-border overflow-hidden shadow-sm">
            <table className="w-full text-sm text-left">
            <thead className="bg-muted/40 text-muted-foreground font-medium">
                <tr>
                <th className="px-4 py-3">Kaynak</th>
                <th className="px-4 py-3">Tip</th>
                <th className="px-4 py-3 text-right">Kullanım</th>
                <th className="px-4 py-3 text-right">Ortalama İlgi</th>
                <th className="px-4 py-3 text-right">Memnuniyet</th>
                <th className="px-4 py-3">Son Kullanım</th>
                </tr>
            </thead>
            <tbody className="divide-y divide-border">
                {stats.map((stat) => (
                <tr key={stat.source_id} className="hover:bg-muted/50 transition-colors">
                    <td className="px-4 py-3 font-medium truncate max-w-[200px]" title={stat.source_name}>{stat.source_name}</td>
                    <td className="px-4 py-3">
                    <Badge variant="outline">{stat.source_type}</Badge>
                    </td>
                    <td className="px-4 py-3 text-right">{stat.times_used}</td>
                    <td className="px-4 py-3 text-right">
                    {(stat.avg_relevance * 100).toFixed(1)}%
                    </td>
                    <td className="px-4 py-3 text-right">
                    {getFeedbackRate(stat.positive_feedback, stat.negative_feedback)}%
                    </td>
                    <td className="px-4 py-3">{new Date(stat.last_used).toLocaleDateString('tr-TR')}</td>
                </tr>
                ))}
            </tbody>
            </table>
        </div>
      </CardContent>
    </Card>
  )
}
