import { useEffect, useState, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Inbox } from 'lucide-react'
import { api } from '@/api/client'
import { uploadPDFSource, uploadTextSource, uploadURLSource } from '@/api/source'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import SourceUploader from '@/components/chatbot/SourceUploader'
import { useToast } from '@/components/ui/toast'
import HeaderActions from '@/features/chatbot/components/HeaderActions'
import { ChatbotSidebar } from '@/features/chatbot/components/ChatbotSidebar'
import EmbeddingCodePanel from '@/features/chatbot/components/EmbeddingCodePanel'
import NewChatbotForm from '@/features/chatbot/components/NewChatbotForm'
import OverviewPanel from '@/features/chatbot/components/OverviewPanel'
import IdentitySection from '@/features/chatbot/components/IdentitySection'
import AppearanceSection from '@/features/chatbot/components/AppearanceSection'
import ColorsSection from '@/features/chatbot/components/ColorsSection'
import SourceList from '@/features/chatbot/components/SourceList'
import { useSourceOps } from '@/features/chatbot/hooks/useSourceOps'
import { usePreview } from '@/features/chatbot/hooks/usePreview'
import { useChatbotForm } from '@/features/chatbot/hooks/useChatbotForm'
import { useToastErrors } from '@/features/chatbot/hooks/useToastErrors'
import PlaygroundPreview from '@/features/chatbot/components/PlaygroundPreview'
import SuggestionsPanel from '@/features/chatbot/components/SuggestionsPanel'
import PendingURLsPanel from '@/features/chatbot/components/PendingURLsPanel'
import URLAdvancedSettings from '@/features/chatbot/components/URLAdvancedSettings'
import BrandingSettings from '@/features/chatbot/components/BrandingSettings'
import GuardrailsSettings from '@/features/chatbot/components/GuardrailsSettings'
import HandoffSettings from '@/features/chatbot/components/HandoffSettings'
import ActionList from '@/features/chatbot/components/ActionList'
import { ChatbotAnalytics } from '@/features/analytics/ChatbotAnalytics'
import { useOrganization } from '@/features/organization/context/OrganizationContext'

// ... existing imports ...

const ChatbotDetailPage = () => {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { currentWorkspace } = useOrganization()
  const prevWorkspaceIdRef = useRef<string | null>(null)
  const [activeTab, setActiveTab] = useState('overview')
  
  const {
    name, setName,
    description, setDescription,
    systemPrompt, setSystemPrompt,
    model, setModel,
    temperature, setTemperature,
    maxTokens, setMaxTokens,
    themeColor, setThemeColor,
    welcomeMessage, setWelcomeMessage,
    position, setPosition,
    botMessageColor, setBotMessageColor,
    userMessageColor, setUserMessageColor,
    botMessageTextColor, setBotMessageTextColor,
    userMessageTextColor, setUserMessageTextColor,
    chatFontFamily, setChatFontFamily,
    chatHeaderColor, setChatHeaderColor,
    chatHeaderTextColor, setChatHeaderTextColor,
    chatBackgroundColor, setChatBackgroundColor,
    botIcon, setBotIcon,
    botDisplayName, setBotDisplayName,
    secureEmbedEnabled, setSecureEmbedEnabled,
    allowedDomains, setAllowedDomains,
    embedSecret, setEmbedSecret,
    suggestionsEnabled, setSuggestionsEnabled,
    suggestedQuestions, setSuggestedQuestions,
    includePaths, setIncludePaths,
    excludePaths, setExcludePaths,
    selectorWhitelist, setSelectorWhitelist,
    discoveryMode, setDiscoveryMode,
    refreshPolicy, setRefreshPolicy,
    refreshFrequency, setRefreshFrequency,
    nextRefreshAt,
    lastRefreshAt,
    hideBranding, setHideBranding,
    customBranding, setCustomBranding,
    setFromServer,
    validate,
    buildPayload,
    confidenceThreshold, setConfidenceThreshold,
    fallbackMessages, setFallbackMessages,
    topicRestrictions, setTopicRestrictions,
    handoffEnabled, setHandoffEnabled,
    handoffType, setHandoffType,
    handoffConfig, setHandoffConfig,
  } = useChatbotForm()
  const [userPlan, setUserPlan] = useState('free')
  const [planConfig, setPlanConfig] = useState<{ branding?: { can_hide_branding?: boolean, can_custom_branding?: boolean } }>({});
  
  const isNew = id === 'new'
  const { sources, refreshSources, pollStatus, handleDeleteSource, handleRefreshSource, refreshingId } = useSourceOps(id, isNew)
  
  // Chat Test State
  const [chatInput, setChatInput] = useState('')
  const [chatHistory, setChatHistory] = useState<{role: 'user' | 'assistant', content: string}[]>([])
  const [chatLoading, setChatLoading] = useState(false)
  
  // Playground State
  const { previewOpen, sessionId } = usePreview()
  const [expandedSection, setExpandedSection] = useState<string | null>('identity')

  

  useEffect(() => {
    try {
      const u = new URL(window.location.href)
      const tab = u.searchParams.get('tab')
      if (tab === 'connect' || tab === 'sources' || tab === 'playground' || tab === 'overview') {
        setActiveTab(tab)
      }
    } catch {}
  }, [])

  useEffect(() => {
    if (currentWorkspace?.id) {
      if (prevWorkspaceIdRef.current && prevWorkspaceIdRef.current !== currentWorkspace.id) {
        navigate('/chatbots')
      }
      prevWorkspaceIdRef.current = currentWorkspace.id
    }
  }, [currentWorkspace, navigate])

  // Fetch user profile and plan config
  const fetchUserProfile = () => {
    api.get('/api/v1/me').then(({ data }) => { 
      const plan = data.plan_code || data.subscription_plan || 'free'
      setUserPlan(plan)
      if (data.config) {
        setPlanConfig(data.config)
      }
    }).catch(() => {})
  }

  useEffect(() => {
    fetchUserProfile()
    
    // Refetch when user returns to tab (in case plan was upgraded)
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        fetchUserProfile()
      }
    }
    document.addEventListener('visibilitychange', handleVisibilityChange)
    
    if (!isNew && id) {
      api.get(`/api/v1/chatbots/${id}`).then(({ data }) => {
        setFromServer(data)
      }).catch(() => {})
      
      refreshSources()
    }
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [id])


  const { toast } = useToast()
  const toasts = useToastErrors()
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [previewRefreshKey, setPreviewRefreshKey] = useState(0)

  const handleSave = async () => {
    if (!validate()) {
      toasts.error('Lütfen bir bot ismi girin.')
      return
    }

    setIsSaving(true)
    const payload = buildPayload()

    try {
      if (isNew) {
        const { data } = await api.post('/api/v1/chatbots', payload)
        toast('Chatbot başarıyla oluşturuldu.', 'success')
        navigate(`/chatbots/${data.id}`)
      } else {
        await api.put(`/api/v1/chatbots/${id}`, payload)
        toast('Değişiklikler kaydedildi.', 'success')
        setPreviewRefreshKey((k) => k + 1)
      }
    } catch (error: any) {
      console.error(error)
      const msg = error.response?.data?.error || 'Bir hata oluştu. Lütfen tekrar deneyin.'
      toasts.error(msg)
    } finally {
      setIsSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Bu chatbotu silmek istediğinize emin misiniz?')) return
    
    setIsDeleting(true)
    try {
      await api.delete(`/api/v1/chatbots/${id}`)
      toast('Chatbot silindi.', 'success')
      navigate('/chatbots')
    } catch {
      toasts.error('Silme işlemi başarısız oldu.')
    } finally {
      setIsDeleting(false)
    }
  }

  

  const handleChat = async (message?: string) => {
    if (chatLoading || (!chatInput.trim() && !message) || !id) return
    const userMsg = message || chatInput
    setChatHistory(prev => [...prev, { role: 'user', content: userMsg }])
    setChatInput('')
    setChatLoading(true)

    try {
      const { data } = await api.post(`/api/v1/chatbots/${id}/chat`, { 
        message: userMsg, 
        session_id: sessionId || 'test-playground' 
      })
      setChatHistory(prev => [...prev, { role: 'assistant', content: data.response }])
    } catch {
      setChatHistory(prev => [...prev, { role: 'assistant', content: 'Bir hata oluştu.' }])
    } finally {
      setChatLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <HeaderActions
        isNew={isNew}
        name={name}
        isDeleting={isDeleting}
        isSaving={isSaving}
        onDelete={handleDelete}
        onSave={handleSave}
      />

      {isNew ? (
        <NewChatbotForm
          name={name}
          description={description}
          onNameChange={setName}
          onDescriptionChange={setDescription}
        />
      ) : (
        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex flex-col lg:flex-row gap-8 items-start">
          <ChatbotSidebar activeTab={activeTab} onTabChange={setActiveTab} />

          <div className="flex-1 w-full min-w-0">
          {/* OVERVIEW TAB */}
          <TabsContent value="overview" className="mt-0 space-y-6">
            <OverviewPanel
              name={name}
              setName={setName}
              systemPrompt={systemPrompt}
              setSystemPrompt={setSystemPrompt}
              model={model}
              setModel={setModel}
              temperature={temperature}
              setTemperature={setTemperature}
              maxTokens={maxTokens}
              setMaxTokens={setMaxTokens}
            />
          </TabsContent>

          {/* ANALYTICS TAB */}
          <TabsContent value="analytics" className="mt-0 space-y-6">
             <ChatbotAnalytics chatbotId={id || ''} />
          </TabsContent>

          <TabsContent value="guardrails" className="mt-0 space-y-6">
            <GuardrailsSettings
              confidenceThreshold={confidenceThreshold}
              setConfidenceThreshold={setConfidenceThreshold}
              fallbackMessages={fallbackMessages}
              setFallbackMessages={setFallbackMessages}
              topicRestrictions={topicRestrictions}
              setTopicRestrictions={setTopicRestrictions}
            />
          </TabsContent>

          <TabsContent value="handoff" className="mt-0 space-y-6">
            <HandoffSettings
              handoffEnabled={handoffEnabled}
              setHandoffEnabled={setHandoffEnabled}
              handoffType={handoffType}
              setHandoffType={setHandoffType}
              handoffConfig={handoffConfig}
              setHandoffConfig={setHandoffConfig}
            />
          </TabsContent>

          <TabsContent value="sources" className="mt-0 space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Bilgi Bankası</CardTitle>
                <CardDescription>Botunuzun cevap verirken kullanacağı kaynakları ekleyin.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
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
                    />
                  }
                />
                
                {/* Pending URLs Panel - shows when discovery mode is 'pending' and there are pending URLs */}
                <PendingURLsPanel
                  chatbotId={id}
                  onSourcesCreated={refreshSources}
                />
                
                {sources.length > 0 ? (
                  <SourceList 
                    sources={sources as any} 
                    userPlan={userPlan}
                    onDelete={handleDeleteSource} 
                    onRefresh={handleRefreshSource}
                    refreshingId={refreshingId}
                  />
                ) : (
                  <div className="rounded-2xl border border-border bg-muted/30 p-10 text-center space-y-3">
                    <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-muted">
                      <Inbox className="w-6 h-6 text-muted-foreground" />
                    </div>
                    <div className="text-sm font-medium text-foreground">Henüz kaynak eklenmemiş</div>
                    <div className="text-xs text-muted-foreground">PDF yükleyin, bir URL ekleyin veya metin girin.</div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* PLAYGROUND TAB */}
          <TabsContent value="playground" className="mt-0 flex flex-col lg:flex-row gap-6 h-auto lg:h-[650px]">
            {/* Settings Column */}
            <div className="w-full lg:w-[320px] flex-shrink-0 flex flex-col gap-4 overflow-y-auto pr-2">
              
              <IdentitySection 
                isExpanded={expandedSection === 'identity'}
                onToggle={() => setExpandedSection(expandedSection === 'identity' ? null : 'identity')}
                botDisplayName={botDisplayName}
                setBotDisplayName={setBotDisplayName}
                botIcon={botIcon}
                setBotIcon={setBotIcon}
                welcomeMessage={welcomeMessage}
                setWelcomeMessage={setWelcomeMessage}
              />

              <AppearanceSection 
                isExpanded={expandedSection === 'appearance'}
                onToggle={() => setExpandedSection(expandedSection === 'appearance' ? null : 'appearance')}
                position={position}
                setPosition={setPosition}
                chatFontFamily={chatFontFamily}
                setChatFontFamily={setChatFontFamily}
                themeColor={themeColor}
                setThemeColor={setThemeColor}
              />

              <ColorsSection 
                isExpanded={expandedSection === 'colors'}
                onToggle={() => setExpandedSection(expandedSection === 'colors' ? null : 'colors')}
                chatBackgroundColor={chatBackgroundColor}
                setChatBackgroundColor={setChatBackgroundColor}
                chatHeaderColor={chatHeaderColor}
                setChatHeaderColor={setChatHeaderColor}
                chatHeaderTextColor={chatHeaderTextColor}
                setChatHeaderTextColor={setChatHeaderTextColor}
                botMessageColor={botMessageColor}
                setBotMessageColor={setBotMessageColor}
                botMessageTextColor={botMessageTextColor}
                setBotMessageTextColor={setBotMessageTextColor}
                userMessageColor={userMessageColor}
                setUserMessageColor={setUserMessageColor}
                userMessageTextColor={userMessageTextColor}
                setUserMessageTextColor={setUserMessageTextColor}
              />

              <BrandingSettings
                isExpanded={expandedSection === 'branding'}
                onToggle={() => setExpandedSection(expandedSection === 'branding' ? null : 'branding')}
                hideBranding={hideBranding}
                setHideBranding={setHideBranding}
                customBranding={customBranding}
                setCustomBranding={setCustomBranding}
                canHideBranding={planConfig.branding?.can_hide_branding ?? false}
                canCustomBranding={planConfig.branding?.can_custom_branding ?? false}
              />

            </div>

          <PlaygroundPreview
            id={id || 'preview'}
            themeColor={themeColor}
            chatHeaderColor={chatHeaderColor}
            chatHeaderTextColor={chatHeaderTextColor}
            botMessageColor={botMessageColor}
            botMessageTextColor={botMessageTextColor}
            userMessageColor={userMessageColor}
            userMessageTextColor={userMessageTextColor}
            chatFontFamily={chatFontFamily}
            position={position}
            botDisplayName={botDisplayName}
            botIcon={botIcon}
            chatBackgroundColor={chatBackgroundColor}
            welcomeMessage={welcomeMessage}
            previewOpen={previewOpen}
            sessionId={sessionId}
            suggestionsEnabled={suggestionsEnabled}
            suggestedQuestions={suggestedQuestions}
            refreshKey={previewRefreshKey}
            hideBranding={hideBranding}
            customBranding={customBranding}
          />
          </TabsContent>

          {/* ACTIONS TAB */}
          <TabsContent value="actions" className="mt-0 space-y-6">
            <ActionList chatbotId={id || ''} />
          </TabsContent>

          {/* CONNECT TAB */}
          <TabsContent value="connect" className="mt-0">
            <EmbeddingCodePanel
              id={id || ''}
              userPlan={userPlan}
              secureEmbedEnabled={secureEmbedEnabled}
              allowedDomains={allowedDomains}
              embedSecret={embedSecret}
              onToggleSecure={(v) => setSecureEmbedEnabled(v)}
              onDomainsChange={(v) => setAllowedDomains(v)}
              onSecretChange={(v) => setEmbedSecret(v)}
              onSecretRefresh={() => setEmbedSecret(Math.random().toString(36).slice(2)+Math.random().toString(36).slice(2))}
            />
          </TabsContent>

          {/* SUGGESTIONS TAB */}
          <TabsContent value="suggestions" className="mt-0 space-y-6">
            <SuggestionsPanel 
              suggestionsEnabled={suggestionsEnabled}
              setSuggestionsEnabled={setSuggestionsEnabled}
              suggestedQuestions={suggestedQuestions}
              setSuggestedQuestions={setSuggestedQuestions}
            />
          </TabsContent>
          </div>
        </Tabs>
      )}
      {import.meta.env.MODE === 'test' && (
        <div className="hidden">
          <button aria-label="Test Chat Send" onClick={() => handleChat('Merhaba')}></button>
          <div data-testid="chat-last-assistant">{chatHistory.filter(m => m.role === 'assistant').slice(-1)[0]?.content || ''}</div>
        </div>
      )}
    </div>
  )
}

export default ChatbotDetailPage
