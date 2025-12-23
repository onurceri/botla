import { useState, useCallback } from 'react'
import {
  Map,
  Search,
  CheckSquare,
  Square,
  AlertCircle,
  Loader2,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { discoverSitemap, bulkCreateSources, SitemapURL } from '@/api/source'
import { cn } from '@/lib/utils'
import { getTurkishErrorMessage } from '@/lib/errorMessages'

interface SitemapImportProps {
  chatbotId: string
  onImportComplete: () => void
}

const SitemapImport = ({ chatbotId, onImportComplete }: SitemapImportProps) => {
  const [sitemapUrl, setSitemapUrl] = useState('')
  const [isExpanded, setIsExpanded] = useState(false)
  const [loading, setLoading] = useState(false)
  const [importing, setImporting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [discoveredUrls, setDiscoveredUrls] = useState<SitemapURL[]>([])
  const [selectedUrls, setSelectedUrls] = useState<Set<string>>(new Set())

  const handleDiscover = useCallback(async () => {
    if (!sitemapUrl.trim()) return

    setLoading(true)
    setError(null)
    setDiscoveredUrls([])
    setSelectedUrls(new Set())

    try {
      const result = await discoverSitemap(chatbotId, sitemapUrl.trim())
      setDiscoveredUrls(result.urls)
      // Select all by default
      setSelectedUrls(new Set(result.urls.map((u) => u.loc)))
    } catch (err: any) {
      setError(getTurkishErrorMessage(err, 'Sitemap okunamadı'))
    } finally {
      setLoading(false)
    }
  }, [chatbotId, sitemapUrl])

  const handleSelectAll = useCallback(() => {
    setSelectedUrls(new Set(discoveredUrls.map((u) => u.loc)))
  }, [discoveredUrls])

  const handleSelectNone = useCallback(() => {
    setSelectedUrls(new Set())
  }, [])

  const handleSelectRecent = useCallback(() => {
    // Select URLs with lastmod within the last 30 days
    const thirtyDaysAgo = new Date()
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

    const recentUrls = discoveredUrls.filter((u) => {
      if (!u.lastmod) return false
      const date = new Date(u.lastmod)
      return date >= thirtyDaysAgo
    })

    if (recentUrls.length === 0) {
      // If no recent URLs, select all
      setSelectedUrls(new Set(discoveredUrls.map((u) => u.loc)))
    } else {
      setSelectedUrls(new Set(recentUrls.map((u) => u.loc)))
    }
  }, [discoveredUrls])

  const toggleUrl = useCallback((url: string) => {
    setSelectedUrls((prev) => {
      const next = new Set(prev)
      if (next.has(url)) {
        next.delete(url)
      } else {
        next.add(url)
      }
      return next
    })
  }, [])

  const handleImport = useCallback(async () => {
    if (selectedUrls.size === 0) return

    setImporting(true)
    setError(null)

    try {
      const result = await bulkCreateSources(chatbotId, Array.from(selectedUrls))

      if (result.errors?.length > 0) {
        setError(`${result.created_count} URL eklendi, ${result.errors.length} hata oluştu`)
      }

      // Clear state and notify parent
      setDiscoveredUrls([])
      setSelectedUrls(new Set())
      setSitemapUrl('')
      setIsExpanded(false)
      onImportComplete()
    } catch (err: any) {
      setError(getTurkishErrorMessage(err, 'İçe aktarma başarısız'))
    } finally {
      setImporting(false)
    }
  }, [chatbotId, selectedUrls, onImportComplete])

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    try {
      const date = new Date(dateStr)
      return date.toLocaleDateString('tr-TR', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      })
    } catch {
      return dateStr
    }
  }

  return (
    <div className="mt-4 border border-border/60 rounded-xl bg-gradient-to-b from-white/40 to-white/20 backdrop-blur overflow-hidden">
      {/* Collapsible Header */}
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between p-3.5 hover:bg-white/40 transition-all duration-200"
      >
        <div className="flex items-center gap-2.5">
          <div className="p-1.5 rounded-lg bg-amber-500/10">
            <Map className="w-3.5 h-3.5 text-amber-500" />
          </div>
          <span className="text-sm font-medium text-foreground">Sitemap'ten İçe Aktar</span>
          {discoveredUrls.length > 0 && (
            <Badge
              variant="secondary"
              className="text-[10px] px-1.5 py-0 h-4 font-medium bg-amber-100 text-amber-600 border-0"
            >
              {discoveredUrls.length} URL
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground hidden sm:inline">
            {isExpanded ? 'Gizle' : 'XML Sitemap içe aktar'}
          </span>
          {isExpanded ? (
            <ChevronDown className="w-4 h-4 text-muted-foreground transition-transform" />
          ) : (
            <ChevronRight className="w-4 h-4 text-muted-foreground transition-transform" />
          )}
        </div>
      </button>

      {/* Expandable Content */}
      {isExpanded && (
        <div className="px-4 pb-4 pt-1 border-t border-border/40 animate-in slide-in-from-top-1 duration-200">
          {/* Help Text */}
          <p className="text-xs text-muted-foreground mb-4 leading-relaxed">
            Sitemap URL'sini girerek web sitenizin tüm sayfalarını otomatik olarak keşfedin.
          </p>

          {/* Sitemap URL Input */}
          <div className="flex gap-2 mb-4">
            <Input
              placeholder="https://example.com/sitemap.xml"
              value={sitemapUrl}
              onChange={(e) => setSitemapUrl(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleDiscover()}
              className="h-9 text-sm bg-white/80 border-border/60 focus:border-amber-300 focus:ring-amber-100 placeholder:text-muted-foreground/60"
              disabled={loading}
            />
            <Button
              type="button"
              onClick={handleDiscover}
              size="sm"
              disabled={loading || !sitemapUrl.trim()}
              className="h-9 px-4 bg-amber-500 hover:bg-amber-600 text-white"
            >
              {loading ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <>
                  <Search className="w-4 h-4 mr-1.5" />
                  Keşfet
                </>
              )}
            </Button>
          </div>

          {/* Error Message */}
          {error && (
            <div className="flex items-center gap-2 p-3 mb-4 rounded-lg bg-rose-50 border border-rose-200 text-rose-600 text-sm">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Discovered URLs */}
          {discoveredUrls.length > 0 && (
            <div className="space-y-3">
              {/* Actions */}
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-xs text-muted-foreground">
                  {discoveredUrls.length} URL bulundu, {selectedUrls.size} seçildi
                </span>
                <div className="flex-1" />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleSelectAll}
                  className="h-7 text-xs"
                >
                  Tümünü Seç
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleSelectNone}
                  className="h-7 text-xs"
                >
                  Hiçbirini Seçme
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleSelectRecent}
                  className="h-7 text-xs"
                >
                  Son 30 Gün
                </Button>
              </div>

              {/* URL List */}
              <div className="max-h-64 overflow-y-auto rounded-lg border border-border/40 bg-white/50">
                {discoveredUrls.map((url, index) => (
                  <button
                    key={url.loc}
                    type="button"
                    onClick={() => toggleUrl(url.loc)}
                    className={cn(
                      'w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-white/80 transition-colors',
                      index !== discoveredUrls.length - 1 && 'border-b border-border/30',
                      selectedUrls.has(url.loc) && 'bg-amber-50/50',
                    )}
                  >
                    {selectedUrls.has(url.loc) ? (
                      <CheckSquare className="w-4 h-4 text-amber-500 flex-shrink-0" />
                    ) : (
                      <Square className="w-4 h-4 text-muted-foreground/40 flex-shrink-0" />
                    )}
                    <div className="flex-1 min-w-0">
                      <p className="text-xs font-mono text-foreground truncate">
                        {url.loc.replace(/^https?:\/\/[^/]+/, '')}
                      </p>
                    </div>
                    <div className="flex items-center gap-2 flex-shrink-0">
                      {url.priority !== undefined && url.priority > 0 && (
                        <Badge
                          variant="outline"
                          className="text-[10px] px-1.5 py-0 h-4 text-muted-foreground border-border/40"
                        >
                          {url.priority.toFixed(1)}
                        </Badge>
                      )}
                      <span className="text-[10px] text-muted-foreground w-20 text-right">
                        {formatDate(url.lastmod)}
                      </span>
                    </div>
                  </button>
                ))}
              </div>

              {/* Import Button */}
              <div className="flex justify-end pt-2">
                <Button
                  type="button"
                  onClick={handleImport}
                  disabled={importing || selectedUrls.size === 0}
                  className="bg-amber-500 hover:bg-amber-600 text-white"
                >
                  {importing ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-1.5 animate-spin" />
                      İçe Aktarılıyor...
                    </>
                  ) : (
                    <>{selectedUrls.size} URL Ekle</>
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Tip */}
          <p className="text-[10px] text-muted-foreground/70 mt-3 flex items-center gap-1.5">
            <span className="inline-block w-1 h-1 rounded-full bg-amber-400"></span>
            Sitemap genellikle sitenizin kök dizininde /sitemap.xml olarak bulunur
          </p>
        </div>
      )}
    </div>
  )
}

export default SitemapImport
