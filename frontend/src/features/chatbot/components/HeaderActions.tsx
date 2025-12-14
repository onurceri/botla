import { Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

type HeaderActionsProps = {
  isNew: boolean
  name: string
  isDeleting: boolean
  isCreating?: boolean
  disabled?: boolean
  onDelete: () => void
  onCreate?: () => void
}

export default function HeaderActions({
  isNew,
  name,
  isDeleting,
  isCreating = false,
  disabled,
  onDelete,
  onCreate,
}: HeaderActionsProps) {
  return (
    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border pb-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">
          {isNew ? 'Yeni Chatbot' : name}
        </h1>
        <p className="text-muted-foreground">
          {isNew ? 'Asistanınızı yapılandırın' : 'Bot ayarlarını ve kaynaklarını yönetin'}
        </p>
      </div>
      <div className="flex items-center gap-2">
        {!isNew && (
          <Button
            variant="destructive"
            size="icon"
            className="mr-2"
            onClick={onDelete}
            isLoading={isDeleting}
            aria-label="Sil"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        )}
        {isNew && onCreate && (
          <Button
            onClick={onCreate}
            className="gap-2"
            isLoading={isCreating}
            disabled={disabled}
            aria-label="Oluştur"
          >
            <Save className="w-4 h-4" />
            Oluştur
          </Button>
        )}
      </div>
    </div>
  )
}
