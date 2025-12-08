import React, { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useToast } from '@/components/ui/toast'
import { createWorkspace } from '@/api/organization'
import { useOrganization } from '../context/OrganizationContext'

interface CreateWorkspaceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export const CreateWorkspaceDialog: React.FC<CreateWorkspaceDialogProps> = ({
  open,
  onOpenChange,
}) => {
  const { toast } = useToast()
  const { currentOrganization, refreshWorkspaces, selectWorkspace } = useOrganization()
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [clientName, setClientName] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value
    setName(val)
    setSlug(val.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, ''))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !slug || !currentOrganization) return

    setIsLoading(true)
    try {
      const ws = await createWorkspace(currentOrganization.id, name, slug, clientName || undefined)
      toast('Çalışma alanı başarıyla oluşturuldu', 'success')
      await refreshWorkspaces()
      selectWorkspace(ws.id)
      onOpenChange(false)
      setName('')
      setSlug('')
      setClientName('')
    } catch (error) {
      console.error(error)
      toast('Çalışma alanı oluşturulamadı', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Çalışma Alanı Oluştur</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="ws-name">Çalışma Alanı Adı</Label>
            <Input
              id="ws-name"
              value={name}
              onChange={handleNameChange}
              placeholder="Marketing"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="ws-slug">URL Kısaltması</Label>
            <Input
              id="ws-slug"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="marketing"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="client-name">Müşteri Adı (İsteğe Bağlı)</Label>
            <Input
              id="client-name"
              value={clientName}
              onChange={(e) => setClientName(e.target.value)}
              placeholder="Client Corp"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              İptal
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Oluşturuluyor...' : 'Oluştur'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
