import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Info } from 'lucide-react'

type Props = {
  id: string
  userPlan: string
  secureEmbedEnabled: boolean
  allowedDomains: string
  embedSecret: string
  onToggleSecure: (v: boolean) => void
  onDomainsChange: (v: string) => void
  onSecretChange: (v: string) => void
  onSecretRefresh: () => void
}

export default function EmbeddingCodePanel({
  id,
  userPlan,
  secureEmbedEnabled,
  allowedDomains,
  embedSecret,
  onToggleSecure,
  onDomainsChange,
  onSecretChange,
  onSecretRefresh,
}: Props) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Web Sitenize Ekleyin</CardTitle>
        <CardDescription>Aşağıdaki kodu sitenizin &lt;body&gt; etiketinin sonuna yapıştırın.</CardDescription>
      </CardHeader>
      <CardContent>
        {userPlan !== 'free' && (
          <div className="mb-4 flex items-center gap-3">
            <label htmlFor="secure-embed-checkbox" className="text-sm font-medium">Güvenli Embed</label>
            <input id="secure-embed-checkbox" type="checkbox" checked={secureEmbedEnabled} onChange={(e) => onToggleSecure(e.target.checked)} />
          </div>
        )}
        {userPlan !== 'free' && secureEmbedEnabled && (
          <div className="grid md:grid-cols-2 gap-4 mb-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">İzinli Alan Adları (virgülle ayırın)</label>
              <Input value={allowedDomains} onChange={(e) => onDomainsChange(e.target.value)} placeholder="example.com, another.com" />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Embed Secret</label>
              <div className="flex gap-2">
                <Input value={embedSecret} onChange={(e) => onSecretChange(e.target.value)} placeholder="Gizli anahtar" />
                <Button type="button" variant="secondary" onClick={onSecretRefresh}>Yenile</Button>
              </div>
            </div>
          </div>
        )}
        <div className="relative group">
          <pre className="bg-muted p-4 rounded-xl text-xs font-mono text-foreground overflow-x-auto border border-border shadow-sm">
            {`<script src="https://cdn.botla.co/widget.js" data-bot="${id}"></script>`}
          </pre>
          <Button 
            size="sm" 
            variant="secondary" 
            className="absolute top-2 right-2 shadow-sm"
            onClick={() => navigator.clipboard.writeText(`<script src="https://cdn.botla.co/widget.js" data-bot="${id}"></script>`)}
          >
            Kopyala
          </Button>
        </div>
        {userPlan === 'free' && (
          <div className="mt-4 text-xs text-muted-foreground">Güvenli embed (izinli alan adı ve secret) özellikleri ücretli planlarda aktif edilir.</div>
        )}
        <div className="mt-4 flex items-center gap-2 text-xs text-muted-foreground">
          <Info className="w-4 h-4" />
          Kodun yüklendiğinden emin olmak için sayfayı yenileyin.
        </div>
      </CardContent>
    </Card>
  )
}
