import { Bot, Cpu, Sparkles, Gauge } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useChatbotContext } from '../../context/ChatbotContext'

export default function OverviewTab() {
  const {
    name, setName,
    customInstruction, setCustomInstruction,
    model, setModel,
    temperature, setTemperature,
    maxTokens, setMaxTokens,
  } = useChatbotContext()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Genel Bakış</h2>
        <p className="text-muted-foreground">
          Chatbotunuzun kimliğini ve temel yapay zeka davranışlarını yapılandırın.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="h-full border-muted-foreground/20 shadow-sm hover:shadow-md transition-shadow">
          <CardHeader>
            <div className="flex items-center gap-2 mb-2">
              <div className="p-2 rounded-lg bg-primary/10 text-primary">
                <Bot className="w-5 h-5" />
              </div>
              <CardTitle>Kimlik</CardTitle>
            </div>
            <CardDescription>
              Botunuzun ismi ve özel talimatları.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="name">Bot İsmi</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Örn: Müşteri Temsilcisi"
                className="bg-background/50"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="customInstruction" className="flex justify-between">
                <span>Özel Talimatlar</span>
                <span className="text-xs text-muted-foreground font-normal">{customInstruction.length} karakter</span>
              </Label>
              <Textarea
                id="customInstruction"
                className="min-h-[300px] resize-none bg-background/50 leading-relaxed font-mono text-sm"
                value={customInstruction}
                onChange={(e) => setCustomInstruction(e.target.value)}
                placeholder="Botunuza özel davranış kuralları ekleyin... Örn: Müşterilere resmi bir dil kullan, fiyat bilgisi verme..."
              />
              <p className="text-xs text-muted-foreground">
                Botunuzun nasıl davranması gerektiğini, tonunu ve özel kurallarını buraya yazın. Dil ve kapsam kuralları otomatik eklenir.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card className="h-full border-muted-foreground/20 shadow-sm hover:shadow-md transition-shadow">
          <CardHeader>
            <div className="flex items-center gap-2 mb-2">
              <div className="p-2 rounded-lg bg-blue-500/10 text-blue-500">
                <Cpu className="w-5 h-5" />
              </div>
              <CardTitle>Model Ayarları</CardTitle>
            </div>
            <CardDescription>
              Yapay zeka modelini ve teknik parametreleri seçin.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-8">
            <div className="space-y-2">
              <Label>Yapay Zeka Modeli</Label>
              <Select value={model} onValueChange={setModel}>
                <SelectTrigger className="bg-background/50 h-11">
                  <SelectValue placeholder="Model seçin" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel>OpenAI</SelectLabel>
                    <SelectItem value="openai:gpt-4o">GPT-4o (Önerilen)</SelectItem>
                    <SelectItem value="openai:gpt-4o-mini">GPT-4o Mini (Hızlı)</SelectItem>
                  </SelectGroup>
                  <SelectGroup>
                    <SelectLabel>Anthropic</SelectLabel>
                    <SelectItem value="anthropic:claude-3-5-sonnet-latest">Claude 3.5 Sonnet</SelectItem>
                    <SelectItem value="anthropic:claude-3-5-haiku-latest">Claude 3.5 Haiku</SelectItem>
                  </SelectGroup>
                  <SelectGroup>
                    <SelectLabel>Google</SelectLabel>
                    <SelectItem value="google:gemini-1.5-pro">Gemini 1.5 Pro</SelectItem>
                    <SelectItem value="google:gemini-1.5-flash">Gemini 1.5 Flash</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-6 pt-6 border-t border-border">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <Label className="flex items-center gap-2">
                    <Sparkles className="w-4 h-4 text-amber-500" />
                    Yaratıcılık (Temperature)
                  </Label>
                  <span className="text-sm font-bold text-primary bg-primary/10 px-3 py-1 rounded-full min-w-[3rem] text-center">
                    {temperature}
                  </span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.1"
                  className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-primary hover:accent-primary/80 transition-all"
                  value={temperature}
                  onChange={(e) => setTemperature(parseFloat(e.target.value))}
                />
                <div className="flex justify-between text-xs text-muted-foreground font-medium">
                  <span>Daha Tutarlı (0.0)</span>
                  <span>Daha Yaratıcı (1.0)</span>
                </div>
              </div>

              <div className="space-y-4 pt-2">
                 <div className="flex items-center justify-between">
                  <Label className="flex items-center gap-2">
                    <Gauge className="w-4 h-4 text-green-500" />
                    Maksimum Token
                  </Label>
                </div>
                 <div className="flex items-center gap-4">
                    <Input
                      type="number"
                      min="1"
                      max="8192"
                      value={maxTokens}
                      onChange={(e) => setMaxTokens(parseInt(e.target.value) || 512)}
                      className="bg-background/50 h-11"
                    />
                 </div>
                 <p className="text-xs text-muted-foreground">
                   Her cevap için üretilecek maksimum kelime/token sayısı.
                 </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
