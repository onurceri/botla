import { useQuery } from '@tanstack/react-query'
import { Users, Building2, Bot, MessageSquare } from 'lucide-react'
import { getOverviewStats } from '@/api/admin'
import { StatsCard } from '@/features/admin/components/StatsCard'
import { HealthPanel } from '@/features/admin/components/HealthPanel'

/**
 * AdminDashboardPage - Main overview page for admin dashboard
 * Shows platform statistics and health status
 */
export function AdminDashboardPage() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['admin', 'stats'],
    queryFn: () => getOverviewStats(),
  })

  // Format numbers for display
  const formatNumber = (val: number | undefined) => (val ?? 0).toLocaleString()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Genel Bakış</h1>
        <p className="text-muted-foreground">
          Platform genel istatistikleri ve sistem durumu.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Toplam Kullanıcı"
          value={formatNumber(stats?.total_users)}
          subtitle={stats?.users_today ? `Bugün +${stats.users_today}` : undefined}
          icon={<Users className="w-5 h-5" />}
          isLoading={isLoading}
        />
        <StatsCard
          title="Organizasyonlar"
          value={formatNumber(stats?.total_organizations)}
          icon={<Building2 className="w-5 h-5" />}
          isLoading={isLoading}
        />
        <StatsCard
          title="Chatbotlar"
          value={formatNumber(stats?.total_chatbots)}
          icon={<Bot className="w-5 h-5" />}
          isLoading={isLoading}
        />
        <StatsCard
          title="Toplam Mesaj"
          value={formatNumber(stats?.total_messages)}
          subtitle={stats?.conversations_today ? `Bugün +${stats.conversations_today}` : undefined}
          icon={<MessageSquare className="w-5 h-5" />}
          isLoading={isLoading}
        />
      </div>

      {/* Panels Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <HealthPanel />
        <div className="bg-card rounded-xl border border-border p-6 flex flex-col items-center justify-center text-center opacity-50">
          <p className="text-sm font-medium text-muted-foreground">Son Aktiviteler</p>
          <p className="text-xs text-muted-foreground mt-1">Yakında</p>
        </div>
      </div>
    </div>
  )
}
