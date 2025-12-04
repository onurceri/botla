import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { 
  Save, 
  Trash2, 
  Play, 
  Code, 
  Settings,
  Database,
  MessageSquare,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  ChevronDown,
  ChevronRight,
  X,
  MessageCircle,
  Palette,
  Type,
  Layout,
  User,
  Inbox,
  Info,
  Bot,
  Plus
} from 'lucide-react'
import { api } from '@/api/client'
import { uploadPDFSource, uploadTextSource, uploadURLSource, listSources, getSourceStatus, deleteSource } from '@/api/source'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import SourceUploader from '@/components/chatbot/SourceUploader'
import { cn } from '@/lib/utils'
import { useToast } from '@/components/ui/toast'
import { WidgetApp } from '@widget/widgetApp'
import '@widget/styles.css'

const ChatbotDetailPage = () => {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState('overview')
  
  // Bot State
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('')
  const [model, setModel] = useState('gpt-3.5-turbo')
  const [temperature, setTemperature] = useState(0.7)
  const [maxTokens, setMaxTokens] = useState(512)
  const [themeColor, setThemeColor] = useState('#a78bfa')
  const [welcomeMessage, setWelcomeMessage] = useState('Merhaba! Size nasıl yardımcı olabilirim?')
  const [position, setPosition] = useState('bottom-right')
  const [botMessageColor, setBotMessageColor] = useState('#fcfcfd')
  const [userMessageColor, setUserMessageColor] = useState('#2e408a')
  const [botMessageTextColor, setBotMessageTextColor] = useState('#030303')
  const [userMessageTextColor, setUserMessageTextColor] = useState('#ffffff')
  const [chatFontFamily, setChatFontFamily] = useState('Inter, sans-serif')
  const [chatHeaderColor, setChatHeaderColor] = useState('#3b82f6')
  const [chatHeaderTextColor, setChatHeaderTextColor] = useState('#ffffff')
  const [chatBackgroundColor, setChatBackgroundColor] = useState('#FFF5E6')
  const [botIcon, setBotIcon] = useState('')
  const [botDisplayName, setBotDisplayName] = useState('')
  const [userPlan, setUserPlan] = useState('free')
  const [secureEmbedEnabled, setSecureEmbedEnabled] = useState(false)
  const [allowedDomains, setAllowedDomains] = useState('')
  const [embedSecret, setEmbedSecret] = useState('')
  const [suggestionsEnabled, setSuggestionsEnabled] = useState(false)
  const [suggestedQuestions, setSuggestedQuestions] = useState<string[]>([])
  
  // Sources State
  const [sources, setSources] = useState<any[]>([])
  
  // Chat Test State
  const [chatInput, setChatInput] = useState('')
  const [chatHistory, setChatHistory] = useState<{role: 'user' | 'assistant', content: string}[]>([])
  const [chatLoading, setChatLoading] = useState(false)
  
  // Playground State
  const [previewOpen, setPreviewOpen] = useState(false)
  const [expandedSection, setExpandedSection] = useState<string | null>('identity')
  const [sessionId, setSessionId] = useState('')

  const isNew = id === 'new'

  useEffect(() => {
    try {
      const u = new URL(window.location.href)
      const tab = u.searchParams.get('tab')
      if (tab === 'connect' || tab === 'sources' || tab === 'playground' || tab === 'overview') {
        setActiveTab(tab)
      }
    } catch {}
    // Generate a unique session ID for the playground on mount
    setSessionId(`playground-${Math.random().toString(36).substring(2, 15)}`)
  }, [])

  useEffect(() => {
    api.get('/api/v1/me').then(({ data }) => { setUserPlan(data.subscription_plan || 'free') }).catch(() => {})
    if (!isNew && id) {
      api.get(`/api/v1/chatbots/${id}`).then(({ data }) => {
        setName(data.name || '')
        setDescription(data.description || '')
        setSystemPrompt(data.system_prompt || '')
        setModel(data.model || 'gpt-3.5-turbo')
        setTemperature(data.temperature ?? 0.7)
        setMaxTokens(data.max_tokens ?? 512)
        setThemeColor(data.theme_color || '#a78bfa')
        setWelcomeMessage(data.welcome_message || '')
        setPosition(data.position || 'bottom-right')
        setBotMessageColor(data.bot_message_color || '#fcfcfd')
        setUserMessageColor(data.user_message_color || '#2e408a')
        setBotMessageTextColor(data.bot_message_text_color || '#030303')
        setUserMessageTextColor(data.user_message_text_color || '#ffffff')
        setChatFontFamily(data.chat_font_family || 'Inter, sans-serif')
        setChatHeaderColor(data.chat_header_color || '#3b82f6')
        setChatHeaderTextColor(data.chat_header_text_color || '#ffffff')
        setChatBackgroundColor(data.chat_background_color || '#FFF5E6')
        setBotIcon(data.bot_icon || '')
        setBotDisplayName(data.bot_display_name || '')
        setAllowedDomains(data.allowed_domains || '')
        setEmbedSecret(data.embed_secret || '')
        setSecureEmbedEnabled(!!data.secure_embed_enabled)
        setSuggestionsEnabled(!!data.suggestions_enabled)
        setSuggestedQuestions(Array.isArray(data.suggested_questions) ? data.suggested_questions : [])
      }).catch(() => {})
      
      refreshSources()
    }
  }, [id])

  const refreshSources = () => {
    if (!isNew && id) {
      listSources(id).then(setSources).catch(() => {})
    }
  }

  const { toast } = useToast()
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleSave = async () => {
    if (!name.trim()) {
      toast('Lütfen bir bot ismi girin.', 'error')
      return
    }

    setIsSaving(true)
    const payload = { 
      name, 
      description,
      system_prompt: systemPrompt, 
      model, 
      temperature, 
      max_tokens: maxTokens,
      theme_color: themeColor,
      welcome_message: welcomeMessage,
      position,
      bot_message_color: botMessageColor,
      user_message_color: userMessageColor,
      bot_message_text_color: botMessageTextColor,
      user_message_text_color: userMessageTextColor,
      chat_font_family: chatFontFamily,
      chat_header_color: chatHeaderColor,
      chat_header_text_color: chatHeaderTextColor,
      chat_background_color: chatBackgroundColor,
      bot_icon: botIcon,
      bot_display_name: botDisplayName,
      secure_embed_enabled: secureEmbedEnabled,
      allowed_domains: secureEmbedEnabled ? allowedDomains : undefined,
      embed_secret: secureEmbedEnabled ? embedSecret : undefined
      , suggestions_enabled: suggestionsEnabled
      , suggested_questions: suggestedQuestions
    }

    try {
      if (isNew) {
        const { data } = await api.post('/api/v1/chatbots', payload)
        toast('Chatbot başarıyla oluşturuldu.', 'success')
        navigate(`/chatbots/${data.id}`)
      } else {
        await api.put(`/api/v1/chatbots/${id}`, payload)
        toast('Değişiklikler kaydedildi.', 'success')
      }
    } catch (error) {
      console.error(error)
      toast('Bir hata oluştu. Lütfen tekrar deneyin.', 'error')
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
    } catch (error) {
      toast('Silme işlemi başarısız oldu.', 'error')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleDeleteSource = async (sourceId: string) => {
    if (!confirm('Bu kaynağı silmek istediğinize emin misiniz?')) return

    try {
      await deleteSource(sourceId)
      toast('Kaynak başarıyla silindi.', 'success')
      refreshSources()
    } catch (error) {
      toast('Kaynak silinirken bir hata oluştu.', 'error')
    }
  }

  const pollStatus = async (sid: string) => {
    // Simple polling logic
    let attempts = 0
    const interval = setInterval(async () => {
      attempts++
      try {
        const s = await getSourceStatus(sid)
        if (s.status !== 'pending' && s.status !== 'processing') {
          clearInterval(interval)
          refreshSources()
        }
      } catch { clearInterval(interval) }
      if (attempts > 60) clearInterval(interval)
    }, 1000)
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
    } catch (error) {
      setChatHistory(prev => [...prev, { role: 'assistant', content: 'Bir hata oluştu.' }])
    } finally {
      setChatLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border pb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-foreground">{isNew ? 'Yeni Chatbot' : name}</h1>
          <p className="text-muted-foreground">{isNew ? 'Asistanınızı yapılandırın' : 'Bot ayarlarını ve kaynaklarını yönetin'}</p>
        </div>
        <div className="flex items-center gap-2">
          {!isNew && (
            <Button 
              variant="destructive" 
              size="icon" 
              className="mr-2"
              onClick={handleDelete}
              isLoading={isDeleting}
            >
              <Trash2 className="w-4 h-4" />
            </Button>
          )}
          <Button onClick={handleSave} className="gap-2" isLoading={isSaving}>
            <Save className="w-4 h-4" />
            {isNew ? 'Oluştur' : 'Değişiklikleri Kaydet'}
          </Button>
        </div>
      </div>

      {isNew ? (
        <Card>
          <CardHeader>
            <CardTitle>Temel Bilgiler</CardTitle>
            <CardDescription>Botunuzu oluşturmak için bir isim verin.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Bot İsmi</label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Örn: Müşteri Temsilcisi" />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Açıklama (Opsiyonel)</label>
              <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Botun amacı nedir?" />
            </div>
          </CardContent>
        </Card>
      ) : (
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <div className="overflow-x-auto pb-2 -mx-4 px-4 md:mx-0 md:px-0 md:pb-0 scrollbar-hide">
            <TabsList className="h-auto w-max flex-nowrap justify-start gap-2 md:w-auto md:flex-wrap">
              <TabsTrigger value="overview" className="gap-2 whitespace-nowrap">
                <Settings className="w-4 h-4" /> Genel
              </TabsTrigger>
              <TabsTrigger value="sources" className="gap-2 whitespace-nowrap">
                <Database className="w-4 h-4" /> Veri Kaynakları
              </TabsTrigger>
              <TabsTrigger value="playground" className="gap-2 whitespace-nowrap">
                <Play className="w-4 h-4" /> Playground
              </TabsTrigger>
              <TabsTrigger value="connect" className="gap-2 whitespace-nowrap">
                <Code className="w-4 h-4" /> Entegrasyon
              </TabsTrigger>
              <TabsTrigger value="suggestions" className="gap-2 whitespace-nowrap">
                <MessageSquare className="w-4 h-4" /> Örnek Sorular
              </TabsTrigger>
            </TabsList>
          </div>

          {/* OVERVIEW TAB */}
          <TabsContent value="overview" className="space-y-6">
            <div className="grid gap-6 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Kimlik & Model</CardTitle>
                  <CardDescription>Bot ismi, model seçimi ve sistem mesajı.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">İsim</label>
                    <Input value={name} onChange={(e) => setName(e.target.value)} />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Model</label>
                    <select 
                      className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      value={model}
                      onChange={(e) => setModel(e.target.value)}
                    >
                      <option value="gpt-3.5-turbo">GPT-3.5 Turbo (Hızlı & Ucuz)</option>
                      <option value="gpt-4">GPT-4 (Akıllı & Pahalı)</option>
                    </select>
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">System Prompt</label>
                    <textarea 
                      className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      value={systemPrompt}
                      onChange={(e) => setSystemPrompt(e.target.value)}
                      placeholder="Sen yardımcı bir asistansın..."
                    />
                    <div className="flex justify-end text-xs text-muted-foreground">{systemPrompt.length} karakter</div>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Model Ayarları</CardTitle>
                  <CardDescription>Yaratıcılık ve token sınırı.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-2">
                    <label className="text-xs font-medium text-muted-foreground uppercase">Yaratıcılık (Temperature): {temperature}</label>
                    <input 
                      type="range" 
                      min="0" 
                      max="1" 
                      step="0.1" 
                      value={temperature} 
                      onChange={(e) => setTemperature(parseFloat(e.target.value))}
                      className="w-full accent-primary"
                    />
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>Tutarlı (0.0)</span>
                      <span>Yaratıcı (1.0)</span>
                    </div>
                  </div>
                  <div className="space-y-2">
                    <label className="text-xs font-medium text-muted-foreground uppercase">Maksimum Token</label>
                    <Input 
                      type="number" 
                      value={maxTokens}
                      onChange={(e) => setMaxTokens(Number(e.target.value))}
                      className="w-full"
                    />
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* SOURCES TAB */}
          <TabsContent value="sources" className="space-y-6">
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
                />
                
                {sources.length > 0 ? (
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
                                "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium",
                                s.status === 'completed' ? "bg-emerald-100 text-emerald-700" :
                                s.status === 'processing' ? "bg-blue-100 text-blue-700" :
                                s.status === 'failed' ? "bg-red-100 text-red-700" :
                                "bg-yellow-100 text-yellow-700"
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
                                onClick={() => handleDeleteSource(s.id)}
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
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
          <TabsContent value="playground" className="flex flex-col lg:flex-row gap-6 h-auto lg:h-[650px]">
            {/* Settings Column */}
            <div className="w-full lg:w-[320px] flex-shrink-0 flex flex-col gap-4 overflow-y-auto pr-2">
              
              {/* Identity Section */}
              <div className="border border-border rounded-xl bg-card overflow-hidden">
                <button 
                  onClick={() => setExpandedSection(expandedSection === 'identity' ? null : 'identity')}
                  className="w-full flex items-center justify-between p-4 bg-white/50 backdrop-blur hover:bg-white/70 transition-colors"
                >
                  <div className="flex items-center gap-2 font-medium">
                    <User className="w-4 h-4 text-primary" />
                    Kimlik
                  </div>
                  {expandedSection === 'identity' ? <ChevronDown className="w-4 h-4 text-muted-foreground" /> : <ChevronRight className="w-4 h-4 text-muted-foreground" />}
                </button>
                {expandedSection === 'identity' && (
                  <div className="p-4 space-y-4 border-t border-border animate-in slide-in-from-top-2 duration-200">
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-muted-foreground uppercase">Bot Görünen Adı</label>
                      <Input value={botDisplayName} onChange={(e) => setBotDisplayName(e.target.value)} placeholder="Örn: Asistan" className="bg-background" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-muted-foreground uppercase">Bot İkon URL</label>
                      <Input value={botIcon} onChange={(e) => setBotIcon(e.target.value)} placeholder="https://..." className="bg-background" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-muted-foreground uppercase">Karşılama Mesajı</label>
                      <Input value={welcomeMessage} onChange={(e) => setWelcomeMessage(e.target.value)} className="bg-background" />
                    </div>
                  </div>
                )}
              </div>

              {/* Appearance Section */}
              <div className="border border-border rounded-xl bg-card overflow-hidden">
                <button 
                  onClick={() => setExpandedSection(expandedSection === 'appearance' ? null : 'appearance')}
                  className="w-full flex items-center justify-between p-4 bg-white/50 backdrop-blur hover:bg-white/70 transition-colors"
                >
                  <div className="flex items-center gap-2 font-medium">
                    <Layout className="w-4 h-4 text-primary" />
                    Görünüm
                  </div>
                  {expandedSection === 'appearance' ? <ChevronDown className="w-4 h-4 text-muted-foreground" /> : <ChevronRight className="w-4 h-4 text-muted-foreground" />}
                </button>
                {expandedSection === 'appearance' && (
                  <div className="p-4 space-y-4 border-t border-border animate-in slide-in-from-top-2 duration-200">
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-muted-foreground uppercase">Konum</label>
                      <select 
                        className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                        value={position}
                        onChange={(e) => setPosition(e.target.value)}
                      >
                        <option value="bottom-right">Sağ Alt</option>
                        <option value="bottom-left">Sol Alt</option>
                      </select>
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-muted-foreground uppercase">Yazı Tipi</label>
                      <select 
                        className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                        value={chatFontFamily}
                        onChange={(e) => setChatFontFamily(e.target.value)}
                      >
                        <option value="Inter, sans-serif">Inter (Modern)</option>
                        <option value="Roboto, sans-serif">Roboto</option>
                        <option value="Open Sans, sans-serif">Open Sans</option>
                        <option value="Lato, sans-serif">Lato</option>
                        <option value="Montserrat, sans-serif">Montserrat</option>
                      </select>
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-muted-foreground uppercase">Ana Renk (Theme)</label>
                      <div className="flex gap-2">
                        <Input type="color" value={themeColor} onChange={(e) => setThemeColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                        <Input value={themeColor} onChange={(e) => setThemeColor(e.target.value)} className="flex-1 bg-background font-mono" />
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {/* Colors Section */}
              <div className="border border-border rounded-xl bg-card overflow-hidden">
                <button 
                  onClick={() => setExpandedSection(expandedSection === 'colors' ? null : 'colors')}
                  className="w-full flex items-center justify-between p-4 bg-white/50 backdrop-blur hover:bg-white/70 transition-colors"
                >
                  <div className="flex items-center gap-2 font-medium">
                    <Palette className="w-4 h-4 text-primary" />
                    Renkler
                  </div>
                  {expandedSection === 'colors' ? <ChevronDown className="w-4 h-4 text-muted-foreground" /> : <ChevronRight className="w-4 h-4 text-muted-foreground" />}
                </button>
                {expandedSection === 'colors' && (
                  <div className="p-4 space-y-4 border-t border-border animate-in slide-in-from-top-2 duration-200">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-muted-foreground uppercase">Chat Arka Plan</label>
                        <div className="flex gap-2 items-center">
                          <Input type="color" value={chatBackgroundColor} onChange={(e) => setChatBackgroundColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                          <Input value={chatBackgroundColor} onChange={(e) => setChatBackgroundColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
                        </div>
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-muted-foreground uppercase">Header</label>
                        <div className="flex gap-2 items-center">
                          <Input type="color" value={chatHeaderColor} onChange={(e) => setChatHeaderColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                          <Input value={chatHeaderColor} onChange={(e) => setChatHeaderColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
                        </div>
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-muted-foreground uppercase">Header Yazı</label>
                        <div className="flex gap-2 items-center">
                          <Input type="color" value={chatHeaderTextColor} onChange={(e) => setChatHeaderTextColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                          <Input value={chatHeaderTextColor} onChange={(e) => setChatHeaderTextColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-muted-foreground uppercase">Bot Mesaj</label>
                        <div className="flex gap-2 items-center">
                          <Input type="color" value={botMessageColor} onChange={(e) => setBotMessageColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                          <Input value={botMessageColor} onChange={(e) => setBotMessageColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
                        </div>
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-muted-foreground uppercase">Bot Yazı</label>
                        <div className="flex gap-2 items-center">
                          <Input type="color" value={botMessageTextColor} onChange={(e) => setBotMessageTextColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                          <Input value={botMessageTextColor} onChange={(e) => setBotMessageTextColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-muted-foreground uppercase">Kullanıcı Mesaj</label>
                        <div className="flex gap-2 items-center">
                          <Input type="color" value={userMessageColor} onChange={(e) => setUserMessageColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                          <Input value={userMessageColor} onChange={(e) => setUserMessageColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
                        </div>
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-muted-foreground uppercase">Kullanıcı Yazı</label>
                        <div className="flex gap-2 items-center">
                          <Input type="color" value={userMessageTextColor} onChange={(e) => setUserMessageTextColor(e.target.value)} className="w-8 h-8 p-0 border-0 rounded-full overflow-hidden cursor-pointer" />
                          <Input value={userMessageTextColor} onChange={(e) => setUserMessageTextColor(e.target.value)} className="flex-1 bg-background font-mono text-xs" />
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              

            </div>

            {/* Preview Column (Mock Browser) */}
            <div className="flex-1 flex flex-col bg-background border border-border rounded-xl shadow-2xl overflow-hidden min-h-[500px]">
              {/* Browser Toolbar */}
              <div className="h-10 bg-white/60 backdrop-blur border-b border-border flex items-center px-4 gap-4">
                <div className="flex gap-2">
                  <div className="w-3 h-3 rounded-full bg-red-500/80" />
                  <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
                  <div className="w-3 h-3 rounded-full bg-green-500/80" />
                </div>
                <div className="flex-1 flex justify-center">
                  <div className="bg-background border border-border rounded-md px-3 py-1 text-xs text-muted-foreground w-64 text-center flex items-center justify-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-emerald-500" />
                    example.com
                  </div>
                </div>
                <div className="w-16" /> {/* Spacer for centering */}
              </div>

              {/* Browser Content */}
              <div className="flex-1 relative bg-slate-50" style={{ backgroundImage: 'radial-gradient(#cbd5e1 1px, transparent 1px)', backgroundSize: '20px 20px' }}>
                
                {/* Mock Page Content */}
                <div className="p-12 max-w-3xl mx-auto space-y-8 opacity-20 pointer-events-none select-none">
                  <div className="h-12 w-48 bg-slate-300 dark:bg-slate-700 rounded-lg" />
                  <div className="space-y-4">
                    <div className="h-64 w-full bg-slate-200 dark:bg-slate-800 rounded-xl" />
                    <div className="space-y-2">
                      <div className="h-4 w-full bg-slate-300 dark:bg-slate-700 rounded" />
                      <div className="h-4 w-5/6 bg-slate-300 dark:bg-slate-700 rounded" />
                      <div className="h-4 w-4/6 bg-slate-300 dark:bg-slate-700 rounded" />
                    </div>
                  </div>
                </div>

                {/* Widget Simulation */}
                  <WidgetApp 
                    chatbotId={id || 'preview'}
                    apiBase={import.meta.env.VITE_API_BASE_URL || ''}
                    themeColor={themeColor}
                    headerColor={chatHeaderColor}
                    headerTextColor={chatHeaderTextColor}
                    botMessageColor={botMessageColor}
                    botMessageTextColor={botMessageTextColor}
                    userMessageColor={userMessageColor}
                    userMessageTextColor={userMessageTextColor}
                    fontFamily={chatFontFamily}
                    position={position as any}
                    botNameOverride={botDisplayName}
                    botIconOverride={botIcon}
                    chatBg={chatBackgroundColor}
                    welcome={welcomeMessage}
                    autoOpen={previewOpen}
                    useOverrides={true}
                    resetSession={true}
                    sessionIdOverride={sessionId}
                    suggestions={suggestionsEnabled ? suggestedQuestions : []}
                    positionStrategy="absolute"
                  />

              </div>
            </div>
          </TabsContent>

          {/* CONNECT TAB */}
          <TabsContent value="connect">
            <Card>
              <CardHeader>
                <CardTitle>Web Sitenize Ekleyin</CardTitle>
                <CardDescription>Aşağıdaki kodu sitenizin &lt;body&gt; etiketinin sonuna yapıştırın.</CardDescription>
              </CardHeader>
              <CardContent>
                {userPlan !== 'free' && (
                  <div className="mb-4 flex items-center gap-3">
                    <label className="text-sm font-medium">Güvenli Embed</label>
                    <input type="checkbox" checked={secureEmbedEnabled} onChange={(e) => setSecureEmbedEnabled(e.target.checked)} />
                  </div>
                )}
                {userPlan !== 'free' && secureEmbedEnabled && (
                <div className="grid md:grid-cols-2 gap-4 mb-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">İzinli Alan Adları (virgülle ayırın)</label>
                    <Input value={allowedDomains} onChange={(e) => setAllowedDomains(e.target.value)} placeholder="example.com, another.com" />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Embed Secret</label>
                    <div className="flex gap-2">
                      <Input value={embedSecret} onChange={(e) => setEmbedSecret(e.target.value)} placeholder="Gizli anahtar" />
                      <Button type="button" variant="secondary" onClick={() => setEmbedSecret(Math.random().toString(36).slice(2)+Math.random().toString(36).slice(2))}>Yenile</Button>
                    </div>
                  </div>
                </div>
                )}
                <div className="relative group">
                  <pre className="bg-muted p-4 rounded-xl text-xs font-mono text-foreground overflow-x-auto border border-border shadow-sm">
                    {`<script src="https://cdn.botla.co/widget.js" data-bot="${id}"></script>`}
                  </pre>
                  <Button 
                    size="sm" 
                    variant="secondary" 
                    className="absolute top-2 right-2 shadow-sm"
                    onClick={() => navigator.clipboard.writeText(`<script src="https://cdn.botla.co/widget.js" data-bot="${id}"></script>`)}
                  >
                    Kopyala
                  </Button>
                </div>
                {userPlan === 'free' && (
                  <div className="mt-4 text-xs text-muted-foreground">Güvenli embed (izinli alan adı ve secret) özellikleri ücretli planlarda aktif edilir.</div>
                )}
                <div className="mt-4 flex items-center gap-2 text-xs text-muted-foreground">
                  <Info className="w-4 h-4" />
                  Kodun yüklendiğinden emin olmak için sayfayı yenileyin.
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* SUGGESTIONS TAB */}
          <TabsContent value="suggestions" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Örnek Sorular</CardTitle>
                <CardDescription>Önerilen soruları gösterin ve düzenleyin.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center gap-3">
                  <label className="text-sm font-medium">Örnek soruları göster</label>
                  <input type="checkbox" checked={suggestionsEnabled} onChange={(e) => setSuggestionsEnabled(e.target.checked)} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Sorular</label>
                  <div className="flex flex-wrap gap-2 min-h-[40px] p-4 rounded-xl border border-border bg-muted/10">
                    {suggestedQuestions.length === 0 && (
                      <div className="w-full text-center text-sm text-muted-foreground py-2">
                        Henüz örnek soru eklenmemiş.
                      </div>
                    )}
                    {suggestedQuestions.map((q, i) => (
                      <div key={i} className="group flex items-center gap-2 px-3 py-1.5 rounded-lg border border-border bg-background shadow-sm hover:border-primary/50 transition-colors">
                        <span className="text-sm">{q}</span>
                        <button 
                          className="text-muted-foreground hover:text-red-600 transition-colors opacity-0 group-hover:opacity-100" 
                          onClick={() => setSuggestedQuestions((prev) => prev.filter((_, idx) => idx !== i))}
                        >
                          <X className="w-3 h-3" />
                        </button>
                      </div>
                    ))}
                  </div>
                  
                  {suggestedQuestions.length < 6 ? (
                    <div className="space-y-2">
                      <div className="flex gap-2">
                        <Input 
                          id="new-question-input"
                          placeholder="Yeni bir soru yazın..." 
                          className="flex-1"
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                              const v = (e.target as HTMLInputElement).value.trim()
                              if (v) {
                                setSuggestedQuestions((prev) => {
                                  const nv = [...prev, v.slice(0, 120)]
                                  const uniq = Array.from(new Set(nv.map(s => s.toLowerCase())))
                                  return nv.filter((s, idx) => uniq.indexOf(s.toLowerCase()) === idx).slice(0, 6)
                                })
                                ;(e.target as HTMLInputElement).value = ''
                              }
                            }
                          }} 
                        />
                        <Button 
                          type="button" 
                          onClick={() => {
                            const input = document.getElementById('new-question-input') as HTMLInputElement
                            const v = input.value.trim()
                            if (v) {
                              setSuggestedQuestions((prev) => {
                                const nv = [...prev, v.slice(0, 120)]
                                const uniq = Array.from(new Set(nv.map(s => s.toLowerCase())))
                                return nv.filter((s, idx) => uniq.indexOf(s.toLowerCase()) === idx).slice(0, 6)
                              })
                              input.value = ''
                            }
                          }}
                        >
                          Ekle
                        </Button>
                      </div>
                      <p className="text-[11px] text-muted-foreground flex items-center gap-1">
                        <span className="inline-block px-1.5 py-0.5 rounded border border-border bg-muted text-[10px] font-mono">Enter</span> 
                        tuşuna basarak veya Ekle butonunu kullanarak ekleyebilirsiniz. (Maks. 6 soru)
                      </p>
                    </div>
                  ) : (
                    <div className="text-xs text-amber-600 bg-amber-50 p-2 rounded border border-amber-100">
                      Maksimum 6 soru limitine ulaştınız. Yeni eklemek için mevcutlardan silmelisiniz.
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      )}
    </div>
  )
}

export default ChatbotDetailPage
