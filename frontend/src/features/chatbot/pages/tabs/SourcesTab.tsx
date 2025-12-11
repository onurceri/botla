import { useParams } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import SourceUploader from '@/components/chatbot/SourceUploader'
import SourceList from '../../components/SourceList'
import URLAdvancedSettings from '../../components/URLAdvancedSettings'
import PendingURLsPanel from '../../components/PendingURLsPanel'
import { useSourceOps } from '../../hooks/useSourceOps'
import { useChatbotContext } from '../../context/ChatbotContext'
import { uploadPDFSource, uploadTextSource, uploadURLSource } from '@/api/source'
import { Inbox, Database, Plus } from 'lucide-react'

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
  } = useChatbotContext()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
       <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Bilgi Bankası</h2>
        <p className="text-muted-foreground">
          Botunuzun soruları cevaplarken kullanacağı kaynakları yönetin.
        </p>
      </div>

      <div className="grid gap-6">
        {/* Upload Section */}
        <Card className="border-muted-foreground/20 shadow-sm">
           <CardHeader>
             <CardTitle className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-primary/10 text-primary">
                    <Plus className="w-5 h-5" />
                </div>
                Yeni Kaynak Ekle
             </CardTitle>
             <CardDescription>
                Web sitesi, PDF dokümanı veya metin içeriği ekleyerek botunuzu eğitin.
             </CardDescription>
           </CardHeader>
           <CardContent>
             <SourceUploader
               onUploadPDF={async (file) => { if(id) { await uploadPDFSource(id, file).then((d) => { refreshSources(); pollStatus(d.id) }) } }}
               onUploadURL={async (u) => { if(id) { await uploadURLSource(id, u).then((d) => { refreshSources(); pollStatus(d.id) }) } }}
               onUploadText={async (t) => { if(id) { await uploadTextSource(id, t).then((d) => { refreshSources(); pollStatus(d.id) }) } }}
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
