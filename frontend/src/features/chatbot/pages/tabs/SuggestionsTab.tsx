import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api } from '@/api/client'
import SuggestionsPanel from '../../components/SuggestionsPanel'
import { useChatbotContext } from '../../context/ChatbotContext'

export default function SuggestionsTab() {
  const { id: chatbotId } = useParams()
  const {
    suggestionsEnabled, setSuggestionsEnabled,
    suggestedQuestions, setSuggestedQuestions,
    allSuggestedQuestions, setAllSuggestedQuestions,
  } = useChatbotContext()

  const [isRegenerating, setIsRegenerating] = useState(false)

  const handleRegenerate = async () => {
    if (!chatbotId) return
    setIsRegenerating(true)
    try {
      await api.post(`/api/v1/chatbots/${chatbotId}/suggestions/regenerate`)
      // Refetch chatbot to get updated suggestions after a short delay
      setTimeout(async () => {
        const { data } = await api.get(`/api/v1/chatbots/${chatbotId}`)
        if (data.suggested_questions) {
          setSuggestedQuestions(() => data.suggested_questions)
        }
        if (data.all_suggested_questions) {
          setAllSuggestedQuestions(() => data.all_suggested_questions)
        }
        setIsRegenerating(false)
      }, 2000) // Wait 2 seconds for background processing
    } catch (err) {
      console.error('Failed to regenerate suggestions:', err)
      setIsRegenerating(false)
    }
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex flex-col gap-2">
          <h2 className="text-2xl font-bold tracking-tight">Konuşma Başlatıcılar</h2>
          <p className="text-muted-foreground">
            Kullanıcıların sohbeti başlatmasını kolaylaştıracak hazır sorular ekleyin.
          </p>
        </div>
        <Button 
          variant="outline" 
          size="sm"
          onClick={handleRegenerate}
          disabled={isRegenerating}
          className="self-start sm:self-auto"
        >
          <RefreshCw className={`w-4 h-4 mr-2 ${isRegenerating ? 'animate-spin' : ''}`} />
          {isRegenerating ? 'Yeniden Üretiliyor...' : 'Yeniden Üret'}
        </Button>
      </div>

      <SuggestionsPanel 
        suggestionsEnabled={suggestionsEnabled}
        setSuggestionsEnabled={setSuggestionsEnabled}
        suggestedQuestions={suggestedQuestions}
        setSuggestedQuestions={setSuggestedQuestions}
        allSuggestedQuestions={allSuggestedQuestions}
      />
    </div>
  )
}
