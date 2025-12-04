import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { User } from 'lucide-react'
import { api } from '@/api/client'

const SettingsPage = () => {
  const [activeSection, setActiveSection] = useState<'profile' | 'plan'>('profile')
  const [userPlan, setUserPlan] = useState<string>('free')
  const [rateLimit, setRateLimit] = useState<{ limit: number | null; remaining: number | null }>({ limit: null, remaining: null })

  useEffect(() => {
    api.get('/api/v1/me')
      .then((res) => {
        const plan = res.data?.subscription_plan || 'free'
        setUserPlan(plan)
        const limit = parseInt(res.headers['x-ratelimit-limit'] || '', 10)
        const remaining = parseInt(res.headers['x-ratelimit-remaining'] || '', 10)
        setRateLimit({
          limit: Number.isFinite(limit) ? limit : null,
          remaining: Number.isFinite(remaining) ? remaining : null,
        })
      })
      .catch(() => {
        setUserPlan('free')
        setRateLimit({ limit: null, remaining: null })
      })
  }, [])
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Ayarlar</h1>
        <p className="text-muted-foreground">Hesap bilgilerinizi görüntüleyin.</p>
      </div>

      <div className="grid gap-6 md:grid-cols-[240px_1fr]">
        {/* Sidebar Navigation for Settings */}
        <nav className="flex flex-col space-y-1">
          <Button
            variant="ghost"
            className={`justify-start ${activeSection === 'profile' ? 'bg-muted text-foreground' : ''}`}
            onClick={() => setActiveSection('profile')}
          >
            <User className="mr-2 h-4 w-4" /> Profil
          </Button>
          <Button
            variant="ghost"
            className={`justify-start ${activeSection === 'plan' ? 'bg-muted text-foreground' : ''}`}
            onClick={() => setActiveSection('plan')}
          >
            {/* CreditCard icon */}
            <svg xmlns="http://www.w3.org/2000/svg" className="mr-2 h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect x="2" y="4" width="20" height="16" rx="2" />
              <line x1="2" y1="10" x2="22" y2="10" />
              <line x1="7" y1="15" x2="11" y2="15" />
            </svg>
            Plan
          </Button>
        </nav>

        {/* Content Area */}
        <div className="space-y-6">
          {activeSection === 'profile' && (
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
          )}

          {activeSection === 'plan' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Plan Özeti</CardTitle>
                  <CardDescription>Kullandığınız plan ve temel özellikler.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-medium">Mevcut Plan</div>
                    <div className="text-sm">{userPlan}</div>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Ücretli planlarda güvenli embed (izinli alan adı ve secret) aktif edilir.
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Kota ve Limitler</CardTitle>
                  <CardDescription>Planınıza bağlı limitler ve anlık durum.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-medium">PDF Yükleme Boyutu</div>
                    <div className="text-sm">50MB</div>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-medium">Güvenli Embed</div>
                    <div className="text-sm">{userPlan !== 'free' ? 'Aktif' : 'Yalnızca ücretli planlarda'}</div>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-medium">Örnek Soru Limiti</div>
                    <div className="text-sm">6</div>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-medium">İstek Hız Limiti</div>
                    <div className="text-sm">
                      {rateLimit.limit !== null && rateLimit.remaining !== null
                        ? `${rateLimit.remaining}/${rateLimit.limit} (kalan/toplam)`
                        : 'Şu an alınamadı'}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default SettingsPage
