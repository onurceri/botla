import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  AreaChart,
  Area
} from 'recharts'
import { 
  MessageSquare, 
  Users, 
  Zap, 
  ArrowUpRight, 
  Plus,
  Bot,
  ThumbsUp
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api } from '@/api/client'
import { getAnalytics } from '@/api/analytics'
import { useOrganization } from '@/features/organization/context/OrganizationContext'

export const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    const data = payload[0]?.payload || {}
    return (
      <div className="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 p-3 rounded-xl shadow-xl min-w-[180px]">
        <p className="text-sm font-medium text-slate-500 dark:text-slate-400 mb-3 pb-2 border-b border-slate-100 dark:border-slate-800">
          {new Date(label).toLocaleDateString('tr-TR', { day: 'numeric', month: 'long', year: 'numeric' })}
        </p>
        <div className="space-y-2">
          {/* Chart data (colored) */}
          {payload.map((entry: any, index: number) => (
            <div key={index} className="flex items-center justify-between gap-4 text-sm">
              <div className="flex items-center gap-2">
                <div 
                  className="w-2 h-2 rounded-full" 
                  style={{ backgroundColor: entry.color }}
                />
                <span className="text-slate-600 dark:text-slate-300">
                  {entry.name}
                </span>
              </div>
              <span className="font-semibold text-slate-900 dark:text-white">
                {entry.value}
              </span>
            </div>
          ))}
          
          {/* Additional data */}
          {(data.tokens > 0 || data.thumbs_up > 0 || data.thumbs_down > 0 || data.handoffs > 0) && (
            <div className="pt-2 mt-2 border-t border-slate-100 dark:border-slate-800 space-y-1.5">
              {data.tokens > 0 && (
                <div className="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
                  <span>Token</span>
                  <span className="font-medium">{data.tokens.toLocaleString()}</span>
                </div>
              )}
              {(data.thumbs_up > 0 || data.thumbs_down > 0) && (
                <div className="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
                  <span>Geri Bildirim</span>
                  <span className="font-medium">
                    👍 {data.thumbs_up} / 👎 {data.thumbs_down}
                  </span>
                </div>
              )}
              {data.handoffs > 0 && (
                <div className="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
                  <span>İnsan Desteği</span>
                  <span className="font-medium">{data.handoffs}</span>
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
    activeBots: 0
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

      try {
        // Fetch Analytics
        const analyticsRaw = await getAnalytics()
        
        // Check if component is still mounted before setting state
        if (!isMounted) return
        
        const analyticsData = Array.isArray(analyticsRaw) ? analyticsRaw : []
        setChartData(analyticsData)

        // Calculate totals from analytics
        const totalConv = analyticsData.reduce((acc: number, curr: any) => acc + (curr?.conversations ?? 0), 0)
        const totalMsg = analyticsData.reduce((acc: number, curr: any) => acc + (curr?.messages ?? 0), 0)
        const totalTok = analyticsData.reduce((acc: number, curr: any) => acc + (curr?.tokens ?? 0), 0)
        const totalPos = analyticsData.reduce((acc: number, curr: any) => acc + (curr?.thumbs_up ?? 0), 0)
        const totalNeg = analyticsData.reduce((acc: number, curr: any) => acc + (curr?.thumbs_down ?? 0), 0)

        // Fetch Bots for count and recent list
        const { data } = await api.get('/api/v1/chatbots')
        
        // Check again before setting state
        if (!isMounted) return
        
        const bots = Array.isArray(data) ? data : []
        setRecentBots(bots.slice(0, 3))
        
        setStats({
          totalConversations: totalConv,
          totalMessages: totalMsg,
          totalTokens: totalTok,
          positiveFeedback: totalPos,
          negativeFeedback: totalNeg,
          activeBots: bots.length
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
    return <div className="p-8 text-center text-muted-foreground">Yükleniyor...</div>
  }

  return (
    <div className="space-y-8">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">Botlarınızın performansına genel bakış.</p>
        </div>
        <Link to="/chatbots/new">
          <Button className="gap-2 shadow-lg shadow-primary/20">
            <Plus className="w-4 h-4" /> Yeni Chatbot
          </Button>
        </Link>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Toplam Konuşma</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalConversations}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Tüm botlarınızdaki toplam
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Toplam Mesaj</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalMessages}</div>
            <p className="text-xs text-muted-foreground mt-1">
              İşlenen toplam mesaj
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Harcanan Token</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{(stats.totalTokens / 1000).toFixed(1)}k</div>
            <p className="text-xs text-muted-foreground mt-1">
              Tahmini maliyet: ${((stats.totalTokens / 1000) * 0.002).toFixed(3)}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memnuniyet</CardTitle>
            <ThumbsUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.positiveFeedback + stats.negativeFeedback > 0 
                ? Math.round((stats.positiveFeedback / (stats.positiveFeedback + stats.negativeFeedback)) * 100)
                : 0}%
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {stats.positiveFeedback} olumlu, {stats.negativeFeedback} olumsuz
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Charts & Recent Activity */}
      <div className="grid gap-4 md:grid-cols-7">
        {/* Main Chart */}
        <Card className="col-span-4">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Aktivite Özeti</CardTitle>
                <p className="text-sm text-muted-foreground mt-1">
                  {chartData.length > 0 ? (
                    <>
                      {new Date(chartData[0].date).toLocaleDateString('tr-TR', { day: 'numeric', month: 'short' })} - {new Date(chartData[chartData.length - 1].date).toLocaleDateString('tr-TR', { day: 'numeric', month: 'short' })}
                    </>
                  ) : (
                    'Son 7 Gün'
                  )}
                </p>
              </div>
              {chartData.length > 0 && (
                <div className="flex items-center gap-1 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 px-2.5 py-1 rounded-full text-xs font-medium">
                  <ArrowUpRight className="w-3.5 h-3.5" />
                  <span>+12.5%</span>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent className="pl-2">
            <div className="min-w-0">
              {chartData.length > 0 ? (
                <ResponsiveContainer width="100%" height={300} minWidth={0}>
                  <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
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
                      stroke="#94a3b8" 
                      fontSize={12} 
                      tickLine={false} 
                      axisLine={false}
                      tickFormatter={formatXAxisTick}
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
              ) : (
                <div className="h-full flex flex-col items-center justify-center text-muted-foreground space-y-2">
                  <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center">
                    <Zap className="w-6 h-6 opacity-50" />
                  </div>
                  <p className="text-sm font-medium">Henüz aktivite yok</p>
                  <p className="text-xs">Botlarınız kullanılmaya başlandığında burada grafik göreceksiniz.</p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Recent Bots */}
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>Son Botlarınız</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentBots.map((bot) => (
                <Link 
                  key={bot.id} 
                  to={`/chatbots/${bot.id}`}
                  className="flex items-center gap-4 p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors cursor-pointer group"
                >
                  <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/10">
                    <Bot className="w-5 h-5" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium leading-none truncate text-foreground">{bot.name}</p>
                    <p className="text-xs text-muted-foreground mt-1 truncate">{bot.model}</p>
                  </div>
                  <div className="h-8 w-8 flex items-center justify-center text-muted-foreground group-hover:text-foreground transition-colors">
                    <ArrowUpRight className="w-4 h-4" />
                  </div>
                </Link>
              ))}
              {recentBots.length === 0 && (
                <div className="text-center text-sm text-muted-foreground py-8">
                  Henüz bir bot oluşturmadınız.
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default DashboardPage
