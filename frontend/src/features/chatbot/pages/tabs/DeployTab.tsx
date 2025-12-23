import { useParams } from 'react-router-dom'
import { Rocket } from 'lucide-react'
import EmbeddingCodePanel from '../../components/EmbeddingCodePanel'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useUpdateSecuritySettings } from '@/hooks/mutations/useChatbotMutations'

export default function DeployTab() {
  const { id = '' } = useParams()
  const {
    secureEmbedEnabled,
    setSecureEmbedEnabled,
    allowedDomains,
    setAllowedDomains,
    embedSecret,
    setEmbedSecret,
    planConfig,
    buildConnectPayload,
  } = useChatbotContext()

  const { mutateAsync: updateSecurity } = useUpdateSecuritySettings(id)

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildConnectPayload(),
    saveFn: (_, payload) => updateSecurity(payload),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <Rocket className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold tracking-tight">Yayınla</h2>
              <p className="text-muted-foreground">
                Botunuzu web sitenize eklemek için kod ve güvenlik ayarları.
              </p>
            </div>
          </div>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
      </div>

      <EmbeddingCodePanel
        id={id}
        secureEmbedPlanEnabled={!!planConfig?.security?.secure_embed_enabled}
        secureEmbedEnabled={secureEmbedEnabled}
        allowedDomains={allowedDomains}
        embedSecret={embedSecret}
        onToggleSecure={setSecureEmbedEnabled}
        onDomainsChange={setAllowedDomains}
        onSecretChange={setEmbedSecret}
        onSecretRefresh={() =>
          setEmbedSecret(Math.random().toString(36).slice(2) + Math.random().toString(36).slice(2))
        }
        onSecretClear={() => setEmbedSecret('')}
      />
    </div>
  )
}
