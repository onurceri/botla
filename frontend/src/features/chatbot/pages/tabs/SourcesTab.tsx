import { useState, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import {
  Database,
  Plus,
  FileText,
  Globe,
  Type as TypeIcon,
  ChevronDown,
  Inbox,
  Search,
  RefreshCw,
  CheckCircle2,
  Clock,
  AlertCircle,
  Sparkles,
  X,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSourceOps } from '../../hooks/useSourceOps'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useUploadSource } from '@/hooks/mutations/useChatbotMutations'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import {
  useUpdateScrapingConfig,
  useUpdateRefresh,
} from '@/hooks/mutations/useChatbotMutations'
import SourceUploader from '@/components/chatbot/SourceUploader'
import URLAdvancedSettings from '../../components/URLAdvancedSettings'
import SourceCard, { Source } from '../../components/SourceCard'
import { Input } from '@/components/ui/input'

type SourceType = 'all' | 'pdf' | 'url' | 'text'
type SourceStatus = 'all' | 'completed' | 'processing' | 'failed'

export default function SourcesTab() {
  const { id = '' } = useParams()
  const isNew = id === 'new'
  const [isUploadExpanded, setIsUploadExpanded] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterType, setFilterType] = useState<SourceType>('all')
  const [filterStatus, setFilterStatus] = useState<SourceStatus>('all')

  const {
    sources,
    refreshSources,
    pollStatus,
    handleDeleteSource,
    handleRefreshSource,
    refreshingId,
  } = useSourceOps(id, isNew)

  const {
    userPlan,
    planConfig,
    discoveryMode,
    setDiscoveryMode,
    refreshPolicy,
    setRefreshPolicy,
    refreshFrequency,
    setRefreshFrequency,
    nextRefreshAt,
    lastRefreshAt,
    includePaths,
    setIncludePaths,
    excludePaths,
    setExcludePaths,
    selectorWhitelist,
    setSelectorWhitelist,
  } = useChatbotContext()

  const { uploadPDF, uploadURL, uploadText } = useUploadSource(id)
  const { mutateAsync: updateScraping } = useUpdateScrapingConfig(id)
  const { mutateAsync: updateRefresh } = useUpdateRefresh(id)

  const {
    isSaving: isScrapingSaving,
    lastSavedAt: scrapingSaved,
    error: scrapingError,
  } = useAutoSave({
    payload: {
      include_paths: includePaths,
      exclude_paths: excludePaths,
      selector_whitelist: selectorWhitelist,
      discovery_mode: discoveryMode,
    },
    saveFn: (_, payload) => updateScraping(payload),
  })

  const {
    isSaving: isRefreshSaving,
    lastSavedAt: refreshSaved,
    error: refreshError,
  } = useAutoSave({
    payload: {
      refresh_policy: refreshPolicy,
      refresh_frequency: refreshFrequency,
    },
    saveFn: (_, payload) => updateRefresh(payload),
  })

  const isSaving = isScrapingSaving || isRefreshSaving
  const lastSavedAt =
    scrapingSaved && refreshSaved
      ? scrapingSaved > refreshSaved
        ? scrapingSaved
        : refreshSaved
      : scrapingSaved || refreshSaved
  const error = scrapingError || refreshError

  const maxFiles = planConfig?.files?.max_files_per_bot || Infinity
  const maxUrls = planConfig?.scraping?.max_urls_per_bot || Infinity

  const currentFiles = sources.filter(
    (s) => s.source_type === 'pdf' || s.source_type === 'text',
  ).length
  const currentUrls = sources.filter((s) => s.source_type === 'url').length

  const isFileLimitReached = currentFiles >= maxFiles
  const isUrlLimitReached = currentUrls >= maxUrls

  const disabledModes: ('pdf' | 'url' | 'text')[] = []
  if (isFileLimitReached) disabledModes.push('pdf', 'text')
  if (isUrlLimitReached) disabledModes.push('url')

  // Filtered sources
  const filteredSources = useMemo(() => {
    return sources.filter((source) => {
      const name = source.original_filename || source.source_url || ''
      if (searchQuery && !name.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false
      }
      if (filterType !== 'all' && source.source_type !== filterType) {
        return false
      }
      if (filterStatus !== 'all') {
        if (filterStatus === 'processing') {
          if (source.status !== 'processing' && source.status !== 'pending' && source.status !== 'queued') {
            return false
          }
        } else if (source.status !== filterStatus) {
          return false
        }
      }
      return true
    })
  }, [sources, searchQuery, filterType, filterStatus])

  // Stats
  const stats = useMemo(() => {
    const completed = sources.filter((s) => s.status === 'completed').length
    const processing = sources.filter(
      (s) => s.status === 'processing' || s.status === 'pending' || s.status === 'queued',
    ).length
    const failed = sources.filter((s) => s.status === 'failed').length
    const totalChunks = sources.reduce((acc, s) => acc + (s.chunk_count || 0), 0)
    return { completed, processing, failed, total: sources.length, totalChunks }
  }, [sources])

  return (
    <div className="space-y-6 animate-in fade-in duration-500 pb-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-2xl bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/25">
            <Database className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-xl font-bold tracking-tight text-slate-900">Bilgi Bankası</h2>
            <p className="text-sm text-muted-foreground">
              Botunuzun öğreneceği kaynakları yönetin
            </p>
          </div>
        </div>
        <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
      </div>

      {/* Stats Cards - Horizontal */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="group relative overflow-hidden p-5 rounded-2xl bg-white border border-slate-200/80 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="absolute top-0 right-0 w-20 h-20 bg-gradient-to-br from-blue-500/10 to-transparent rounded-bl-[60px]" />
          <div className="flex items-center gap-4">
            <div className="p-2.5 rounded-xl bg-blue-500/10 group-hover:bg-blue-500/15 transition-colors">
              <Database className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-3xl font-bold text-slate-900">{stats.total}</p>
              <p className="text-xs text-slate-500 font-medium">Toplam Kaynak</p>
            </div>
          </div>
        </div>

        <div className="group relative overflow-hidden p-5 rounded-2xl bg-white border border-slate-200/80 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="absolute top-0 right-0 w-20 h-20 bg-gradient-to-br from-emerald-500/10 to-transparent rounded-bl-[60px]" />
          <div className="flex items-center gap-4">
            <div className="p-2.5 rounded-xl bg-emerald-500/10 group-hover:bg-emerald-500/15 transition-colors">
              <CheckCircle2 className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-3xl font-bold text-emerald-600">{stats.completed}</p>
              <p className="text-xs text-slate-500 font-medium">Hazır</p>
            </div>
          </div>
        </div>

        <div className="group relative overflow-hidden p-5 rounded-2xl bg-white border border-slate-200/80 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="absolute top-0 right-0 w-20 h-20 bg-gradient-to-br from-amber-500/10 to-transparent rounded-bl-[60px]" />
          <div className="flex items-center gap-4">
            <div className="p-2.5 rounded-xl bg-amber-500/10 group-hover:bg-amber-500/15 transition-colors">
              <Clock className="w-5 h-5 text-amber-600 animate-pulse" />
            </div>
            <div>
              <p className="text-3xl font-bold text-amber-600">{stats.processing}</p>
              <p className="text-xs text-slate-500 font-medium">İşleniyor</p>
            </div>
          </div>
        </div>

        <div className="group relative overflow-hidden p-5 rounded-2xl bg-white border border-slate-200/80 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="absolute top-0 right-0 w-20 h-20 bg-gradient-to-br from-violet-500/10 to-transparent rounded-bl-[60px]" />
          <div className="flex items-center gap-4">
            <div className="p-2.5 rounded-xl bg-violet-500/10 group-hover:bg-violet-500/15 transition-colors">
              <Sparkles className="w-5 h-5 text-violet-600" />
            </div>
            <div>
              <p className="text-3xl font-bold text-violet-600">{stats.totalChunks.toLocaleString()}</p>
              <p className="text-xs text-slate-500 font-medium">Veri Parçası</p>
            </div>
          </div>
        </div>
      </div>

      {/* Upload Section - Full Width */}
      <div className="rounded-2xl border border-slate-200/80 bg-white shadow-sm overflow-hidden">
        {/* Upload Header - Always visible */}
        <button
          onClick={() => setIsUploadExpanded(!isUploadExpanded)}
          className="w-full flex items-center justify-between p-5 hover:bg-slate-50/50 transition-colors"
        >
          <div className="flex items-center gap-4">
            <div className="p-2.5 rounded-xl bg-gradient-to-br from-primary/10 to-primary/5">
              <Plus className="w-5 h-5 text-primary" />
            </div>
            <div className="text-left">
              <h3 className="font-semibold text-slate-900">Yeni Kaynak Ekle</h3>
              <p className="text-sm text-slate-500">PDF, web sitesi veya metin içeriği ekleyin</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            {/* Limit Pills */}
            <div className="hidden sm:flex items-center gap-2">
              {maxFiles !== Infinity && (
                <span className={cn(
                  "flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium",
                  isFileLimitReached 
                    ? "bg-red-100 text-red-600" 
                    : "bg-slate-100 text-slate-600"
                )}>
                  <FileText className="w-3.5 h-3.5" />
                  {currentFiles}/{maxFiles}
                </span>
              )}
              {maxUrls !== Infinity && (
                <span className={cn(
                  "flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium",
                  isUrlLimitReached 
                    ? "bg-red-100 text-red-600" 
                    : "bg-slate-100 text-slate-600"
                )}>
                  <Globe className="w-3.5 h-3.5" />
                  {currentUrls}/{maxUrls}
                </span>
              )}
            </div>
            <div className={cn(
              "p-2 rounded-lg transition-transform duration-300",
              isUploadExpanded && "rotate-180"
            )}>
              <ChevronDown className="w-5 h-5 text-slate-400" />
            </div>
          </div>
        </button>

        {/* Upload Content - Collapsible */}
        {isUploadExpanded && (
          <div className="px-5 pb-6 pt-2 border-t border-slate-100 animate-in fade-in slide-in-from-top-2 duration-300">
            {(isFileLimitReached || isUrlLimitReached) && (
              <div className="mb-5 p-4 rounded-xl bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-200/60 flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-amber-800">Limit Uyarısı</p>
                  <p className="text-xs text-amber-700 mt-0.5">
                    {isFileLimitReached && isUrlLimitReached
                      ? 'Dosya ve URL limitlerinize ulaştınız.'
                      : isFileLimitReached
                        ? 'Dosya yükleme limitinize ulaştınız.'
                        : 'URL ekleme limitinize ulaştınız.'}
                    {' '}Yeni kaynak eklemek için mevcut kaynakları silin veya planınızı yükseltin.
                  </p>
                </div>
              </div>
            )}

            <SourceUploader
              disabledModes={disabledModes}
              onUploadPDF={async (file) => {
                if (id) {
                  await uploadPDF.mutateAsync(file).then((d) => {
                    refreshSources()
                    pollStatus(d.id)
                  })
                }
              }}
              onUploadURL={async (u) => {
                if (id) {
                  await uploadURL.mutateAsync(u).then((d) => {
                    refreshSources()
                    pollStatus(d.id)
                  })
                }
              }}
              onUploadText={async (t) => {
                if (id) {
                  await uploadText.mutateAsync(t).then((d) => {
                    refreshSources()
                    pollStatus(d.id)
                  })
                }
              }}
              maxFileSizeMB={planConfig.files?.max_size_mb}
              maxTextLength={planConfig.files?.max_text_length}
              extraUrlSettings={
                <URLAdvancedSettings
                  discoveryMode={discoveryMode}
                  setDiscoveryMode={setDiscoveryMode}
                  refreshPolicy={refreshPolicy}
                  refreshFrequency={refreshFrequency}
                  nextRefreshAt={nextRefreshAt}
                  lastRefreshAt={lastRefreshAt}
                  onRefreshPolicyChange={setRefreshPolicy}
                  onRefreshFrequencyChange={setRefreshFrequency}
                  includePaths={includePaths}
                  setIncludePaths={setIncludePaths}
                  excludePaths={excludePaths}
                  setExcludePaths={setExcludePaths}
                  selectorWhitelist={selectorWhitelist}
                  setSelectorWhitelist={setSelectorWhitelist}
                  chatbotId={id}
                  onImportComplete={refreshSources}
                  planScrapingConfig={planConfig.scraping}
                  planRefreshConfig={planConfig.refresh}
                />
              }
            />
          </div>
        )}
      </div>

      {/* Source List Section - Full Width */}
      <div className="rounded-2xl border border-slate-200/80 bg-white shadow-sm overflow-hidden">
        {/* List Header with Search & Filters */}
        <div className="p-5 border-b border-slate-100">
          <div className="flex flex-col lg:flex-row lg:items-center gap-4">
            {/* Title */}
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-xl bg-blue-500/10">
                <Database className="w-4 h-4 text-blue-600" />
              </div>
              <h3 className="font-semibold text-slate-900">
                Kaynak Listesi
                {sources.length > 0 && (
                  <span className="ml-2 text-sm font-normal text-slate-500">
                    ({sources.length})
                  </span>
                )}
              </h3>
            </div>

            {/* Search & Filters */}
            <div className="flex-1 flex flex-col sm:flex-row gap-3">
              {/* Search */}
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <Input
                  placeholder="Kaynak ara..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 h-10 bg-slate-50/50 border-slate-200 rounded-xl"
                />
                {searchQuery && (
                  <button 
                    onClick={() => setSearchQuery('')}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
                  >
                    <X className="w-4 h-4" />
                  </button>
                )}
              </div>

              {/* Filter Pills */}
              <div className="flex flex-wrap gap-2">
                {(['all', 'pdf', 'url', 'text'] as const).map((type) => (
                  <button
                    key={type}
                    onClick={() => setFilterType(type)}
                    className={cn(
                      'inline-flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium transition-all duration-200',
                      filterType === type
                        ? 'bg-primary text-primary-foreground shadow-sm'
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200',
                    )}
                  >
                    {type === 'all' && 'Tümü'}
                    {type === 'pdf' && <><FileText className="w-3.5 h-3.5" /> PDF</>}
                    {type === 'url' && <><Globe className="w-3.5 h-3.5" /> URL</>}
                    {type === 'text' && <><TypeIcon className="w-3.5 h-3.5" /> Metin</>}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Status Filters - Second Row */}
          <div className="mt-4 flex flex-wrap items-center gap-2">
            <span className="text-xs font-medium text-slate-500 mr-1">Durum:</span>
            {(['all', 'completed', 'processing', 'failed'] as const).map((status) => (
              <button
                key={status}
                onClick={() => setFilterStatus(status)}
                className={cn(
                  'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200',
                  filterStatus === status
                    ? status === 'completed' 
                      ? 'bg-emerald-100 text-emerald-700'
                      : status === 'processing'
                        ? 'bg-blue-100 text-blue-700'
                        : status === 'failed'
                          ? 'bg-red-100 text-red-700'
                          : 'bg-primary/10 text-primary'
                    : 'bg-slate-50 text-slate-500 hover:bg-slate-100',
                )}
              >
                {status === 'completed' && <CheckCircle2 className="w-3 h-3" />}
                {status === 'processing' && <RefreshCw className={cn("w-3 h-3", filterStatus === status && "animate-spin")} />}
                {status === 'failed' && <AlertCircle className="w-3 h-3" />}
                {status === 'all' ? 'Tümü' : status === 'completed' ? 'Tamamlandı' : status === 'processing' ? 'İşleniyor' : 'Başarısız'}
              </button>
            ))}

            {/* Active filter count */}
            {(searchQuery || filterType !== 'all' || filterStatus !== 'all') && (
              <div className="ml-auto flex items-center gap-2 text-xs text-slate-500">
                <span className="font-medium">{filteredSources.length} / {sources.length}</span>
                <button
                  onClick={() => {
                    setSearchQuery('')
                    setFilterType('all')
                    setFilterStatus('all')
                  }}
                  className="text-primary hover:text-primary/80 font-medium flex items-center gap-1"
                >
                  <X className="w-3 h-3" />
                  Temizle
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Source Grid */}
        <div className="p-5">
          {filteredSources.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {filteredSources.map((source) => (
                <SourceCard
                  key={source.id}
                  source={source as Source}
                  userPlan={userPlan}
                  onDelete={handleDeleteSource}
                  onRefresh={handleRefreshSource}
                  isRefreshing={refreshingId === source.id}
                />
              ))}
            </div>
          ) : sources.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="relative mb-6">
                <div className="absolute inset-0 bg-gradient-to-br from-blue-500/20 to-indigo-500/20 rounded-full blur-2xl scale-150" />
                <div className="relative p-6 rounded-3xl bg-gradient-to-br from-slate-100 to-slate-50 border border-slate-200/60 shadow-sm">
                  <Inbox className="w-12 h-12 text-slate-400" />
                </div>
              </div>
              <h3 className="text-lg font-semibold text-slate-900 mb-2">
                Henüz kaynak eklenmemiş
              </h3>
              <p className="text-sm text-slate-500 max-w-md">
                Yukarıdaki <span className="font-medium text-primary">"Yeni Kaynak Ekle"</span> bölümünden 
                PDF, web sitesi veya metin içeriği ekleyebilirsiniz.
              </p>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="p-4 rounded-2xl bg-slate-100/50 mb-4">
                <Search className="w-8 h-8 text-slate-400" />
              </div>
              <h3 className="text-base font-semibold text-slate-900 mb-1">
                Sonuç bulunamadı
              </h3>
              <p className="text-sm text-slate-500">
                Arama kriterlerinize uygun kaynak yok.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
