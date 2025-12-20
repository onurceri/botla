import { useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useInfiniteQuery } from '@tanstack/react-query'
import { getSourceChunks } from '@/api/source'
import { Loader2, Search, Database } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

interface ChunkInspectorProps {
  sourceId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function ChunkInspector({ sourceId, open, onOpenChange }: ChunkInspectorProps) {
  const [search, setSearch] = useState('')

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError
  } = useInfiniteQuery({
    queryKey: ['sourceChunks', sourceId],
    queryFn: ({ pageParam }) => getSourceChunks(sourceId, 20, pageParam as string | undefined),
    getNextPageParam: (lastPage) => lastPage.next_cursor || undefined,
    initialPageParam: undefined as string | undefined,
    enabled: open && !!sourceId,
  })

  const chunks = data?.pages.flatMap((page) => page.chunks) || []
  
  // Client-side filtering for loaded chunks
  const filteredChunks = chunks.filter(c => 
    !search || c.payload.original_text.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl h-[80vh] flex flex-col p-0 gap-0">
        <DialogHeader className="p-6 pb-4 border-b">
          <DialogTitle className="flex items-center gap-2">
            <Database className="w-5 h-5 text-primary" />
            Kaynak Parçaları
          </DialogTitle>
          <DialogDescription>
            Bu kaynaktan çıkarılan metin parçalarını inceleyin.
          </DialogDescription>
          <div className="pt-4 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input 
              placeholder="Yüklenen parçalarda ara..." 
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9"
            />
          </div>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto p-6 space-y-4" id="chunk-scroll-container">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-2">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
              <span>Parçalar yükleniyor...</span>
            </div>
          ) : isError ? (
            <div className="flex flex-col items-center justify-center h-full text-destructive gap-2">
              <span className="font-medium">Parçalar yüklenemedi.</span>
              <Button variant="outline" onClick={() => window.location.reload()}>Tekrar Dene</Button>
            </div>
          ) : filteredChunks.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
              {search ? 'Arama sonucu bulunamadı.' : 'Bu kaynak için parça bulunamadı.'}
            </div>
          ) : (
            <div className="space-y-4">
              {filteredChunks.map((chunk) => (
                <div key={chunk.id} className="bg-muted/30 border border-border rounded-lg p-4 text-sm relative group hover:border-primary/50 transition-colors">
                  <div className="absolute top-3 right-3 text-xs text-muted-foreground font-mono bg-muted px-1.5 py-0.5 rounded opacity-50 group-hover:opacity-100 transition-opacity">
                    #{chunk.payload.chunk_index}
                  </div>
                  <p className="whitespace-pre-wrap leading-relaxed text-foreground/90 font-mono text-xs">
                    {chunk.payload.original_text}
                  </p>
                  <div className="mt-3 pt-3 border-t border-border/50 flex justify-between items-center text-xs text-muted-foreground">
                    <span>ID: {chunk.id.slice(0, 8)}...</span>
                    <span>Score: {chunk.score.toFixed(4)}</span>
                  </div>
                </div>
              ))}
              
              {hasNextPage && (
                <div className="pt-4 flex justify-center">
                  <Button 
                    variant="outline" 
                    onClick={() => fetchNextPage()} 
                    disabled={isFetchingNextPage}
                    className="w-full"
                  >
                    {isFetchingNextPage ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin mr-2" />
                        Yükleniyor...
                      </>
                    ) : (
                      'Daha Fazla Yükle'
                    )}
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
