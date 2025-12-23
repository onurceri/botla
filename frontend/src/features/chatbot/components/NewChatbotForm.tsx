import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

type Props = {
  name: string
  description: string
  onNameChange: (v: string) => void
  onDescriptionChange: (v: string) => void
}

export default function NewChatbotForm({
  name,
  description,
  onNameChange,
  onDescriptionChange,
}: Props) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Temel Bilgiler</CardTitle>
        <CardDescription>Botunuzu oluşturmak için bir isim verin.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Bot İsmi</label>
          <Input
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            placeholder="Örn: Müşteri Temsilcisi"
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Açıklama (Opsiyonel)</label>
          <Input
            value={description}
            onChange={(e) => onDescriptionChange(e.target.value)}
            placeholder="Botun amacı nedir?"
          />
        </div>
      </CardContent>
    </Card>
  )
}
