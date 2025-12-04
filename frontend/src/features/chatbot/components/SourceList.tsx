import { Button } from '@/components/ui/button'
import { CheckCircle2, RefreshCw, AlertCircle, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'

type Source = { id: string; source_type: string; original_filename?: string; source_url?: string; status: string; chunk_count: number }

export default function SourceList({ sources, onDelete }: { sources: Source[]; onDelete: (id: string) => void }) {
  return (
    <div className="rounded-2xl border border-border overflow-hidden shadow-sm">
      <table className="w-full text-sm text-left">
        <thead className="bg-muted/40 text-muted-foreground font-medium">
          <tr>
            <th className="px-4 py-3">Tip</th>
            <th className="px-4 py-3">Kaynak Adı</th>
            <th className="px-4 py-3">Durum</th>
            <th className="px-4 py-3">Parçalar</th>
            <th className="px-4 py-3 text-right">İşlem</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {sources.map((s) => (
            <tr key={s.id} className="hover:bg-muted/50 transition-colors">
              <td className="px-4 py-3 uppercase text-xs font-bold text-muted-foreground">{s.source_type}</td>
              <td className="px-4 py-3 font-medium truncate max-w-[200px] text-foreground" title={s.original_filename || s.source_url}>
                {s.original_filename || s.source_url}
              </td>
              <td className="px-4 py-3">
                <span className={cn(
                  'inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium',
                  s.status === 'completed' ? 'bg-emerald-100 text-emerald-700' :
                  s.status === 'processing' ? 'bg-blue-100 text-blue-700' :
                  s.status === 'failed' ? 'bg-red-100 text-red-700' :
                  'bg-yellow-100 text-yellow-700'
                )}>
                  {s.status === 'completed' && <CheckCircle2 className="w-3 h-3" />}
                  {s.status === 'processing' && <RefreshCw className="w-3 h-3 animate-spin" />}
                  {s.status === 'failed' && <AlertCircle className="w-3 h-3" />}
                  {s.status}
                </span>
              </td>
              <td className="px-4 py-3 text-muted-foreground">{s.chunk_count}</td>
              <td className="px-4 py-3 text-right">
                <Button 
                  variant="ghost" 
                  size="icon" 
                  className="h-8 w-8 text-muted-foreground hover:text-destructive"
                  aria-label="Kaynağı Sil"
                  onClick={() => onDelete(s.id)}
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

