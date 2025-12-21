import { useParams } from 'react-router-dom'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useUpdateBasicInfo, useUpdateModelSettings } from '@/hooks/mutations/useChatbotMutations'
import IdentityModelSection from './sections/IdentityModelSection'
import GuardrailsSection from './sections/GuardrailsSection'
import HandoffSection from './sections/HandoffSection'

export default function SettingsTab() {
  const { id = '' } = useParams()
  const {
    name, setName,
    customInstruction, setCustomInstruction,
    model, setModel,
    temperature, setTemperature,
    maxTokens, setMaxTokens,
    availableModels,
  } = useChatbotContext()

  const { mutateAsync: updateBasicInfo } = useUpdateBasicInfo(id)
  const { mutateAsync: updateModelSettings } = useUpdateModelSettings(id)

  const { isSaving: isBasicInfoSaving, lastSavedAt: basicInfoSavedAt, error: basicInfoError } = useAutoSave({
    payload: { name, description: null, custom_instruction: customInstruction, language: 'tr-TR' },
    saveFn: (_, payload) => updateBasicInfo(payload),
    enabled: !!name.trim(),
  })

  const { isSaving: isModelSaving, lastSavedAt: modelSavedAt, error: modelError } = useAutoSave({
    payload: { model, temperature, max_tokens: maxTokens },
    saveFn: (_, payload) => updateModelSettings(payload),
    enabled: !!model,
  })

  const isSaving = isBasicInfoSaving || isModelSaving
  const lastSavedAt = basicInfoSavedAt && modelSavedAt
    ? (basicInfoSavedAt > modelSavedAt ? basicInfoSavedAt : modelSavedAt)
    : (basicInfoSavedAt || modelSavedAt)
  const error = basicInfoError || modelError

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Ayarlar</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Chatbotunuzun kimliğini, AI davranışlarını ve güvenlik kurallarını yapılandırın.
        </p>
      </div>

      <div className="space-y-6">
        <IdentityModelSection
          name={name}
          setName={setName}
          customInstruction={customInstruction}
          setCustomInstruction={setCustomInstruction}
          model={model}
          setModel={setModel}
          temperature={temperature}
          setTemperature={setTemperature}
          maxTokens={maxTokens}
          setMaxTokens={setMaxTokens}
          availableModels={availableModels}
        />

        <GuardrailsSection chatbotId={id} />

        <HandoffSection chatbotId={id} />
      </div>
    </div>
  )
}
