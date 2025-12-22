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
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { 
  Users, 
  MessageSquare, 
  Zap, 
  ThumbsUp, 
  Calendar,
  ArrowUpRight,
  Activity
} from 'lucide-react'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { getChatbotAnalyticsOverview, getChatbotAnalyticsTrends } from '@/api/analytics'
import { formatXAxisTick, formatYAxisTick } from '@/pages/DashboardPage'
import { SourceUsageStats } from './SourceUsageStats'
import { ErrorBoundary } from '@/components/ui/error-boundary'

interface ChatbotAnalyticsProps {
  chatbotId: string
}

interface OverviewData {
  total_conversations: number
  total_messages: number
  total_tokens_used: number
  avg_positive_feedback?: number
  positive_feedback?: number
  negative_feedback?: number
}

interface TrendData {
  date: string
  total_conversations: number
  total_messages: number
}

interface TrendResponse {
  daily: TrendData[]
}

const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    const labelDate = new Date(label)
    return (
      <div className="glass-card p-3 border border-border/50 shadow-xl rounded-lg bg-background/95 backdrop-blur-sm">
        <p className="text-sm font-medium mb-2 border-b border-border/50 pb-1">
          {!Number.isNaN(labelDate.getTime())
            ? labelDate.toLocaleDateString('tr-TR', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
            : String(label)}
        </p>
        <div className="space-y-1">
          {payload.map((entry: any) => (
            <div key={entry.name} className="flex items-center gap-2 text-sm">
              <div 
                className="w-2 h-2 rounded-full" 
                style={{ backgroundColor: entry.color }}
              />
              <span className="text-muted-foreground capitalize">
                {entry.name === 'total_conversations' ? 'Konuşma' : 'Mesaj'}:
              </span>
              <span className="font-bold font-mono">
                {entry.value}
              </span>
            </div>
          ))}
        </div>
      </div>
    )
  }
  return null
}

export function ChatbotAnalytics({ chatbotId }: ChatbotAnalyticsProps) {
  const [overview, setOverview] = useState<OverviewData | null>(null)
  const [trends, setTrends] = useState<TrendData[]>([])
  const [loading, setLoading] = useState(true)
  const [days, setDays] = useState('30')
  const [messagesColor, setMessagesColor] = useState('var(--color-primary)')

  useEffect(() => {
    const raw = getComputedStyle(document.documentElement).getPropertyValue('--color-primary').trim()
    if (raw) {
      setMessagesColor(raw)
    }
  }, [])

  useEffect(() => {
    async function loadData() {
      if (!chatbotId) return
      setLoading(true)
      try {
        const [overviewData, trendsData] = await Promise.all([
          getChatbotAnalyticsOverview(chatbotId),
          getChatbotAnalyticsTrends(chatbotId, parseInt(days))
        ])
        setOverview(overviewData)
        const daily =
          Array.isArray(trendsData)
            ? (trendsData as TrendData[])
            : Array.isArray((trendsData as TrendResponse | null | undefined)?.daily)
              ? (trendsData as TrendResponse).daily
              : []
        setTrends(daily)
      } catch (error) {
        console.error(error)
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [chatbotId, days])

  const messagesGradientId = `analytics-messages-${chatbotId}`
  const conversationsGradientId = `analytics-conversations-${chatbotId}`

  const StatCard = ({ title, value, icon: Icon, subtext, trend }: any) => (
    <Card className="overflow-hidden relative group hover:shadow-md transition-all duration-300 border-l-4 border-l-primary/20">
      <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
        <Icon className="w-24 h-24" />
      </div>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 z-10 relative">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        <div className="p-2 bg-primary/10 rounded-full text-primary">
          <Icon className="h-4 w-4" />
        </div>
      </CardHeader>
      <CardContent className="z-10 relative">
        <div className="text-2xl font-bold tracking-tight">{value}</div>
        <div className="flex items-center text-xs text-muted-foreground mt-1">
          {trend && (
            <span className="text-green-600 flex items-center mr-2 font-medium bg-green-500/10 px-1.5 py-0.5 rounded">
              <ArrowUpRight className="w-3 h-3 mr-1" />
              {trend}
            </span>
          )}
          <span>{subtext}</span>
        </div>
      </CardContent>
    </Card>
  )

  if (loading && !overview) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[1, 2, 3, 4].map(i => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="h-20 bg-muted/50" />
            <CardContent className="h-12" />
          </Card>
        ))}
        <Card className="col-span-4 h-[300px] animate-pulse bg-muted/20" />
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Date Filter & Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h3 className="text-lg font-medium tracking-tight">Genel Bakış</h3>
          <p className="text-sm text-muted-foreground">Botunuzun performans metrikleri ve kullanım istatistikleri.</p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={days} onValueChange={setDays}>
            <SelectTrigger className="w-[180px]">
              <Calendar className="mr-2 h-4 w-4" />
              <SelectValue placeholder="Zaman Aralığı" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="7">Son 7 Gün</SelectItem>
              <SelectItem value="30">Son 30 Gün</SelectItem>
              <SelectItem value="90">Son 3 Ay</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Toplam Mesaj"
          value={overview?.total_messages?.toLocaleString() || 0}
          icon={MessageSquare}
          subtext="Tüm zamanlar"
          trend="12%"
        />
        <StatCard
          title="Konuşmalar"
          value={overview?.total_conversations?.toLocaleString() || 0}
          icon={Users}
          subtext="Aktif oturumlar"
          trend="5%"
        />
        <StatCard
          title="Memnuniyet"
          value={
            overview?.avg_positive_feedback !== undefined
              ? `${Math.round(overview.avg_positive_feedback * 100)}%`
              : overview?.positive_feedback !== undefined && overview?.negative_feedback !== undefined
                ? overview.positive_feedback + overview.negative_feedback > 0
                  ? `${Math.round((overview.positive_feedback / (overview.positive_feedback + overview.negative_feedback)) * 100)}%`
                  : '0%'
                : '-'
          }
          icon={ThumbsUp}
          subtext="Kullanıcı oyları"
        />
        <StatCard
          title="Token Kullanımı"
          value={overview?.total_tokens_used?.toLocaleString() || 0}
          icon={Zap}
          subtext="Toplam tüketim"
        />
      </div>

      {/* Main Chart */}
      <Card className="col-span-4 overflow-hidden border-none shadow-lg bg-gradient-to-br from-background to-muted/20">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5 text-primary" />
                Etkileşim Trendleri
              </CardTitle>
              <CardDescription>
                Günlük konuşma ve mesaj trafiği analizi
              </CardDescription>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-primary" />
                <span>Mesajlar</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-purple-500" />
                <span>Konuşmalar</span>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent className="pl-0">
          <div className="w-full mt-4 min-w-0">
            <ErrorBoundary>
              <ResponsiveContainer
                width="100%"
                height={350}
                minWidth={0}
                minHeight={0}
                initialDimension={{ width: 800, height: 350 }}
                debounce={50}
              >
                <AreaChart data={trends} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id={messagesGradientId} x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor={messagesColor} stopOpacity={0.3}/>
                      <stop offset="95%" stopColor={messagesColor} stopOpacity={0}/>
                    </linearGradient>
                    <linearGradient id={conversationsGradientId} x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <XAxis 
                    dataKey="date" 
                    tickFormatter={formatXAxisTick} 
                    stroke="#888888" 
                    fontSize={12} 
                    tickLine={false} 
                    axisLine={false}
                    minTickGap={30}
                  />
                  <YAxis 
                    tickFormatter={formatYAxisTick} 
                    stroke="#888888" 
                    fontSize={12} 
                    tickLine={false} 
                    axisLine={false}
                    tickCount={5}
                  />
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border))" opacity={0.4} />
                  <Tooltip content={<CustomTooltip />} />
                  <Area 
                    type="monotone" 
                    dataKey="total_messages" 
                    stroke={messagesColor} 
                    strokeWidth={3}
                    fillOpacity={1} 
                    fill={`url(#${messagesGradientId})`} 
                    activeDot={{ r: 6, strokeWidth: 0, className: "animate-ping" }}
                  />
                  <Area 
                    type="monotone" 
                    dataKey="total_conversations" 
                    stroke="#8b5cf6" 
                    strokeWidth={3}
                    fillOpacity={1} 
                    fill={`url(#${conversationsGradientId})`} 
                  />
                </AreaChart>
              </ResponsiveContainer>
            </ErrorBoundary>
          </div>
        </CardContent>
      </Card>

      {/* Source Stats */}
      <SourceUsageStats chatbotId={chatbotId} />
    </div>
  )
}
