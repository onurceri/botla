import { useParams } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import SourceUploader from '@/components/chatbot/SourceUploader'
import SourceList from '../../components/SourceList'
import URLAdvancedSettings from '../../components/URLAdvancedSettings'
import PendingURLsPanel from '../../components/PendingURLsPanel'
import { useSourceOps } from '../../hooks/useSourceOps'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useUploadSource } from '@/hooks/mutations/useChatbotMutations'
import { Inbox, Database, Plus } from 'lucide-react'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useUpdateScrapingConfig, useUpdateRefresh } from '@/hooks/mutations/useChatbotMutations'

export default function SourcesTab() {
  const { id = '' } = useParams()
  const isNew = id === 'new'
  const { 
    sources, refreshSources, pollStatus, handleDeleteSource, handleRefreshSource, refreshingId 
  } = useSourceOps(id, isNew)

  const {
    userPlan,
    planConfig,
    discoveryMode, setDiscoveryMode,
    refreshPolicy, setRefreshPolicy,
    refreshFrequency, setRefreshFrequency,
    nextRefreshAt,
    lastRefreshAt,
    includePaths, setIncludePaths,
    excludePaths, setExcludePaths,
    selectorWhitelist, setSelectorWhitelist,
    buildSourceSettingsPayload,
  } = useChatbotContext()

  const { uploadPDF, uploadURL, uploadText } = useUploadSource(id)
  const { mutateAsync: updateScraping } = useUpdateScrapingConfig(id)
  const { mutateAsync: updateRefresh } = useUpdateRefresh(id)

  const { isSaving: isScrapingSaving, lastSavedAt: scrapingSaved, error: scrapingError } = useAutoSave({
    payload: { 
      include_paths: includePaths,
      exclude_paths: excludePaths,
      selector_whitelist: selectorWhitelist,
      discovery_mode: discoveryMode 
    },
    saveFn: (id, payload) => updateScraping(payload)
  })

  const { isSaving: isRefreshSaving, lastSavedAt: refreshSaved, error: refreshError } = useAutoSave({
    payload: {
      refresh_policy: refreshPolicy,
      refresh_frequency: refreshFrequency
    },
    saveFn: (id, payload) => updateRefresh(payload)
  })

  const isSaving = isScrapingSaving || isRefreshSaving
  const lastSavedAt = scrapingSaved && refreshSaved
    ? (scrapingSaved > refreshSaved ? scrapingSaved : refreshSaved)
    : (scrapingSaved || refreshSaved)
  const error = scrapingError || refreshError

  const maxFiles = planConfig?.files?.max_files_per_bot || Infinity
  const maxUrls = planConfig?.scraping?.max_urls_per_bot || Infinity
  
  const currentFiles = sources.filter(s => s.source_type === 'pdf' || s.source_type === 'text').length
  const currentUrls = sources.filter(s => s.source_type === 'url').length
  
  const isFileLimitReached = currentFiles >= maxFiles
  const isUrlLimitReached = currentUrls >= maxUrls
  
  const disabledModes: ('pdf' | 'url' | 'text')[] = []
  if (isFileLimitReached) disabledModes.push('pdf', 'text')
  if (isUrlLimitReached) disabledModes.push('url')

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Bilgi Bankası</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Botunuzun soruları cevaplarken kullanacağı kaynakları yönetin.
        </p>
      </div>

      <div className="grid gap-6">
        {/* Upload Section */}
        <Card className="border-muted-foreground/20 shadow-sm">
           <CardHeader>
             <div className="flex justify-between items-start">
               <div>
                 <CardTitle className="flex items-center gap-2">
                    <div className="p-2 rounded-lg bg-primary/10 text-primary">
                        <Plus className="w-5 h-5" />
                    </div>
                    Yeni Kaynak Ekle
                 </CardTitle>
                 <CardDescription>
                    Web sitesi, PDF dokümanı veya metin içeriği ekleyerek botunuzu eğitin.
                 </CardDescription>
               </div>
               <div className="text-right flex flex-col gap-1">
                 {maxFiles !== Infinity && (
                   <span className={`text-xs font-medium ${isFileLimitReached ? 'text-destructive' : 'text-muted-foreground'}`}>
                     Dosya: {currentFiles} / {maxFiles}
                   </span>
                 )}
                 {maxUrls !== Infinity && (
                   <span className={`text-xs font-medium ${isUrlLimitReached ? 'text-destructive' : 'text-muted-foreground'}`}>
                     URL: {currentUrls} / {maxUrls}
                   </span>
                 )}
               </div>
             </div>
           </CardHeader>
           <CardContent>
             {(isFileLimitReached || isUrlLimitReached) && (
               <div className="mb-4 p-3 rounded-lg bg-destructive/10 text-destructive text-sm font-medium">
                 {isFileLimitReached && isUrlLimitReached 
                   ? 'Dosya ve URL limitlerinize ulaştınız. Yeni kaynak eklemek için mevcut kaynakları silin veya planınızı yükseltin.'
                   : isFileLimitReached
                     ? 'Dosya yükleme limitinize ulaştınız. Yeni dosya yüklemek için mevcut dosyalardan silin veya planınızı yükseltin.'
                     : 'URL ekleme limitinize ulaştınız. Yeni URL eklemek için mevcut URL\'lerden silin veya planınızı yükseltin.'
                 }
               </div>
             )}
             <SourceUploader
               disabledModes={disabledModes}
               onUploadPDF={async (file) => { 
                 if(id) { 
                   await uploadPDF.mutateAsync(file).then((d) => { refreshSources(); pollStatus(d.id) }) 
                 } 
               }}
               onUploadURL={async (u) => { 
                 if(id) { 
                   await uploadURL.mutateAsync(u).then((d) => { refreshSources(); pollStatus(d.id) }) 
                 } 
               }}
               onUploadText={async (t) => { 
                 if(id) { 
                   await uploadText.mutateAsync(t).then((d) => { refreshSources(); pollStatus(d.id) }) 
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
           </CardContent>
        </Card>

        <PendingURLsPanel
          chatbotId={id}
          onSourcesCreated={refreshSources}
        />

        {/* Source List Section */}
        <Card className="border-muted-foreground/20 shadow-sm">
           <CardHeader>
             <CardTitle className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-blue-500/10 text-blue-500">
                    <Database className="w-5 h-5" />
                </div>
                Ekli Kaynaklar
             </CardTitle>
             <CardDescription>
                Botunuzun şu anda kullandığı tüm veri kaynakları.
             </CardDescription>
           </CardHeader>
           <CardContent>
             {sources.length > 0 ? (
               <SourceList 
                 sources={sources as any} 
                 userPlan={userPlan}
                 onDelete={handleDeleteSource} 
                 onRefresh={handleRefreshSource}
                 refreshingId={refreshingId}
               />
             ) : (
               <div className="rounded-xl border border-dashed border-muted-foreground/25 bg-muted/30 p-10 text-center space-y-3">
                 <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-muted shadow-sm">
                   <Inbox className="w-6 h-6 text-muted-foreground" />
                 </div>
                 <div className="space-y-1">
                    <div className="text-sm font-medium text-foreground">Henüz kaynak eklenmemiş</div>
                    <div className="text-xs text-muted-foreground">Yukarıdaki alandan ilk kaynağınızı ekleyin.</div>
                 </div>
               </div>
             )}
           </CardContent>
        </Card>
      </div>
    </div>
  )
}
