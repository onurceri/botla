import { useParams } from 'react-router-dom'
import ActionList from '../../components/ActionList'

export default function ActionsTab() {
  const { id = '' } = useParams()
  return (
    <div className="space-y-6 animate-in fade-in duration-500 max-w-5xl mx-auto">
      <div className="flex flex-col gap-2 border-b pb-6">
        <h2 className="text-2xl font-bold tracking-tight">Aksiyonlar ve Entegrasyonlar</h2>
        <p className="text-muted-foreground text-lg">
          Botunuzu sadece konuşan bir asistan olmaktan çıkarıp, iş yapan bir çalışana dönüştürün.
        </p>
      </div>
      <ActionList chatbotId={id} />
    </div>
  )
}
