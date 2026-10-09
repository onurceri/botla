import { useParams } from 'react-router-dom'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import {
  useUpdateBasicInfo,
  useUpdateModelSettings,
  useUpdateSecuritySettings,
} from '@/hooks/mutations/useChatbotMutations'
import IdentityModelSection from './sections/IdentityModelSection'
import VisibilitySection from './sections/VisibilitySection'
import { Settings2 } from 'lucide-react'

export default function SettingsTab() {
  const { id = '' } = useParams()
  const { name, customInstruction, model, temperature, maxTokens, buildVisibilityPayload } =
    useChatbotContext()

  const { mutateAsync: updateBasicInfo } = useUpdateBasicInfo(id)
  const { mutateAsync: updateModelSettings } = useUpdateModelSettings(id)
  const { mutateAsync: updateSecuritySettings } = useUpdateSecuritySettings(id)

  const {
    isSaving: isBasicInfoSaving,
    lastSavedAt: basicInfoSavedAt,
    error: basicInfoError,
  } = useAutoSave({
    payload: { name, description: null, custom_instruction: customInstruction, language: 'tr-TR' },
    saveFn: (_, payload) => updateBasicInfo(payload),
    enabled: !!name.trim(),
  })

  const {
    isSaving: isModelSaving,
    lastSavedAt: modelSavedAt,
    error: modelError,
  } = useAutoSave({
    payload: { model, temperature, max_tokens: maxTokens },
    saveFn: (_, payload) => updateModelSettings(payload),
    enabled: !!model,
  })

  const {
    isSaving: isVisibilitySaving,
    lastSavedAt: visibilitySavedAt,
    error: visibilityError,
  } = useAutoSave({
    payload: buildVisibilityPayload(),
    saveFn: (_, payload) => updateSecuritySettings(payload),
  })

  const isSaving = isBasicInfoSaving || isModelSaving || isVisibilitySaving
  const lastSavedAt =
    [basicInfoSavedAt, modelSavedAt, visibilitySavedAt]
      .filter((d): d is Date => d !== null)
      .sort((a, b) => b.getTime() - a.getTime())[0] ?? null
  const error = basicInfoError || modelError || visibilityError

  return (
    <div className="space-y-6 animate-in fade-in duration-500 pb-10">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-slate-900/5 text-slate-900">
              <Settings2 className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-xl font-bold tracking-tight text-slate-900">Ayarlar</h2>
              <p className="text-sm text-slate-500 font-medium">
                Chatbot davranışlarını ve yeteneklerini yönetin
              </p>
            </div>
          </div>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
      </div>

      {/* Main Content */}
      <div className="space-y-6">
        <IdentityModelSection />
        <VisibilitySection />
      </div>
    </div>
  )
}
