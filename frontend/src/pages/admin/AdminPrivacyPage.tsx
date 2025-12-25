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
import { Badge } from '@/components/ui/badge'
import { format } from 'date-fns'
import { Loader2, Check, X } from 'lucide-react'
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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Gizlilik Talepleri</h2>
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

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
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
                <TableCell colSpan={6} className="h-24 text-center">
                  <Loader2 className="mx-auto h-6 w-6 animate-spin text-muted-foreground" />
                </TableCell>
              </TableRow>
            ) : data?.data?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                  Kayıt bulunamadı.
                </TableCell>
              </TableRow>
            ) : (
              data?.data?.map((request: PrivacyRequest) => (
                <TableRow key={request.id}>
                  <TableCell>
                    {format(new Date(request.created_at), 'dd.MM.yyyy HH:mm')}
                  </TableCell>
                  <TableCell className="text-sm">
                    <div className="flex flex-col">
                      <span className="font-medium">{request.user_email}</span>
                      <span className="font-mono text-[10px] text-muted-foreground">{request.user_id}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {request.request_type === 'export' ? 'Veri Aktarımı' : 
                       request.request_type === 'deletion' ? 'Hesap Silme' : 'Veri Düzeltme'}
                    </Badge>
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">
                    {request.reason || '-'}
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-col gap-1">
                      <Badge
                        variant={
                          request.status === 'completed'
                            ? 'secondary'
                            : request.status === 'denied'
                            ? 'destructive'
                            : request.status === 'processing'
                            ? 'default'
                            : 'outline'
                        }
                      >
                        {request.status === 'completed' ? 'Tamamlandı' :
                         request.status === 'denied' ? 'Reddedildi' :
                         request.status === 'processing' ? 'İşleniyor' : 'Bekliyor'}
                      </Badge>
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
                          className="text-blue-600 hover:underline font-medium text-left"
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
                          className="h-8 w-8 p-0 text-green-600 hover:bg-green-50 hover:text-green-700"
                          onClick={() => handleApprove(request.id)}
                          disabled={processMutation.isPending}
                        >
                          <Check className="h-4 w-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          className="h-8 w-8 p-0 text-red-600 hover:bg-red-50 hover:text-red-700"
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
      </div>

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
