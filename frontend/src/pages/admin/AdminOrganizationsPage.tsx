/**
 * AdminOrganizationsPage - Organization management page
 * Lists all organizations on the platform
 */
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, Building2, Users, Bot } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { tr } from 'date-fns/locale'
import * as adminApi from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PlanBadge, normalizePlanId } from '@/components/ui/plan-badge'

export function AdminOrganizationsPage() {
  const [search, setSearch] = useState('')
  const [planFilter, setPlanFilter] = useState('')
  const [offset, setOffset] = useState(0)
  const limit = 20

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin', 'organizations', { search, planFilter, offset, limit }],
    queryFn: () =>
      adminApi.listOrganizations({
        name: search || undefined,
        plan_id: planFilter || undefined,
        limit,
        offset,
      }),
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setOffset(0)
  }

  const organizations = data?.organizations ?? []
  const total = data?.total ?? 0
  const hasNextPage = offset + limit < total
  const hasPrevPage = offset > 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Organizasyonlar</h1>
        <p className="text-muted-foreground">
          Platformdaki tüm organizasyonları görüntüle ve yönet. Toplam: {total}
        </p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <form onSubmit={handleSearch} className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Organizasyon adı ile ara..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </form>
        <select
          value={planFilter}
          onChange={(e) => {
            setPlanFilter(e.target.value)
            setOffset(0)
          }}
          className="px-4 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
        >
          <option value="">Tüm Planlar</option>
          <option value="free">Free</option>
          <option value="pro">Pro</option>
          <option value="business">Business</option>
        </select>
      </div>

      {/* Organizations Grid */}
      {isLoading ? (
        <div className="p-8 text-center text-muted-foreground">Yükleniyor...</div>
      ) : error ? (
        <div className="p-8 text-center text-destructive">
          Hata: {(error as Error).message}
        </div>
      ) : organizations.length === 0 ? (
        <div className="p-8 text-center text-muted-foreground">
          Organizasyon bulunamadı.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {organizations.map((org) => (
            <Card key={org.id} className="hover:border-primary/50 transition-colors">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                      <Building2 className="w-5 h-5 text-primary" />
                    </div>
                    <div>
                      <CardTitle className="text-base">{org.name}</CardTitle>
                      <p className="text-sm text-muted-foreground">@{org.slug}</p>
                    </div>
                  </div>
                  <PlanBadge 
                    plan={normalizePlanId(org.plan_id)} 
                    size="sm" 
                    variant="soft" 
                  />
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="flex items-center gap-4 text-sm text-muted-foreground mb-4">
                  <div className="flex items-center gap-1">
                    <Users className="w-4 h-4" />
                    <span>{org.user_count ?? 0}</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Bot className="w-4 h-4" />
                    <span>{org.chatbot_count ?? 0}</span>
                  </div>
                </div>

                <div className="flex items-center justify-between pt-3 border-t border-border">
                  <span className="text-xs text-muted-foreground">
                    {formatDistanceToNow(new Date(org.created_at), {
                      addSuffix: true,
                      locale: tr,
                    })}
                  </span>
                  <Button variant="ghost" size="sm">
                    Detaylar
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Pagination */}
      {total > limit && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            {offset + 1} - {Math.min(offset + limit, total)} / {total} organizasyon
          </span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setOffset(Math.max(0, offset - limit))}
              disabled={!hasPrevPage}
            >
              Önceki
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setOffset(offset + limit)}
              disabled={!hasNextPage}
            >
              Sonraki
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
