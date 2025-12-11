import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Shield, MessageSquareWarning, Ban, Plus, X, Gauge, AlertTriangle, Zap } from 'lucide-react'

type FallbackMessages = {
  no_info_found?: string
  error_message?: string
  handoff_message?: string
}

type TopicConfig = {
  allowed_topics?: string[]
  blocked_topics?: string[]
  blocked_message?: string
}

type ThresholdConfig = {
  high_threshold: number
  medium_threshold: number
  fallback_mode: 'smart' | 'static' | 'escalate'
  show_confidence_warning: boolean
}

interface GuardrailsSettingsProps {
  confidenceThreshold: number
  setConfidenceThreshold: (v: number) => void
  thresholdConfig: ThresholdConfig
  setThresholdConfig: (v: ThresholdConfig) => void
  fallbackMessages: FallbackMessages | null
  setFallbackMessages: (v: FallbackMessages | null) => void
  topicRestrictions: TopicConfig | null
  setTopicRestrictions: (v: TopicConfig | null) => void
  // Feature flags
  canCustomizeThresholds?: boolean
  canUseSmartFallback?: boolean
  canUseEscalateFallback?: boolean
  canManageTopics?: boolean
  canCustomizeMessages?: boolean
}

export default function GuardrailsSettings({
  confidenceThreshold,
  setConfidenceThreshold,
  thresholdConfig,
  setThresholdConfig,
  fallbackMessages,
  setFallbackMessages,
  topicRestrictions,
  setTopicRestrictions,
  canCustomizeThresholds = false,
  canUseSmartFallback = true,
  canUseEscalateFallback = false,
  canManageTopics = false,
  canCustomizeMessages = false
}: GuardrailsSettingsProps) {
  
  const [newAllowedTopic, setNewAllowedTopic] = useState('')
  const [newBlockedTopic, setNewBlockedTopic] = useState('')

  const handleConfidenceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setConfidenceThreshold(parseFloat(e.target.value))
  }

  const updateThresholdConfig = (key: keyof ThresholdConfig, value: any) => {
    setThresholdConfig({
      ...thresholdConfig,
      [key]: value
    })
  }

  const updateFallbackMessage = (key: keyof FallbackMessages, value: string) => {
    const current = fallbackMessages || {}
    setFallbackMessages({
      ...current,
      [key]: value
    })
  }

  const toggleTopicRestrictions = (enabled: boolean) => {
    if (enabled) {
      setTopicRestrictions({
        allowed_topics: [],
        blocked_topics: [],
        blocked_message: "Üzgünüm, bu konuda konuşamıyorum."
      })
    } else {
      setTopicRestrictions(null)
    }
  }

  const addTopic = (type: 'allowed' | 'blocked') => {
    if (!topicRestrictions) return
    
    if (type === 'allowed' && newAllowedTopic.trim()) {
      setTopicRestrictions({
        ...topicRestrictions,
        allowed_topics: [...(topicRestrictions.allowed_topics || []), newAllowedTopic.trim()]
      })
      setNewAllowedTopic('')
    } else if (type === 'blocked' && newBlockedTopic.trim()) {
      setTopicRestrictions({
        ...topicRestrictions,
        blocked_topics: [...(topicRestrictions.blocked_topics || []), newBlockedTopic.trim()]
      })
      setNewBlockedTopic('')
    }
  }

  const removeTopic = (type: 'allowed' | 'blocked', index: number) => {
    if (!topicRestrictions) return

    if (type === 'allowed') {
      const newTopics = [...(topicRestrictions.allowed_topics || [])]
      newTopics.splice(index, 1)
      setTopicRestrictions({ ...topicRestrictions, allowed_topics: newTopics })
    } else {
      const newTopics = [...(topicRestrictions.blocked_topics || [])]
      newTopics.splice(index, 1)
      setTopicRestrictions({ ...topicRestrictions, blocked_topics: newTopics })
    }
  }

  return (
    <div className="space-y-6">
      <Tabs defaultValue="thresholds" className="w-full">
        <TabsList className="w-full grid grid-cols-3">
          <TabsTrigger value="thresholds" className="text-xs sm:text-sm">Eşleşme & Güven</TabsTrigger>
          <TabsTrigger value="messages" className="text-xs sm:text-sm">Mesajlar</TabsTrigger>
          <TabsTrigger value="restrictions" className="text-xs sm:text-sm">Konu Kontrolü</TabsTrigger>
        </TabsList>

        {/* --- TAB 1: THRESHOLDS (Cevap Güvenilirlik) --- */}
        <TabsContent value="thresholds" className="mt-4 space-y-4 animate-in fade-in-50 duration-300">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <Gauge className="w-5 h-5 text-primary" />
                <CardTitle className="text-lg">Cevap Güvenilirlik Ayarları</CardTitle>
              </div>
              <CardDescription>
                Botunuzun cevap verme veya bulamama kararlarını yönetin.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-8">
              
              {/* Visualization Bar */}
              <div className="space-y-2 pt-2">
                <div className="relative h-6 w-full rounded-full overflow-hidden flex text-[10px] font-bold text-white shadow-inner">
                  {/* Red Zone */}
                  <div 
                    style={{ width: `${thresholdConfig.medium_threshold * 100}%` }} 
                    className="bg-red-500/80 flex items-center justify-center transition-all duration-300"
                  >
                    {thresholdConfig.medium_threshold > 0.15 && "YOK"}
                  </div>
                  {/* Yellow Zone */}
                  <div 
                    style={{ width: `${(thresholdConfig.high_threshold - thresholdConfig.medium_threshold) * 100}%` }} 
                    className="bg-yellow-500/80 flex items-center justify-center transition-all duration-300"
                  >
                     {(thresholdConfig.high_threshold - thresholdConfig.medium_threshold) > 0.15 && "ŞÜPHELİ"}
                  </div>
                  {/* Green Zone */}
                  <div 
                    style={{ width: `${(1 - thresholdConfig.high_threshold) * 100}%` }} 
                    className="bg-green-500/80 flex items-center justify-center transition-all duration-300"
                  >
                    {(1 - thresholdConfig.high_threshold) > 0.15 && "GÜVENİLİR"}
                  </div>
                </div>
                <div className="flex justify-between text-xs text-muted-foreground px-1">
                  <span>0%</span>
                  <div className="flex-1 relative mx-2 h-4">
                     <div 
                        className="absolute -translate-x-1/2 transition-all duration-300 font-medium text-yellow-600" 
                        style={{ left: `${thresholdConfig.medium_threshold * 100}%` }}
                     >
                        {(thresholdConfig.medium_threshold * 100).toFixed(0)}%
                     </div>
                     <div 
                        className="absolute -translate-x-1/2 transition-all duration-300 font-medium text-green-600" 
                        style={{ left: `${thresholdConfig.high_threshold * 100}%` }}
                     >
                        {(thresholdConfig.high_threshold * 100).toFixed(0)}%
                     </div>
                  </div>
                  <span>100%</span>
                </div>
              </div>

              <div className="grid gap-6 md:grid-cols-2">
                {/* Medium Threshold Slider */}
                <div className="space-y-4 p-4 border rounded-xl bg-card/50">
                    <div className="flex justify-between items-start">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <AlertTriangle className="w-4 h-4 text-yellow-500" />
                          <label className="text-sm font-medium">Alt Sınır</label>
                        </div>
                        <p className="text-[10px] text-muted-foreground">Bu eşiğin altı "Bulunamadı" sayılır.</p>
                      </div>
                      <Badge variant="outline" className="font-mono">
                        {(thresholdConfig.medium_threshold * 100).toFixed(0)}%
                      </Badge>
                    </div>
                    
                    <input
                      type="range"
                      min="10"
                      max="50"
                      step="5"
                      value={thresholdConfig.medium_threshold * 100}
                      onChange={(e) => updateThresholdConfig('medium_threshold', parseInt(e.target.value) / 100)}
                      disabled={!canCustomizeThresholds}
                      className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-yellow-500 disabled:opacity-50"
                    />
                </div>

                {/* High Threshold Slider */}
                <div className="space-y-4 p-4 border rounded-xl bg-card/50">
                    <div className="flex justify-between items-start">
                      <div className="space-y-1">
                         <div className="flex items-center gap-2">
                            <Zap className="w-4 h-4 text-green-500" />
                            <label className="text-sm font-medium">Üst Sınır</label>
                         </div>
                         <p className="text-[10px] text-muted-foreground">Bu eşiğin üstü "Güvenilir" sayılır.</p>
                      </div>
                      <Badge variant="outline" className="font-mono">
                        {(thresholdConfig.high_threshold * 100).toFixed(0)}%
                      </Badge>
                    </div>

                    <input
                      type="range"
                      min="30"
                      max="90"
                      step="5"
                      value={thresholdConfig.high_threshold * 100}
                      onChange={(e) => updateThresholdConfig('high_threshold', parseInt(e.target.value) / 100)}
                      disabled={!canCustomizeThresholds}
                      className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-green-500 disabled:opacity-50"
                    />
                </div>
              </div>

              {/* Settings Group */}
              <div className="space-y-4 pt-4 border-t">
                 <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <label className="text-sm font-medium">Belirsizlik Uyarısı</label>
                      <p className="text-xs text-muted-foreground">Sarı bölge sonuçlarında kullanıcıya uyarı göster.</p>
                    </div>
                    <Switch
                      checked={thresholdConfig.show_confidence_warning}
                      onCheckedChange={(v) => updateThresholdConfig('show_confidence_warning', v)}
                    />
                 </div>

                 <div className="space-y-3 pt-2">
                    <label className="text-sm font-medium">Eşleşme Bulunamadığında (Kırmızı Bölge)</label>
                    <Select 
                      value={thresholdConfig.fallback_mode} 
                      onValueChange={(v: 'smart' | 'static' | 'escalate') => updateThresholdConfig('fallback_mode', v)}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Fallback modu seçin" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="static">
                          <div className="flex items-center gap-2">
                            <span>📝 Sabit Mesaj Göster</span>
                            <span className="text-xs text-muted-foreground ml-auto pl-2">Standart</span>
                          </div>
                        </SelectItem>
                        <SelectItem value="smart" disabled={!canUseSmartFallback}>
                          <div className="flex items-center gap-2">
                            <span>🤖 Akıllı Yönlendirme Yap</span>
                            {!canUseSmartFallback && <Badge variant="secondary" className="scale-90">Pro</Badge>}
                          </div>
                        </SelectItem>
                        <SelectItem value="escalate" disabled={!canUseEscalateFallback}>
                          <div className="flex items-center gap-2">
                            <span>👤 İnsan Desteğine Aktar</span>
                            {!canUseEscalateFallback && <Badge variant="secondary" className="scale-90">Ent</Badge>}
                          </div>
                        </SelectItem>
                      </SelectContent>
                    </Select>
                 </div>
              </div>

              {!canCustomizeThresholds && (
                <div className="bg-primary/5 border border-primary/20 rounded-lg p-3 text-sm flex items-center gap-3">
                  <div className="p-2 bg-primary/10 rounded-full text-primary">
                     <Shield className="w-4 h-4" />
                  </div>
                  <div>
                    <span className="font-semibold text-primary block">Pro Özelliği</span>
                    <span className="text-muted-foreground text-xs">Bu ayarları özelleştirmek için planınızı yükseltin.</span>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* --- TAB 2: MESSAGES (Yedek Mesajlar) --- */}
        <TabsContent value="messages" className="mt-4 animate-in fade-in-50 duration-300">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <MessageSquareWarning className="w-5 h-5 text-primary" />
                <CardTitle className="text-lg">Yedek Mesajlar</CardTitle>
              </div>
              <CardDescription>
                Beklenmedik durumlarda botun vereceği standart cevaplar.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium flex items-center gap-2">
                    Bilgi Bulunamadı Mesajı
                    {thresholdConfig.fallback_mode === 'static' && <Badge variant="secondary" className="text-[10px]">Aktif</Badge>}
                    {!canCustomizeMessages && <Badge variant="outline" className="text-[10px] text-muted-foreground border-primary/20 bg-primary/5">Plan Yükselt</Badge>}
                  </label>
                  <Textarea
                    placeholder="Üzgünüm, bu konuda bilgim yok."
                    className="min-h-[100px]"
                    disabled={!canCustomizeMessages}
                    value={fallbackMessages?.no_info_found || ''}
                    onChange={(e) => updateFallbackMessage('no_info_found', e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">
                    "Sabit Mesaj" modu veya genel arama başarısız olduğunda gösterilir.
                  </p>
                </div>
                
                <div className="border-t pt-4 space-y-2">
                  <label className="text-sm font-medium">Hata Mesajı</label>
                  <Textarea
                    placeholder="Bir hata oluştu, lütfen tekrar deneyin."
                    className="min-h-[80px]"
                    disabled={!canCustomizeMessages}
                    value={fallbackMessages?.error_message || ''}
                    onChange={(e) => updateFallbackMessage('error_message', e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">
                    Sistem hatası durumunda gösterilecek mesaj.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* --- TAB 3: RESTRICTIONS (Konu Kısıtlamaları) --- */}
        <TabsContent value="restrictions" className="mt-4 animate-in fade-in-50 duration-300">
           <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Ban className="w-5 h-5 text-primary" />
                  <CardTitle className="text-lg">Konu Sınırlamaları</CardTitle>
                  {!canManageTopics && <Badge variant="secondary" className="text-[10px]">Pro</Badge>}
                </div>
                <Switch
                  checked={!!topicRestrictions}
                  onCheckedChange={toggleTopicRestrictions}
                  disabled={!canManageTopics}
                />
              </div>
              <CardDescription>
                Botun konuşabileceği ve konuşamayacağı konuları kesin olarak belirleyin.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {!topicRestrictions ? (
                <div className="text-center py-8 text-muted-foreground text-sm border border-dashed rounded-lg bg-muted/20">
                  Konu sınırlaması aktif değil. Bot kaynaklar dahilindeki her soruya cevap verir.
                  <br/>
                  <Button variant="link" onClick={() => toggleTopicRestrictions(true)} disabled={!canManageTopics} className="mt-2 text-primary">
                    Sınırlamaları Aktifleştir
                  </Button>
                </div>
              ) : (
                <div className="space-y-6 animate-in slide-in-from-top-2 duration-200">
                  {/* Blocked Topics */}
                  <div className="space-y-3">
                    <label className="text-sm font-medium">Yasaklı Konular (Blacklist)</label>
                    <div className="flex gap-2">
                      <Input 
                        placeholder="Örn: Siyaset, Rakip Firmalar" 
                        value={newBlockedTopic}
                        onChange={(e) => setNewBlockedTopic(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addTopic('blocked'))}
                      />
                      <Button type="button" size="sm" onClick={() => addTopic('blocked')} disabled={!newBlockedTopic.trim()}>
                        <Plus className="w-4 h-4" />
                      </Button>
                    </div>
                    {(!topicRestrictions.blocked_topics || topicRestrictions.blocked_topics.length === 0) && (
                      <p className="text-[11px] text-muted-foreground italic">Henüz yasaklı konu eklenmedi.</p>
                    )}
                    <div className="flex flex-wrap gap-2">
                      {topicRestrictions.blocked_topics?.map((topic, i) => (
                        <Badge key={i} variant="destructive" className="gap-1 pr-1 pl-2 py-1">
                          {topic}
                          <button type="button" onClick={() => removeTopic('blocked', i)} className="hover:bg-red-600 rounded-full p-0.5 ml-1">
                            <X className="w-3 h-3" />
                          </button>
                        </Badge>
                      ))}
                    </div>
                  </div>

                  <div className="border-t border-border" />

                  {/* Allowed Topics (Whitelist) */}
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                       <label className="text-sm font-medium">İzin Verilen Konular (Whitelist)</label>
                       <Badge variant="secondary" className="text-[10px]">Opsiyonel</Badge>
                    </div>
                    
                    <div className="flex gap-2">
                      <Input 
                        placeholder="Örn: Ürün Özellikleri, Fiyatlandırma" 
                        value={newAllowedTopic}
                        onChange={(e) => setNewAllowedTopic(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addTopic('allowed'))}
                      />
                      <Button type="button" size="sm" onClick={() => addTopic('allowed')} disabled={!newAllowedTopic.trim()}>
                        <Plus className="w-4 h-4" />
                      </Button>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {topicRestrictions.allowed_topics?.map((topic, i) => (
                        <Badge key={i} variant="secondary" className="gap-1 pr-1 pl-2 py-1 bg-primary/10 text-primary border-primary/20">
                          {topic}
                          <button type="button" onClick={() => removeTopic('allowed', i)} className="hover:bg-primary/20 rounded-full p-0.5 ml-1">
                            <X className="w-3 h-3" />
                          </button>
                        </Badge>
                      ))}
                    </div>
                    <p className="text-[11px] text-muted-foreground bg-blue-50 dark:bg-blue-950/30 p-3 rounded text-blue-800 dark:text-blue-300">
                      <strong>Not:</strong> Whitelist kullanıldığında, bot <u>SADECE</u> burada belirtilen konular hakkında konuşur. Diğer tüm konular reddedilir.
                    </p>
                  </div>

                  <div className="border-t border-border" />

                  {/* Blocked Message */}
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Engelleme Mesajı</label>
                    <Input
                      placeholder="Üzgünüm, bu konuda konuşamıyorum."
                      value={topicRestrictions.blocked_message || ''}
                      onChange={(e) => setTopicRestrictions({...topicRestrictions, blocked_message: e.target.value})}
                    />
                    <p className="text-[11px] text-muted-foreground">Kullanıcı yasaklı bir konuya girdiğinde bu mesaj gösterilir.</p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
      
      {/* Legacy Confidence Threshold (hidden but functional for backward compat) */}
      <input type="hidden" value={confidenceThreshold} onChange={handleConfidenceChange} />
    </div>
  )
}

