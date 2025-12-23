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
import { createOrganization } from '@/api/organization'
import { useOrganization } from '../context/OrganizationContext'

interface CreateOrganizationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export const CreateOrganizationDialog: React.FC<CreateOrganizationDialogProps> = ({
  open,
  onOpenChange,
}) => {
  const { toast } = useToast()
  const { refreshOrganizations, selectOrganization } = useOrganization()
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value
    setName(val)
    // Auto-generate slug from name if slug hasn't been manually edited
    // For simplicity, we'll just always suggest a slug based on name
    setSlug(
      val
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-|-$/g, ''),
    )
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !slug) return

    setIsLoading(true)
    try {
      const org = await createOrganization(name, slug)
      toast('Organizasyon başarıyla oluşturuldu', 'success')
      await refreshOrganizations()
      await selectOrganization(org.id)
      onOpenChange(false)
      setName('')
      setSlug('')
    } catch (error) {
      console.error(error)
      toast('Organizasyon oluşturulamadı', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Organizasyon Oluştur</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="org-name">Organizasyon Adı</Label>
            <Input
              id="org-name"
              value={name}
              onChange={handleNameChange}
              placeholder="Acme Corp"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="org-slug">URL Kısaltması</Label>
            <Input
              id="org-slug"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="acme-corp"
              required
            />
            <p className="text-sm text-muted-foreground">
              Organizasyon URL'i için benzersiz tanımlayıcı.
            </p>
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
