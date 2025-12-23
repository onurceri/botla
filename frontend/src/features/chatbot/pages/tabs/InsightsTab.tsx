import { useParams } from 'react-router-dom'
import { BarChart3, Inbox } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ChatbotAnalytics } from '@/features/analytics/ChatbotAnalytics'
import HandoffRequestsTab from './HandoffRequestsTab'
import { useState } from 'react'
import { useChatbotContext } from '../../context/ChatbotContext'

export default function InsightsTab() {
  const { id = '' } = useParams()
  const [activeSection, setActiveSection] = useState<'analytics' | 'requests'>('analytics')
  const { planConfig } = useChatbotContext()

  const canUseHandoff = planConfig?.guardrails?.can_use_escalate_fallback

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-primary/10 text-primary">
            <BarChart3 className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Raporlar</h2>
            <p className="text-muted-foreground">Performans analizleri ve destek talepleri.</p>
          </div>
        </div>
      </div>

      <Tabs
        value={activeSection}
        onValueChange={(v) => setActiveSection(v as 'analytics' | 'requests')}
        className="w-full"
      >
        <TabsList className="w-full justify-start bg-muted/50 p-1 h-auto">
          <TabsTrigger
            value="analytics"
            className="flex items-center gap-2 px-4 py-2.5 data-[state=active]:bg-background"
          >
            <BarChart3 className="w-4 h-4" />
            Analizler
          </TabsTrigger>
          {canUseHandoff && (
            <TabsTrigger
              value="requests"
              className="flex items-center gap-2 px-4 py-2.5 data-[state=active]:bg-background"
            >
              <Inbox className="w-4 h-4" />
              Destek Talepleri
            </TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="analytics" className="mt-6">
          <ChatbotAnalytics chatbotId={id} />
        </TabsContent>

        {canUseHandoff && (
          <TabsContent value="requests" className="mt-6">
            <HandoffRequestsTab />
          </TabsContent>
        )}
      </Tabs>
    </div>
  )
}
