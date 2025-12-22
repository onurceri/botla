import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Check, MessageSquare, Sparkles, PlusCircle, Trash2, Power } from 'lucide-react'

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
    <div className="space-y-6">
      {/* Status Toggle */}
      <div className="flex items-center justify-between p-3.5 bg-slate-50/50 rounded-2xl border border-slate-100 transition-all">
        <div className="flex items-center gap-2.5">
          <div className={`p-1.5 rounded-lg transition-colors ${suggestionsEnabled ? 'bg-emerald-100 text-emerald-600' : 'bg-slate-200 text-slate-400'}`}>
            <Power className="w-3.5 h-3.5" />
          </div>
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
            Öneri Soruları
          </span>
        </div>
        
        <button
          type="button"
          onClick={() => setSuggestionsEnabled(!suggestionsEnabled)}
          className={`relative inline-flex h-5 w-10 items-center rounded-full transition-all duration-300 ${
            suggestionsEnabled ? 'bg-emerald-500 shadow-[0_2px_8px_rgba(16,185,129,0.3)]' : 'bg-slate-200'
          } hover:scale-105 active:scale-95 cursor-pointer`}
        >
          <span
            className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow-sm transition-transform duration-300 ${
              suggestionsEnabled ? 'translate-x-[22px]' : 'translate-x-1'
            }`}
          />
        </button>
      </div>

      <div className={`space-y-6 transition-all duration-500 ${!suggestionsEnabled ? 'opacity-40 grayscale pointer-events-none blur-[1px]' : ''}`}>
        {/* AI-Generated Questions Section */}
        {hasGeneratedQuestions && (
          <div className="space-y-3">
            <div className="flex items-center justify-between px-1">
              <div className="flex items-center gap-2">
                <Sparkles className="w-3.5 h-3.5 text-amber-500" />
                <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400">
                  AI Önerileri
                </label>
              </div>
              <span className="text-[9px] font-bold text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full border border-slate-200/50">
                {suggestedQuestions.filter(q => allSuggestedQuestions.includes(q)).length}/{allSuggestedQuestions.length} SEÇİLİ
              </span>
            </div>
            
            <div className="grid grid-cols-1 gap-2">
              {allSuggestedQuestions.map((q, i) => {
                const isSelected = suggestedQuestions.includes(q)
                return (
                  <label 
                    key={i} 
                    className={`
                      group flex items-center gap-3 p-3 rounded-xl border transition-all duration-300 cursor-pointer
                      ${isSelected 
                        ? 'border-primary/30 bg-primary/[0.03] shadow-sm' 
                        : 'border-slate-100 bg-white hover:border-slate-200 hover:bg-slate-50/50'
                      }
                    `}
                  >
                    <div className={`
                      flex items-center justify-center w-5 h-5 rounded-md border-2 transition-all duration-300
                      ${isSelected 
                        ? 'bg-primary border-primary text-white' 
                        : 'border-slate-200 bg-white group-hover:border-slate-300'
                      }
                    `}>
                      {isSelected && <Check className="w-3 h-3 stroke-[3]" />}
                    </div>
                    <input 
                      type="checkbox" 
                      className="sr-only"
                      checked={isSelected}
                      onChange={(e) => toggleQuestion(q, e.target.checked)}
                    />
                    <span className={`text-[12px] font-medium leading-tight transition-colors ${isSelected ? 'text-slate-900' : 'text-slate-600'}`}>
                      {q}
                    </span>
                  </label>
                )
              })}
            </div>
          </div>
        )}

        {/* Custom Questions Section */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 px-1">
            <PlusCircle className="w-3.5 h-3.5 text-primary" />
            <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400">
              Kendi Sorularınız
            </label>
          </div>

          <div className="flex flex-col gap-2">
            {customQuestions.map((q, i) => (
              <div key={i} className="group flex items-center gap-3 p-3 rounded-xl border border-slate-100 bg-white hover:border-slate-200 transition-all duration-300">
                <div className="w-1 h-1 rounded-full bg-primary/40" />
                <span className="text-[12px] font-medium text-slate-700 flex-1 truncate">{q}</span>
                <button 
                  className="p-1.5 rounded-lg text-slate-400 hover:text-rose-500 hover:bg-rose-50 transition-all opacity-0 group-hover:opacity-100" 
                  onClick={() => setSuggestedQuestions((prev) => prev.filter((_, idx) => prev.indexOf(q) !== idx || idx !== prev.indexOf(q)))}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            ))}

            <div className="flex gap-2 mt-1">
              <div className="relative flex-1">
                <Input 
                  id="new-question-input"
                  placeholder="Soru ekle..." 
                  className="bg-slate-50/50 border-slate-100 rounded-xl h-9 px-4 text-[12px] focus:bg-white transition-all focus:ring-1 focus:ring-primary/20"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      addCustomQuestion((e.target as HTMLInputElement).value)
                      ;(e.target as HTMLInputElement).value = ''
                    }
                  }} 
                />
              </div>
              <Button 
                type="button" 
                onClick={() => {
                  const input = document.getElementById('new-question-input') as HTMLInputElement
                  addCustomQuestion(input.value)
                  input.value = ''
                }}
                className="h-9 px-4 rounded-xl font-bold text-[10px] tracking-wider uppercase shadow-sm"
              >
                EKLE
              </Button>
            </div>
          </div>
        </div>

        {/* Empty State */}
        {!hasGeneratedQuestions && customQuestions.length === 0 && (
          <div className="p-8 text-center bg-slate-50/50 rounded-2xl border border-dashed border-slate-200 flex flex-col items-center gap-3">
            <div className="p-2.5 rounded-full bg-white shadow-sm border border-slate-100">
              <MessageSquare className="w-6 h-6 text-slate-200" />
            </div>
            <div>
              <p className="text-[12px] font-bold text-slate-900">Henüz soru yok</p>
              <p className="text-[11px] text-slate-500 mt-0.5 leading-relaxed">
                Manuel soru ekleyebilir veya AI'ın üretmesini sağlayabilirsiniz.
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
