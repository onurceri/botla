import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { CheckCircle2, RefreshCw, AlertCircle, Trash2, FileText, Database } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import ChunkInspector from './ChunkInspector'

type Source = { 
  id: string; 
  source_type: string; 
  original_filename?: string; 
  source_url?: string; 
  status: string; 
  chunk_count: number; 
  capability_summary?: string;
  error_message?: string;
}

interface SourceListProps {
  sources: Source[]
  userPlan: string
  onDelete: (id: string) => void
  onRefresh: (id: string) => void
  refreshingId?: string
}

const StatusBadge = ({ status, errorMessage }: { status: string, errorMessage?: string }) => {
  const content = (
     <span className={cn(
        'inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium shrink-0',
        status === 'completed' ? 'bg-emerald-100 text-emerald-700' :
        status === 'processing' ? 'bg-blue-100 text-blue-700' :
        status === 'failed' ? 'bg-red-100 text-red-700' :
        'bg-yellow-100 text-yellow-700',
        status === 'failed' && errorMessage && 'cursor-help'
      )}>
        {status === 'completed' && <CheckCircle2 className="w-3 h-3" />}
        {status === 'processing' && <RefreshCw className="w-3 h-3 animate-spin" />}
        {status === 'failed' && <AlertCircle className="w-3 h-3" />}
        {status === 'pending' && <RefreshCw className="w-3 h-3 animate-pulse" />}
        {status}
      </span>
  )

  if (status === 'failed' && errorMessage) {
    return (
      <TooltipProvider delayDuration={0}>
        <Tooltip>
          <TooltipTrigger asChild>{content}</TooltipTrigger>
          <TooltipContent side="top">
            <p className="max-w-xs break-words">{errorMessage}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }
  return content
}

export default function SourceList({ sources, userPlan, onDelete, onRefresh, refreshingId }: SourceListProps) {
  const [inspectingSourceId, setInspectingSourceId] = useState<string | null>(null)
  const canRefresh = userPlan !== 'free'

  return (
    <div className="space-y-4">
      <ChunkInspector 
        sourceId={inspectingSourceId || ''} 
        open={!!inspectingSourceId} 
        onOpenChange={(open) => !open && setInspectingSourceId(null)} 
      />

      {/* Mobile View: Cards */}
      <div className="md:hidden space-y-4">
        {sources.map((s) => (
          <Card key={s.id}>
            <CardHeader className="pb-2">
              <div className="flex justify-between items-start">
                <div className="space-y-1">
                  <div className="text-xs font-bold text-muted-foreground uppercase">{s.source_type}</div>
                  <CardTitle className="text-base break-all line-clamp-2" title={s.original_filename || s.source_url}>
                    {s.original_filename || s.source_url}
                  </CardTitle>
                </div>
                <StatusBadge status={s.status} errorMessage={s.error_message} />
              </div>
            </CardHeader>
            <CardContent className="pb-2 text-sm text-muted-foreground">
               <div className="flex items-center gap-2">
                  <span className="font-medium text-foreground">{s.chunk_count}</span>
                  <span>Parça</span>
               </div>
               <div className="flex flex-wrap gap-2 mt-2">
                 <Button 
                    variant="link" 
                    className="p-0 h-auto text-xs text-muted-foreground hover:text-primary"
                    onClick={() => setInspectingSourceId(s.id)}
                 >
                    <Database className="w-3 h-3 mr-1" />
                    Parçalar
                 </Button>
                 {s.capability_summary && (
                    <Dialog>
                      <DialogTrigger asChild>
                        <Button variant="link" className="p-0 h-auto text-xs text-muted-foreground hover:text-primary">
                          <FileText className="w-3 h-3 mr-1" />
                          Özet
                        </Button>
                      </DialogTrigger>
                      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                        <DialogHeader>
                          <DialogTitle>Kaynak Özeti</DialogTitle>
                        </DialogHeader>
                        <div className="mt-4 text-sm text-foreground whitespace-pre-wrap">
                          {s.capability_summary}
                        </div>
                      </DialogContent>
                    </Dialog>
                 )}
               </div>
            </CardContent>
            <CardFooter className="pt-2 flex justify-end gap-2 border-t">
              {s.source_type === 'url' && (
                <Button 
                  variant="outline" 
                  size="sm" 
                  disabled={!canRefresh || s.status === 'pending' || s.status === 'processing' || refreshingId === s.id}
                  onClick={() => onRefresh(s.id)}
                  className="h-8"
                >
                  <RefreshCw className={cn("w-3.5 h-3.5 mr-2", refreshingId === s.id && "animate-spin")} />
                  Yenile
                </Button>
              )}
              <Button 
                variant="destructive" 
                size="sm"
                onClick={() => onDelete(s.id)}
                className="h-8"
              >
                <Trash2 className="w-3.5 h-3.5 mr-2" />
                Sil
              </Button>
            </CardFooter>
          </Card>
        ))}
        {sources.length === 0 && (
           <div className="text-center py-8 text-muted-foreground bg-muted/20 rounded-xl border border-dashed">
             Henüz kaynak eklenmemiş.
           </div>
        )}
      </div>

      {/* Desktop View: Table */}
      <div className="hidden md:block rounded-2xl border border-border overflow-hidden shadow-sm">
        <table className="w-full text-sm text-left">
          <thead className="bg-muted/40 text-muted-foreground font-medium">
            <tr>
              <th className="px-4 py-3">Tip</th>
              <th className="px-4 py-3">Kaynak Adı</th>
              <th className="px-4 py-3">Durum</th>
              <th className="px-4 py-3">Parçalar</th>
              <th className="px-4 py-3">Özet</th>
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
                  <StatusBadge status={s.status} errorMessage={s.error_message} />
                </td>
                <td className="px-4 py-3 text-muted-foreground">{s.chunk_count}</td>
                <td className="px-4 py-3">
                  {s.capability_summary ? (
                    <Dialog>
                      <DialogTrigger asChild>
                         <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-primary" title="Özeti Gör">
                           <FileText className="w-4 h-4" />
                         </Button>
                      </DialogTrigger>
                      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                        <DialogHeader>
                           <DialogTitle>Kaynak Özeti</DialogTitle>
                        </DialogHeader>
                        <div className="mt-4 text-sm text-foreground whitespace-pre-wrap">
                             {s.capability_summary}
                        </div>
                      </DialogContent>
                    </Dialog>
                  ) : (
                    <span className="text-muted-foreground/30 text-xs">-</span>
                  )}
                </td>
                <td className="px-4 py-3 text-right space-x-1">
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-8 w-8 text-muted-foreground hover:text-primary"
                    aria-label="Parçaları İncele"
                    title="Parçaları İncele"
                    onClick={() => setInspectingSourceId(s.id)}
                  >
                    <Database className="w-4 h-4" />
                  </Button>
                  {s.source_type === 'url' && (
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span>
                            <Button 
                              variant="ghost" 
                              size="icon" 
                              className="h-8 w-8 text-muted-foreground hover:text-primary disabled:opacity-50"
                              aria-label="Kaynağı Yenile"
                              disabled={!canRefresh || s.status === 'pending' || s.status === 'processing' || refreshingId === s.id}
                              onClick={() => onRefresh(s.id)}
                            >
                              <RefreshCw className={cn("w-4 h-4", refreshingId === s.id && "animate-spin")} />
                            </Button>
                          </span>
                        </TooltipTrigger>
                        {!canRefresh && (
                          <TooltipContent>
                            <p>Yenileme özelliği ücretli planlarda aktiftir</p>
                          </TooltipContent>
                        )}
                      </Tooltip>
                    </TooltipProvider>
                  )}
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
             {sources.length === 0 && (
               <tr>
                 <td colSpan={6} className="text-center py-8 text-muted-foreground">
                   Henüz kaynak eklenmemiş.
                 </td>
               </tr>
             )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

