import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
} from 'recharts'
import {
  MessageSquare,
  Users,
  Zap,
  ArrowUpRight,
  Plus,
  Bot,
  ThumbsUp,
  Activity,
} from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api } from '@/api/client'
import { getAnalytics } from '@/api/analytics'
import { useOrganization } from '@/features/organization/context/OrganizationContext'

export const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    const data = payload[0]?.payload || {}
    return (
      <div className="glass-card p-3 border border-border/50 shadow-xl rounded-lg bg-background/95 backdrop-blur-sm">
        <p className="text-sm font-medium mb-2 border-b border-border/50 pb-1">
          {new Date(label).toLocaleDateString('tr-TR', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric',
          })}
        </p>
        <div className="space-y-1">
          {payload.map((entry: any) => (
            <div key={entry.name} className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 rounded-full" style={{ backgroundColor: entry.color }} />
              <span className="text-muted-foreground capitalize">
                {entry.name === 'conversations'
                  ? 'Konuşma'
                  : entry.name === 'messages'
                    ? 'Mesaj'
                    : entry.name}
              </span>
              <span className="font-bold font-mono">{entry.value}</span>
            </div>
          ))}

          {/* Additional data */}
          {(data.tokens > 0 || data.thumbs_up > 0 || data.thumbs_down > 0 || data.handoffs > 0) && (
            <div className="pt-2 mt-2 border-t border-border/50 space-y-1">
              {data.tokens > 0 && (
                <div className="flex items-center justify-between gap-4 text-xs">
                  <span className="text-muted-foreground">Token</span>
                  <span className="font-mono">{data.tokens.toLocaleString()}</span>
                </div>
              )}
              {(data.thumbs_up > 0 || data.thumbs_down > 0) && (
                <div className="flex items-center justify-between gap-4 text-xs">
                  <span className="text-muted-foreground">Geri Bildirim</span>
                  <span className="font-mono">
                    👍 {data.thumbs_up} / 👎 {data.thumbs_down}
                  </span>
                </div>
              )}
              {data.handoffs > 0 && (
                <div className="flex items-center justify-between gap-4 text-xs">
                  <span className="text-muted-foreground">İnsan Desteği</span>
                  <span className="font-mono">{data.handoffs}</span>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    )
  }
  return null
}

export const formatXAxisTick = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString('tr-TR', { day: '2-digit', month: 'short' })
}

export const formatYAxisTick = (value: number) => `${value}`

const DashboardPage = () => {
  const { currentWorkspace, isLoading: isOrgLoading } = useOrganization()
  const [stats, setStats] = useState({
    totalConversations: 0,
    totalMessages: 0,
    totalTokens: 0,
    positiveFeedback: 0,
    negativeFeedback: 0,
    activeBots: 0,
  })
  const [chartData, setChartData] = useState<any[]>([])
  const [recentBots, setRecentBots] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Abort controller to cancel in-flight requests on cleanup
    const abortController = new AbortController()
    let isMounted = true

    const fetchData = async () => {
      // Wait for org context to be ready (at least tried to load)
      if (isOrgLoading) return
      if (!currentWorkspace) return

      try {
        // Fetch Analytics
        const analyticsRaw = await getAnalytics()

        // Check if component is still mounted before setting state
        if (!isMounted) return

        const analyticsData = Array.isArray(analyticsRaw) ? analyticsRaw : []
        setChartData(analyticsData)

        // Calculate totals from analytics
        const totalConv = analyticsData.reduce(
          (acc: number, curr: any) => acc + (curr?.conversations ?? 0),
          0,
        )
        const totalMsg = analyticsData.reduce(
          (acc: number, curr: any) => acc + (curr?.messages ?? 0),
          0,
        )
        const totalTok = analyticsData.reduce(
          (acc: number, curr: any) => acc + (curr?.tokens ?? 0),
          0,
        )
        const totalPos = analyticsData.reduce(
          (acc: number, curr: any) => acc + (curr?.thumbs_up ?? 0),
          0,
        )
        const totalNeg = analyticsData.reduce(
          (acc: number, curr: any) => acc + (curr?.thumbs_down ?? 0),
          0,
        )

        let bots: any[] = []
        try {
          const { data } = await api.get('/api/v1/chatbots')
          bots = Array.isArray(data) ? data : []
        } catch (error) {
          console.error('Failed to fetch chatbots', error)
          bots = []
        }

        if (!isMounted) return

        setRecentBots(bots.slice(0, 3))

        setStats({
          totalConversations: totalConv,
          totalMessages: totalMsg,
          totalTokens: totalTok,
          positiveFeedback: totalPos,
          negativeFeedback: totalNeg,
          activeBots: bots.length,
        })
      } catch (error) {
        // Ignore abort errors
        if (error instanceof Error && error.name === 'AbortError') return
        console.error('Failed to fetch dashboard data', error)
      } finally {
        if (isMounted) {
          setLoading(false)
        }
      }
    }

    fetchData()

    // Cleanup function
    return () => {
      isMounted = false
      abortController.abort()
    }
  }, [currentWorkspace, isOrgLoading])

  if (loading) {
    return (
      <div className="space-y-8 animate-pulse">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-2">
            <div className="h-8 w-48 bg-muted rounded"></div>
            <div className="h-4 w-64 bg-muted rounded"></div>
          </div>
          <div className="h-10 w-32 bg-muted rounded"></div>
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="h-32 bg-muted/50 border-0" />
          ))}
        </div>
        <div className="grid gap-4 grid-cols-1 lg:grid-cols-7">
          <Card className="lg:col-span-4 h-96 bg-muted/50 border-0" />
          <Card className="lg:col-span-3 h-96 bg-muted/50 border-0" />
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
            Dashboard
          </h1>
          <p className="text-muted-foreground">
            Botlarınızın son 30 günlük performansına genel bakış.
          </p>
        </div>
        <Link to="/dashboard/chatbots/new">
          <Button className="gap-2 shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all duration-300">
            <Plus className="w-4 h-4" /> Yeni Chatbot
          </Button>
        </Link>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <div className="glass-card rounded-xl overflow-hidden relative group hover:shadow-md transition-all duration-300 border-l-4 border-l-blue-500/50">
          <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <Users className="w-24 h-24" />
          </div>
          <div className="p-6 pb-2 relative z-10 flex justify-between items-center">
            <h3 className="tracking-tight text-sm font-medium text-muted-foreground">
              Toplam Konuşma
            </h3>
            <div className="p-2 bg-blue-500/10 rounded-full text-blue-500">
              <Users className="h-4 w-4" />
            </div>
          </div>
          <div className="p-6 pt-0 relative z-10">
            <div className="text-2xl font-bold tracking-tight">{stats.totalConversations}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Son 30 günde tüm botlarınızdaki toplam
            </p>
          </div>
        </div>

        <div className="glass-card rounded-xl overflow-hidden relative group hover:shadow-md transition-all duration-300 border-l-4 border-l-purple-500/50">
          <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <MessageSquare className="w-24 h-24" />
          </div>
          <div className="p-6 pb-2 relative z-10 flex justify-between items-center">
            <h3 className="tracking-tight text-sm font-medium text-muted-foreground">
              Toplam Mesaj
            </h3>
            <div className="p-2 bg-purple-500/10 rounded-full text-purple-500">
              <MessageSquare className="h-4 w-4" />
            </div>
          </div>
          <div className="p-6 pt-0 relative z-10">
            <div className="text-2xl font-bold tracking-tight">{stats.totalMessages}</div>
            <p className="text-xs text-muted-foreground mt-1">Son 30 günde işlenen toplam mesaj</p>
          </div>
        </div>

        <div className="glass-card rounded-xl overflow-hidden relative group hover:shadow-md transition-all duration-300 border-l-4 border-l-amber-500/50">
          <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <Zap className="w-24 h-24" />
          </div>
          <div className="p-6 pb-2 relative z-10 flex justify-between items-center">
            <h3 className="tracking-tight text-sm font-medium text-muted-foreground">
              Harcanan Token
            </h3>
            <div className="p-2 bg-amber-500/10 rounded-full text-amber-500">
              <Zap className="h-4 w-4" />
            </div>
          </div>
          <div className="p-6 pt-0 relative z-10">
            <div className="text-2xl font-bold tracking-tight">
              {(stats.totalTokens / 1000).toFixed(1)}k
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Son 30 günlük tahmini maliyet: ${((stats.totalTokens / 1000) * 0.002).toFixed(3)}
            </p>
          </div>
        </div>

        <div className="glass-card rounded-xl overflow-hidden relative group hover:shadow-md transition-all duration-300 border-l-4 border-l-emerald-500/50">
          <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <ThumbsUp className="w-24 h-24" />
          </div>
          <div className="p-6 pb-2 relative z-10 flex justify-between items-center">
            <h3 className="tracking-tight text-sm font-medium text-muted-foreground">Memnuniyet</h3>
            <div className="p-2 bg-emerald-500/10 rounded-full text-emerald-500">
              <ThumbsUp className="h-4 w-4" />
            </div>
          </div>
          <div className="p-6 pt-0 relative z-10">
            <div className="text-2xl font-bold tracking-tight">
              {stats.positiveFeedback + stats.negativeFeedback > 0
                ? Math.round(
                    (stats.positiveFeedback / (stats.positiveFeedback + stats.negativeFeedback)) *
                      100,
                  )
                : 0}
              %
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Son 30 günde: {stats.positiveFeedback} olumlu, {stats.negativeFeedback} olumsuz
            </p>
          </div>
        </div>
      </div>

      {/* Charts & Recent Activity */}
      <div className="grid gap-4 grid-cols-1 lg:grid-cols-7">
        {/* Main Chart */}
        <div className="lg:col-span-4 glass-card rounded-xl p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h3 className="font-semibold text-lg">Aktivite Özeti</h3>
              <p className="text-sm text-muted-foreground">
                {chartData.length > 0 ? (
                  <>
                    {new Date(chartData[0].date).toLocaleDateString('tr-TR', {
                      day: 'numeric',
                      month: 'short',
                    })}{' '}
                    -{' '}
                    {new Date(chartData[chartData.length - 1].date).toLocaleDateString('tr-TR', {
                      day: 'numeric',
                      month: 'short',
                    })}
                  </>
                ) : (
                  'Son 30 Gün'
                )}
              </p>
            </div>
            {chartData.length > 0 && (
              <div className="flex items-center gap-1 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 px-2.5 py-1 rounded-full text-xs font-medium border border-emerald-500/20">
                <Activity className="w-3.5 h-3.5" />
                <span>Canlı</span>
              </div>
            )}
          </div>

          <div className="min-w-0">
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300} minWidth={0}>
                <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorMessages" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="colorConversations" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid
                    strokeDasharray="3 3"
                    stroke="hsl(var(--border))"
                    vertical={false}
                    strokeOpacity={0.5}
                  />
                  <XAxis
                    dataKey="date"
                    stroke="hsl(var(--muted-foreground))"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={formatXAxisTick}
                    minTickGap={30}
                  />
                  <YAxis
                    stroke="hsl(var(--muted-foreground))"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={formatYAxisTick}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Area
                    type="monotone"
                    dataKey="messages"
                    name="messages"
                    stroke="#8b5cf6"
                    strokeWidth={2}
                    fillOpacity={1}
                    fill="url(#colorMessages)"
                  />
                  <Area
                    type="monotone"
                    dataKey="conversations"
                    name="conversations"
                    stroke="#3b82f6"
                    strokeWidth={2}
                    fillOpacity={1}
                    fill="url(#colorConversations)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[300px] flex items-center justify-center border-2 border-dashed border-muted rounded-lg">
                <p className="text-muted-foreground">Henüz aktivite yok</p>
              </div>
            )}
          </div>
        </div>

        {/* Recent Chatbots */}
        <div className="lg:col-span-3 glass-card rounded-xl p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="font-semibold text-lg">Son Chatbotlar</h3>
            <Link to="/dashboard/chatbots" className="text-xs text-primary hover:underline">
              Tümünü Gör
            </Link>
          </div>

          <div className="space-y-4">
            {recentBots.length > 0 ? (
              recentBots.map((bot) => (
                <Link
                  key={bot.id}
                  to={`/dashboard/chatbots/${bot.id}`}
                  className="flex items-center justify-between p-4 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors group border border-border/50"
                >
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10 rounded-full text-primary group-hover:scale-110 transition-transform">
                      <Bot className="h-4 w-4" />
                    </div>
                    <div>
                      <h4 className="font-medium text-sm group-hover:text-primary transition-colors">
                        {bot.name}
                      </h4>
                      <p className="text-xs text-muted-foreground">{bot.model}</p>
                    </div>
                  </div>
                  <ArrowUpRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                </Link>
              ))
            ) : (
              <div className="text-center py-12 border-2 border-dashed border-muted rounded-lg">
                <p className="text-muted-foreground text-sm">Henüz chatbot oluşturulmadı</p>
                <Link to="/dashboard/chatbots/new">
                  <Button variant="link" className="text-primary mt-2">
                    İlk botunuzu oluşturun
                  </Button>
                </Link>
              </div>
            )}
          </div>

          <div className="mt-6 pt-6 border-t border-border/50">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Toplam Aktif Bot</span>
              <span className="font-bold">{stats.activeBots}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default DashboardPage
