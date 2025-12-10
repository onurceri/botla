import { useEffect, useState } from 'react'
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, MessageSquare, Zap, ThumbsUp } from 'lucide-react'
import { getChatbotAnalyticsOverview, getChatbotAnalyticsTrends } from '@/api/analytics'
import { CustomTooltip, formatXAxisTick, formatYAxisTick } from '@/pages/DashboardPage'
import { SourceUsageStats } from './SourceUsageStats'

interface ChatbotAnalyticsProps {
  chatbotId: string
}

export const ChatbotAnalytics = ({ chatbotId }: ChatbotAnalyticsProps) => {
  const [overview, setOverview] = useState<any>(null)
  const [trends, setTrends] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [overviewData, trendsData] = await Promise.all([
          getChatbotAnalyticsOverview(chatbotId),
          getChatbotAnalyticsTrends(chatbotId, 30) // Default 30 days
        ])
        setOverview(overviewData)
        // Transform trends data to match Dashboard format for CustomTooltip compatibility
        const normalizedTrends = (trendsData?.daily || []).map((item: any) => ({
          date: item.date,
          messages: item.total_messages || 0,
          conversations: item.total_conversations || 0,
          tokens: item.total_tokens_used || 0,
          thumbs_up: item.thumbs_up_count || 0,
          thumbs_down: item.thumbs_down_count || 0,
          handoffs: item.handoff_count || 0
        }))
        setTrends(normalizedTrends)
      } catch (error) {
        console.error('Failed to fetch analytics:', error)
      } finally {
        setLoading(false)
      }
    }

    if (chatbotId) {
      fetchData()
    }
  }, [chatbotId])

  if (loading) {
    return <div className="p-8 text-center text-muted-foreground">Analizler yükleniyor...</div>
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Toplam Mesaj</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{overview?.total_messages || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Konuşmalar</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{overview?.total_conversations || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memnuniyet Oranı</CardTitle>
            <ThumbsUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {overview?.feedback_rate ? `${overview.feedback_rate.toFixed(1)}%` : '-'}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {overview?.positive_feedback || 0} olumlu, {overview?.negative_feedback || 0} olumsuz
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Toplam Token</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{overview?.total_tokens_used?.toLocaleString() || 0}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Aktivite Grafiği (Son 30 Gün)</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trends} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorMessages" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.2}/>
                    <stop offset="95%" stopColor="#f59e0b" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="colorConversations" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.2}/>
                    <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} strokeOpacity={0.5} />
                <XAxis 
                  dataKey="date" 
                  tickFormatter={formatXAxisTick} 
                  stroke="#94a3b8"
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                  dy={10}
                />
                <YAxis 
                  stroke="#94a3b8"
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={formatYAxisTick}
                  dx={-10}
                />
                <Tooltip content={<CustomTooltip />} />
                <Area 
                  type="monotone" 
                  dataKey="conversations" 
                  name="Konuşma"
                  stroke="#8b5cf6" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorConversations)" 
                  dot={false}
                  activeDot={{ r: 6, strokeWidth: 0, fill: '#8b5cf6' }}
                />
                <Area 
                  type="monotone" 
                  dataKey="messages" 
                  name="Mesaj"
                  stroke="#f59e0b" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorMessages)" 
                  dot={false}
                  activeDot={{ r: 6, strokeWidth: 0, fill: '#f59e0b' }}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      <SourceUsageStats chatbotId={chatbotId} />
    </div>
  )
}
