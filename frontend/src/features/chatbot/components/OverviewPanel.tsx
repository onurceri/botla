import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useChatbotContext, type ModelInfo } from '../context/ChatbotContext'

type Props = {
  name: string
  setName: (v: string) => void
  customInstruction: string
  setCustomInstruction: (v: string) => void
  model: string
  setModel: (v: string) => void
  temperature: number
  setTemperature: (v: number) => void
  maxTokens: number
  setMaxTokens: (v: number) => void
}

export default function OverviewPanel({
  name,
  setName,
  customInstruction,
  setCustomInstruction,
  model,
  setModel,
  temperature,
  setTemperature,
  maxTokens,
  setMaxTokens,
}: Props) {
  // Get available models from context (fetched from backend)
  const { availableModels } = useChatbotContext()

  return (
    <div className="grid gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Kimlik</CardTitle>
          <CardDescription>Bot ismi ve özel talimatlar.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">İsim</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Özel Talimatlar</label>
            <textarea
              className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              value={customInstruction}
              onChange={(e) => setCustomInstruction(e.target.value)}
              placeholder="Botunuza özel davranış kuralları ekleyin..."
            />
            <div className="flex justify-end text-xs text-muted-foreground">
              {customInstruction.length} karakter
            </div>
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
              {availableModels.length > 0 ? (
                availableModels.map((m: ModelInfo) => (
                  <option key={m.id} value={m.id}>
                    {m.name}
                  </option>
                ))
              ) : (
                <option value={model}>{model}</option>
              )}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">
                Yaratıcılık (Temperature): {temperature}
              </label>
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
