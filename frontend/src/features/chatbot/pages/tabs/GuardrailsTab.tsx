import GuardrailsSettings from '../../components/GuardrailsSettings'
import { useChatbotContext } from '../../context/ChatbotContext'

export default function GuardrailsTab() {
  const {
    confidenceThreshold, setConfidenceThreshold,
    fallbackMessages, setFallbackMessages,
    topicRestrictions, setTopicRestrictions,
    thresholdConfig, setThresholdConfig,
    planConfig
  } = useChatbotContext()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Güvenlik ve Sınırlar</h2>
        <p className="text-muted-foreground">
          Botunuzun hangi konularda cevap vereceğini ve güven eşiklerini yapılandırın.
        </p>
      </div>

      <GuardrailsSettings
        confidenceThreshold={confidenceThreshold}
        setConfidenceThreshold={setConfidenceThreshold}
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
    </div>
  )
}
