import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Headphones, Mail, AlertCircle, Info, CheckCircle2, Lock, Shield } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

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
  canUseHandoff?: boolean
}

export default function HandoffSettings({
  handoffEnabled,
  setHandoffEnabled,
  handoffConfig,
  setHandoffConfig,
  canUseHandoff = false,
}: Props) {
  const handleConfigChange = (field: keyof HandoffConfig, value: string) => {
    setHandoffConfig({
      ...(handoffConfig || {}),
      [field]: value,
    })
  }

  return (
    <div className="grid gap-6">
      <Card className="border-muted-foreground/20 shadow-sm overflow-hidden">
        <CardHeader className="bg-muted/30 pb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
                <div className="p-2.5 bg-violet-500/10 rounded-xl text-violet-600 ring-1 ring-violet-500/20">
                    <Headphones className="w-6 h-6" />
                </div>
                <div>
                    <div className="flex items-center gap-2">
                      <CardTitle className="text-xl">İnsan Desteği Ayarları</CardTitle>
                      {!canUseHandoff && <Badge variant="secondary" className="scale-90">Ent</Badge>}
                    </div>
                    <CardDescription className="text-base mt-1">
                        Bot cevaplayamadığında veya kullanıcı istediğinde insan operatöre devir süreçlerini yönetin.
                    </CardDescription>
                </div>
            </div>
            <div className="flex items-center gap-3 bg-background px-4 py-2 rounded-full border shadow-sm">
                <span className="text-sm font-medium">Aktif</span>
                <Switch
                  checked={handoffEnabled}
                  onCheckedChange={setHandoffEnabled}
                  disabled={!canUseHandoff}
                />
                {!canUseHandoff && <Lock className="w-4 h-4 text-muted-foreground" />}
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-6">
          {!canUseHandoff && (
            <div className="mb-6 bg-amber-500/5 border border-amber-500/20 rounded-xl p-4 flex items-center gap-4">
                <div className="p-2.5 bg-amber-500/10 rounded-full text-amber-600 shrink-0">
                    <Shield className="w-5 h-5" />
                </div>
                <div>
                    <span className="font-semibold text-amber-900 dark:text-amber-500 block">Enterprise Özelliği</span>
                    <span className="text-muted-foreground text-sm">İnsan desteği özelliğini kullanmak için planınızı yükseltin.</span>
                </div>
                <Button variant="outline" size="sm" className="ml-auto border-amber-500/20 hover:bg-amber-500/10 text-amber-600">Yükselt</Button>
            </div>
          )}

          {!handoffEnabled ? (
            <div className="flex flex-col items-center justify-center py-12 text-center border-2 border-dashed rounded-2xl bg-muted/10 space-y-4">
               <div className="p-4 bg-muted rounded-full">
                  <Headphones className="w-8 h-8 text-muted-foreground/50" />
               </div>
               <div className="max-w-md space-y-2">
                 <h3 className="text-lg font-semibold">İnsan Desteği Kapalı</h3>
                 <p className="text-muted-foreground">
                    Şu anda kullanıcılar bir temsilciye bağlanma talebinde bulunamaz. Sadece bot yanıtları aktiftir.
                 </p>
               </div>
            </div>
          ) : (
            <div className="space-y-8 animate-in slide-in-from-top-4 duration-300">
              
              {/* Channel Selection (Only Email for now) */}
              <div className="space-y-4">
                 <label className="text-base font-semibold text-foreground flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-primary"></span>
                    İletişim Kanalı
                 </label>
                 <div className="grid md:grid-cols-2 gap-4">
                    <div className="relative flex items-center gap-4 p-4 border-2 border-primary bg-primary/5 rounded-xl cursor-pointer shadow-sm">
                       <div className="p-3 bg-background rounded-full border shadow-sm text-primary">
                          <Mail className="w-5 h-5" />
                       </div>
                       <div>
                          <p className="font-semibold text-foreground">E-posta Bildirimi</p>
                          <p className="text-xs text-muted-foreground mt-0.5">Konuşma dökümünü e-posta ile gönder.</p>
                       </div>
                       <div className="absolute top-4 right-4">
                          <CheckCircle2 className="w-5 h-5 text-primary fill-primary/10" />
                       </div>
                    </div>
                    
                    <div className="flex items-center gap-4 p-4 border rounded-xl opacity-50 cursor-not-allowed bg-muted/20">
                        <div className="p-3 bg-background rounded-full border shadow-sm text-muted-foreground">
                            <Headphones className="w-5 h-5" />
                        </div>
                        <div>
                            <p className="font-semibold text-foreground">Canlı Destek (Yakında)</p>
                            <p className="text-xs text-muted-foreground mt-0.5">Intercom, Zendesk vb. entegrasyonlar.</p>
                        </div>
                    </div>
                 </div>
              </div>

              <div className="border-t border-border/50" />

              {/* Email Configuration */}
              <div className="space-y-6">
                 <div className="space-y-1">
                    <label className="text-base font-semibold text-foreground">E-posta Yapılandırması</label>
                    <p className="text-sm text-muted-foreground">Destek taleplerinin iletileceği adres ve format ayarları.</p>
                 </div>

                 <div className="grid gap-6 md:grid-cols-2">
                    <div className="space-y-3">
                      <label className="text-sm font-medium flex items-center gap-2">
                         Alıcı E-posta Adresi <span className="text-red-500">*</span>
                      </label>
                      <Input
                        className="h-11 bg-background/50"
                        type="email"
                        placeholder="destek@firmaniz.com"
                        value={handoffConfig?.email_to || ''}
                        onChange={(e) => handleConfigChange('email_to', e.target.value)}
                      />
                      <p className="text-xs text-muted-foreground">
                        Tüm destek talepleri bu adrese yönlendirilir.
                      </p>
                    </div>

                    <div className="space-y-3">
                      <label className="text-sm font-medium">E-posta Başlığı (Opsiyonel)</label>
                      <Input
                        className="h-11 bg-background/50"
                        type="text"
                        placeholder="[Botla] Yeni Destek Talebi - {bot_name}"
                        value={handoffConfig?.email_subject || ''}
                        onChange={(e) => handleConfigChange('email_subject', e.target.value)}
                      />
                      <p className="text-xs text-muted-foreground">
                        Boş bırakılırsa varsayılan şablon kullanılır.
                      </p>
                    </div>
                 </div>

                 {/* Warning if no email configured */}
                  {!handoffConfig?.email_to && (
                    <div className="flex items-start gap-3 p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-700 dark:text-amber-400">
                      <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
                      <div className="text-sm">
                        <p className="font-semibold">E-posta adresi gerekli</p>
                        <p className="mt-1 opacity-90">
                          İnsan desteği özelliğinin çalışması için geçerli bir e-posta adresi girmeniz gerekmektedir. Aksi takdirde talepler iletilemez.
                        </p>
                      </div>
                    </div>
                  )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Info Card */}
      <Card className="bg-gradient-to-br from-muted/50 to-muted/10 border-muted-foreground/10">
        <CardContent className="p-6">
          <div className="flex items-start gap-4">
             <div className="p-2 bg-background rounded-lg shadow-sm text-primary border">
                 <Info className="w-5 h-5" />
             </div>
             <div className="space-y-4">
                <h4 className="text-base font-semibold">Süreç Nasıl İşler?</h4>
                <div className="grid md:grid-cols-3 gap-6">
                    <div className="space-y-2">
                        <div className="flex items-center gap-2 text-sm font-medium text-foreground">
                            <span className="flex items-center justify-center w-6 h-6 rounded-full bg-primary/10 text-primary text-xs">1</span>
                            Talep Oluşturma
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                            Kullanıcı sohbet esnasında "İnsan Desteği" butonuna tıklar veya temsilciyle görüşmek istediğini belirtir.
                        </p>
                    </div>
                    <div className="space-y-2">
                        <div className="flex items-center gap-2 text-sm font-medium text-foreground">
                            <span className="flex items-center justify-center w-6 h-6 rounded-full bg-primary/10 text-primary text-xs">2</span>
                            Bildirim Gönderimi
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                            Tüm konuşma geçmişi ve kullanıcı bilgileri, belirlediğiniz e-posta adresine anında raporlanır.
                        </p>
                    </div>
                    <div className="space-y-2">
                        <div className="flex items-center gap-2 text-sm font-medium text-foreground">
                            <span className="flex items-center justify-center w-6 h-6 rounded-full bg-primary/10 text-primary text-xs">3</span>
                            Kullanıcı Bilgilendirmesi
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                            Kullanıcıya talebinin alındığı ve en kısa sürede dönüş yapılacağı bilgisi verilir.
                        </p>
                    </div>
                </div>
             </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
