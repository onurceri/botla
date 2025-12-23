import { useParams } from 'react-router-dom'
import { ChatbotAnalytics } from '@/features/analytics/ChatbotAnalytics'

export default function AnalyticsTab() {
  const { id = '' } = useParams()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Analizler</h2>
        <p className="text-muted-foreground">
          Botunuzun performansını, kullanıcı etkileşimlerini ve kaynak kullanımını takip edin.
        </p>
      </div>

      <ChatbotAnalytics chatbotId={id} />
    </div>
  )
}
