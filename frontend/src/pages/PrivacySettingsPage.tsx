import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { useToast } from '@/components/ui/toast'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Loader2, Download, Trash2, Shield, Edit3 } from 'lucide-react'

interface Consents {
  marketing: boolean
  analytics: boolean
  personalization: boolean
  third_party: boolean
}

interface PrivacyRequest {
  id: string
  request_type: string
  status: string
  export_url?: string | null
  export_expires_at?: string | null
  denial_reason?: string | null
  created_at: string
}

export default function PrivacySettingsPage() {
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const [deleteReason, setDeleteReason] = useState('')
  const [correctionReason, setCorrectionReason] = useState('')
  const [exportRequestId, setExportRequestId] = useState<string | null>(null)

  // 1. Get Consents
  const { data: consents, isLoading: consentsLoading } = useQuery<Consents>({
    queryKey: ['privacy', 'consents'],
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me/privacy/consents')
      return data
    },
  })

  // 2. Update Consents
  const updateConsentMutation = useMutation({
    mutationFn: async (vars: Partial<Consents>) => {
      await api.patch('/api/v1/me/privacy/consents', vars)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['privacy', 'consents'] })
      toast('Gizlilik tercihleriniz güncellendi.', 'success')
    },
    onError: () => {
      toast('Tercihler güncellenirken bir hata oluştu.', 'error')
    },
  })

  // 3. Data Export
  const exportMutation = useMutation({
    mutationFn: async () => {
      const { data } = await api.post('/api/v1/me/privacy/export')
      return data
    },
    onSuccess: (data: PrivacyRequest) => {
      setExportRequestId(data.id)
      toast('Veri dışa aktarma talebiniz alındı.', 'success')
    },
    onError: () => {
      toast('Dışa aktarma talebi oluşturulurken bir hata oluştu.', 'error')
    },
  })

  const exportRequestQuery = useQuery<PrivacyRequest>({
    queryKey: ['privacy', 'export-request', exportRequestId],
    queryFn: async () => {
      const { data } = await api.get(`/api/v1/me/privacy/requests/${exportRequestId}`)
      return data
    },
    enabled: !!exportRequestId,
    refetchInterval: (query) => {
      const status = (query.state.data as PrivacyRequest | undefined)?.status
      if (status === 'completed' || status === 'denied') return false
      return 2000
    },
  })
  const exportRequest = exportRequestQuery.data

  const downloadExport = async () => {
    if (!exportRequestId) return
    try {
      const res = await api.get(`/api/v1/me/privacy/requests/${exportRequestId}/download`, {
        responseType: 'blob',
      })

      let filename = 'export.json'
      const cd = res.headers['content-disposition']
      if (typeof cd === 'string') {
        const m = /filename="([^"]+)"/.exec(cd)
        if (m?.[1]) filename = m[1]
      }

      const url = window.URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
    } catch {
      toast('Dışa aktarma indirme sırasında bir hata oluştu.', 'error')
    }
  }

  // 4. Data Correction
  const correctionMutation = useMutation({
    mutationFn: async () => {
      await api.post('/api/v1/me/privacy/correction', { reason: correctionReason })
    },
    onSuccess: () => {
      toast('Veri düzeltme talebiniz alındı.', 'success')
      setCorrectionReason('')
    },
    onError: () => {
      toast('Düzeltme talebi oluşturulurken bir hata oluştu.', 'error')
    },
  })

  // 5. Account Deletion
  const deleteAccountMutation = useMutation({
    mutationFn: async () => {
      await api.post('/api/v1/me/privacy/delete', { reason: deleteReason || 'User requested deletion' })
    },
    onSuccess: () => {
      toast('Hesap silme talebiniz alındı.', 'success')
    },
    onError: () => {
      toast('Hesap silme talebi oluşturulurken bir hata oluştu.', 'error')
    },
  })

  const handleConsentChange = (key: keyof Consents, value: boolean) => {
    updateConsentMutation.mutate({ [key]: value })
  }

  if (consentsLoading) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="container max-w-3xl py-10 space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold">Gizlilik Ayarları</h1>
        <p className="text-muted-foreground">
          Veri işleme izinlerinizi, veri dışa aktarma ve hesap silme işlemlerinizi buradan yönetebilirsiniz.
        </p>
      </div>

      {/* 1. Consents */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="w-5 h-5" />
            İzinler ve Tercihler
          </CardTitle>
          <CardDescription>
            Hangi verilerinizin nasıl işleneceğine karar verin.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between space-x-2">
            <Label htmlFor="marketing" className="flex flex-col space-y-1">
              <span>Pazarlama İletişimi</span>
              <span className="font-normal text-xs text-muted-foreground">
                Kampanya ve duyurulardan haberdar olmak istiyorum.
              </span>
            </Label>
            <Switch
              id="marketing"
              checked={consents?.marketing}
              onCheckedChange={(checked) => handleConsentChange('marketing', checked)}
            />
          </div>
          <div className="flex items-center justify-between space-x-2">
            <Label htmlFor="analytics" className="flex flex-col space-y-1">
              <span>Analitik Veriler</span>
              <span className="font-normal text-xs text-muted-foreground">
                Hizmet kalitesini artırmak için anonim kullanım verilerimi paylaş.
              </span>
            </Label>
            <Switch
              id="analytics"
              checked={consents?.analytics}
              onCheckedChange={(checked) => handleConsentChange('analytics', checked)}
            />
          </div>
          <div className="flex items-center justify-between space-x-2">
            <Label htmlFor="personalization" className="flex flex-col space-y-1">
              <span>Kişiselleştirme</span>
              <span className="font-normal text-xs text-muted-foreground">
                Bana özel içerik ve öneriler sunulmasını kabul ediyorum.
              </span>
            </Label>
            <Switch
              id="personalization"
              checked={consents?.personalization}
              onCheckedChange={(checked) => handleConsentChange('personalization', checked)}
            />
          </div>
          <div className="flex items-center justify-between space-x-2">
            <Label htmlFor="third_party" className="flex flex-col space-y-1">
              <span>Üçüncü Taraf Paylaşımı</span>
              <span className="font-normal text-xs text-muted-foreground">
                Verilerimin iş ortakları ve üçüncü taraflarla paylaşılmasını kabul ediyorum.
              </span>
            </Label>
            <Switch
              id="third_party"
              checked={consents?.third_party}
              onCheckedChange={(checked) => handleConsentChange('third_party', checked)}
            />
          </div>
        </CardContent>
      </Card>

      {/* 2. Data Export */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Download className="w-5 h-5" />
            Veri Dışa Aktarma
          </CardTitle>
          <CardDescription>
            Tüm kişisel verilerinizin bir kopyasını indirin. İşlem tamamlandığında size e-posta iletilecektir.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant="outline"
            onClick={() => exportMutation.mutate()}
            disabled={exportMutation.isPending}
          >
            {exportMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Hazırlanıyor...
              </>
            ) : (
              'Verilerimi İndir'
            )}
          </Button>

          {exportRequestId && (
            <div className="mt-4 space-y-2 text-sm text-muted-foreground">
              <div>Durum: {exportRequest?.status ?? 'pending'}</div>
              {exportRequest?.status === 'completed' && (
                <Button variant="secondary" onClick={downloadExport}>
                  İndirmeyi Başlat
                </Button>
              )}
              {exportRequest?.status === 'denied' && (
                <div>Talebiniz reddedildi.</div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 3. Data Correction */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Edit3 className="w-5 h-5" />
            Veri Düzeltme
          </CardTitle>
          <CardDescription>
            Kişisel verilerinizde bir yanlışlık olduğunu düşünüyorsanız düzeltme talebinde bulunun.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="correction-reason">Düzeltme Detayları</Label>
            <Textarea
              id="correction-reason"
              placeholder="Hangi verinin nasıl düzeltilmesini istediğinizi açıklayın..."
              value={correctionReason}
              onChange={(e) => setCorrectionReason(e.target.value)}
            />
          </div>
          <Button
            variant="outline"
            onClick={() => correctionMutation.mutate()}
            disabled={correctionMutation.isPending || !correctionReason.trim()}
          >
            {correctionMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Gönderiliyor...
              </>
            ) : (
              'Düzeltme Talebi Gönder'
            )}
          </Button>
        </CardContent>
      </Card>

      {/* 4. Delete Account */}
      <Card className="border-destructive/50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-destructive">
            <Trash2 className="w-5 h-5" />
            Hesabı Sil
          </CardTitle>
          <CardDescription>
            Hesabınızı ve tüm verilerinizi kalıcı olarak silin. Bu işlem geri alınamaz.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive">Hesabımı Sil</Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Emin misiniz?</AlertDialogTitle>
                <AlertDialogDescription>
                  Bu işlem geri alınamaz. Hesabınız ve tüm verileriniz sunucularımızdan kalıcı olarak silinecektir.
                </AlertDialogDescription>
                <div className="mt-4">
                  <Label htmlFor="reason" className="text-right">
                    Silme Nedeni (İsteğe bağlı)
                  </Label>
                  <Textarea
                    id="reason"
                    placeholder="Bize neden ayrıldığınızı söylemek ister misiniz?"
                    value={deleteReason}
                    onChange={(e) => setDeleteReason(e.target.value)}
                    className="mt-2"
                  />
                </div>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>İptal</AlertDialogCancel>
                <AlertDialogAction
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  onClick={() => deleteAccountMutation.mutate()}
                >
                  {deleteAccountMutation.isPending ? 'Siliniyor...' : 'Evet, Hesabımı Sil'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </CardContent>
      </Card>
    </div>
  )
}
