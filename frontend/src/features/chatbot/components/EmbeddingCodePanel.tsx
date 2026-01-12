import { useState } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { PlanBadge } from '@/components/ui/plan-badge'
import {
  Code2,
  Copy,
  Check,
  Shield,
  Globe,
  Key,
  ChevronDown,
  ChevronUp,
  Info,
  RefreshCw,
  Lock,
  Trash2,
} from 'lucide-react'

type Props = {
  id: string
  secureEmbedPlanEnabled: boolean
  secureEmbedEnabled: boolean
  allowedDomains: string
  embedSecret: string
  onToggleSecure: (v: boolean) => void
  onDomainsChange: (v: string) => void
  onSecretChange: (v: string) => void
  onSecretRefresh: () => void
  onSecretClear: () => void
}

export default function EmbeddingCodePanel({
  id,
  secureEmbedPlanEnabled,
  secureEmbedEnabled,
  allowedDomains,
  embedSecret,
  onToggleSecure,
  onDomainsChange,
  onSecretChange,
  onSecretRefresh,
  onSecretClear,
}: Props) {
  const [copied, setCopied] = useState(false)
  const [advancedOpen, setAdvancedOpen] = useState(false)

  const widgetScriptUrl =
    import.meta.env.VITE_WIDGET_SCRIPT_URL || 'https://widget.botla.app/widget.js'
  const embedCode = `<script src="${widgetScriptUrl}" data-bot="${id}"></script>`

  const handleCopy = () => {
    navigator.clipboard.writeText(embedCode)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-6">
      {/* Embed Code Section */}
      <Card className="border-border/60 shadow-sm overflow-hidden">
        <CardHeader className="bg-muted/30 pb-4">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary/10 rounded-xl">
              <Code2 className="h-5 w-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg">Kodu Web Sitenize Ekleyin</CardTitle>
              <CardDescription>
                Bu kodu sitenizin{' '}
                <code className="text-xs bg-muted px-1 py-0.5 rounded">&lt;body&gt;</code>{' '}
                etiketinin kapanışından hemen önce yapıştırın.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="pt-5">
          <div className="relative">
            <div className="flex items-center justify-between px-4 py-2 bg-muted/50 rounded-t-xl border border-b-0 border-border">
              <div className="flex items-center gap-2 text-xs text-muted-foreground font-medium">
                <Code2 className="h-3.5 w-3.5" />
                HTML
              </div>
              <Button
                size="sm"
                variant="ghost"
                className="h-7 gap-1.5 text-xs"
                onClick={handleCopy}
              >
                {copied ? (
                  <>
                    <Check className="h-3.5 w-3.5 text-green-600" />
                    Kopyalandı
                  </>
                ) : (
                  <>
                    <Copy className="h-3.5 w-3.5" />
                    Kopyala
                  </>
                )}
              </Button>
            </div>
            <pre className="bg-foreground/5 p-4 rounded-b-xl text-sm font-mono text-foreground overflow-x-auto border border-border">
              {embedCode}
            </pre>
          </div>
        </CardContent>
      </Card>

      {/* Security & Access Control Section */}
      <Card className="border-border/60 shadow-sm overflow-hidden">
        <CardHeader className="bg-muted/30 pb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2.5 bg-primary/10 rounded-xl">
                <Shield className="h-5 w-5 text-primary" />
              </div>
              <div>
                <CardTitle className="text-lg">Güvenlik ve Erişim Kontrolü</CardTitle>
                <CardDescription>
                  Botunuzun sadece yetkili sitelerde veya kullanıcılarda çalışmasını sağlayın.
                </CardDescription>
              </div>
            </div>

            {secureEmbedPlanEnabled && (
              <div className="flex items-center gap-3">
                <span
                  className={`text-sm font-medium ${secureEmbedEnabled ? 'text-primary' : 'text-muted-foreground'}`}
                >
                  {secureEmbedEnabled ? 'Aktif' : 'Pasif'}
                </span>
                <Switch
                  id="secure-embed-toggle"
                  checked={secureEmbedEnabled}
                  onCheckedChange={onToggleSecure}
                  aria-label="Güvenli Embed"
                />
              </div>
            )}
          </div>
        </CardHeader>

        <CardContent className="pt-5 space-y-5">
          {!secureEmbedPlanEnabled && (
            <div className="flex items-start gap-3 p-4 bg-muted/50 rounded-xl border border-border">
              <Lock className="h-5 w-5 text-muted-foreground mt-0.5" />
              <div>
                <p className="text-sm text-muted-foreground">
                  Güvenli embed (izinli alan adı ve token doğrulama) özellikleri ücretli planlarda
                  aktif edilir.
                </p>
              </div>
            </div>
          )}

          {secureEmbedPlanEnabled && secureEmbedEnabled && (
            <>
              {/* Domain Restriction */}
              <div className="p-4 bg-card rounded-xl border border-border space-y-3">
                <div className="flex items-center gap-2">
                  <Globe className="h-4 w-4 text-primary" />
                  <label className="text-sm font-medium">Alan Adı Kısıtlaması</label>
                </div>
                <p className="text-xs text-muted-foreground">
                  Widget'ın çalışmasına izin verilen domainleri belirtin.
                </p>
                <Input
                  value={allowedDomains}
                  onChange={(e) => {
                    // Sanitize: remove quotes from each domain
                    const sanitized = e.target.value
                      .split(',')
                      .map((d) => d.trim().replace(/^["']|["']$/g, ''))
                      .join(', ')
                    onDomainsChange(sanitized)
                  }}
                  placeholder="ornek.com, digersite.com"
                  className="bg-background"
                />
                <p className="text-xs text-muted-foreground flex items-center gap-1.5">
                  <Info className="h-3 w-3" />
                  Virgülle ayırarak birden fazla alan adı girebilirsiniz.
                </p>
              </div>

              {/* Advanced: JWT Authentication */}
              <div className="border border-border rounded-xl overflow-hidden">
                <button
                  type="button"
                  onClick={() => setAdvancedOpen(!advancedOpen)}
                  className="w-full flex items-center justify-between p-4 bg-muted/30 hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <Key className="h-4 w-4 text-primary" />
                    <span className="text-sm font-medium">Gelişmiş: Token Doğrulama (JWT)</span>
                    <PlanBadge plan="pro" size="xs" variant="solid" />
                  </div>
                  {advancedOpen ? (
                    <ChevronUp className="h-4 w-4 text-muted-foreground" />
                  ) : (
                    <ChevronDown className="h-4 w-4 text-muted-foreground" />
                  )}
                </button>

                {advancedOpen && (
                  <div className="p-4 space-y-4 border-t border-border">
                    <div className="p-3 bg-muted/30 rounded-lg border border-border">
                      <p className="text-xs text-muted-foreground flex items-start gap-2">
                        <Lock className="h-3.5 w-3.5 mt-0.5 flex-shrink-0" />
                        Bu yöntem, sunucunuzun her oturum için benzersiz bir token oluşturmasını
                        gerektirir. Sadece yazılım geliştirme bilginiz varsa veya geliştirici
                        ekibiniz varsa kullanın.
                      </p>
                    </div>

                    <div className="space-y-3">
                      <label className="text-sm font-medium">Embed Secret</label>
                      <Input
                        value={embedSecret}
                        onChange={(e) => onSecretChange(e.target.value)}
                        placeholder="Gizli anahtar henüz oluşturulmadı"
                        className="bg-background font-mono text-sm"
                        readOnly
                      />
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={onSecretRefresh}
                          className="gap-2"
                        >
                          <RefreshCw className="h-3.5 w-3.5" />
                          Yenile
                        </Button>
                        {embedSecret && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={onSecretClear}
                            className="gap-2 text-destructive border-destructive/30 hover:text-destructive hover:bg-destructive/10"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                            Temizle
                          </Button>
                        )}
                      </div>
                      {embedSecret && (
                        <div className="mt-3 p-3 bg-muted/50 rounded-lg border border-border">
                          <p className="text-sm text-muted-foreground flex items-start gap-2">
                            <Info className="h-4 w-4 mt-0.5 flex-shrink-0 text-foreground" />
                            <span>
                              <strong className="text-foreground">Token doğrulaması aktif.</strong>
                              <br />
                              Sadece alan adı kısıtlaması kullanmak istiyorsanız, yukarıdaki
                              "Temizle" butonuna tıklayın.
                            </span>
                          </p>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
