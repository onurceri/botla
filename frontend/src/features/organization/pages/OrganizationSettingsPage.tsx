import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useOrganization } from '../context/OrganizationContext'
import {
  updateOrganization,
  deleteOrganization,
  getMembers,
  addMember,
  removeMember,
  updateMemberRole,
  Member,
} from '@/api/organization'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useToast } from '@/components/ui/toast'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { User, Shield, ShieldAlert, Crown } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { getTurkishErrorMessage } from '@/lib/errorMessages'

type CallerRole = 'owner' | 'admin' | 'member'

export const OrganizationSettingsPage: React.FC = () => {
  const { currentOrganization, refreshOrganizations } = useOrganization()
  const navigate = useNavigate()
  const { toast } = useToast()

  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [members, setMembers] = useState<Member[]>([])
  const [callerRole, setCallerRole] = useState<CallerRole>('member')
  const [currentUserId, setCurrentUserId] = useState<string>('')
  const [loadingMembers, setLoadingMembers] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('member')
  const [isInviteOpen, setIsInviteOpen] = useState(false)

  // Permission checks based on caller's role
  const canManageMembers = callerRole === 'owner' || callerRole === 'admin'
  const canUpdateOrg = callerRole === 'owner' || callerRole === 'admin'
  const canDeleteOrg = callerRole === 'owner'

  // Check if caller can change a specific user's role
  const canChangeUserRole = (member: Member) => {
    if (callerRole === 'member') return false
    if (callerRole === 'owner') return true
    // Admins can only change member roles, not other admins or owners
    return member.role === 'member'
  }

  // Check if caller can remove a specific member
  const canRemoveMember = (member: Member) => {
    if (!canManageMembers) return false
    // Can't remove self through this UI
    if (member.user_id === currentUserId) return false
    // Admins can't remove owners
    if (callerRole === 'admin' && member.role === 'owner') return false
    return true
  }

  // Check available roles for the role selector
  const getAvailableRolesForMember = (_member: Member) => {
    if (callerRole === 'owner') {
      return ['member', 'admin', 'owner']
    }
    // Admins can only set to member or admin
    return ['member', 'admin']
  }

  useEffect(() => {
    if (currentOrganization) {
      setName(currentOrganization.name)
      setSlug(currentOrganization.slug)
      loadMembers()
    }
  }, [currentOrganization])

  const loadMembers = async () => {
    if (!currentOrganization) return
    setLoadingMembers(true)
    try {
      const response = await getMembers(currentOrganization.id)
      setMembers(response.members)
      setCallerRole(response.caller_role)
      
      // Find current user ID from members list using caller_role
      const currentMember = response.members.find(m => m.role === response.caller_role)
      if (currentMember) {
        setCurrentUserId(currentMember.user_id)
      }
      // Actually we need all members with the same role, let's get it from auth context instead
      // For now, we'll identify current user by finding the owner if the response owner matches
    } catch (error) {
      console.error(error)
      toast('Üyeler yüklenemedi', 'error')
    } finally {
      setLoadingMembers(false)
    }
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!currentOrganization) return
    try {
      await updateOrganization(currentOrganization.id, name, slug)
      toast('Organizasyon başarıyla güncellendi', 'success')
      await refreshOrganizations()
    } catch (error: any) {
      console.error(error)
      toast(getTurkishErrorMessage(error, 'Organizasyon güncellenemedi'), 'error')
    }
  }

  const handleDelete = async () => {
    if (!currentOrganization) return
    if (!confirm('Bu organizasyonu silmek istediğinize emin misiniz? Bu işlem geri alınamaz.')) return
    try {
      await deleteOrganization(currentOrganization.id)
      toast('Organizasyon silindi', 'success')
      await refreshOrganizations()
      navigate('/dashboard')
    } catch (error: any) {
      console.error(error)
      toast(getTurkishErrorMessage(error, 'Organizasyon silinemedi'), 'error')
    }
  }

  const handleInvite = async () => {
    if (!currentOrganization) return
    try {
      await addMember(currentOrganization.id, inviteEmail, inviteRole)
      toast('Üye başarıyla eklendi', 'success')
      setIsInviteOpen(false)
      setInviteEmail('')
      loadMembers()
    } catch (error: any) {
      console.error(error)
      toast(getTurkishErrorMessage(error, 'Üye eklenemedi'), 'error')
    }
  }

  const handleRemoveMember = async (userId: string) => {
    if (!currentOrganization) return
    if (!confirm('Bu üyeyi çıkarmak istediğinize emin misiniz?')) return
    try {
      await removeMember(currentOrganization.id, userId)
      toast('Üye çıkarıldı', 'success')
      loadMembers()
    } catch (error: any) {
      console.error(error)
      toast(getTurkishErrorMessage(error, 'Üye çıkarılamadı'), 'error')
    }
  }

  const handleUpdateRole = async (userId: string, newRole: string) => {
    if (!currentOrganization) return
    try {
      await updateMemberRole(currentOrganization.id, userId, newRole)
      toast('Rol güncellendi', 'success')
      loadMembers()
    } catch (error: any) {
      console.error(error)
      toast(getTurkishErrorMessage(error, 'Rol güncellenemedi'), 'error')
    }
  }

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'owner': return <Crown className="h-4 w-4 text-yellow-500" />
      case 'admin': return <ShieldAlert className="h-4 w-4 text-blue-500" />
      default: return <Shield className="h-4 w-4 text-gray-400" />
    }
  }

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case 'owner': return 'default'
      case 'admin': return 'secondary'
      default: return 'outline'
    }
  }

  const getRoleLabel = (role: string) => {
    switch (role) {
      case 'owner': return 'Sahip'
      case 'admin': return 'Yönetici'
      default: return 'Üye'
    }
  }

  if (!currentOrganization) return <div>Yükleniyor...</div>

  return (
    <div className="container mx-auto py-8 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Organizasyon Ayarları</h1>
          <p className="text-muted-foreground">Organizasyon tercihlerinizi ve üyelerinizi yönetin.</p>
        </div>
        <Badge variant={getRoleBadgeVariant(callerRole)} className="flex items-center gap-1">
          {getRoleIcon(callerRole)}
          {getRoleLabel(callerRole)}
        </Badge>
      </div>

      <Tabs defaultValue="general" className="w-full">
        <TabsList>
          <TabsTrigger value="general">Genel</TabsTrigger>
          <TabsTrigger value="members">Üyeler</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Genel Bilgiler</CardTitle>
              <CardDescription>
                {canUpdateOrg 
                  ? 'Organizasyonunuzun adını ve URL kısaltmasını güncelleyin.'
                  : 'Organizasyon bilgilerini görüntülüyorsunuz. Düzenleme için yönetici veya sahip yetkiniz olmalıdır.'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleUpdate} className="space-y-4">
                <div className="grid w-full max-w-sm items-center gap-1.5">
                  <Label htmlFor="name">Ad</Label>
                  <Input 
                    id="name" 
                    value={name} 
                    onChange={(e) => setName(e.target.value)} 
                    required 
                    disabled={!canUpdateOrg}
                  />
                </div>
                <div className="grid w-full max-w-sm items-center gap-1.5">
                  <Label htmlFor="slug">URL Kısaltması</Label>
                  <Input 
                    id="slug" 
                    value={slug} 
                    onChange={(e) => setSlug(e.target.value)} 
                    required 
                    disabled={!canUpdateOrg}
                  />
                </div>
                {canUpdateOrg && (
                  <Button type="submit">Değişiklikleri Kaydet</Button>
                )}
              </form>
            </CardContent>
          </Card>

          {canDeleteOrg && (
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
                    <p className="font-medium text-red-900">Organizasyonu Sil</p>
                    <p className="text-sm text-red-700">
                      Bu organizasyonu ve bağlı tüm verileri (chatbotlar, üyeler, ayarlar) kalıcı olarak siler.
                    </p>
                  </div>
                  <Button variant="destructive" onClick={handleDelete} className="shrink-0">
                    Organizasyonu Sil
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="members" className="space-y-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>Üyeler</CardTitle>
                <CardDescription>
                  {canManageMembers 
                    ? 'Bu organizasyona kimlerin erişebileceğini yönetin.'
                    : 'Organizasyon üyelerini görüntülüyorsunuz.'}
                </CardDescription>
              </div>
              {canManageMembers && (
                <Dialog open={isInviteOpen} onOpenChange={setIsInviteOpen}>
                  <DialogTrigger asChild>
                    <Button>Üye Ekle</Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Yeni Üye Ekle</DialogTitle>
                      <DialogDescription>
                        Eklemek istediğiniz kullanıcının e-posta adresini girin. Kullanıcının zaten bir hesabı olmalıdır.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                      <div className="space-y-2">
                        <Label htmlFor="email">E-posta</Label>
                        <Input id="email" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} placeholder="kullanici@ornek.com" />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="role">Rol</Label>
                        <Select value={inviteRole} onValueChange={setInviteRole}>
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="member">Üye</SelectItem>
                            <SelectItem value="admin">Yönetici</SelectItem>
                            {callerRole === 'owner' && (
                              <SelectItem value="owner">Sahip</SelectItem>
                            )}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    <DialogFooter>
                      <Button onClick={handleInvite}>Üye Ekle</Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              )}
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {members.map((member) => (
                  <div key={member.id} className="flex items-center justify-between p-4 border rounded-lg">
                    <div className="flex items-center gap-4">
                      <Avatar>
                        <AvatarImage src={member.user.avatar_url || undefined} />
                        <AvatarFallback>
                          {member.user.full_name || member.user.email ? (
                            (member.user.full_name || member.user.email || '').charAt(0).toUpperCase()
                          ) : (
                            <User className="h-4 w-4" />
                          )}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <p className="font-medium">{member.user.full_name || member.user.email}</p>
                        <p className="text-sm text-muted-foreground">{member.user.email}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {canChangeUserRole(member) ? (
                        <Select value={member.role} onValueChange={(role) => handleUpdateRole(member.user_id, role)}>
                          <SelectTrigger className="w-[130px]">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {getAvailableRolesForMember(member).map((role) => (
                              <SelectItem key={role} value={role}>
                                <div className="flex items-center gap-2">
                                  {getRoleIcon(role)}
                                  {getRoleLabel(role)}
                                </div>
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      ) : (
                        <Badge variant={getRoleBadgeVariant(member.role)} className="flex items-center gap-1">
                          {getRoleIcon(member.role)}
                          {getRoleLabel(member.role)}
                        </Badge>
                      )}
                      {canRemoveMember(member) && (
                        <Button variant="ghost" size="sm" onClick={() => handleRemoveMember(member.user_id)} className="text-red-500 hover:text-red-600">
                          Çıkar
                        </Button>
                      )}
                    </div>
                  </div>
                ))}
                {members.length === 0 && !loadingMembers && (
                    <p className="text-center text-muted-foreground py-4">Üye bulunamadı.</p>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
