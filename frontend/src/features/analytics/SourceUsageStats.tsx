import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { getSourceUsageStats } from '@/api/analytics'
import {
  FileText,
  Globe,
  AlignLeft,
  ThumbsUp,
  ThumbsDown,
  Activity,
  Clock,
  BarChart3,
} from 'lucide-react'

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

  const getIconForType = (type: string) => {
    switch (type.toLowerCase()) {
      case 'pdf':
        return <FileText className="h-4 w-4" />
      case 'url':
        return <Globe className="h-4 w-4" />
      case 'text':
        return <AlignLeft className="h-4 w-4" />
      default:
        return <FileText className="h-4 w-4" />
    }
  }

  const getFeedbackRate = (pos: number, neg: number) => {
    const total = pos + neg
    return total > 0 ? Math.round((pos / total) * 100) : 0
  }

  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="h-20 bg-muted/50" />
            <CardContent className="h-32" />
          </Card>
        ))}
      </div>
    )
  }

  if (stats.length === 0) {
    return (
      <Card className="bg-muted/30 border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-12 text-center">
          <BarChart3 className="h-10 w-10 text-muted-foreground/50 mb-3" />
          <h3 className="font-medium text-muted-foreground">Henüz veri yok</h3>
          <p className="text-sm text-muted-foreground/80 mt-1">
            Botunuz kaynakları kullandıkça burada istatistikler görünecektir.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium tracking-tight">Kaynak Performansı</h3>
        <Badge variant="outline" className="text-muted-foreground font-normal">
          Son 30 Gün
        </Badge>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {stats.map((stat) => (
          <Card
            key={stat.source_id}
            className="overflow-hidden hover:shadow-md transition-all duration-300 group"
          >
            <CardHeader className="p-4 bg-muted/30 pb-3">
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-2 min-w-0">
                  <div className="p-2 bg-background rounded-md shadow-sm text-primary">
                    {getIconForType(stat.source_type)}
                  </div>
                  <div className="min-w-0">
                    <h4 className="font-medium text-sm truncate" title={stat.source_name}>
                      {stat.source_name}
                    </h4>
                    <span className="text-xs text-muted-foreground capitalize flex items-center gap-1">
                      {stat.source_type}
                    </span>
                  </div>
                </div>
                {stat.times_used > 10 && (
                  <Badge
                    variant="secondary"
                    className="bg-green-500/10 text-green-600 hover:bg-green-500/20 border-0 text-[10px] px-1.5"
                  >
                    Popüler
                  </Badge>
                )}
              </div>
            </CardHeader>
            <CardContent className="p-4 pt-3 space-y-4">
              <div className="flex items-end justify-between">
                <div>
                  <div className="text-2xl font-bold tracking-tight">{stat.times_used}</div>
                  <div className="text-xs text-muted-foreground font-medium">Kullanım</div>
                </div>
                <div className="text-right">
                  <div className="text-sm font-medium flex items-center justify-end gap-1 text-muted-foreground">
                    <Clock className="h-3 w-3" />
                    {new Date(stat.last_used).toLocaleDateString('tr-TR', {
                      day: 'numeric',
                      month: 'short',
                    })}
                  </div>
                  <div className="text-[10px] text-muted-foreground/80">Son Erişim</div>
                </div>
              </div>

              <div className="space-y-3 pt-1">
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Activity className="h-3 w-3" /> İlgi Düzeyi
                    </span>
                    <span className="font-medium">{(stat.avg_relevance * 100).toFixed(0)}%</span>
                  </div>
                  <Progress value={stat.avg_relevance * 100} className="h-1.5 bg-muted" />
                </div>

                <div className="flex items-center gap-4 pt-1 border-t border-border/50 mt-2">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <ThumbsUp className="h-3.5 w-3.5 text-green-600/70" />
                    <span className="font-medium text-foreground">{stat.positive_feedback}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <ThumbsDown className="h-3.5 w-3.5 text-red-600/70" />
                    <span className="font-medium text-foreground">{stat.negative_feedback}</span>
                  </div>
                  <div className="ml-auto text-xs text-muted-foreground">
                    {getFeedbackRate(stat.positive_feedback, stat.negative_feedback)}% Memnuniyet
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
