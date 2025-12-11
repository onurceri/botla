import HandoffSettings from '../../components/HandoffSettings'
import { useChatbotContext } from '../../context/ChatbotContext'

export default function HandoffTab() {
  const {
    handoffEnabled, setHandoffEnabled,
    handoffType, setHandoffType,
    handoffConfig, setHandoffConfig,
  } = useChatbotContext()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">İnsan Devri</h2>
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
      />
    </div>
  )
}
