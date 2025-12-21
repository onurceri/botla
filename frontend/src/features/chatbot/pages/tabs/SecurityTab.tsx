import { useParams } from 'react-router-dom'
import { Shield } from 'lucide-react'
import GuardrailsSettings from '../../components/GuardrailsSettings'
import HandoffSettings from '../../components/HandoffSettings'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useUpdateGuardrails, useUpdateHandoff } from '@/hooks/mutations/useChatbotMutations'

export default function SecurityTab() {
  const { id = '' } = useParams()
  const {
    fallbackMessages, setFallbackMessages,
    topicRestrictions, setTopicRestrictions,
    thresholdConfig, setThresholdConfig,
    handoffEnabled, setHandoffEnabled,
    handoffType, setHandoffType,
    handoffConfig, setHandoffConfig,
    planConfig,
    buildGuardrailsPayload,
    buildHandoffPayload,
  } = useChatbotContext()

  const { mutateAsync: updateGuardrails } = useUpdateGuardrails(id)
  const { mutateAsync: updateHandoff } = useUpdateHandoff(id)

  const { isSaving: isGuardrailsSaving, lastSavedAt: guardrailsSavedAt, error: guardrailsError } = useAutoSave({
    payload: buildGuardrailsPayload(),
    saveFn: (_, payload) => updateGuardrails(payload),
  })

  const { isSaving: isHandoffSaving, lastSavedAt: handoffSavedAt, error: handoffError } = useAutoSave({
    payload: buildHandoffPayload(),
    saveFn: (_, payload) => updateHandoff(payload),
  })

  const isSaving = isGuardrailsSaving || isHandoffSaving
  const lastSavedAt = guardrailsSavedAt && handoffSavedAt
    ? (guardrailsSavedAt > handoffSavedAt ? guardrailsSavedAt : handoffSavedAt)
    : (guardrailsSavedAt || handoffSavedAt)
  const error = guardrailsError || handoffError

  const canUseHandoff = planConfig?.guardrails?.can_use_escalate_fallback

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-600">
              <Shield className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold tracking-tight">Güvenlik</h2>
              <p className="text-muted-foreground">
                Konu kısıtlamaları, eşikler ve insan desteği ayarları.
              </p>
            </div>
          </div>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
      </div>

      <div className="space-y-8">
        {/* Guardrails Section */}
        <Card className="border-muted-foreground/20 shadow-sm">
          <CardHeader>
            <CardTitle>Konu Kısıtlamaları ve Eşikler</CardTitle>
            <CardDescription>
              Botunuzun hangi konularda cevap vereceğini ve güven eşiklerini belirleyin.
            </CardDescription>
          </CardHeader>
          <CardContent>
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
        </Card>

        {/* Handoff Section */}
        <Card className="border-muted-foreground/20 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              İnsan Desteği
              {handoffEnabled && canUseHandoff && (
                <span className="px-2 py-0.5 text-xs font-medium bg-green-500/10 text-green-600 rounded-full">
                  Aktif
                </span>
              )}
            </CardTitle>
            <CardDescription>
              Botun cevap veremediği veya kullanıcının talep ettiği durumlarda konuşmayı insana yönlendirin.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <HandoffSettings
              handoffEnabled={handoffEnabled}
              setHandoffEnabled={setHandoffEnabled}
              handoffType={handoffType}
              setHandoffType={setHandoffType}
              handoffConfig={handoffConfig}
              setHandoffConfig={setHandoffConfig}
              canUseHandoff={canUseHandoff}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
