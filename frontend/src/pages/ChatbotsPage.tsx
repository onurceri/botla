import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Bot, MoreHorizontal, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from '@/components/ui/card'
import { useOrganization } from '@/features/organization/context/OrganizationContext'
import { useChatbots } from '@/hooks/queries/useChatbots'

const ChatbotsPage = () => {
  const { currentWorkspace, isLoading: isOrgLoading } = useOrganization()
  const [openMenuId, setOpenMenuId] = useState<string | number | null>(null)

  // Use React Query for chatbots list
  const {
    data: bots = [],
    isLoading: botsLoading,
    error: botsError,
  } = useChatbots(!!currentWorkspace && !isOrgLoading)
  // console.log('ChatbotsPage render:', { botsLoading, isOrgLoading, hasWorkspace: !!currentWorkspace, botsLength: bots.length })

  useEffect(() => {
    if (botsError) {
      console.error(botsError)
    }
  }, [botsError])

  useEffect(() => {
    const onDocClick = (e: MouseEvent) => {
      if (!openMenuId) return
      const menuEl = document.querySelector(`[data-menu="${openMenuId}"]`)
      const triggerEl = document.querySelector(`[data-menu-trigger="${openMenuId}"]`)
      const target = e.target as Node
      if (menuEl && triggerEl) {
        if (!menuEl.contains(target) && !triggerEl.contains(target)) {
          setOpenMenuId(null)
        }
      } else {
        setOpenMenuId(null)
      }
    }
    document.addEventListener('mousedown', onDocClick)
    return () => document.removeEventListener('mousedown', onDocClick)
  }, [openMenuId])

  const handleMenuToggle = (id: string | number) => {
    setOpenMenuId((curr) => (curr === id ? null : id))
  }

  const handleDelete = async (id: string | number) => {
    try {
      await api.delete(`/api/v1/chatbots/${id}`)
      // Note: In a full migration, we'd use a mutation hook here that invalidates the chatbots query
      // For now, we'll rely on manual refetch or optimistic update
    } catch (err) {
      console.error(err)
    } finally {
      setOpenMenuId(null)
    }
  }

  const loading = isOrgLoading || botsLoading

  if (isOrgLoading || (loading && currentWorkspace)) {
    return (
      <div className="space-y-8">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-2">
            <div className="h-8 w-48 bg-muted animate-pulse rounded-md" />
            <div className="h-4 w-64 bg-muted animate-pulse rounded-md" />
          </div>
          <div className="h-10 w-32 bg-muted animate-pulse rounded-md" />
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader className="pb-2">
                <div className="w-12 h-12 bg-muted rounded-xl" />
              </CardHeader>
              <CardContent>
                <div className="h-6 w-3/4 bg-muted rounded mb-2" />
                <div className="h-4 w-full bg-muted rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  if (!currentWorkspace) {
    return (
      <div className="flex flex-col items-center justify-center h-[50vh] space-y-4">
        <h2 className="text-2xl font-semibold">Çalışma Alanı Seçin</h2>
        <p className="text-muted-foreground">
          Chatbotlarınızı görüntülemek için lütfen bir çalışma alanı seçin veya oluşturun.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Chatbotlarım</h1>
          <p className="text-muted-foreground">Tüm yapay zeka asistanlarınızı buradan yönetin.</p>
        </div>
        <Link to="/dashboard/chatbots/new">
          <Button className="gap-2 shadow-lg shadow-primary/20">
            <Plus className="w-4 h-4" /> Yeni Oluştur
          </Button>
        </Link>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {bots.map((bot: any) => (
          <Card
            key={bot.id}
            className="group hover:shadow-xl hover:-translate-y-1 hover:border-primary/50 transition-all duration-300"
          >
            <CardHeader className="relative flex flex-row items-start justify-between space-y-0 pb-2">
              <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center text-primary border border-primary/10 group-hover:scale-110 transition-transform duration-300">
                <Bot className="w-6 h-6" />
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 -mr-2 text-muted-foreground"
                onClick={() => handleMenuToggle(bot.id)}
                data-menu-trigger={bot.id}
              >
                <MoreHorizontal className="w-4 h-4" />
              </Button>
              {openMenuId === bot.id && (
                <div
                  className="absolute z-50 right-0 top-10 min-w-[160px] rounded-md border bg-popover text-popover-foreground shadow-md"
                  data-menu={bot.id}
                >
                  <div className="p-1">
                    <button
                      className="flex w-full items-center gap-2 rounded-sm px-2 py-2 text-sm text-destructive hover:bg-destructive/10"
                      onClick={() => handleDelete(bot.id)}
                    >
                      <Trash2 className="h-4 w-4" /> Sil
                    </button>
                  </div>
                </div>
              )}
            </CardHeader>
            <CardContent className="pt-4">
              <CardTitle className="text-xl mb-2">{bot.name}</CardTitle>
              <CardDescription className="line-clamp-2 min-h-[2.5rem]">
                {bot.description || 'Açıklama yok.'}
              </CardDescription>

              <div className="mt-4 flex flex-wrap gap-2">
                <span className="inline-flex items-center rounded-full border border-border bg-muted px-2.5 py-0.5 text-xs font-semibold text-muted-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
                  {bot.model}
                </span>
                <span className="inline-flex items-center rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2.5 py-0.5 text-xs font-semibold text-emerald-600">
                  Aktif
                </span>
              </div>
            </CardContent>
            <CardFooter className="pt-0">
              <Link to={`/dashboard/chatbots/${bot.id}`} className="w-full">
                <Button
                  variant="outline"
                  className="w-full group-hover:bg-primary group-hover:text-primary-foreground group-hover:border-primary transition-all"
                >
                  Yönet
                </Button>
              </Link>
            </CardFooter>
          </Card>
        ))}

        {/* Empty State / Create New Card */}
        <Link
          to="/dashboard/chatbots/new"
          className="group relative flex flex-col items-center justify-center gap-4 rounded-xl border-2 border-dashed border-border bg-muted/30 p-8 text-center hover:border-primary/50 hover:bg-muted/50 transition-all duration-300"
        >
          <div className="rounded-full bg-muted p-4 group-hover:scale-110 transition-transform duration-300">
            <Plus className="h-8 w-8 text-muted-foreground group-hover:text-primary" />
          </div>
          <div className="space-y-1">
            <h3 className="font-semibold text-foreground">Yeni Chatbot Ekle</h3>
            <p className="text-sm text-muted-foreground">Özelleştirilmiş bir asistan oluşturun</p>
          </div>
        </Link>
      </div>
    </div>
  )
}

export default ChatbotsPage
