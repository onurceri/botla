import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import Markdown from 'markdown-to-jsx'
import { getHandoffRequests, getHandoffRequestDetail, updateHandoffStatus, HandoffRequest, HandoffRequestDetail } from '@/api/handoff'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Inbox, Mail, Clock, Check, Copy, User, Bot, Loader2, Shield } from 'lucide-react'
import { useChatbotContext } from '../../context/ChatbotContext'

// Simple relative time formatter
function formatRelativeTime(date: Date): string {
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)
  
  if (diffMins < 1) return 'az önce'
  if (diffMins < 60) return `${diffMins} dakika önce`
  if (diffHours < 24) return `${diffHours} saat önce`
  if (diffDays < 7) return `${diffDays} gün önce`
  return date.toLocaleDateString('tr-TR')
}

const statusLabels: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' }> = {
  pending: { label: 'Bekliyor', variant: 'default' },
  assigned: { label: 'Atandı', variant: 'secondary' },
  resolved: { label: 'Çözüldü', variant: 'outline' },
}

export default function HandoffRequestsTab() {
  const { id: chatbotId } = useParams<{ id: string }>()
  const { planConfig } = useChatbotContext()
  const [requests, setRequests] = useState<HandoffRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedRequest, setSelectedRequest] = useState<HandoffRequestDetail | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [updating, setUpdating] = useState(false)

  const canUseHandoff = planConfig?.guardrails?.can_use_escalate_fallback

  const loadRequests = async () => {
    if (!chatbotId || !canUseHandoff) return
    setLoading(true)
    try {
      const data = await getHandoffRequests(chatbotId)
      setRequests(data)
    } catch (e) {
      console.error('Failed to load handoff requests', e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadRequests()
  }, [chatbotId, canUseHandoff])

  if (canUseHandoff === false) {
    return (
      <div className="flex flex-col items-center justify-center py-24 text-center space-y-6">
        <div className="p-6 bg-amber-500/10 rounded-full">
            <Shield className="w-12 h-12 text-amber-600" />
        </div>
        <div className="max-w-md space-y-2">
            <h3 className="text-xl font-bold text-foreground">Bu Özellik Planınızda Mevcut Değil</h3>
            <p className="text-muted-foreground">
                İnsan desteği taleplerini görüntülemek ve yönetmek için Enterprise plana geçiş yapmanız gerekmektedir.
            </p>
        </div>
        <Button className="bg-amber-600 hover:bg-amber-700 text-white">
            Planı Yükselt
        </Button>
      </div>
    )
  }

  const openDetail = async (request: HandoffRequest) => {
    if (!chatbotId) return
    setDetailLoading(true)
    setDetailOpen(true)
    try {
      const detail = await getHandoffRequestDetail(chatbotId, request.id)
      setSelectedRequest(detail)
    } catch (e) {
      console.error('Failed to load request detail', e)
    } finally {
      setDetailLoading(false)
    }
  }

  const handleStatusChange = async (status: string) => {
    if (!chatbotId || !selectedRequest) return
    setUpdating(true)
    try {
      await updateHandoffStatus(chatbotId, selectedRequest.request.id, status)
      // Update local state
      setSelectedRequest(prev => prev ? { ...prev, request: { ...prev.request, status: status as HandoffRequest['status'] } } : null)
      setRequests(prev => prev.map(r => r.id === selectedRequest.request.id ? { ...r, status: status as HandoffRequest['status'] } : r))
    } catch (e) {
      console.error('Failed to update status', e)
    } finally {
      setUpdating(false)
    }
  }

  const copyEmail = () => {
    if (selectedRequest?.request.user_email) {
      navigator.clipboard.writeText(selectedRequest.request.user_email)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Destek Talepleri</h2>
        <p className="text-muted-foreground">
          Kullanıcılardan gelen insan desteği taleplerini görüntüleyin ve yönetin.
        </p>
      </div>

      {requests.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <Inbox className="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 className="text-lg font-medium mb-2">Henüz destek talebi yok</h3>
            <p className="text-muted-foreground text-sm">
              Kullanıcılar insan desteği istediğinde burada görünecektir.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {requests.map((request) => (
            <Card 
              key={request.id} 
              className="cursor-pointer hover:shadow-md transition-shadow"
              onClick={() => openDetail(request)}
            >
              <CardContent className="p-4">
                <div className="flex items-center justify-between gap-4">
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="p-2 rounded-full bg-primary/10">
                      <Inbox className="h-4 w-4 text-primary" />
                    </div>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        {request.user_email ? (
                          <span className="font-medium flex items-center gap-1 truncate">
                            <Mail className="h-3.5 w-3.5" />
                            {request.user_email}
                          </span>
                        ) : (
                          <span className="text-muted-foreground text-sm">E-posta bekleniyor</span>
                        )}
                      </div>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Clock className="h-3 w-3" />
                        {formatRelativeTime(new Date(request.created_at))}
                      </div>
                    </div>
                  </div>
                  <Badge variant={statusLabels[request.status]?.variant || 'outline'}>
                    {statusLabels[request.status]?.label || request.status}
                  </Badge>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Detail Dialog */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Destek Talebi Detayı</DialogTitle>
            <DialogDescription>
              Konuşma geçmişi ve kullanıcı bilgilerini inceleyin.
            </DialogDescription>
          </DialogHeader>

          {detailLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : selectedRequest ? (
            <div className="space-y-6">
              {/* User Info */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium">Kullanıcı Bilgileri</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {selectedRequest.request.user_email ? (
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Mail className="h-4 w-4 text-muted-foreground" />
                        <span>{selectedRequest.request.user_email}</span>
                      </div>
                      <Button variant="ghost" size="sm" onClick={copyEmail}>
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">E-posta adresi henüz paylaşılmamış</p>
                  )}
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Clock className="h-4 w-4" />
                    <span>Talep: {new Date(selectedRequest.request.created_at).toLocaleString('tr-TR')}</span>
                  </div>
                </CardContent>
              </Card>

              {/* Status Update */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium">Durum</CardTitle>
                </CardHeader>
                <CardContent>
                  <Select 
                    value={selectedRequest.request.status} 
                    onValueChange={handleStatusChange}
                    disabled={updating}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="pending">
                        <div className="flex items-center gap-2">
                          <Clock className="h-4 w-4" /> Bekliyor
                        </div>
                      </SelectItem>
                      <SelectItem value="assigned">
                        <div className="flex items-center gap-2">
                          <User className="h-4 w-4" /> Atandı
                        </div>
                      </SelectItem>
                      <SelectItem value="resolved">
                        <div className="flex items-center gap-2">
                          <Check className="h-4 w-4" /> Çözüldü
                        </div>
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </CardContent>
              </Card>

              {/* Conversation */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium">Konuşma Geçmişi</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3 max-h-[300px] overflow-y-auto">
                    {(!selectedRequest.messages || selectedRequest.messages.length === 0) ? (
                      <p className="text-sm text-muted-foreground text-center py-4">
                        Konuşma geçmişi bulunamadı
                      </p>
                    ) : (
                      selectedRequest.messages.map((msg) => (
                        <div
                          key={msg.id}
                          className={`flex gap-3 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}
                        >
                          <div className={`p-1.5 rounded-full h-fit ${msg.role === 'user' ? 'bg-primary/10' : 'bg-muted'}`}>
                            {msg.role === 'user' ? (
                              <User className="h-3.5 w-3.5 text-primary" />
                            ) : (
                              <Bot className="h-3.5 w-3.5 text-muted-foreground" />
                            )}
                          </div>
                          <div
                            className={`rounded-lg px-3 py-2 max-w-[80%] ${
                              msg.role === 'user'
                                ? 'bg-primary text-primary-foreground'
                                : 'bg-muted'
                            }`}
                          >
                            <div className="text-sm markdown-content">
                              <Markdown>{msg.content}</Markdown>
                            </div>
                            <p className="text-xs opacity-60 mt-1">
                              {new Date(msg.created_at).toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit' })}
                            </p>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
    </div>
  )
}
