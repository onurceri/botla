import { useParams } from 'react-router-dom'
import { Zap } from 'lucide-react'
import ActionList from '../../components/ActionList'

export default function ActionsTab() {
  const { id = '' } = useParams()
  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-orange-500/10 text-orange-500">
            <Zap className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Aksiyonlar</h2>
            <p className="text-muted-foreground">
              Botunuzu API'larla entegre edin ve akıllı işlemler gerçekleştirin.
            </p>
          </div>
        </div>
      </div>
      <ActionList chatbotId={id} />
    </div>
  )
}
