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
    <div className="flex items-center justify-between gap-4 mb-2">
      <div className="flex flex-col">
        <div className="flex items-center gap-2">
          <h1 className="text-xl lg:text-2xl font-bold tracking-tight text-foreground truncate max-w-[180px] sm:max-w-[400px]">
            {isNew ? 'Yeni Chatbot' : name}
          </h1>
          {!isNew && (
            <span className="hidden sm:inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-100 text-emerald-700 border border-emerald-200 uppercase tracking-wider">
              Aktif
            </span>
          )}
        </div>
        <p className="text-[10px] lg:text-xs text-muted-foreground font-medium opacity-70">
          {isNew ? 'Asistanınızı yapılandırın' : 'Ayarlar ve Kaynaklar'}
        </p>
      </div>
      
      <div className="flex items-center gap-2">
        {!isNew && (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
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
            size="sm"
            className="h-9 px-4 gap-2 shadow-sm"
            isLoading={isCreating}
            disabled={disabled}
            aria-label="Oluştur"
          >
            <Save className="w-4 h-4" />
            <span className="hidden sm:inline">Oluştur</span>
          </Button>
        )}
      </div>
    </div>
  )
}
