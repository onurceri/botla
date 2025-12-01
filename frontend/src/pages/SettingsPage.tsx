import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { User } from 'lucide-react'

const SettingsPage = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Ayarlar</h1>
        <p className="text-muted-foreground">Hesap bilgilerinizi görüntüleyin.</p>
      </div>

      <div className="grid gap-6 md:grid-cols-[240px_1fr]">
        {/* Sidebar Navigation for Settings */}
        <nav className="flex flex-col space-y-1">
          <Button variant="ghost" className="justify-start bg-muted text-foreground">
            <User className="mr-2 h-4 w-4" /> Profil
          </Button>
        </nav>

        {/* Content Area */}
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Profil Bilgileri</CardTitle>
              <CardDescription>Profil bilgileri şu an yalnızca görüntülenebilir.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Ad Soyad</label>
                <Input defaultValue="Onur Ceri" disabled />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Email</label>
                <Input defaultValue="onur@example.com" disabled />
              </div>
            </CardContent>
          </Card>

          
        </div>
      </div>
    </div>
  )
}

export default SettingsPage
