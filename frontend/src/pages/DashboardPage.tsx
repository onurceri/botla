import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { 
  BarChart, 
  Bar, 
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
  Bot
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api } from '@/api/client'
import { getAnalytics } from '@/api/analytics'

const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    return (
      <div className="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 p-3 rounded-xl shadow-xl">
        <p className="text-sm font-medium text-slate-500 dark:text-slate-400 mb-2">
          {new Date(label).toLocaleDateString('tr-TR', { day: 'numeric', month: 'long', year: 'numeric' })}
        </p>
        <div className="space-y-1">
          {payload.map((entry: any, index: number) => (
            <div key={index} className="flex items-center gap-2 text-sm">
              <div 
                className="w-2 h-2 rounded-full" 
                style={{ backgroundColor: entry.color }}
              />
              <span className="font-medium text-slate-700 dark:text-slate-200">
                {entry.name}:
              </span>
              <span className="font-bold text-slate-900 dark:text-white">
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

const DashboardPage = () => {
  const [stats, setStats] = useState({
    totalConversations: 0,
    totalMessages: 0,
    activeBots: 0
  })
  const [chartData, setChartData] = useState<any[]>([])
  const [recentBots, setRecentBots] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch Analytics
        const analyticsData = await getAnalytics()
        setChartData(analyticsData)

        // Calculate totals from analytics
        const totalConv = analyticsData.reduce((acc: number, curr: any) => acc + curr.conversations, 0)
        const totalMsg = analyticsData.reduce((acc: number, curr: any) => acc + curr.messages, 0)

        // Fetch Bots for count and recent list
        const { data: bots } = await api.get('/api/v1/chatbots')
        setRecentBots(bots.slice(0, 3)) // Take first 3
        
        setStats({
          totalConversations: totalConv,
          totalMessages: totalMsg,
          activeBots: bots.length
        })
      } catch (error) {
        console.error('Failed to fetch dashboard data', error)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

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
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Toplam Konuşma</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalConversations}</div>
            <p className="text-xs text-muted-foreground mt-1 flex items-center">
              <span className="text-emerald-400 flex items-center mr-1">
                <ArrowUpRight className="w-3 h-3 mr-0.5" /> +12%
              </span>
              geçen haftaya göre
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
            <p className="text-xs text-muted-foreground mt-1 flex items-center">
              <span className="text-emerald-400 flex items-center mr-1">
                <ArrowUpRight className="w-3 h-3 mr-0.5" /> +8%
              </span>
              geçen haftaya göre
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Aktif Botlar</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.activeBots}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Sistemde kayıtlı asistanlarınız
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
            <div className="h-[300px]">
              {chartData.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
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
                      tickFormatter={(value) => {
                        const date = new Date(value);
                        return date.toLocaleDateString('tr-TR', { day: '2-digit', month: 'short' });
                      }}
                      dy={10}
                    />
                    <YAxis 
                      stroke="#94a3b8" 
                      fontSize={12} 
                      tickLine={false} 
                      axisLine={false} 
                      tickFormatter={(value) => `${value}`}
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
                <div key={bot.id} className="flex items-center gap-4 p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors">
                  <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/10">
                    <Bot className="w-5 h-5" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium leading-none truncate text-foreground">{bot.name}</p>
                    <p className="text-xs text-muted-foreground mt-1 truncate">{bot.model}</p>
                  </div>
                  <Link to={`/chatbots/${bot.id}`}>
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                      <ArrowUpRight className="w-4 h-4" />
                    </Button>
                  </Link>
                </div>
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
