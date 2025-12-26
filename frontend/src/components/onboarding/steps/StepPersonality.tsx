/**
 * Step 3: Personality Configuration
 * Allows the user to customize the chatbot's behavior and welcome message.
 */

import { Palette } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

interface StepPersonalityProps {
  systemPrompt: string
  welcomeMessage: string
  onSystemPromptChange: (prompt: string) => void
  onWelcomeMessageChange: (message: string) => void
}

export function StepPersonality({
  systemPrompt,
  welcomeMessage,
  onSystemPromptChange,
  onWelcomeMessageChange,
}: StepPersonalityProps) {
  return (
    <div className="space-y-6 animate-fade-up">
      <div className="text-center mb-8">
        <div
          className="inline-flex items-center justify-center w-16 h-16 rounded-2xl 
                      bg-primary/10 mb-4"
        >
          <Palette className="w-8 h-8 text-primary" />
        </div>
        <h2 className="heading-md text-foreground mb-2">Kişiliğini Belirleyin</h2>
        <p className="body-sm">Botunuzun nasıl davranacağını ve konuşacağını ayarlayın</p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground" htmlFor="systemPrompt">
            Sistem Talimatı
          </label>
          <Textarea
            id="systemPrompt"
            placeholder="Botunuzun nasıl davranacağını açıklayın..."
            value={systemPrompt}
            onChange={(e) => onSystemPromptChange(e.target.value)}
            className="min-h-[100px] rounded-xl border-border/50 bg-white/50 
                     focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
          />
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground" htmlFor="welcomeMessage">
            Karşılama Mesajı
          </label>
          <Input
            id="welcomeMessage"
            placeholder="Merhaba! Size nasıl yardımcı olabilirim?"
            value={welcomeMessage}
            onChange={(e) => onWelcomeMessageChange(e.target.value)}
            className="h-12 rounded-xl border-border/50 bg-white/50 
                     focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
          />
        </div>
      </div>
    </div>
  )
}
