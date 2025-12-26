/**
 * AdminUsersPage - User management page for platform admins
 * Lists all users with search, filter, and management capabilities
 */
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Search, MoreHorizontal, Shield, ShieldOff, UserX } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { tr } from 'date-fns/locale'
import * as adminApi from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/DropdownMenu'
import { useToast } from '@/components/ui/toast'
import { PlanBadge, normalizePlanId } from '@/components/ui/plan-badge'
import { StatusBadge } from '@/components/ui/status-badge'

export function AdminUsersPage() {
  const [search, setSearch] = useState('')
  const [planFilter, setPlanFilter] = useState('')
  const [offset, setOffset] = useState(0)
  const limit = 20

  const queryClient = useQueryClient()
  const { toast } = useToast()

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin', 'users', { search, planFilter, offset, limit }],
    queryFn: () =>
      adminApi.listUsers({
        email: search || undefined,
        plan_id: planFilter || undefined,
        limit,
        offset,
      }),
  })

  const updateUserMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Parameters<typeof adminApi.updateUser>[1] }) =>
      adminApi.updateUser(id, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
      toast('Kullanıcı güncellendi.', 'success')
    },
    onError: () => {
      toast('Kullanıcı güncellenirken hata oluştu.', 'error')
    },
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setOffset(0)
  }

  const toggleAdminStatus = (user: adminApi.AdminUser) => {
    updateUserMutation.mutate({
      id: user.id,
      updates: { is_platform_admin: !user.is_platform_admin },
    })
  }

  const users = data?.users ?? []
  const total = data?.total ?? 0
  const hasNextPage = offset + limit < total
  const hasPrevPage = offset > 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Kullanıcılar</h1>
        <p className="text-muted-foreground">
          Platformdaki tüm kullanıcıları görüntüle ve yönet. Toplam: {total}
        </p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <form onSubmit={handleSearch} className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="E-posta ile ara..."
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

      {/* Users Table */}
      <Card>
        <CardHeader className="pb-3 border-b">
          <CardTitle className="text-sm font-medium">Kullanıcı Listesi</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-8 text-center text-muted-foreground">Yükleniyor...</div>
          ) : error ? (
            <div className="p-8 text-center text-destructive">Hata: {(error as Error).message}</div>
          ) : users.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground">Kullanıcı bulunamadı.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-muted/50 text-muted-foreground">
                  <tr className="text-left">
                    <th className="px-4 py-3 font-medium">E-posta</th>
                    <th className="px-4 py-3 font-medium">İsim</th>
                    <th className="px-4 py-3 font-medium">Plan</th>
                    <th className="px-4 py-3 font-medium">Durum</th>
                    <th className="px-4 py-3 font-medium">Kayıt</th>
                    <th className="px-4 py-3 font-medium text-right">İşlemler</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {users.map((user) => (
                    <tr key={user.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{user.email}</span>
                          {user.is_platform_admin && (
                            <span className="px-1.5 py-0.5 text-[10px] bg-primary/10 text-primary rounded-full flex items-center gap-1 font-medium">
                              <Shield className="w-3 h-3" />
                              Admin
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {user.full_name || '-'}
                      </td>
                      <td className="px-4 py-3">
                        <PlanBadge 
                          plan={normalizePlanId(user.plan_id)} 
                          size="sm" 
                          variant="soft" 
                        />
                      </td>
                      <td className="px-4 py-3">
                        <StatusBadge status="active" size="sm" />
                      </td>
                      <td className="px-4 py-3 text-sm text-muted-foreground">
                        {formatDistanceToNow(new Date(user.created_at), {
                          addSuffix: true,
                          locale: tr,
                        })}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreHorizontal className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => toggleAdminStatus(user)}>
                              {user.is_platform_admin ? (
                                <>
                                  <ShieldOff className="w-4 h-4 mr-2" />
                                  Admin Yetkisini Kaldır
                                </>
                              ) : (
                                <>
                                  <Shield className="w-4 h-4 mr-2" />
                                  Admin Yap
                                </>
                              )}
                            </DropdownMenuItem>
                            <DropdownMenuItem className="text-destructive">
                              <UserX className="w-4 h-4 mr-2" />
                              Hesabı Askıya Al
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
        {/* Pagination */}
        {total > limit && (
          <div className="p-4 border-t flex items-center justify-between">
            <span className="text-xs text-muted-foreground">
              {offset + 1} - {Math.min(offset + limit, total)} / {total} kullanıcı
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
      </Card>
    </div>
  )
}
