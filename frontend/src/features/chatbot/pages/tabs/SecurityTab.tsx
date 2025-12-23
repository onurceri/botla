import { useParams } from 'react-router-dom'
import { ShieldCheck } from 'lucide-react'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useUpdateGuardrails, useUpdateHandoff } from '@/hooks/mutations/useChatbotMutations'
import GuardrailsSection from './sections/GuardrailsSection'
import HandoffSection from './sections/HandoffSection'

export default function SecurityTab() {
  const { id = '' } = useParams()
  const { buildGuardrailsPayload, buildHandoffPayload } = useChatbotContext()

  const { mutateAsync: updateGuardrails } = useUpdateGuardrails(id)
  const { mutateAsync: updateHandoff } = useUpdateHandoff(id)

  const {
    isSaving: isGuardrailsSaving,
    lastSavedAt: guardrailsSavedAt,
    error: guardrailsError,
  } = useAutoSave({
    payload: buildGuardrailsPayload(),
    saveFn: (_, payload) => updateGuardrails(payload),
  })

  const {
    isSaving: isHandoffSaving,
    lastSavedAt: handoffSavedAt,
    error: handoffError,
  } = useAutoSave({
    payload: buildHandoffPayload(),
    saveFn: (_, payload) => updateHandoff(payload),
  })

  const isSaving = isGuardrailsSaving || isHandoffSaving
  const lastSavedAt =
    guardrailsSavedAt && handoffSavedAt
      ? guardrailsSavedAt > handoffSavedAt
        ? guardrailsSavedAt
        : handoffSavedAt
      : guardrailsSavedAt || handoffSavedAt
  const error = guardrailsError || handoffError

  return (
    <div className="space-y-6 animate-in fade-in duration-500 pb-10">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-slate-900/5 text-slate-900">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-xl font-bold tracking-tight text-slate-900">
                Güvenlik ve İnsan Desteği
              </h2>
              <p className="text-sm text-slate-500 font-medium">
                Bot güvenlik ayarlarını ve insan desteği süreçlerini yönetin
              </p>
            </div>
          </div>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
        <GuardrailsSection />
        <div className="lg:sticky lg:top-6">
          <HandoffSection />
        </div>
      </div>
    </div>
  )
}
