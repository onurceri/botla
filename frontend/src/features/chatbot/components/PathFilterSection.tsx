import { Plus, X, Filter, CheckCircle2, XCircle, ChevronDown, ChevronRight, Code2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { useState } from 'react'

interface PathFilterSectionProps {
  includePaths: string[]
  setIncludePaths: (paths: string[]) => void
  excludePaths: string[]
  setExcludePaths: (paths: string[]) => void
  selectorWhitelist: string[]
  setSelectorWhitelist: (selectors: string[]) => void
}

const PathFilterSection = ({
  includePaths,
  setIncludePaths,
  excludePaths,
  setExcludePaths,
  selectorWhitelist,
  setSelectorWhitelist,
}: PathFilterSectionProps) => {
  const [newIncludePath, setNewIncludePath] = useState('')
  const [newExcludePath, setNewExcludePath] = useState('')
  const [newSelector, setNewSelector] = useState('')
  const [isExpanded, setIsExpanded] = useState(false)

  const handleAddIncludePath = () => {
    const trimmed = newIncludePath.trim()
    if (trimmed && !includePaths.includes(trimmed)) {
      setIncludePaths([...includePaths, trimmed])
      setNewIncludePath('')
    }
  }

  const handleAddExcludePath = () => {
    const trimmed = newExcludePath.trim()
    if (trimmed && !excludePaths.includes(trimmed)) {
      setExcludePaths([...excludePaths, trimmed])
      setNewExcludePath('')
    }
  }

  const handleAddSelector = () => {
    const trimmed = newSelector.trim()
    if (trimmed && !selectorWhitelist.includes(trimmed)) {
      setSelectorWhitelist([...selectorWhitelist, trimmed])
      setNewSelector('')
    }
  }

  const handleRemoveIncludePath = (index: number) => {
    setIncludePaths(includePaths.filter((_, i) => i !== index))
  }

  const handleRemoveExcludePath = (index: number) => {
    setExcludePaths(excludePaths.filter((_, i) => i !== index))
  }

  const handleRemoveSelector = (index: number) => {
    setSelectorWhitelist(selectorWhitelist.filter((_, i) => i !== index))
  }

  const handleIncludeKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddIncludePath()
    }
  }

  const handleExcludeKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddExcludePath()
    }
  }

  const handleSelectorKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddSelector()
    }
  }

  const totalFilters = includePaths.length + excludePaths.length + selectorWhitelist.length

  return (
    <div className="mt-4 border border-border/60 rounded-xl bg-gradient-to-b from-white/40 to-white/20 backdrop-blur overflow-hidden">
      {/* Collapsible Header */}
      <button 
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between p-3.5 hover:bg-white/40 transition-all duration-200"
      >
        <div className="flex items-center gap-2.5">
          <div className="p-1.5 rounded-lg bg-blue-500/10">
            <Filter className="w-3.5 h-3.5 text-blue-500" />
          </div>
          <span className="text-sm font-medium text-foreground">Gelişmiş Tarama Ayarları</span>
          {totalFilters > 0 && (
            <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4 font-medium bg-blue-100 text-blue-600 border-0">
              {totalFilters} ayar
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground hidden sm:inline">
            {isExpanded ? 'Gizle' : 'Ayarları göster'}
          </span>
          {isExpanded 
            ? <ChevronDown className="w-4 h-4 text-muted-foreground transition-transform" /> 
            : <ChevronRight className="w-4 h-4 text-muted-foreground transition-transform" />
          }
        </div>
      </button>

      {/* Expandable Content */}
      {isExpanded && (
        <div className="px-4 pb-4 pt-1 border-t border-border/40 animate-in slide-in-from-top-1 duration-200">
          {/* Help Text */}
          <p className="text-xs text-muted-foreground mb-4 leading-relaxed">
            URL filtreleri ve CSS seçicileri ile tarama davranışını özelleştirin.
          </p>

          {/* Path Filters Section */}
          <div className="space-y-5">
            {/* URL Filters */}
            <div className="space-y-4">
              <div className="flex items-center gap-2 mb-3">
                <Filter className="w-3.5 h-3.5 text-blue-500" />
                <span className="text-xs font-semibold text-foreground">URL Filtreleme</span>
                <span className="text-[10px] text-muted-foreground ml-auto">Wildcard (*) kullanabilirsiniz</span>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
                {/* Include Paths */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                    <span className="text-xs font-semibold text-foreground">Dahil Et</span>
                    <span className="text-[10px] text-muted-foreground ml-auto">Boş = Tümü</span>
                  </div>
                  
                  <div className="flex gap-2">
                    <Input
                      placeholder="/blog/*, /docs/*"
                      value={newIncludePath}
                      onChange={(e) => setNewIncludePath(e.target.value)}
                      onKeyDown={handleIncludeKeyDown}
                      className="h-9 text-sm bg-white/80 border-border/60 focus:border-emerald-300 focus:ring-emerald-100 placeholder:text-muted-foreground/60"
                    />
                    <Button
                      type="button"
                      onClick={handleAddIncludePath}
                      size="sm"
                      variant="outline"
                      disabled={!newIncludePath.trim()}
                      className="h-9 w-9 p-0 border-border/60 hover:bg-emerald-50 hover:border-emerald-200 hover:text-emerald-600 disabled:opacity-40"
                    >
                      <Plus className="w-4 h-4" />
                    </Button>
                  </div>

                  <div className="flex flex-wrap gap-1.5 min-h-[28px]">
                    {includePaths.length === 0 ? (
                      <span className="text-xs text-muted-foreground/60 italic">Tüm sayfalar dahil edilecek</span>
                    ) : (
                      includePaths.map((path, index) => (
                        <Badge 
                          key={index} 
                          variant="outline" 
                          className="pl-2.5 pr-1 gap-1.5 text-xs font-normal bg-emerald-50/80 text-emerald-700 border-emerald-200/80 hover:bg-emerald-100 transition-colors"
                        >
                          <span className="font-mono text-[11px]">{path}</span>
                          <button
                            type="button"
                            onClick={() => handleRemoveIncludePath(index)}
                            className="hover:bg-emerald-200/60 rounded-full p-0.5 transition-colors"
                          >
                            <X className="w-3 h-3" />
                          </button>
                        </Badge>
                      ))
                    )}
                  </div>
                </div>

                {/* Exclude Paths */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <XCircle className="w-4 h-4 text-rose-400" />
                    <span className="text-xs font-semibold text-foreground">Hariç Tut</span>
                  </div>
                  
                  <div className="flex gap-2">
                    <Input
                      placeholder="/admin/*, /login/*"
                      value={newExcludePath}
                      onChange={(e) => setNewExcludePath(e.target.value)}
                      onKeyDown={handleExcludeKeyDown}
                      className="h-9 text-sm bg-white/80 border-border/60 focus:border-rose-300 focus:ring-rose-100 placeholder:text-muted-foreground/60"
                    />
                    <Button
                      type="button"
                      onClick={handleAddExcludePath}
                      size="sm"
                      variant="outline"
                      disabled={!newExcludePath.trim()}
                      className="h-9 w-9 p-0 border-border/60 hover:bg-rose-50 hover:border-rose-200 hover:text-rose-600 disabled:opacity-40"
                    >
                      <Plus className="w-4 h-4" />
                    </Button>
                  </div>

                  <div className="flex flex-wrap gap-1.5 min-h-[28px]">
                    {excludePaths.length === 0 ? (
                      <span className="text-xs text-muted-foreground/60 italic">Hiçbir sayfa hariç tutulmayacak</span>
                    ) : (
                      excludePaths.map((path, index) => (
                        <Badge 
                          key={index} 
                          variant="outline" 
                          className="pl-2.5 pr-1 gap-1.5 text-xs font-normal bg-rose-50/80 text-rose-600 border-rose-200/80 hover:bg-rose-100 transition-colors"
                        >
                          <span className="font-mono text-[11px]">{path}</span>
                          <button
                            type="button"
                            onClick={() => handleRemoveExcludePath(index)}
                            className="hover:bg-rose-200/60 rounded-full p-0.5 transition-colors"
                          >
                            <X className="w-3 h-3" />
                          </button>
                        </Badge>
                      ))
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* CSS Selector Section */}
            <div className="pt-4 border-t border-border/30 space-y-3">
              <div className="flex items-center gap-2 mb-3">
                <Code2 className="w-3.5 h-3.5 text-violet-500" />
                <span className="text-xs font-semibold text-foreground">İçerik Seçici (CSS Selector)</span>
                <span className="text-[10px] text-muted-foreground ml-auto">Boş = Tüm sayfa</span>
              </div>
              
              <p className="text-xs text-muted-foreground leading-relaxed mb-3">
                Sadece belirtilen HTML elementlerinden metin çıkarılır. Menü, footer gibi gereksiz alanları hariç tutmak için kullanın.
              </p>

              <div className="flex gap-2">
                <Input
                  placeholder=".content, #article-body, main article"
                  value={newSelector}
                  onChange={(e) => setNewSelector(e.target.value)}
                  onKeyDown={handleSelectorKeyDown}
                  className="h-9 text-sm bg-white/80 border-border/60 focus:border-violet-300 focus:ring-violet-100 placeholder:text-muted-foreground/60 font-mono"
                />
                <Button
                  type="button"
                  onClick={handleAddSelector}
                  size="sm"
                  variant="outline"
                  disabled={!newSelector.trim()}
                  className="h-9 w-9 p-0 border-border/60 hover:bg-violet-50 hover:border-violet-200 hover:text-violet-600 disabled:opacity-40"
                >
                  <Plus className="w-4 h-4" />
                </Button>
              </div>

              <div className="flex flex-wrap gap-1.5 min-h-[28px]">
                {selectorWhitelist.length === 0 ? (
                  <span className="text-xs text-muted-foreground/60 italic">Tüm sayfa içeriği kullanılacak</span>
                ) : (
                  selectorWhitelist.map((selector, index) => (
                    <Badge 
                      key={index} 
                      variant="outline" 
                      className="pl-2.5 pr-1 gap-1.5 text-xs font-normal bg-violet-50/80 text-violet-700 border-violet-200/80 hover:bg-violet-100 transition-colors"
                    >
                      <span className="font-mono text-[11px]">{selector}</span>
                      <button
                        type="button"
                        onClick={() => handleRemoveSelector(index)}
                        className="hover:bg-violet-200/60 rounded-full p-0.5 transition-colors"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </Badge>
                  ))
                )}
              </div>

              <p className="text-[10px] text-muted-foreground/70 mt-2">
                💡 İpucu: Tarayıcıda sağ tık → Öğeyi Denetle ile CSS selector bulabilirsiniz
              </p>
            </div>
          </div>

          {/* Subtle info footer */}
          <div className="mt-4 pt-3 border-t border-border/30">
            <p className="text-[10px] text-muted-foreground/70 flex items-center gap-1.5">
              <span className="inline-block w-1 h-1 rounded-full bg-amber-400"></span>
              Hariç tutma kuralları dahil etme kurallarından önceliklidir
            </p>
          </div>
        </div>
      )}
    </div>
  )
}

export default PathFilterSection

