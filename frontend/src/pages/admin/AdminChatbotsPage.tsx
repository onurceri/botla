/**
 * AdminChatbotsPage - Chatbot management page
 * Lists all chatbots across the platform with management capabilities
 */
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Search, Bot, RefreshCw, MessageSquare, Database, MoreHorizontal } from 'lucide-react'
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

export function AdminChatbotsPage() {
  const [search, setSearch] = useState('')
  const [offset, setOffset] = useState(0)
  const limit = 20

  const queryClient = useQueryClient()
  const { toast } = useToast()

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin', 'chatbots', { search, offset, limit }],
    queryFn: () =>
      adminApi.listChatbots({
        name: search || undefined,
        limit,
        offset,
      }),
  })

  const forceRefreshMutation = useMutation({
    mutationFn: adminApi.forceRefreshChatbot,
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'chatbots'] })
      toast(
        `${result.sources_reset} kaynak sıfırlandı, ${result.sources_queued} kaynak kuyruğa eklendi.`,
        'success'
      )
    },
    onError: () => {
      toast('Chatbot yenilenirken bir hata oluştu.', 'error')
    },
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setOffset(0)
  }

  const chatbots = data?.chatbots ?? []
  const total = data?.total ?? 0
  const hasNextPage = offset + limit < total
  const hasPrevPage = offset > 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Chatbotlar</h1>
        <p className="text-muted-foreground">
          Platform genelindeki tüm chatbotları görüntüle ve yönet. Toplam: {total}
        </p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <form onSubmit={handleSearch} className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Chatbot adı ile ara..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </form>
      </div>

      {/* Chatbots Table */}
      <Card>
        <CardHeader className="pb-3 border-b">
          <CardTitle className="text-sm font-medium">Chatbot Listesi</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-8 text-center text-muted-foreground">Yükleniyor...</div>
          ) : error ? (
            <div className="p-8 text-center text-destructive">Hata: {(error as Error).message}</div>
          ) : chatbots.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground">Chatbot bulunamadı.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-muted/50 text-muted-foreground">
                  <tr className="text-left">
                    <th className="px-4 py-3 font-medium">Chatbot</th>
                    <th className="px-4 py-3 font-medium">Sahip</th>
                    <th className="px-4 py-3 font-medium">Organizasyon</th>
                    <th className="px-4 py-3 font-medium">İstatistikler</th>
                    <th className="px-4 py-3 font-medium">Oluşturulma</th>
                    <th className="px-4 py-3 font-medium text-right">İşlemler</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {chatbots.map((chatbot) => (
                    <tr key={chatbot.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
                            <Bot className="w-4 h-4 text-primary" />
                          </div>
                          <span className="font-medium">{chatbot.name}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {chatbot.owner_email}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {chatbot.organization_name || '-'}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-4 text-muted-foreground">
                          <div className="flex items-center gap-1.5" title="Kaynak sayısı">
                            <Database className="w-4 h-4" />
                            <span className="font-medium text-foreground">{chatbot.source_count}</span>
                          </div>
                          <div className="flex items-center gap-1.5" title="Mesaj sayısı">
                            <MessageSquare className="w-4 h-4" />
                            <span className="font-medium text-foreground">{chatbot.message_count}</span>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {formatDistanceToNow(new Date(chatbot.created_at), {
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
                            <DropdownMenuItem
                              onClick={() => forceRefreshMutation.mutate(chatbot.id)}
                              disabled={forceRefreshMutation.isPending}
                            >
                              <RefreshCw
                                className={`w-4 h-4 mr-2 ${
                                  forceRefreshMutation.isPending ? 'animate-spin' : ''
                                }`}
                              />
                              Zorla Yenile
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
              {offset + 1} - {Math.min(offset + limit, total)} / {total} chatbot
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
