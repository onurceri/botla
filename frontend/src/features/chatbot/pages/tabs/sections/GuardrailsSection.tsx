import { ChevronDown, ChevronUp, Shield } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useState } from 'react'
import GuardrailsSettings from '../../../components/GuardrailsSettings'
import { useChatbotContext } from '../../../context/ChatbotContext'
import { useAutoSave } from '../../../hooks/useAutoSave'
import { useUpdateGuardrails } from '@/hooks/mutations/useChatbotMutations'

interface GuardrailsSectionProps {
  chatbotId: string
}

export default function GuardrailsSection({ chatbotId }: GuardrailsSectionProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const {
    fallbackMessages, setFallbackMessages,
    topicRestrictions, setTopicRestrictions,
    thresholdConfig, setThresholdConfig,
    planConfig,
    buildGuardrailsPayload,
  } = useChatbotContext()

  const { mutateAsync: updateGuardrails } = useUpdateGuardrails(chatbotId)

  useAutoSave({
    payload: buildGuardrailsPayload(),
    saveFn: (_, payload) => updateGuardrails(payload),
  })

  // Summary for collapsed view - just show a simple message
  const summaryText = 'Konu kısıtlamaları ve eşik değerleri'

  return (
    <Card className="border-muted-foreground/20 shadow-sm hover:shadow-md transition-shadow">
      <CardHeader 
        className="cursor-pointer select-none"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-600">
              <Shield className="w-5 h-5" />
            </div>
            <div>
              <CardTitle className="text-lg">Güvenlik Kuralları</CardTitle>
              <CardDescription className="mt-0.5">
                {isExpanded ? 'Konu kısıtlamaları ve eşik değerleri' : summaryText}
              </CardDescription>
            </div>
          </div>
          <Button variant="ghost" size="icon" className="shrink-0">
            {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </Button>
        </div>
      </CardHeader>
      
      {isExpanded && (
        <CardContent className="pt-0 animate-in slide-in-from-top-2 duration-200">
          <GuardrailsSettings
            thresholdConfig={thresholdConfig}
            setThresholdConfig={setThresholdConfig}
            fallbackMessages={fallbackMessages}
            setFallbackMessages={setFallbackMessages}
            topicRestrictions={topicRestrictions}
            setTopicRestrictions={setTopicRestrictions}
            
            canCustomizeThresholds={planConfig?.guardrails?.can_customize_thresholds}
            canUseSmartFallback={planConfig?.guardrails?.can_use_smart_fallback}
            canUseEscalateFallback={planConfig?.guardrails?.can_use_escalate_fallback}
            canManageTopics={planConfig?.guardrails?.can_manage_topics}
            canCustomizeMessages={planConfig?.guardrails?.can_customize_messages}
          />
        </CardContent>
      )}
    </Card>
  )
}
