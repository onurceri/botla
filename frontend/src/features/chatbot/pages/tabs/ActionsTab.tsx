import { useParams } from 'react-router-dom'
import ActionList from '../../components/ActionList'

export default function ActionsTab() {
  const { id = '' } = useParams()
  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Aksiyonlar</h2>
        <p className="text-muted-foreground">
          Botunuzun dış sistemlerle entegre olmasını ve fonksiyon çağırmasını sağlayın.
        </p>
      </div>
      <ActionList chatbotId={id} />
    </div>
  )
}
