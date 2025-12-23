import { usePlan, useProfile } from '@/hooks/queries/useProfile'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { PlanBadge, PlanTier } from '@/components/ui/plan-badge'

const ProfilePage = () => {
  const { data: profile, isLoading: profileLoading } = useProfile()
  const { data: plan, isLoading: planLoading } = usePlan()

  const loading = profileLoading || planLoading

  const fullName = profile?.full_name || ''
  const email = profile?.email || ''
  const userPlan = plan?.code || 'free'

  if (loading) {
    return <div className="p-8 text-center text-muted-foreground">Yükleniyor...</div>
  }

  return (
    <div className="max-w-4xl space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Profil</h1>
        <p className="text-muted-foreground mt-1">Kişisel bilgilerinizi görüntüleyin.</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Profil Bilgileri</CardTitle>
          <CardDescription>Kişisel bilgileriniz ve iletişim tercihleriniz.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center gap-6">
            <div className="h-20 w-20 rounded-full bg-muted flex items-center justify-center text-2xl font-bold text-muted-foreground uppercase">
              {fullName ? fullName.substring(0, 2) : email.substring(0, 2)}
            </div>
            <div className="space-y-1">
              <h3 className="font-medium text-lg">{fullName || 'İsimsiz Kullanıcı'}</h3>
              <p className="text-sm text-muted-foreground">{email}</p>
              <PlanBadge plan={userPlan as PlanTier} size="md" className="mt-2" />
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Ad Soyad
              </label>
              <Input value={fullName} disabled className="bg-muted/50" />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Email Adresi
              </label>
              <Input value={email} disabled className="bg-muted/50" />
            </div>
          </div>
        </CardContent>
        <CardFooter className="border-t bg-muted/50 px-6 py-4">
          <p className="text-xs text-muted-foreground">
            Profil bilgilerinizi değiştirmek için lütfen sistem yöneticisi ile iletişime geçin.
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}

export default ProfilePage
