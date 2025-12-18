import { useState } from 'react'
import { 
  Settings2, 
  ChevronDown, 
  Link2, 
  Zap, 
  Clock, 
  Ban,
  RefreshCw,
  Calendar,
  CalendarDays,
  Filter,
  CheckCircle2,
  XCircle,
  Code2,
  Map,
  Search,
  CheckSquare,
  Square,
  AlertCircle,
  Loader2,
  Plus,
  X
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { discoverSitemap, bulkCreateSources, SitemapURL } from '@/api/source'
import { cn } from '@/lib/utils'
import { getTurkishErrorMessage } from '@/lib/errorMessages'

type DiscoveryMode = 'auto' | 'pending' | 'disabled'
type RefreshPolicy = 'manual' | 'auto'
type RefreshFrequency = 'daily' | 'weekly' | 'monthly'

interface URLAdvancedSettingsProps {
  // Discovery Mode
  discoveryMode: DiscoveryMode
  setDiscoveryMode: (mode: DiscoveryMode) => void
  // Refresh Settings
  refreshPolicy: RefreshPolicy
  refreshFrequency: RefreshFrequency | null
  nextRefreshAt: string | null
  lastRefreshAt: string | null
  onRefreshPolicyChange: (policy: RefreshPolicy) => void
  onRefreshFrequencyChange: (frequency: RefreshFrequency) => void
  // Path Filters
  includePaths: string[]
  setIncludePaths: (paths: string[]) => void
  excludePaths: string[]
  setExcludePaths: (paths: string[]) => void
  selectorWhitelist: string[]
  setSelectorWhitelist: (selectors: string[]) => void
  // Sitemap
  chatbotId: string
  onImportComplete: () => void
  planScrapingConfig?: { max_pages_per_crawl?: number; max_urls_per_bot?: number; dynamic_enabled?: boolean }
  planRefreshConfig?: { enabled: boolean; max_monthly: number }
}

type SectionKey = 'discovery' | 'refresh' | 'filters' | 'sitemap' | null

export default function URLAdvancedSettings({
  discoveryMode,
  setDiscoveryMode,
  refreshPolicy,
  refreshFrequency,
  nextRefreshAt,
  lastRefreshAt,
  onRefreshPolicyChange,
  onRefreshFrequencyChange,
  includePaths,
  setIncludePaths,
  excludePaths,
  setExcludePaths,
  selectorWhitelist,
  setSelectorWhitelist,
  chatbotId,
  onImportComplete,
  planScrapingConfig,
  planRefreshConfig,
}: URLAdvancedSettingsProps) {
  const [expandedSection, setExpandedSection] = useState<SectionKey>(null)
  
  // Filter state
  const [newIncludePath, setNewIncludePath] = useState('')
  const [newExcludePath, setNewExcludePath] = useState('')
  const [newSelector, setNewSelector] = useState('')
  
  // Sitemap state
  const [sitemapUrl, setSitemapUrl] = useState('')
  const [sitemapLoading, setSitemapLoading] = useState(false)
  const [sitemapImporting, setSitemapImporting] = useState(false)
  const [sitemapError, setSitemapError] = useState<string | null>(null)
  const [discoveredUrls, setDiscoveredUrls] = useState<SitemapURL[]>([])
  const [selectedUrls, setSelectedUrls] = useState<Set<string>>(new Set())

  // Check if URL discovery is enabled based on plan
  const isDiscoveryEnabled = (planScrapingConfig?.max_pages_per_crawl ?? 0) > 0

  const toggleSection = (section: SectionKey) => {
    setExpandedSection(expandedSection === section ? null : section)
  }

  // Discovery mode options - with disabled state based on plan
  const discoveryModes = [
    { 
      value: 'auto' as const, 
      label: 'Otomatik', 
      icon: Zap, 
      color: 'text-emerald-500', 
      bg: 'bg-emerald-50', 
      border: 'border-emerald-200',
      requiresDiscovery: true,
      description: 'Keşfedilen URL\'ler otomatik eklenir'
    },
    { 
      value: 'pending' as const, 
      label: 'Onay Bekle', 
      icon: Clock, 
      color: 'text-amber-500', 
      bg: 'bg-amber-50', 
      border: 'border-amber-200',
      requiresDiscovery: true,
      description: 'URL\'ler onayınızı bekler'
    },
    { 
      value: 'disabled' as const, 
      label: 'Kapalı', 
      icon: Ban, 
      color: 'text-gray-400', 
      bg: 'bg-gray-50', 
      border: 'border-gray-200',
      requiresDiscovery: false,
      description: 'Alt sayfa keşfi yapılmaz'
    },
  ]

  // Refresh frequency options
  const frequencies = [
    { value: 'daily' as const, label: 'Günlük', icon: Clock },
    { value: 'weekly' as const, label: 'Haftalık', icon: CalendarDays },
    { value: 'monthly' as const, label: 'Aylık', icon: Calendar },
  ]

  // Helper functions for filters
  const handleAddPath = (type: 'include' | 'exclude') => {
    if (type === 'include') {
      const trimmed = newIncludePath.trim()
      if (trimmed && !includePaths.includes(trimmed)) {
        setIncludePaths([...includePaths, trimmed])
        setNewIncludePath('')
      }
    } else {
      const trimmed = newExcludePath.trim()
      if (trimmed && !excludePaths.includes(trimmed)) {
        setExcludePaths([...excludePaths, trimmed])
        setNewExcludePath('')
      }
    }
  }

  const handleAddSelector = () => {
    const trimmed = newSelector.trim()
    if (trimmed && !selectorWhitelist.includes(trimmed)) {
      setSelectorWhitelist([...selectorWhitelist, trimmed])
      setNewSelector('')
    }
  }

  // Sitemap functions
  const handleDiscoverSitemap = async () => {
    if (!sitemapUrl.trim()) return
    setSitemapLoading(true)
    setSitemapError(null)
    setDiscoveredUrls([])
    setSelectedUrls(new Set())
    try {
      const result = await discoverSitemap(chatbotId, sitemapUrl.trim())
      setDiscoveredUrls(result.urls)
      setSelectedUrls(new Set(result.urls.map(u => u.loc)))
    } catch (err: any) {
      setSitemapError(getTurkishErrorMessage(err, 'Sitemap okunamadı'))
    } finally {
      setSitemapLoading(false)
    }
  }

  const handleImportSitemap = async () => {
    if (selectedUrls.size === 0) return
    setSitemapImporting(true)
    setSitemapError(null)
    try {
      await bulkCreateSources(chatbotId, Array.from(selectedUrls))
      setDiscoveredUrls([])
      setSelectedUrls(new Set())
      setSitemapUrl('')
      onImportComplete()
    } catch (err: any) {
      setSitemapError(getTurkishErrorMessage(err, 'İçe aktarma başarısız'))
    } finally {
      setSitemapImporting(false)
    }
  }

  const formatDate = (dateStr: string | null): string => {
    if (!dateStr) return '-'
    try {
      return new Date(dateStr).toLocaleDateString('tr-TR', { day: 'numeric', month: 'short', year: 'numeric' })
    } catch {
      return '-'
    }
  }

  const totalFilters = includePaths.length + excludePaths.length + selectorWhitelist.length

  // Section Header Component
  const SectionHeader = ({ 
    section, 
    icon: Icon, 
    title, 
    badge,
    color = 'text-blue-500',
    bgColor = 'bg-blue-50'
  }: { 
    section: SectionKey
    icon: any
    title: string
    badge?: string | number
    color?: string
    bgColor?: string
  }) => (
    <button
      type="button"
      onClick={() => toggleSection(section)}
      className={cn(
        "w-full flex items-center justify-between p-3 rounded-lg transition-all",
        expandedSection === section 
          ? "bg-white shadow-sm" 
          : "hover:bg-white/60"
      )}
    >
      <div className="flex items-center gap-2.5">
        <div className={cn("p-1.5 rounded-lg", bgColor)}>
          <Icon className={cn("w-3.5 h-3.5", color)} />
        </div>
        <span className="text-sm font-medium text-gray-700">{title}</span>
        {badge !== undefined && badge !== 0 && (
          <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4 font-medium">
            {badge}
          </Badge>
        )}
      </div>
      <ChevronDown className={cn(
        "w-4 h-4 text-gray-400 transition-transform",
        expandedSection === section && "rotate-180"
      )} />
    </button>
  )

  return (
    <div className="mt-4 rounded-xl border border-gray-200/80 bg-gradient-to-b from-gray-50/80 to-white/60 backdrop-blur overflow-hidden">
      {/* Main Header */}
      <div className="px-4 py-3 border-b border-gray-100 bg-white/40">
        <div className="flex items-center gap-2">
          <Settings2 className="w-4 h-4 text-gray-500" />
          <span className="text-sm font-semibold text-gray-700">URL Tarama Ayarları</span>
        </div>
        <p className="text-xs text-gray-500 mt-1">
          Web sitesi kaynaklarının nasıl taranacağını ve güncelleneceğini yapılandırın
        </p>
      </div>

      {/* Accordion Sections */}
      <div className="p-2 space-y-1">
        
        {/* 1. Discovery Mode Section */}
        <div>
          <SectionHeader 
            section="discovery" 
            icon={Link2} 
            title="Sayfa Keşif Modu"
            badge={!isDiscoveryEnabled ? 'Pro' : (discoveryMode !== 'auto' ? discoveryMode === 'pending' ? 'Onay' : 'Kapalı' : undefined)}
            color="text-indigo-500"
            bgColor="bg-indigo-50"
          />
          {expandedSection === 'discovery' && (
            <div className="px-3 pb-3 pt-2 animate-in fade-in slide-in-from-top-1 duration-200">
              <p className="text-xs text-gray-500 mb-3">
                Bir URL eklediğinizde, sayfadaki bağlantıların nasıl işleneceğini belirleyin.
              </p>
              
              {/* Show upgrade notice when discovery is disabled by plan */}
              {!isDiscoveryEnabled && (
                <div className="mb-3 p-2.5 rounded-lg bg-amber-50 border border-amber-200">
                  <div className="flex items-center gap-2 text-amber-700">
                    <AlertCircle className="w-3.5 h-3.5 flex-shrink-0" />
                    <span className="text-xs font-medium">Alt sayfa keşfi Pro planda aktif</span>
                  </div>
                  <p className="text-[10px] text-amber-600 mt-1 ml-5">
                    Pro plana yükselterek eklediğiniz URL'lerdeki tüm alt sayfaları otomatik keşfedebilirsiniz.
                  </p>
                </div>
              )}
              
              <div className="grid grid-cols-3 gap-2">
                {discoveryModes.map((mode) => {
                  const Icon = mode.icon
                  const isSelected = discoveryMode === mode.value
                  const isDisabled = mode.requiresDiscovery && !isDiscoveryEnabled
                  return (
                    <button
                      key={mode.value}
                      type="button"
                      onClick={() => !isDisabled && setDiscoveryMode(mode.value)}
                      disabled={isDisabled}
                      className={cn(
                        "flex flex-col items-center p-3 rounded-lg border-2 transition-all relative",
                        isDisabled 
                          ? "border-gray-100 bg-gray-50 cursor-not-allowed opacity-60"
                          : isSelected 
                            ? `${mode.border} ${mode.bg}` 
                            : "border-gray-100 bg-white hover:border-gray-200"
                      )}
                    >
                      {isDisabled && (
                        <Badge 
                          variant="outline" 
                          className="absolute -top-2 -right-2 text-[8px] px-1.5 py-0 h-4 bg-violet-100 text-violet-700 border-violet-200"
                        >
                          Pro
                        </Badge>
                      )}
                      <Icon className={cn("w-4 h-4 mb-1.5", isDisabled ? "text-gray-300" : isSelected ? mode.color : "text-gray-400")} />
                      <span className={cn("text-xs font-medium", isDisabled ? "text-gray-400" : isSelected ? "text-gray-700" : "text-gray-500")}>
                        {mode.label}
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>
          )}
        </div>


        {/* 2. Auto Refresh Section */}
        <div>
          <SectionHeader 
            section="refresh" 
            icon={RefreshCw} 
            title="Otomatik Yenileme"
            badge={planRefreshConfig?.enabled === false ? 'Pro' : (refreshPolicy === 'auto' ? (refreshFrequency || undefined) : undefined)}
            color="text-blue-500"
            bgColor="bg-blue-50"
          />
          {expandedSection === 'refresh' && (
            <div className="px-3 pb-3 pt-2 animate-in fade-in slide-in-from-top-1 duration-200">
              <p className="text-xs text-gray-500 mb-3">
                URL kaynaklarınızın otomatik olarak güncellenme ayarları
              </p>

              {planRefreshConfig?.enabled === false && (
                <div className="mb-3 p-2.5 rounded-lg bg-blue-50 border border-blue-200">
                  <div className="flex items-center gap-2 text-blue-700">
                    <AlertCircle className="w-3.5 h-3.5 flex-shrink-0" />
                    <span className="text-xs font-medium">Otomatik yenileme Pro planda aktif</span>
                  </div>
                  <p className="text-[10px] text-blue-600 mt-1 ml-5">
                    Pro plana yükselterek kaynaklarınızın her zaman güncel kalmasını sağlayabilirsiniz.
                  </p>
                </div>
              )}
              
              {/* Policy Toggle */}
              <div className="grid grid-cols-2 gap-2 mb-3">
                <button
                  type="button"
                  onClick={() => onRefreshPolicyChange('manual')}
                  className={cn(
                    "p-3 rounded-lg border-2 transition-all text-center",
                    refreshPolicy === 'manual'
                      ? "border-blue-200 bg-blue-50"
                      : "border-gray-100 bg-white hover:border-gray-200"
                  )}
                >
                  <span className={cn("text-xs font-medium", refreshPolicy === 'manual' ? "text-blue-700" : "text-gray-500")}>
                    Manuel
                  </span>
                </button>
                <button
                  type="button"
                  disabled={planRefreshConfig?.enabled === false}
                  onClick={() => {
                    if (planRefreshConfig?.enabled === false) return
                    onRefreshPolicyChange('auto')
                    if (!refreshFrequency) onRefreshFrequencyChange('weekly')
                  }}
                  className={cn(
                    "p-3 rounded-lg border-2 transition-all text-center relative",
                    refreshPolicy === 'auto'
                      ? "border-blue-200 bg-blue-50"
                      : "border-gray-100 bg-white hover:border-gray-200",
                    planRefreshConfig?.enabled === false && "opacity-60 cursor-not-allowed bg-gray-50 border-gray-100"
                  )}
                >
                  {planRefreshConfig?.enabled === false && (
                    <Badge variant="outline" className="absolute -top-2 -right-2 text-[8px] px-1.5 py-0 h-4 bg-violet-100 text-violet-700 border-violet-200">
                      Pro
                    </Badge>
                  )}
                  <span className={cn("text-xs font-medium", refreshPolicy === 'auto' ? "text-blue-700" : "text-gray-500")}>
                    Otomatik
                  </span>
                </button>
              </div>

              {/* Frequency Selector (only when auto) */}
              {refreshPolicy === 'auto' && (
                <>
                  <div className="grid grid-cols-3 gap-2 mb-3">
                    {frequencies.map((freq) => {
                      const Icon = freq.icon
                      const isSelected = refreshFrequency === freq.value
                      return (
                        <button
                          key={freq.value}
                          type="button"
                          onClick={() => onRefreshFrequencyChange(freq.value)}
                          className={cn(
                            "flex flex-col items-center p-2.5 rounded-lg border transition-all",
                            isSelected 
                              ? "border-blue-300 bg-blue-50 text-blue-700" 
                              : "border-gray-100 bg-white text-gray-500 hover:border-gray-200"
                          )}
                        >
                          <Icon className="w-3.5 h-3.5 mb-1" />
                          <span className="text-[11px] font-medium">{freq.label}</span>
                        </button>
                      )
                    })}
                  </div>
                  
                  {/* Status */}
                  <div className="bg-gray-50 rounded-lg p-2.5 space-y-1.5 text-xs">
                    <div className="flex justify-between">
                      <span className="text-gray-500">Son yenileme</span>
                      <span className="text-gray-700">{formatDate(lastRefreshAt)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Sonraki</span>
                      <span className="text-gray-700">{formatDate(nextRefreshAt)}</span>
                    </div>
                  </div>
                </>
              )}
            </div>
          )}
        </div>

        {/* 3. URL Filters & CSS Selectors */}
        <div>
          <SectionHeader 
            section="filters" 
            icon={Filter} 
            title="URL Filtreleri"
            badge={totalFilters > 0 ? `${totalFilters} filtre` : undefined}
            color="text-violet-500"
            bgColor="bg-violet-50"
          />
          {expandedSection === 'filters' && (
            <div className="px-3 pb-3 pt-2 animate-in fade-in slide-in-from-top-1 duration-200 space-y-4">
              <p className="text-xs text-gray-500">
                URL filtreleri ve CSS seçicileriyle tarama davranışını özelleştirin.
              </p>

              {/* Include Paths */}
              <div className="space-y-2">
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                  <span className="text-xs font-medium text-gray-700">Dahil Et</span>
                </div>
                <div className="flex gap-2">
                  <Input
                    placeholder="/blog/*, /docs/*"
                    value={newIncludePath}
                    onChange={(e) => setNewIncludePath(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleAddPath('include')}
                    className="h-8 text-xs"
                  />
                  <Button type="button" size="sm" variant="outline" onClick={() => handleAddPath('include')} className="h-8 w-8 p-0">
                    <Plus className="w-3.5 h-3.5" />
                  </Button>
                </div>
                <div className="flex flex-wrap gap-1 min-h-[24px]">
                  {includePaths.map((path, i) => (
                    <Badge key={i} variant="outline" className="text-[10px] bg-emerald-50 text-emerald-700 border-emerald-200 gap-1 pr-1">
                      <span className="font-mono">{path}</span>
                      <button type="button" onClick={() => setIncludePaths(includePaths.filter((_, idx) => idx !== i))} className="hover:text-emerald-900">
                        <X className="w-2.5 h-2.5" />
                      </button>
                    </Badge>
                  ))}
                  {includePaths.length === 0 && <span className="text-[10px] text-gray-400 italic">Tüm sayfalar dahil</span>}
                </div>
              </div>

              {/* Exclude Paths */}
              <div className="space-y-2">
                <div className="flex items-center gap-1.5">
                  <XCircle className="w-3.5 h-3.5 text-rose-400" />
                  <span className="text-xs font-medium text-gray-700">Hariç Tut</span>
                </div>
                <div className="flex gap-2">
                  <Input
                    placeholder="/admin/*, /login/*"
                    value={newExcludePath}
                    onChange={(e) => setNewExcludePath(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleAddPath('exclude')}
                    className="h-8 text-xs"
                  />
                  <Button type="button" size="sm" variant="outline" onClick={() => handleAddPath('exclude')} className="h-8 w-8 p-0">
                    <Plus className="w-3.5 h-3.5" />
                  </Button>
                </div>
                <div className="flex flex-wrap gap-1 min-h-[24px]">
                  {excludePaths.map((path, i) => (
                    <Badge key={i} variant="outline" className="text-[10px] bg-rose-50 text-rose-600 border-rose-200 gap-1 pr-1">
                      <span className="font-mono">{path}</span>
                      <button type="button" onClick={() => setExcludePaths(excludePaths.filter((_, idx) => idx !== i))} className="hover:text-rose-900">
                        <X className="w-2.5 h-2.5" />
                      </button>
                    </Badge>
                  ))}
                  {excludePaths.length === 0 && <span className="text-[10px] text-gray-400 italic">Hiçbir sayfa hariç tutulmayacak</span>}
                </div>
              </div>

              {/* CSS Selectors */}
              <div className="pt-3 border-t border-gray-100 space-y-2">
                <div className="flex items-center gap-1.5">
                  <Code2 className="w-3.5 h-3.5 text-violet-500" />
                  <span className="text-xs font-medium text-gray-700">CSS Seçiciler</span>
                </div>
                <p className="text-[10px] text-gray-500">Sadece belirtilen elementlerden içerik çıkarılır</p>
                <div className="flex gap-2">
                  <Input
                    placeholder=".content, #article, main"
                    value={newSelector}
                    onChange={(e) => setNewSelector(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleAddSelector()}
                    className="h-8 text-xs font-mono"
                  />
                  <Button type="button" size="sm" variant="outline" onClick={handleAddSelector} className="h-8 w-8 p-0">
                    <Plus className="w-3.5 h-3.5" />
                  </Button>
                </div>
                <div className="flex flex-wrap gap-1 min-h-[24px]">
                  {selectorWhitelist.map((sel, i) => (
                    <Badge key={i} variant="outline" className="text-[10px] bg-violet-50 text-violet-700 border-violet-200 gap-1 pr-1">
                      <span className="font-mono">{sel}</span>
                      <button type="button" onClick={() => setSelectorWhitelist(selectorWhitelist.filter((_, idx) => idx !== i))} className="hover:text-violet-900">
                        <X className="w-2.5 h-2.5" />
                      </button>
                    </Badge>
                  ))}
                  {selectorWhitelist.length === 0 && <span className="text-[10px] text-gray-400 italic">Tüm sayfa içeriği kullanılacak</span>}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 4. Sitemap Import */}
        <div>
          <SectionHeader 
            section="sitemap" 
            icon={Map} 
            title="Sitemap İçe Aktar"
            badge={discoveredUrls.length > 0 ? `${discoveredUrls.length} URL` : undefined}
            color="text-amber-500"
            bgColor="bg-amber-50"
          />
          {expandedSection === 'sitemap' && (
            <div className="px-3 pb-3 pt-2 animate-in fade-in slide-in-from-top-1 duration-200">
              <p className="text-xs text-gray-500 mb-3">
                Sitemap URL'sini girerek tüm sayfaları otomatik olarak keşfedin.
              </p>
              
              <div className="flex gap-2 mb-3">
                <Input
                  placeholder="https://site.com/sitemap.xml"
                  value={sitemapUrl}
                  onChange={(e) => setSitemapUrl(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleDiscoverSitemap()}
                  className="h-8 text-xs"
                  disabled={sitemapLoading}
                />
                <Button
                  type="button"
                  size="sm"
                  onClick={handleDiscoverSitemap}
                  disabled={sitemapLoading || !sitemapUrl.trim()}
                  className="h-8 px-3 bg-amber-500 hover:bg-amber-600 text-white"
                >
                  {sitemapLoading ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Search className="w-3.5 h-3.5" />}
                </Button>
              </div>

              {sitemapError && (
                <div className="flex items-center gap-2 p-2 mb-3 rounded-lg bg-rose-50 text-rose-600 text-xs">
                  <AlertCircle className="w-3.5 h-3.5" />
                  <span>{sitemapError}</span>
                </div>
              )}

              {discoveredUrls.length > 0 && (
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-gray-500">{selectedUrls.size}/{discoveredUrls.length} seçildi</span>
                    <div className="flex gap-1">
                      <Button type="button" variant="ghost" size="sm" onClick={() => setSelectedUrls(new Set(discoveredUrls.map(u => u.loc)))} className="h-6 text-[10px] px-2">Tümü</Button>
                      <Button type="button" variant="ghost" size="sm" onClick={() => setSelectedUrls(new Set())} className="h-6 text-[10px] px-2">Hiçbiri</Button>
                    </div>
                  </div>
                  
                  <div className="max-h-40 overflow-y-auto rounded-lg border border-gray-100 bg-white">
                    {discoveredUrls.map((url, i) => (
                      <button
                        key={url.loc}
                        type="button"
                        onClick={() => {
                          const next = new Set(selectedUrls)
                          next.has(url.loc) ? next.delete(url.loc) : next.add(url.loc)
                          setSelectedUrls(next)
                        }}
                        className={cn(
                          "w-full flex items-center gap-2 px-2 py-1.5 text-left hover:bg-gray-50 transition-colors",
                          i !== discoveredUrls.length - 1 && "border-b border-gray-50",
                          selectedUrls.has(url.loc) && "bg-amber-50/50"
                        )}
                      >
                        {selectedUrls.has(url.loc) ? (
                          <CheckSquare className="w-3.5 h-3.5 text-amber-500 flex-shrink-0" />
                        ) : (
                          <Square className="w-3.5 h-3.5 text-gray-300 flex-shrink-0" />
                        )}
                        <span className="text-[10px] font-mono text-gray-600 truncate flex-1">
                          {url.loc.replace(/^https?:\/\/[^/]+/, '')}
                        </span>
                      </button>
                    ))}
                  </div>

                  <Button
                    type="button"
                    onClick={handleImportSitemap}
                    disabled={sitemapImporting || selectedUrls.size === 0}
                    className="w-full h-8 bg-amber-500 hover:bg-amber-600 text-white text-xs"
                  >
                    {sitemapImporting ? (
                      <><Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin" />İçe Aktarılıyor...</>
                    ) : (
                      <>{selectedUrls.size} URL Ekle</>
                    )}
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
