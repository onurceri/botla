import { useParams } from 'react-router-dom'
import EmbeddingCodePanel from '../../components/EmbeddingCodePanel'
import { useChatbotContext } from '../../context/ChatbotContext'

export default function ConnectTab() {
  const { id = '' } = useParams()
  const {
    secureEmbedEnabled, setSecureEmbedEnabled,
    allowedDomains, setAllowedDomains,
    embedSecret, setEmbedSecret,
    planConfig
  } = useChatbotContext()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Bağlantı & Entegrasyon</h2>
        <p className="text-muted-foreground">
          Botunuzu web sitenize eklemek için gerekli kodları ve güvenlik ayarlarını yönetin.
        </p>
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
        onSecretRefresh={() => setEmbedSecret(Math.random().toString(36).slice(2)+Math.random().toString(36).slice(2))}
      />
    </div>
  )
}
