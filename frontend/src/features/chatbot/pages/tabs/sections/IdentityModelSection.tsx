import { Bot, Cpu, Sparkles, Gauge, Activity, BrainCircuit, Lock } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useChatbotContext } from '../../../context/ChatbotContext'

export default function IdentityModelSection() {
  const {
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
    availableModels,
    planConfig,
  } = useChatbotContext()

  // Get limits from plan (default to standard limits if not set or zero)
  const minTokens = planConfig?.chat?.min_response_token_limit || 20
  const maxTokensLimit = planConfig?.chat?.max_response_token_limit || 8192

  return (
    <div className="bg-white rounded-[24px] border border-slate-200/60 shadow-sm overflow-hidden flex flex-col h-full">
      {/* Header */}
      <div className="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-blue-500/10 text-blue-600 ring-1 ring-blue-500/20 shadow-sm">
            <BrainCircuit className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-bold tracking-tight text-slate-900 uppercase">
              Akil ve Kimlik
            </h3>
            <p className="text-[11px] text-slate-500 font-medium">
              Temel yapılandırma ve model ayarları
            </p>
          </div>
        </div>
      </div>

      <div className="p-6 lg:p-8 space-y-8 flex-1">
        {/* Top Row: Name and Model */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 lg:gap-8">
          <div className="space-y-3">
            <label
              htmlFor="name"
              className="flex items-center gap-2 text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1"
            >
              <Bot className="w-3.5 h-3.5" />
              Bot İsmi
            </label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Örn: Asistan"
              className="h-12 rounded-xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-primary/20 transition-all font-medium text-slate-900 px-4"
            />
          </div>

          <div className="space-y-3">
            <label className="flex items-center gap-2 text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">
              <Cpu className="w-3.5 h-3.5" />
              Yapay Zeka Modeli
            </label>
            <Select value={model} onValueChange={setModel}>
              <SelectTrigger className="h-12 rounded-xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-primary/20 transition-all font-medium text-slate-900 px-4">
                <SelectValue placeholder="Model seçin" />
              </SelectTrigger>
              <SelectContent className="max-h-[300px]">
                {availableModels && availableModels.length > 0 ? (
                  availableModels.map((m) => (
                    <SelectItem key={m.id} value={m.id} className="font-medium">
                      {m.name}
                    </SelectItem>
                  ))
                ) : (
                  <SelectItem value="loading" disabled>
                    Yükleniyor...
                  </SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Instructions */}
        <div className="space-y-3 flex flex-col flex-1">
          <div className="flex items-center justify-between ml-1">
            <label
              htmlFor="customInstruction"
              className="flex items-center gap-2 text-[11px] font-bold text-slate-500 uppercase tracking-widest"
            >
              <Activity className="w-3.5 h-3.5" />
              Özel Talimatlar / Sistem Mesajı
            </label>
            <span className="text-[10px] font-bold text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full tabular-nums">
              {customInstruction.length} karakter
            </span>
          </div>
          <div className="relative group">
            <Textarea
              id="customInstruction"
              className="min-h-[340px] w-full rounded-2xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-primary/20 transition-all font-mono text-sm leading-relaxed p-6 resize-y"
              value={customInstruction}
              onChange={(e) => setCustomInstruction(e.target.value)}
              placeholder="# Bot Kimliği&#10;- Sen yardımsever bir asistansın.&#10;&#10;# Kurallar&#10;- Kullanıcıya her zaman kibar davran.&#10;- Bilmediğin konularda spekülasyon yapma."
            />
            <div className="absolute bottom-4 right-4 pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity">
              <span className="text-[10px] text-slate-400 font-medium bg-white/80 backdrop-blur px-2 py-1 rounded-lg border border-slate-100">
                Markdown Desteklenir
              </span>
            </div>
          </div>
          <p className="text-[12px] text-slate-400 ml-1 font-medium">
            Botunuzun nasıl davranması gerektiğini, tonunu ve kısıtlamalarını buraya yazın. Ne kadar
            detaylı olursanız o kadar iyi sonuç alırsınız.
          </p>
        </div>

        {/* Technical Footer */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 pt-8 border-t border-slate-100/80">
          {/* Temperature */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-[11px] font-bold text-slate-500 uppercase tracking-widest">
                <Sparkles className="w-3.5 h-3.5 text-amber-500" />
                Yaratıcılık
              </label>
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-lg bg-amber-500/10 text-amber-600 flex items-center justify-center text-xs font-bold tabular-nums">
                  {temperature}
                </div>
              </div>
            </div>
            <div className="relative pt-1 pb-1">
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                className="w-full h-1.5 bg-slate-100 rounded-lg appearance-none cursor-pointer accent-amber-500 hover:accent-amber-600 transition-all"
                value={temperature}
                onChange={(e) => setTemperature(parseFloat(e.target.value))}
              />
              <div className="flex justify-between mt-2">
                <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                  Tutarlı
                </span>
                <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                  Yaratıcı
                </span>
              </div>
            </div>
            <p className="text-[11px] text-slate-400 ml-1 leading-normal">
              Bu ayar botun kaynaklara sadık kalma seviyesini etkiler. Düşük değerler daha tutarlı
              ve kaynağa bağlı cevaplar üretir.
            </p>
          </div>

          {/* Tokens */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-[11px] font-bold text-slate-500 uppercase tracking-widest">
                <Gauge className="w-3.5 h-3.5 text-green-500" />
                Maksimum Uzunluk
              </label>
              <div className="flex items-center gap-2">
                <div
                  className={`flex items-center gap-1.5 text-[10px] font-bold px-2 py-1 rounded-md ${maxTokens > maxTokensLimit || maxTokens < minTokens ? 'bg-red-50 text-red-600' : 'bg-slate-100 text-slate-500'}`}
                >
                  <Lock className="w-3 h-3" />
                  <span className="tabular-nums">
                    Limit: {minTokens} - {maxTokensLimit}
                  </span>
                </div>
              </div>
            </div>
            <div className="relative group">
              <Input
                type="number"
                min={minTokens}
                max={maxTokensLimit}
                value={maxTokens}
                onChange={(e) => {
                  // Allow empty string to be handled, but convert to int
                  const val = parseInt(e.target.value)
                  if (!isNaN(val)) setMaxTokens(val)
                }}
                className={`h-11 rounded-xl border-slate-200 focus:bg-white focus:ring-2 focus:ring-primary/20 transition-all pl-11 pr-20 font-mono text-sm ${
                  maxTokens > maxTokensLimit || maxTokens < minTokens
                    ? 'bg-red-50/50 border-red-200 text-red-600 focus:ring-red-500/20'
                    : 'bg-slate-50'
                }`}
              />
              <div className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400">
                <Gauge className="w-4 h-4" />
              </div>
              <div className="absolute right-3.5 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400 uppercase pointer-events-none pr-6">
                Tokens
              </div>
            </div>
            <p className="text-[11px] text-slate-400 ml-1 leading-normal">
              Her cevap için üretilecek maksimum kelime sayısı.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
