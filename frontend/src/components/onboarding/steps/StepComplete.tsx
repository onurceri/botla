/**
 * Step 4: Completion
 * Shows success message and next steps after bot creation.
 */

import { CheckCircle2, Sparkles } from 'lucide-react'

interface StepCompleteProps {
  botName: string
}

const NEXT_STEPS = [
  "Test Alanı'nda botunuzu test edin",
  'Daha fazla kaynak ekleyerek bilgi tabanını genişletin',
  'Web sitenize embed kodu ile entegre edin',
] as const

export function StepComplete({ botName }: StepCompleteProps) {
  return (
    <div className="text-center animate-fade-up">
      <div
        className="inline-flex items-center justify-center w-20 h-20 rounded-full 
                    bg-success/10 mb-6"
      >
        <CheckCircle2 className="w-10 h-10 text-success" />
      </div>
      <h2 className="heading-md text-foreground mb-3">Tebrikler! 🎉</h2>
      <p className="body-lg mb-8">
        <span className="font-semibold text-foreground">{botName}</span> başarıyla oluşturuldu
        ve kullanıma hazır.
      </p>

      <div className="glass-panel p-6 text-left mb-6">
        <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-primary" />
          Şimdi Yapabilecekleriniz
        </h3>
        <ul className="space-y-3 text-sm text-muted-foreground">
          {NEXT_STEPS.map((step, index) => (
            <li key={index} className="flex items-start gap-3">
              <span className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 text-xs font-medium text-primary">
                {index + 1}
              </span>
              <span>{step}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
