import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { User, CreditCard, Bell, Shield } from 'lucide-react'

const SettingsPage = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Ayarlar</h1>
        <p className="text-muted-foreground">Hesap tercihlerinizi ve aboneliğinizi yönetin.</p>
      </div>

      <div className="grid gap-6 md:grid-cols-[240px_1fr]">
        {/* Sidebar Navigation for Settings */}
        <nav className="flex flex-col space-y-1">
          <Button variant="ghost" className="justify-start bg-muted text-foreground">
            <User className="mr-2 h-4 w-4" /> Profil
          </Button>
          <Button variant="ghost" className="justify-start hover:bg-muted">
            <CreditCard className="mr-2 h-4 w-4" /> Faturalandırma
          </Button>
          <Button variant="ghost" className="justify-start hover:bg-muted">
            <Bell className="mr-2 h-4 w-4" /> Bildirimler
          </Button>
          <Button variant="ghost" className="justify-start hover:bg-muted">
            <Shield className="mr-2 h-4 w-4" /> Güvenlik
          </Button>
        </nav>

        {/* Content Area */}
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Profil Bilgileri</CardTitle>
              <CardDescription>Kişisel bilgilerinizi güncelleyin.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Ad Soyad</label>
                <Input defaultValue="Onur Ceri" />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Email</label>
                <Input defaultValue="onur@example.com" disabled />
                <p className="text-xs text-muted-foreground">Email adresi değiştirilemez.</p>
              </div>
              <div className="flex justify-end">
                <Button>Kaydet</Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Şifre Değiştir</CardTitle>
              <CardDescription>Hesap güvenliğinizi sağlamak için güçlü bir şifre kullanın.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Mevcut Şifre</label>
                <Input type="password" />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Yeni Şifre</label>
                <Input type="password" />
              </div>
              <div className="flex justify-end">
                <Button variant="outline">Şifreyi Güncelle</Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

export default SettingsPage
