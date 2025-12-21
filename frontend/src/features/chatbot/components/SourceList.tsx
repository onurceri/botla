import { Inbox } from 'lucide-react'
import SourceCard, { Source } from './SourceCard'

interface SourceListProps {
  sources: Source[]
  userPlan: string
  onDelete: (id: string) => void
  onRefresh: (id: string) => void
  refreshingId?: string
}

export default function SourceList({ 
  sources, 
  userPlan, 
  onDelete, 
  onRefresh, 
  refreshingId 
}: SourceListProps) {
  if (sources.length === 0) {
    return (
      <div 
        className="rounded-xl border border-dashed border-muted-foreground/25 bg-muted/30 p-10 text-center space-y-3"
        data-testid="empty-state"
      >
        <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-muted shadow-sm">
          <Inbox className="w-6 h-6 text-muted-foreground" />
        </div>
        <div className="space-y-1">
          <div className="text-sm font-medium text-foreground">Henüz kaynak eklenmemiş</div>
          <div className="text-xs text-muted-foreground">Yukarıdaki alandan kaynaklar ekleyebilirsiniz.</div>
        </div>
      </div>
    )
  }

  return (
    <div 
      className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
      data-testid="source-list"
    >
      {sources.map((source) => (
        <SourceCard
          key={source.id}
          source={source}
          userPlan={userPlan}
          onDelete={onDelete}
          onRefresh={onRefresh}
          isRefreshing={refreshingId === source.id}
        />
      ))}
    </div>
  )
}
