import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useOrganization } from '../context/OrganizationContext'
import { updateWorkspace, deleteWorkspace } from '@/api/organization'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useToast } from '@/components/ui/toast'

export const WorkspaceSettingsPage: React.FC = () => {
  const { currentOrganization, currentWorkspace, workspaces, refreshWorkspaces } = useOrganization()
  const navigate = useNavigate()
  const { toast } = useToast()

  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [clientName, setClientName] = useState('')

  useEffect(() => {
    if (currentWorkspace) {
      setName(currentWorkspace.name)
      setSlug(currentWorkspace.slug)
      setClientName(currentWorkspace.client_name || '')
    }
  }, [currentWorkspace])

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!currentOrganization || !currentWorkspace) return
    try {
      await updateWorkspace(currentOrganization.id, currentWorkspace.id, name, slug, clientName || undefined)
      toast('Çalışma alanı başarıyla güncellendi', 'success')
      await refreshWorkspaces()
    } catch (error) {
      console.error(error)
      toast('Çalışma alanı güncellenemedi', 'error')
    }
  }

  const handleDelete = async () => {
    if (!currentOrganization || !currentWorkspace) return
    
    if (workspaces.length <= 1) {
      toast('En az bir çalışma alanı bulunmalıdır. Silmek için önce yeni bir çalışma alanı oluşturun.', 'error')
      return
    }

    if (!confirm('Bu çalışma alanını silmek istediğinize emin misiniz? Bu işlem geri alınamaz.')) return
    try {
      await deleteWorkspace(currentOrganization.id, currentWorkspace.id)
      toast('Çalışma alanı silindi', 'success')
      await refreshWorkspaces()
      navigate('/dashboard')
    } catch (error: any) {
      console.error(error)
      const errorMessage = error.response?.data?.error || 'Çalışma alanı silinemedi'
      // Translate backend error if possible
      const translatedError = errorMessage === 'cannot delete the last workspace in the organization'
        ? 'Organizasyondaki son çalışma alanı silinemez'
        : errorMessage
      toast(translatedError, 'error')
    }
  }

  if (!currentWorkspace) return <div>Ayarları yönetmek için bir çalışma alanı seçin.</div>

  return (
    <div className="container mx-auto py-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Çalışma Alanı Ayarları</h1>
        <p className="text-muted-foreground">Çalışma alanı tercihlerinizi yönetin.</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Genel Bilgiler</CardTitle>
          <CardDescription>Çalışma alanınızın adını ve URL kısaltmasını güncelleyin.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleUpdate} className="space-y-4">
            <div className="grid w-full max-w-sm items-center gap-1.5">
              <Label htmlFor="name">Ad</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
            </div>
            <div className="grid w-full max-w-sm items-center gap-1.5">
              <Label htmlFor="slug">URL Kısaltması</Label>
              <Input id="slug" value={slug} onChange={(e) => setSlug(e.target.value)} required />
            </div>
            <div className="grid w-full max-w-sm items-center gap-1.5">
              <Label htmlFor="clientName">Müşteri Adı (İsteğe Bağlı)</Label>
              <Input id="clientName" value={clientName} onChange={(e) => setClientName(e.target.value)} placeholder="Örn. Acme Corp" />
            </div>
            <Button type="submit">Değişiklikleri Kaydet</Button>
          </form>
        </CardContent>
      </Card>

      <Card className="border-red-200 bg-red-50/50">
        <CardHeader>
          <CardTitle className="text-red-900">Kritik İşlemler</CardTitle>
          <CardDescription className="text-red-700">
            Bu alandaki işlemler geri alınamaz ve kalıcı veri kaybına yol açabilir. Lütfen dikkatli olun.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 p-4 border border-red-200 rounded-lg bg-white/50">
            <div className="space-y-1">
              <p className="font-medium text-red-900">Çalışma Alanını Sil</p>
              <p className="text-sm text-red-700">
                Bu çalışma alanını ve bağlı tüm verileri (chatbotlar, kaynaklar, konuşma geçmişleri) kalıcı olarak siler.
              </p>
            </div>
            <Button 
              variant="destructive" 
              onClick={handleDelete} 
              className="shrink-0"
              disabled={workspaces.length <= 1}
              title={workspaces.length <= 1 ? "En az bir çalışma alanı kalmalıdır" : undefined}
            >
              Çalışma Alanını Sil
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
