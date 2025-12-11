import SuggestionsPanel from '../../components/SuggestionsPanel'
import { useChatbotContext } from '../../context/ChatbotContext'

export default function SuggestionsTab() {
  const {
    suggestionsEnabled, setSuggestionsEnabled,
    suggestedQuestions, setSuggestedQuestions,
  } = useChatbotContext()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Konuşma Başlatıcılar</h2>
        <p className="text-muted-foreground">
          Kullanıcıların sohbeti başlatmasını kolaylaştıracak hazır sorular ekleyin.
        </p>
      </div>

      <SuggestionsPanel 
        suggestionsEnabled={suggestionsEnabled}
        setSuggestionsEnabled={setSuggestionsEnabled}
        suggestedQuestions={suggestedQuestions}
        setSuggestedQuestions={setSuggestedQuestions}
      />
    </div>
  )
}
