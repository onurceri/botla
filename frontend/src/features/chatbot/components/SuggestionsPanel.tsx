import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { X, Check } from 'lucide-react'

type Props = {
  suggestionsEnabled: boolean
  setSuggestionsEnabled: (v: boolean) => void
  suggestedQuestions: string[]
  setSuggestedQuestions: (updater: (prev: string[]) => string[]) => void
  allSuggestedQuestions: string[]
}

export default function SuggestionsPanel({ 
  suggestionsEnabled, 
  setSuggestionsEnabled, 
  suggestedQuestions, 
  setSuggestedQuestions,
  allSuggestedQuestions 
}: Props) {
  // Toggle a question's visibility
  const toggleQuestion = (question: string, checked: boolean) => {
    if (checked) {
      setSuggestedQuestions((prev) => {
        if (prev.includes(question)) return prev
        return [...prev, question]
      })
    } else {
      setSuggestedQuestions((prev) => prev.filter((q) => q !== question))
    }
  }

  // Add a custom question
  const addCustomQuestion = (value: string) => {
    const v = value.trim()
    if (v) {
      setSuggestedQuestions((prev) => {
        const nv = [...prev, v.slice(0, 120)]
        const uniq = Array.from(new Set(nv.map(s => s.toLowerCase())))
        return nv.filter((s, idx) => uniq.indexOf(s.toLowerCase()) === idx)
      })
    }
  }

  // Combine generated and custom questions for display
  const customQuestions = suggestedQuestions.filter(q => !allSuggestedQuestions.includes(q))
  const hasGeneratedQuestions = allSuggestedQuestions.length > 0

  return (
    <Card>
      <CardHeader>
        <CardTitle>Örnek Sorular</CardTitle>
        <CardDescription>Önerilen soruları gösterin ve düzenleyin.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium">Örnek soruları göster</label>
          <input type="checkbox" checked={suggestionsEnabled} onChange={(e) => setSuggestionsEnabled(e.target.checked)} />
        </div>

        {/* AI-Generated Questions Section */}
        {hasGeneratedQuestions && (
          <div className="space-y-3">
            <label className="text-sm font-medium text-muted-foreground">
              AI Tarafından Üretilen ({suggestedQuestions.filter(q => allSuggestedQuestions.includes(q)).length}/{allSuggestedQuestions.length} seçili)
            </label>
            <div className="grid gap-2">
              {allSuggestedQuestions.map((q, i) => {
                const isSelected = suggestedQuestions.includes(q)
                return (
                  <label 
                    key={i} 
                    className={`
                      flex items-center gap-3 p-3 rounded-lg border cursor-pointer
                      transition-all duration-200
                      ${isSelected 
                        ? 'border-primary bg-primary/5 shadow-sm' 
                        : 'border-border hover:border-primary/30 hover:bg-muted/30'
                      }
                    `}
                  >
                    <div className={`
                      flex items-center justify-center w-5 h-5 rounded border-2 transition-all
                      ${isSelected 
                        ? 'bg-primary border-primary text-primary-foreground' 
                        : 'border-muted-foreground/40'
                      }
                    `}>
                      {isSelected && <Check className="w-3 h-3" />}
                    </div>
                    <input 
                      type="checkbox" 
                      className="sr-only"
                      checked={isSelected}
                      onChange={(e) => toggleQuestion(q, e.target.checked)}
                    />
                    <span className="text-sm flex-1">{q}</span>
                  </label>
                )
              })}
            </div>
          </div>
        )}

        {/* Custom Questions Section */}
        {customQuestions.length > 0 && (
          <div className="space-y-3">
            <label className="text-sm font-medium text-muted-foreground">
              Özel Sorularınız ({customQuestions.length})
            </label>
            <div className="flex flex-wrap gap-2">
              {customQuestions.map((q, i) => (
                <div key={i} className="group flex items-center gap-2 px-3 py-1.5 rounded-lg border border-border bg-background shadow-sm hover:border-primary/50 transition-colors">
                  <span className="text-sm">{q}</span>
                  <button 
                    className="text-muted-foreground hover:text-red-600 transition-colors opacity-0 group-hover:opacity-100" 
                    onClick={() => setSuggestedQuestions((prev) => prev.filter((_, idx) => prev.indexOf(q) !== idx || idx !== prev.indexOf(q)))}
                  >
                    <X className="w-3 h-3" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Empty State */}
        {!hasGeneratedQuestions && customQuestions.length === 0 && (
          <div className="text-center text-sm text-muted-foreground py-6 border rounded-lg bg-muted/10">
            Henüz örnek soru yok. Kaynak ekleyerek otomatik üretilebilir veya manuel ekleyebilirsiniz.
          </div>
        )}

        {/* Add Custom Question */}
        <div className="space-y-2 pt-2 border-t">
          <label className="text-sm font-medium">Özel Soru Ekle</label>
          <div className="flex gap-2">
            <Input 
              id="new-question-input"
              placeholder="Yeni bir soru yazın..." 
              className="flex-1"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  addCustomQuestion((e.target as HTMLInputElement).value)
                  ;(e.target as HTMLInputElement).value = ''
                }
              }} 
            />
            <Button 
              type="button" 
              onClick={() => {
                const input = document.getElementById('new-question-input') as HTMLInputElement
                addCustomQuestion(input.value)
                input.value = ''
              }}
            >
              Ekle
            </Button>
          </div>
          <p className="text-[11px] text-muted-foreground flex items-center gap-1">
            <span className="inline-block px-1.5 py-0.5 rounded border border-border bg-muted text-[10px] font-mono">Enter</span> 
            tuşuna basarak veya Ekle butonunu kullanarak ekleyebilirsiniz.
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
