import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

type Props = {
  name: string
  setName: (v: string) => void
  model: string
  setModel: (v: string) => void
  systemPrompt: string
  setSystemPrompt: (v: string) => void
  temperature: number
  setTemperature: (v: number) => void
  maxTokens: number
  setMaxTokens: (v: number) => void
}

export default function OverviewPanel({ name, setName, model, setModel, systemPrompt, setSystemPrompt, temperature, setTemperature, maxTokens, setMaxTokens }: Props) {
  return (
    <div className="grid gap-6 md:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Kimlik & Model</CardTitle>
          <CardDescription>Bot ismi, model seçimi ve sistem mesajı.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">İsim</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Model</label>
            <select 
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              value={model}
              onChange={(e) => setModel(e.target.value)}
            >
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo (Hızlı & Ucuz)</option>
              <option value="gpt-4">GPT-4 (Akıllı & Pahalı)</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">System Prompt</label>
            <textarea 
              className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              value={systemPrompt}
              onChange={(e) => setSystemPrompt(e.target.value)}
              placeholder="Sen yardımcı bir asistansın..."
            />
            <div className="flex justify-end text-xs text-muted-foreground">{systemPrompt.length} karakter</div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Model Ayarları</CardTitle>
          <CardDescription>Yaratıcılık ve token sınırı.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <label className="text-xs font-medium text-muted-foreground uppercase">Yaratıcılık (Temperature): {temperature}</label>
            <input 
              type="range" 
              min="0" 
              max="1" 
              step="0.1" 
              value={temperature} 
              onChange={(e) => setTemperature(parseFloat(e.target.value))}
              className="w-full accent-primary"
            />
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Tutarlı (0.0)</span>
              <span>Yaratıcı (1.0)</span>
            </div>
          </div>
          <div className="space-y-2">
            <label htmlFor="max-token-input" className="text-xs font-medium text-muted-foreground uppercase">Maksimum Token</label>
            <Input 
              type="number" 
              value={maxTokens}
              onChange={(e) => setMaxTokens(Number(e.target.value))}
              className="w-full"
              id="max-token-input"
            />
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
