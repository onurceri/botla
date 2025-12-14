import HandoffSettings from '../../components/HandoffSettings'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

export default function HandoffTab() {
  const {
    handoffEnabled, setHandoffEnabled,
    handoffType, setHandoffType,
    handoffConfig, setHandoffConfig,
    planConfig,
    buildHandoffPayload,
  } = useChatbotContext()

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildHandoffPayload(),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">İnsan Devri</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Botun cevap veremediği veya kullanıcının talep ettiği durumlarda konuşmayı insana yönlendirin.
        </p>
      </div>

      <HandoffSettings
        handoffEnabled={handoffEnabled}
        setHandoffEnabled={setHandoffEnabled}
        handoffType={handoffType}
        setHandoffType={setHandoffType}
        handoffConfig={handoffConfig}
        setHandoffConfig={setHandoffConfig}
        canUseHandoff={planConfig?.guardrails?.can_use_escalate_fallback}
      />
    </div>
  )
}
