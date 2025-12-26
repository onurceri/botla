/**
 * Step 1: Bot Name Input
 * Allows the user to name their chatbot.
 */

import { Bot } from 'lucide-react'
import { Input } from '@/components/ui/input'

interface StepBotNameProps {
  botName: string
  onBotNameChange: (name: string) => void
}

export function StepBotName({ botName, onBotNameChange }: StepBotNameProps) {
  return (
    <div className="space-y-6 animate-fade-up">
      <div className="text-center mb-8">
        <div
          className="inline-flex items-center justify-center w-16 h-16 rounded-2xl 
                      bg-primary/10 mb-4"
        >
          <Bot className="w-8 h-8 text-primary" />
        </div>
        <h2 className="heading-md text-foreground mb-2">Botunuza İsim Verin</h2>
        <p className="body-sm">Bu isim dashboard'da ve widget'ta görünecektir</p>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-foreground" htmlFor="botName">
          Bot Adı
        </label>
        <Input
          id="botName"
          placeholder="Örn: Müşteri Destek Botu"
          value={botName}
          onChange={(e) => onBotNameChange(e.target.value)}
          className="h-12 rounded-xl border-border/50 bg-white/50 
                   focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
        />
        <p className="text-xs text-muted-foreground">Minimum 2 karakter</p>
      </div>
    </div>
  )
}
