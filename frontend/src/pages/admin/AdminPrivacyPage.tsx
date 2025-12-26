import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { 
  listPrivacyRequests, 
  processPrivacyRequest, 
  PrivacyRequest,
  getPrivacyDownloadURL 
} from '@/api/admin'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { useToast } from '@/components/ui/toast'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { format } from 'date-fns'
import { Loader2, Check, X, Download, FileText, Trash2, Edit } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { StatusBadge } from '@/components/ui/status-badge'

// Request type badge config
const requestTypeConfig: Record<string, { label: string; icon: React.ReactNode; className: string }> = {
  export: {
    label: 'Veri Aktarımı',
    icon: <Download className="w-3.5 h-3.5" />,
    className: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  },
  deletion: {
    label: 'Hesap Silme',
    icon: <Trash2 className="w-3.5 h-3.5" />,
    className: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  },
  correction: {
    label: 'Veri Düzeltme',
    icon: <Edit className="w-3.5 h-3.5" />,
    className: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
  },
}

export function AdminPrivacyPage() {
  const [status, setStatus] = useState<string>('pending')
  const [rejectDialog, setRejectDialog] = useState<{ id: string; open: boolean }>({
    id: '',
    open: false,
  })
  const [rejectReason, setRejectReason] = useState('')
  
  const { toast } = useToast()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'privacy', { status }],
    queryFn: () => listPrivacyRequests({ status: status === 'all' ? undefined : status }),
  })

  const processMutation = useMutation({
    mutationFn: async ({ id, action, reason }: { id: string; action: 'approve' | 'deny'; reason?: string }) => {
      await processPrivacyRequest(id, action, reason)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'privacy'] })
      toast('İşlem başarılı.', 'success')
      setRejectDialog({ id: '', open: false })
      setRejectReason('')
    },
    onError: () => {
      toast('İşlem sırasında bir hata oluştu.', 'error')
    },
  })

  const handleApprove = (id: string) => {
    if (confirm('Bu talebi onaylamak istediğinize emin misiniz?')) {
      processMutation.mutate({ id, action: 'approve' })
    }
  }

  const handleDownload = async (id: string) => {
    try {
      const { url } = await getPrivacyDownloadURL(id)
      window.open(url, '_blank')
    } catch (error) {
      console.error('Failed to get download URL:', error)
      alert('İndirme bağlantısı alınamadı.')
    }
  }

  const handleRejectClick = (id: string) => {
    setRejectDialog({ id, open: true })
  }

  const confirmReject = () => {
    if (!rejectReason) {
      toast('Lütfen bir ret nedeni girin.', 'error')
      return
    }
    processMutation.mutate({ id: rejectDialog.id, action: 'deny', reason: rejectReason })
  }

  const getRequestTypeBadge = (type: string) => {
    const config = requestTypeConfig[type] || requestTypeConfig['export']
    return (
      <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 text-xs font-medium rounded-full border ${config.className}`}>
        {config.icon}
        {config.label}
      </span>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Gizlilik Talepleri</h1>
          <p className="text-muted-foreground">
            Kullanıcıların veri silme ve dışa aktarma taleplerini yönetin.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Durum Seçin" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="pending">Bekleyenler</SelectItem>
              <SelectItem value="processing">İşlenenler</SelectItem>
              <SelectItem value="completed">Tamamlananlar</SelectItem>
              <SelectItem value="denied">Reddedilenler</SelectItem>
              <SelectItem value="all">Tümü</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <Card>
        <CardHeader className="pb-3 border-b">
          <CardTitle className="text-sm font-medium">Talep Listesi</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/50 hover:bg-muted/50">
                <TableHead>Tarih</TableHead>
                <TableHead>Kullanıcı</TableHead>
                <TableHead>Talep Tipi</TableHead>
                <TableHead>Açıklama</TableHead>
                <TableHead>Durum</TableHead>
                <TableHead>İşlem Bilgisi</TableHead>
                <TableHead className="text-right">İşlemler</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={7} className="h-24 text-center">
                    <Loader2 className="mx-auto h-6 w-6 animate-spin text-muted-foreground" />
                  </TableCell>
                </TableRow>
              ) : data?.data?.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="h-24 text-center">
                    <div className="flex flex-col items-center justify-center gap-2">
                      <FileText className="w-8 h-8 text-muted-foreground" />
                      <span className="text-muted-foreground">Kayıt bulunamadı.</span>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                data?.data?.map((request: PrivacyRequest) => (
                  <TableRow key={request.id}>
                    <TableCell className="text-muted-foreground">
                      {format(new Date(request.created_at), 'dd.MM.yyyy HH:mm')}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="font-medium">{request.user_email}</span>
                        <span className="font-mono text-[10px] text-muted-foreground">{request.user_id}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {getRequestTypeBadge(request.request_type)}
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">
                      {request.reason || '-'}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-1">
                        <StatusBadge status={request.status} size="sm" />
                        {request.status === 'denied' && request.denial_reason && (
                          <span className="text-[10px] text-destructive italic max-w-[150px] truncate">
                            Ret: {request.denial_reason}
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-1 text-[11px] text-muted-foreground">
                        {request.processed_at && (
                          <span>
                            İşlem: {format(new Date(request.processed_at), 'dd.MM HH:mm')}
                          </span>
                        )}
                        {request.completed_at && (
                          <span>
                            Tamam: {format(new Date(request.completed_at), 'dd.MM HH:mm')}
                          </span>
                        )}
                        {request.export_url && request.status === 'completed' && (
                          <button 
                            onClick={() => handleDownload(request.id)}
                            className="text-primary hover:underline font-medium text-left"
                          >
                            Veriyi İndir
                          </button>
                        )}
                        {!request.processed_at && !request.completed_at && '-'}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      {request.status === 'pending' && (
                        <div className="flex justify-end gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            className="h-8 w-8 p-0 text-green-600 hover:bg-green-50 hover:text-green-700 dark:hover:bg-green-900/20"
                            onClick={() => handleApprove(request.id)}
                            disabled={processMutation.isPending}
                          >
                            <Check className="h-4 w-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            className="h-8 w-8 p-0 text-red-600 hover:bg-red-50 hover:text-red-700 dark:hover:bg-red-900/20"
                            onClick={() => handleRejectClick(request.id)}
                            disabled={processMutation.isPending}
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={rejectDialog.open} onOpenChange={(open) => setRejectDialog(prev => ({ ...prev, open }))}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Talebi Reddet</DialogTitle>
            <DialogDescription>
              Bu talebi neden reddettiğinizi belirtin. Bu açıklama kullanıcıya iletilebilir.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="reject-reason">Ret Nedeni</Label>
              <Textarea
                id="reject-reason"
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                placeholder="Örn: Mükerrer talep..."
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRejectDialog({ id: '', open: false })}>
              İptal
            </Button>
            <Button variant="destructive" onClick={confirmReject} disabled={processMutation.isPending}>
              {processMutation.isPending ? 'İşleniyor...' : 'Reddet'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
