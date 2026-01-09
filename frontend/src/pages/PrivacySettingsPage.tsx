import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { AxiosError } from 'axios'
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
import { getStatusLabel, privacy } from '@/i18n/privacy'

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

interface PrivacyRequestsResponse {
  data: PrivacyRequest[]
  total: number
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

  // 2. Fetch existing privacy requests on page load
  const { data: existingRequests } = useQuery<PrivacyRequestsResponse>({
    queryKey: ['privacy', 'requests'],
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me/privacy/requests')
      return data
    },
  })

  // Find active export request from existing requests
  useEffect(() => {
    if (existingRequests?.data) {
      const activeExport = existingRequests.data.find(
        (req) =>
          req.request_type === 'export' &&
          (req.status === 'pending' || req.status === 'processing' || req.status === 'completed'),
      )
      if (activeExport && !exportRequestId) {
        setExportRequestId(activeExport.id)
      }
    }
  }, [existingRequests, exportRequestId])

  // Check if there's an active (pending/processing) export request
  const hasActiveExportRequest = existingRequests?.data?.some(
    (req) =>
      req.request_type === 'export' && (req.status === 'pending' || req.status === 'processing'),
  )

  // 3. Update Consents
  const updateConsentMutation = useMutation({
    mutationFn: async (vars: Partial<Consents>) => {
      await api.patch('/api/v1/me/privacy/consents', vars)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['privacy', 'consents'] })
      toast(privacy.toast.consentsUpdated, 'success')
    },
    onError: () => {
      toast(privacy.toast.consentsError, 'error')
    },
  })

  // 4. Data Export
  const exportMutation = useMutation({
    mutationFn: async () => {
      const { data } = await api.post('/api/v1/me/privacy/export')
      return data
    },
    onSuccess: (data: PrivacyRequest) => {
      setExportRequestId(data.id)
      queryClient.invalidateQueries({ queryKey: ['privacy', 'requests'] })
      toast(privacy.toast.exportRequested, 'success')
    },
    onError: (error: AxiosError) => {
      if (error.response?.status === 409) {
        toast(privacy.toast.exportActiveExists, 'error')
      } else {
        toast(privacy.toast.exportError, 'error')
      }
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
      toast(privacy.toast.downloadError, 'error')
    }
  }

  // 5. Data Correction
  const correctionMutation = useMutation({
    mutationFn: async () => {
      await api.post('/api/v1/me/privacy/correction', { reason: correctionReason })
    },
    onSuccess: () => {
      toast(privacy.toast.correctionRequested, 'success')
      setCorrectionReason('')
    },
    onError: () => {
      toast(privacy.toast.correctionError, 'error')
    },
  })

  // 6. Account Deletion
  const deleteAccountMutation = useMutation({
    mutationFn: async () => {
      await api.post('/api/v1/me/privacy/delete', {
        reason: deleteReason || 'User requested deletion',
      })
    },
    onSuccess: () => {
      toast(privacy.toast.deleteRequested, 'success')
    },
    onError: () => {
      toast(privacy.toast.deleteError, 'error')
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
        <h1 className="text-3xl font-bold">{privacy.page.title}</h1>
        <p className="text-muted-foreground">{privacy.page.description}</p>
      </div>

      {/* 1. Consents */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="w-5 h-5" />
            {privacy.consents.title}
          </CardTitle>
          <CardDescription>{privacy.consents.description}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between space-x-2">
            <Label htmlFor="marketing" className="flex flex-col space-y-1">
              <span>{privacy.consents.marketing.label}</span>
              <span className="font-normal text-xs text-muted-foreground">
                {privacy.consents.marketing.description}
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
              <span>{privacy.consents.analytics.label}</span>
              <span className="font-normal text-xs text-muted-foreground">
                {privacy.consents.analytics.description}
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
              <span>{privacy.consents.personalization.label}</span>
              <span className="font-normal text-xs text-muted-foreground">
                {privacy.consents.personalization.description}
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
              <span>{privacy.consents.thirdParty.label}</span>
              <span className="font-normal text-xs text-muted-foreground">
                {privacy.consents.thirdParty.description}
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
            {privacy.export.title}
          </CardTitle>
          <CardDescription>{privacy.export.description}</CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant="outline"
            onClick={() => exportMutation.mutate()}
            disabled={exportMutation.isPending || hasActiveExportRequest}
          >
            {exportMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {privacy.export.preparing}
              </>
            ) : (
              privacy.export.button
            )}
          </Button>

          {exportRequestId && (
            <div className="mt-4 space-y-2 text-sm text-muted-foreground">
              <div>
                {privacy.export.statusLabel}: {getStatusLabel(exportRequest?.status)}
              </div>
              {exportRequest?.status === 'completed' && (
                <Button variant="secondary" onClick={downloadExport}>
                  {privacy.export.downloadButton}
                </Button>
              )}
              {exportRequest?.status === 'denied' && <div>{privacy.export.denied}</div>}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 3. Data Correction */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Edit3 className="w-5 h-5" />
            {privacy.correction.title}
          </CardTitle>
          <CardDescription>{privacy.correction.description}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="correction-reason">{privacy.correction.label}</Label>
            <Textarea
              id="correction-reason"
              placeholder={privacy.correction.placeholder}
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
                {privacy.correction.sending}
              </>
            ) : (
              privacy.correction.button
            )}
          </Button>
        </CardContent>
      </Card>

      {/* 4. Delete Account */}
      <Card className="border-destructive/50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-destructive">
            <Trash2 className="w-5 h-5" />
            {privacy.delete.title}
          </CardTitle>
          <CardDescription>{privacy.delete.description}</CardDescription>
        </CardHeader>
        <CardContent>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive">{privacy.delete.button}</Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{privacy.delete.dialog.title}</AlertDialogTitle>
                <AlertDialogDescription>{privacy.delete.dialog.description}</AlertDialogDescription>
                <div className="mt-4">
                  <Label htmlFor="reason" className="text-right">
                    {privacy.delete.dialog.reasonLabel}
                  </Label>
                  <Textarea
                    id="reason"
                    placeholder={privacy.delete.dialog.reasonPlaceholder}
                    value={deleteReason}
                    onChange={(e) => setDeleteReason(e.target.value)}
                    className="mt-2"
                  />
                </div>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>{privacy.delete.dialog.cancel}</AlertDialogCancel>
                <AlertDialogAction
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  onClick={() => deleteAccountMutation.mutate()}
                >
                  {deleteAccountMutation.isPending
                    ? privacy.delete.dialog.deleting
                    : privacy.delete.dialog.confirmButton}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </CardContent>
      </Card>
    </div>
  )
}
