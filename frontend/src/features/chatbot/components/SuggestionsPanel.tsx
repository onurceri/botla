import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { X } from 'lucide-react'

type Props = {
  suggestionsEnabled: boolean
  setSuggestionsEnabled: (v: boolean) => void
  suggestedQuestions: string[]
  setSuggestedQuestions: (updater: (prev: string[]) => string[]) => void
}

export default function SuggestionsPanel({ suggestionsEnabled, setSuggestionsEnabled, suggestedQuestions, setSuggestedQuestions }: Props) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Örnek Sorular</CardTitle>
        <CardDescription>Önerilen soruları gösterin ve düzenleyin.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium">Örnek soruları göster</label>
          <input type="checkbox" checked={suggestionsEnabled} onChange={(e) => setSuggestionsEnabled(e.target.checked)} />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Sorular</label>
          <div className="flex flex-wrap gap-2 min-h-[40px] p-4 rounded-xl border border-border bg-muted/10">
            {suggestedQuestions.length === 0 && (
              <div className="w-full text-center text-sm text-muted-foreground py-2">
                Henüz örnek soru eklenmemiş.
              </div>
            )}
            {suggestedQuestions.map((q, i) => (
              <div key={i} className="group flex items-center gap-2 px-3 py-1.5 rounded-lg border border-border bg-background shadow-sm hover:border-primary/50 transition-colors">
                <span className="text-sm">{q}</span>
                <button 
                  className="text-muted-foreground hover:text-red-600 transition-colors opacity-0 group-hover:opacity-100" 
                  onClick={() => setSuggestedQuestions((prev) => prev.filter((_, idx) => idx !== i))}
                >
                  <X className="w-3 h-3" />
                </button>
              </div>
            ))}
          </div>
          {suggestedQuestions.length < 6 ? (
            <div className="space-y-2">
              <div className="flex gap-2">
                <Input 
                  id="new-question-input"
                  placeholder="Yeni bir soru yazın..." 
                  className="flex-1"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      const v = (e.target as HTMLInputElement).value.trim()
                      if (v) {
                        setSuggestedQuestions((prev) => {
                          const nv = [...prev, v.slice(0, 120)]
                          const uniq = Array.from(new Set(nv.map(s => s.toLowerCase())))
                          return nv.filter((s, idx) => uniq.indexOf(s.toLowerCase()) === idx).slice(0, 6)
                        })
                        ;(e.target as HTMLInputElement).value = ''
                      }
                    }
                  }} 
                />
                <Button 
                  type="button" 
                  onClick={() => {
                    const input = document.getElementById('new-question-input') as HTMLInputElement
                    const v = input.value.trim()
                    if (v) {
                      setSuggestedQuestions((prev) => {
                        const nv = [...prev, v.slice(0, 120)]
                        const uniq = Array.from(new Set(nv.map(s => s.toLowerCase())))
                        return nv.filter((s, idx) => uniq.indexOf(s.toLowerCase()) === idx).slice(0, 6)
                      })
                      input.value = ''
                    }
                  }}
                >
                  Ekle
                </Button>
              </div>
              <p className="text-[11px] text-muted-foreground flex items-center gap-1">
                <span className="inline-block px-1.5 py-0.5 rounded border border-border bg-muted text-[10px] font-mono">Enter</span> 
                tuşuna basarak veya Ekle butonunu kullanarak ekleyebilirsiniz. (Maks. 6 soru)
              </p>
            </div>
          ) : (
            <div className="text-xs text-amber-600 bg-amber-50 p-2 rounded border border-amber-100">
              Maksimum 6 soru limitine ulaştınız. Yeni eklemek için mevcutlardan silmelisiniz.
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

