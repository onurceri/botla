import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

type Props = {
  name: string
  setName: (v: string) => void
  systemPrompt: string
  setSystemPrompt: (v: string) => void
  model: string
  setModel: (v: string) => void
  temperature: number
  setTemperature: (v: number) => void
  maxTokens: number
  setMaxTokens: (v: number) => void
}

export default function OverviewPanel({ 
  name, setName, 
  systemPrompt, setSystemPrompt,
  model, setModel,
  temperature, setTemperature,
  maxTokens, setMaxTokens
}: Props) {
  return (
    <div className="grid gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Kimlik</CardTitle>
          <CardDescription>Bot ismi ve sistem mesajı.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">İsim</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
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
          <CardDescription>Yapay zeka modelini ve parametrelerini yapılandırın.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Model</label>
            <select
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              value={model}
              onChange={(e) => setModel(e.target.value)}
            >
              <optgroup label="OpenAI">
                <option value="openai:gpt-4o">GPT-4o</option>
                <option value="openai:gpt-4o-mini">GPT-4o Mini</option>
              </optgroup>
              <optgroup label="Anthropic">
                <option value="anthropic:claude-3-5-sonnet-latest">Claude 3.5 Sonnet</option>
                <option value="anthropic:claude-3-5-haiku-latest">Claude 3.5 Haiku</option>
              </optgroup>
              <optgroup label="Google">
                <option value="google:gemini-1.5-pro">Gemini 1.5 Pro</option>
                <option value="google:gemini-1.5-flash">Gemini 1.5 Flash</option>
              </optgroup>
            </select>
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Yaratıcılık (Temperature): {temperature}</label>
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                className="w-full"
                value={temperature}
                onChange={(e) => setTemperature(parseFloat(e.target.value))}
              />
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">Maksimum Token</label>
              <Input 
                type="number" 
                min="1" 
                max="4096" 
                value={maxTokens} 
                onChange={(e) => setMaxTokens(parseInt(e.target.value) || 512)} 
              />
            </div>
          </div>
        </CardContent>
      </Card>

    </div>
  )
}
