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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Loader2, Download, Trash2, Shield, Edit3, ChevronLeft, ChevronRight, Clock, Eye } from 'lucide-react'
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
  reason?: string | null
  export_url?: string | null
  export_expires_at?: string | null
  denial_reason?: string | null
  created_at: string
  completed_at?: string | null
}

interface PrivacyRequestsResponse {
  data: PrivacyRequest[]
  total: number
  page: number
  limit: number
  last_completed_at?: string
  next_available_at?: string
}

export default function PrivacySettingsPage() {
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const [deleteReason, setDeleteReason] = useState('')
  const [correctionReason, setCorrectionReason] = useState('')
  const [exportRequestId, setExportRequestId] = useState<string | null>(null)
  const [exportPage, setExportPage] = useState(1)
  const [correctionPage, setCorrectionPage] = useState(1)
  const [limit] = useState(10)
  const [selectedCorrectionRequest, setSelectedCorrectionRequest] = useState<PrivacyRequest | null>(null)

  // 1. Get Consents
  const { data: consents, isLoading: consentsLoading } = useQuery<Consents>({
    queryKey: ['privacy', 'consents'],
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me/privacy/consents')
      return data
    },
  })

  // 2. Fetch export requests with pagination
  const { data: exportRequests, isLoading: exportRequestsLoading } = useQuery<PrivacyRequestsResponse>({
    queryKey: ['privacy', 'requests', 'export', exportPage, limit],
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me/privacy/requests', {
        params: { type: 'export', page: exportPage, limit },
      })
      return data
    },
  })

  // 3. Fetch correction requests with pagination
  const { data: correctionRequests, isLoading: correctionRequestsLoading } = useQuery<PrivacyRequestsResponse>({
    queryKey: ['privacy', 'requests', 'correction', correctionPage, limit],
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me/privacy/requests', {
        params: { type: 'correction', page: correctionPage, limit },
      })
      return data
    },
  })

  // Rate limit info from backend (includes soft-deleted requests)
  const nextExportAvailableTime = exportRequests?.next_available_at
    ? new Date(exportRequests.next_available_at)
    : null
  const isExportRateLimited = nextExportAvailableTime ? new Date() < nextExportAvailableTime : false

  const nextCorrectionAvailableTime = correctionRequests?.next_available_at
    ? new Date(correctionRequests.next_available_at)
    : null
  const isCorrectionRateLimited = nextCorrectionAvailableTime
    ? new Date() < nextCorrectionAvailableTime
    : false



  // Find active export request from export requests
  useEffect(() => {
    if (exportRequests?.data) {
      const activeExport = exportRequests.data.find(
        (req) => req.status === 'pending' || req.status === 'processing' || req.status === 'completed',
      )
      if (activeExport && !exportRequestId) {
        setExportRequestId(activeExport.id)
      }
    }
  }, [exportRequests, exportRequestId])

  // Check if there's an active (pending/processing) export request
  const hasActiveExportRequest = exportRequests?.data?.some(
    (req) => req.status === 'pending' || req.status === 'processing',
  )

  // Check if there's an active (pending/processing) correction request
  const hasActiveCorrectionRequest = correctionRequests?.data?.some(
    (req) => req.status === 'pending' || req.status === 'processing',
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
      } else if (error.response?.status === 429) {
        toast(privacy.toast.exportRateLimit, 'error')
      } else {
        toast(privacy.toast.exportError, 'error')
      }
    },
  })

  const downloadExportRequest = async (requestId: string) => {
    try {
      const res = await api.get(`/api/v1/me/privacy/requests/${requestId}/download`, {
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
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (error) {
      console.error('Download failed:', error)
    }
  }

  // 5. Data Correction
  const correctionMutation = useMutation({
    mutationFn: async () => {
      await api.post('/api/v1/me/privacy/correction', { reason: correctionReason })
    },
    onSuccess: () => {
      setCorrectionReason('')
      queryClient.invalidateQueries({ queryKey: ['privacy', 'requests', 'correction'] })
      toast(privacy.toast.correctionRequested, 'success')
    },
    onError: (error: AxiosError) => {
      if (error.response?.status === 409) {
        toast('Zaten bekleyen bir düzeltme talebiniz var.', 'error')
      } else if (error.response?.status === 429) {
        toast(privacy.export.rateLimit.message, 'error')
      } else {
        toast(privacy.toast.correctionError, 'error')
      }
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
        <CardContent className="space-y-6">
          <div>
            <Button
              variant="outline"
              onClick={() => exportMutation.mutate()}
              disabled={exportMutation.isPending || hasActiveExportRequest || isExportRateLimited}
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

            {isExportRateLimited && nextExportAvailableTime && (
              <div className="mt-4 flex items-start gap-3 rounded-lg bg-amber-50 border border-amber-200 p-4">
                <Clock className="mt-0.5 h-5 w-5 flex-shrink-0 text-amber-600" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-amber-900">
                    {privacy.export.rateLimit.message}
                  </p>
                  <p className="mt-1 text-sm text-amber-700">
                    {privacy.export.rateLimit.nextAvailable.replace(
                      '{time}',
                      nextExportAvailableTime.toLocaleString('tr-TR', {
                        dateStyle: 'medium',
                        timeStyle: 'short',
                      }),
                    )}
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Export History Table */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold">{privacy.export.history.title}</h3>
            {exportRequestsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : exportRequests?.data && exportRequests.data.length > 0 ? (
              <>
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{privacy.export.history.status}</TableHead>
                        <TableHead>{privacy.export.history.date}</TableHead>
                        <TableHead className="text-right">{privacy.export.history.actions}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {exportRequests.data.map((request) => (
                        <TableRow key={request.id}>
                          <TableCell>
                            <span
                              className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                request.status === 'completed'
                                  ? 'bg-green-100 text-green-800'
                                  : request.status === 'processing'
                                    ? 'bg-blue-100 text-blue-800'
                                    : request.status === 'pending'
                                      ? 'bg-yellow-100 text-yellow-800'
                                      : request.status === 'denied'
                                        ? 'bg-red-100 text-red-800'
                                        : 'bg-gray-100 text-gray-800'
                              }`}
                            >
                              {getStatusLabel(request.status)}
                            </span>
                          </TableCell>
                          <TableCell>
                            {new Date(request.created_at).toLocaleString('tr-TR', {
                              dateStyle: 'medium',
                              timeStyle: 'short',
                            })}
                          </TableCell>
                          <TableCell className="text-right">
                            {request.status === 'completed' && (
                              <Button
                                variant="secondary"
                                size="sm"
                                onClick={() => downloadExportRequest(request.id)}
                                className="h-8"
                                title={privacy.export.downloadButton}
                              >
                                <Download className="h-3 w-3" />
                              </Button>
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>

                {/* Pagination */}
                {exportRequests.total > limit && (
                  <div className="flex items-center justify-between px-2">
                    <div className="text-sm text-muted-foreground">
                      Sayfa {exportRequests.page} / {Math.ceil(exportRequests.total / limit)}
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setExportPage((p) => Math.max(1, p - 1))}
                        disabled={exportPage === 1}
                      >
                        <ChevronLeft className="h-4 w-4" />
                        Önceki
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setExportPage((p) => p + 1)}
                        disabled={exportPage >= Math.ceil(exportRequests.total / limit)}
                      >
                        Sonraki
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
              </>
            ) : (
              <p className="text-sm text-muted-foreground py-4">{privacy.export.history.empty}</p>
            )}
          </div>
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
        <CardContent className="space-y-6">
          <div>
            <div className="space-y-2 mb-4">
              <Label htmlFor="correction-reason">{privacy.correction.label}</Label>
              <Textarea
                id="correction-reason"
                placeholder={privacy.correction.placeholder}
                value={correctionReason}
                onChange={(e) => setCorrectionReason(e.target.value.slice(0, privacy.correction.maxLength))}
                maxLength={privacy.correction.maxLength}
                className="resize-y max-h-64"
              />
              <p className="text-xs text-muted-foreground text-right">
                {privacy.correction.charCount
                  .replace('{count}', String(correctionReason.length))
                  .replace('{max}', String(privacy.correction.maxLength))}
              </p>
            </div>
            <Button
              variant="outline"
              onClick={() => correctionMutation.mutate()}
              disabled={correctionMutation.isPending || !correctionReason.trim() || hasActiveCorrectionRequest || isCorrectionRateLimited}
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

            {isCorrectionRateLimited && nextCorrectionAvailableTime && (
              <div className="mt-4 flex items-start gap-3 rounded-lg bg-amber-50 border border-amber-200 p-4">
                <Clock className="mt-0.5 h-5 w-5 flex-shrink-0 text-amber-600" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-amber-900">
                    {privacy.correction.rateLimit.message}
                  </p>
                  <p className="mt-1 text-sm text-amber-700">
                    {privacy.correction.rateLimit.nextAvailable.replace(
                      '{time}',
                      nextCorrectionAvailableTime.toLocaleString('tr-TR', {
                        dateStyle: 'medium',
                        timeStyle: 'short',
                      }),
                    )}
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Correction History Table */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold">{privacy.correction.history.title}</h3>
            {correctionRequestsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : correctionRequests?.data && correctionRequests.data.length > 0 ? (
              <>
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{privacy.correction.history.status}</TableHead>
                        <TableHead>{privacy.correction.history.date}</TableHead>
                        <TableHead className="text-right">{privacy.correction.history.actions}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {correctionRequests.data.map((request) => (
                        <TableRow key={request.id}>
                          <TableCell>
                            <span
                              className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                request.status === 'completed'
                                  ? 'bg-green-100 text-green-800'
                                  : request.status === 'processing'
                                    ? 'bg-blue-100 text-blue-800'
                                    : request.status === 'pending'
                                      ? 'bg-yellow-100 text-yellow-800'
                                      : request.status === 'denied'
                                        ? 'bg-red-100 text-red-800'
                                        : 'bg-gray-100 text-gray-800'
                              }`}
                            >
                              {getStatusLabel(request.status)}
                            </span>
                          </TableCell>
                          <TableCell>
                            {new Date(request.created_at).toLocaleString('tr-TR', {
                              dateStyle: 'medium',
                              timeStyle: 'short',
                            })}
                          </TableCell>
                          <TableCell className="text-right">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8"
                              onClick={() => setSelectedCorrectionRequest(request)}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>

                {/* Pagination for correction requests */}
                {correctionRequests.total > limit && (
                  <div className="flex items-center justify-between px-2">
                    <div className="text-sm text-muted-foreground">
                      Sayfa {correctionRequests.page} / {Math.ceil(correctionRequests.total / limit)}
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCorrectionPage((p) => Math.max(1, p - 1))}
                        disabled={correctionPage === 1}
                      >
                        <ChevronLeft className="h-4 w-4" />
                        Önceki
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCorrectionPage((p) => p + 1)}
                        disabled={correctionPage >= Math.ceil(correctionRequests.total / limit)}
                      >
                        Sonraki
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
              </>
            ) : (
              <p className="text-sm text-muted-foreground py-4">{privacy.correction.history.empty}</p>
            )}
          </div>
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

      {/* Correction Request Detail Modal */}
      <Dialog open={!!selectedCorrectionRequest} onOpenChange={(open) => !open && setSelectedCorrectionRequest(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{privacy.correction.history.detailsTitle}</DialogTitle>
            <DialogDescription>
              {selectedCorrectionRequest && new Date(selectedCorrectionRequest.created_at).toLocaleString('tr-TR', {
                dateStyle: 'long',
                timeStyle: 'short',
              })}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label className="text-sm font-medium">{privacy.correction.history.status}</Label>
              <div className="mt-1">
                <span
                  className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    selectedCorrectionRequest?.status === 'completed'
                      ? 'bg-green-100 text-green-800'
                      : selectedCorrectionRequest?.status === 'processing'
                        ? 'bg-blue-100 text-blue-800'
                        : selectedCorrectionRequest?.status === 'pending'
                          ? 'bg-yellow-100 text-yellow-800'
                          : selectedCorrectionRequest?.status === 'denied'
                            ? 'bg-red-100 text-red-800'
                            : 'bg-gray-100 text-gray-800'
                  }`}
                >
                  {getStatusLabel(selectedCorrectionRequest?.status)}
                </span>
              </div>
            </div>
            <div>
              <Label className="text-sm font-medium">{privacy.correction.history.reason}</Label>
              <p className="mt-1 text-sm text-muted-foreground whitespace-pre-wrap">
                {selectedCorrectionRequest?.reason || '-'}
              </p>
            </div>
            {selectedCorrectionRequest?.denial_reason && (
              <div>
                <Label className="text-sm font-medium text-destructive">Red Sebebi</Label>
                <p className="mt-1 text-sm text-muted-foreground whitespace-pre-wrap">
                  {selectedCorrectionRequest.denial_reason}
                </p>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
