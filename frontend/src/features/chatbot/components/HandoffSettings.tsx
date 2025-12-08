import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Headphones, Mail, AlertCircle } from 'lucide-react'

type HandoffConfig = {
  email_to?: string
  email_subject?: string
}

type Props = {
  handoffEnabled: boolean
  setHandoffEnabled: (v: boolean) => void
  handoffType: 'email'
  setHandoffType: (v: 'email') => void
  handoffConfig: HandoffConfig | null
  setHandoffConfig: (v: HandoffConfig | null) => void
}

export default function HandoffSettings({
  handoffEnabled,
  setHandoffEnabled,
  handoffConfig,
  setHandoffConfig,
}: Props) {
  const handleConfigChange = (field: keyof HandoffConfig, value: string) => {
    setHandoffConfig({
      ...(handoffConfig || {}),
      [field]: value,
    })
  }

  return (
    <div className="grid gap-6">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Headphones className="w-5 h-5 text-primary" />
            <CardTitle>İnsan Desteği Ayarları</CardTitle>
          </div>
          <CardDescription>
            Bot cevaplayamadığında veya kullanıcı istediğinde insan operatöre devir ayarları.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Enable Toggle */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <label className="text-sm font-medium">İnsan Desteğini Aktifleştir</label>
              <p className="text-xs text-muted-foreground">
                Kullanıcılar asistandan istediğinde temsilciye aktarım talebinde bulunabilir.
              </p>
            </div>
            <Switch
              checked={handoffEnabled}
              onCheckedChange={setHandoffEnabled}
            />
          </div>

          {handoffEnabled && (
            <>
              <div className="border-t pt-6 space-y-4">
                {/* Handoff Type - Email only for now */}
                <div className="flex items-center gap-3 p-4 border rounded-lg bg-primary/5">
                  <Mail className="w-5 h-5 text-primary" />
                  <div>
                    <p className="text-sm font-medium">E-posta ile Bildirim</p>
                    <p className="text-xs text-muted-foreground">
                      Kullanıcı destek istediğinde konuşma dökümü e-posta ile gönderilir.
                    </p>
                  </div>
                </div>

                {/* Email Configuration */}
                <div className="space-y-4 pt-2">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Alıcı E-posta Adresi</label>
                    <Input
                      type="email"
                      placeholder="destek@firmaniz.com"
                      value={handoffConfig?.email_to || ''}
                      onChange={(e) => handleConfigChange('email_to', e.target.value)}
                    />
                    <p className="text-xs text-muted-foreground">
                      Destek talepleri bu adrese gönderilir.
                    </p>
                  </div>

                  <div className="space-y-2">
                    <label className="text-sm font-medium">E-posta Başlığı (Opsiyonel)</label>
                    <Input
                      type="text"
                      placeholder="[Botla] Yeni Destek Talebi - {bot_name}"
                      value={handoffConfig?.email_subject || ''}
                      onChange={(e) => handleConfigChange('email_subject', e.target.value)}
                    />
                    <p className="text-xs text-muted-foreground">
                      Boş bırakılırsa varsayılan başlık kullanılır.
                    </p>
                  </div>
                </div>

                {/* Warning if no email configured */}
                {!handoffConfig?.email_to && (
                  <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800">
                    <AlertCircle className="w-5 h-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
                    <div className="text-sm text-amber-800 dark:text-amber-200">
                      <p className="font-medium">E-posta adresi gerekli</p>
                      <p className="text-xs mt-1">
                        İnsan desteği özelliğinin çalışması için bir e-posta adresi girmeniz gerekiyor.
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Info Card */}
      <Card className="bg-muted/30">
        <CardContent className="pt-6">
          <div className="space-y-3">
            <h4 className="text-sm font-medium">Nasıl Çalışır?</h4>
            <ul className="text-sm text-muted-foreground space-y-2">
              <li className="flex items-start gap-2">
                <span className="text-primary font-bold">1.</span>
                <span>Kullanıcı chat widget&apos;ında &quot;İnsan Desteği İste&quot; butonuna basar.</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary font-bold">2.</span>
                <span>Konuşma dökümü otomatik olarak belirtilen e-posta adresine gönderilir.</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary font-bold">3.</span>
                <span>Kullanıcıya talebinin alındığı ve en kısa sürede iletişime geçileceği bildirilir.</span>
              </li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
