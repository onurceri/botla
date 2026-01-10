import { useParams } from 'react-router-dom'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import SuggestionsPanel from '../../components/SuggestionsPanel'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useRegenerateSuggestions } from '@/hooks/mutations/useChatbotMutations'

export default function SuggestionsTab() {
  const { id: chatbotId } = useParams()
  const {
    suggestionsEnabled,
    setSuggestionsEnabled,
    suggestedQuestions,
    manualQuestions,
    setManualQuestions,
    buildSuggestionsPayload,
  } = useChatbotContext()

  const regenerate = useRegenerateSuggestions(chatbotId!)

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildSuggestionsPayload(),
  })

  const handleRegenerate = () => {
    regenerate.mutate()
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-4">
            <h2 className="text-2xl font-bold tracking-tight">Konuşma Başlatıcılar</h2>
            <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
          </div>
          <p className="text-muted-foreground">
            Kullanıcıların sohbeti başlatmasını kolaylaştıracak hazır sorular ekleyin.
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleRegenerate}
          disabled={regenerate.isPending}
          className="self-start sm:self-auto"
        >
          <RefreshCw className={`w-4 h-4 mr-2 ${regenerate.isPending ? 'animate-spin' : ''}`} />
          {regenerate.isPending ? 'Yeniden Üretiliyor...' : 'Yeniden Üret'}
        </Button>
      </div>

      <SuggestionsPanel
        suggestionsEnabled={suggestionsEnabled}
        setSuggestionsEnabled={setSuggestionsEnabled}
        suggestedQuestions={suggestedQuestions}
        manualQuestions={manualQuestions}
        setManualQuestions={setManualQuestions}
        maxManualQuestions={3}
      />
    </div>
  )
}
