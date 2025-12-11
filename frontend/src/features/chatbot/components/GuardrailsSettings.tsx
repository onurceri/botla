import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Shield, MessageSquareWarning, Ban, Plus, X, Gauge, AlertTriangle, Zap, CheckCircle2 } from 'lucide-react'
import { cn } from '@/lib/utils'

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
        <TabsList className="w-full grid grid-cols-1 sm:grid-cols-3 h-auto sm:h-12 p-1 bg-muted/50 rounded-xl gap-1 sm:gap-0">
          <TabsTrigger value="thresholds" className="rounded-lg data-[state=active]:bg-background data-[state=active]:shadow-sm transition-all py-2 sm:py-1">
            <Gauge className="w-4 h-4 mr-2" />
            Eşleşme & Güven
          </TabsTrigger>
          <TabsTrigger value="messages" className="rounded-lg data-[state=active]:bg-background data-[state=active]:shadow-sm transition-all py-2 sm:py-1">
            <MessageSquareWarning className="w-4 h-4 mr-2" />
            Mesajlar
          </TabsTrigger>
          <TabsTrigger value="restrictions" className="rounded-lg data-[state=active]:bg-background data-[state=active]:shadow-sm transition-all py-2 sm:py-1">
            <Ban className="w-4 h-4 mr-2" />
            Konu Kontrolü
          </TabsTrigger>
        </TabsList>

        {/* --- TAB 1: THRESHOLDS (Cevap Güvenilirlik) --- */}
        <TabsContent value="thresholds" className="mt-6 space-y-6 animate-in fade-in-50 duration-300">
          <Card className="border-muted-foreground/20 shadow-sm overflow-hidden">
            <CardHeader className="bg-muted/30 pb-6">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-primary/10 rounded-xl text-primary ring-1 ring-primary/20">
                    <Gauge className="w-6 h-6" />
                </div>
                <div>
                    <CardTitle className="text-xl">Cevap Güvenilirlik Ayarları</CardTitle>
                    <CardDescription className="text-base mt-1">
                        Botunuzun cevap verme kararlılığını ve güven eşiklerini yönetin.
                    </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-8 p-6">
              
              {/* Visualization Bar */}
              <div className="space-y-3 bg-muted/30 p-6 rounded-2xl border border-border/50">
                <div className="flex justify-between items-end mb-2">
                    <span className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">Güven Skor Dağılımı</span>
                </div>
                <div className="relative h-12 w-full rounded-xl overflow-hidden flex shadow-sm ring-1 ring-black/5">
                  {/* Red Zone */}
                  <div 
                    style={{ width: `${thresholdConfig.medium_threshold * 100}%` }} 
                    className="bg-red-500 flex flex-col items-center justify-center transition-all duration-300 relative group overflow-hidden"
                  >
                    <span className="text-xs font-bold text-white/90 drop-shadow-md">YOK</span>
                    <span className="text-[10px] text-white/80 font-medium hidden group-hover:block absolute bottom-1 whitespace-nowrap">Cevap Vermez</span>
                  </div>
                  {/* Yellow Zone */}
                  <div 
                    style={{ width: `${(thresholdConfig.high_threshold - thresholdConfig.medium_threshold) * 100}%` }} 
                    className="bg-amber-400 flex flex-col items-center justify-center transition-all duration-300 relative group overflow-hidden"
                  >
                     <span className="text-xs font-bold text-white/95 drop-shadow-md">ŞÜPHELİ</span>
                     <span className="text-[10px] text-white/90 font-medium hidden group-hover:block absolute bottom-1 whitespace-nowrap">Kontrol Eder</span>
                  </div>
                  {/* Green Zone */}
                  <div 
                    style={{ width: `${(1 - thresholdConfig.high_threshold) * 100}%` }} 
                    className="bg-emerald-500 flex flex-col items-center justify-center transition-all duration-300 relative group overflow-hidden"
                  >
                    <span className="text-xs font-bold text-white/90 drop-shadow-md">GÜVENİLİR</span>
                    <span className="text-[10px] text-white/80 font-medium hidden group-hover:block absolute bottom-1 whitespace-nowrap">Direkt Cevaplar</span>
                  </div>
                </div>
                
                {/* Scale Indicators */}
                <div className="relative h-6 w-full text-xs font-medium text-muted-foreground select-none">
                     <span className="absolute left-0 top-0">0%</span>
                     <span className="absolute right-0 top-0">100%</span>
                     
                     <div 
                        className="absolute -translate-x-1/2 top-0 flex flex-col items-center transition-all duration-300" 
                        style={{ left: `${thresholdConfig.medium_threshold * 100}%` }}
                     >
                        <div className="h-2 w-px bg-red-500 mb-1"></div>
                        <span className="text-red-600 font-bold bg-red-50 px-1.5 py-0.5 rounded border border-red-100">
                            {(thresholdConfig.medium_threshold * 100).toFixed(0)}%
                        </span>
                     </div>
                     
                     <div 
                        className="absolute -translate-x-1/2 top-0 flex flex-col items-center transition-all duration-300" 
                        style={{ left: `${thresholdConfig.high_threshold * 100}%` }}
                     >
                        <div className="h-2 w-px bg-emerald-500 mb-1"></div>
                        <span className="text-emerald-600 font-bold bg-emerald-50 px-1.5 py-0.5 rounded border border-emerald-100">
                            {(thresholdConfig.high_threshold * 100).toFixed(0)}%
                        </span>
                     </div>
                </div>
              </div>

              <div className="grid gap-6 md:grid-cols-2">
                {/* Medium Threshold Slider */}
                <div className="space-y-5 p-5 border rounded-2xl bg-card hover:border-red-200 transition-colors shadow-sm">
                    <div className="flex justify-between items-start">
                      <div className="space-y-1.5">
                        <div className="flex items-center gap-2">
                          <div className="p-1.5 bg-red-100 text-red-600 rounded-lg">
                             <AlertTriangle className="w-4 h-4" />
                          </div>
                          <label className="font-semibold text-foreground">Alt Sınır (Red)</label>
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                            Bu skorun altındaki eşleşmeler tamamen reddedilir ve "Bilgi Yok" mesajı gösterilir.
                        </p>
                      </div>
                    </div>
                    
                    <div className="pt-2">
                        <input
                          type="range"
                          min="10"
                          max="50"
                          step="5"
                          value={thresholdConfig.medium_threshold * 100}
                          onChange={(e) => updateThresholdConfig('medium_threshold', parseInt(e.target.value) / 100)}
                          disabled={!canCustomizeThresholds}
                          className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-red-500 hover:accent-red-600 transition-all"
                        />
                    </div>
                </div>

                {/* High Threshold Slider */}
                <div className="space-y-5 p-5 border rounded-2xl bg-card hover:border-emerald-200 transition-colors shadow-sm">
                    <div className="flex justify-between items-start">
                      <div className="space-y-1.5">
                         <div className="flex items-center gap-2">
                            <div className="p-1.5 bg-emerald-100 text-emerald-600 rounded-lg">
                                <CheckCircle2 className="w-4 h-4" />
                            </div>
                            <label className="font-semibold text-foreground">Üst Sınır (Kabul)</label>
                         </div>
                         <p className="text-xs text-muted-foreground leading-relaxed">
                            Bu skorun üzerindeki eşleşmeler güvenilir kabul edilir ve doğrudan kullanıcıya sunulur.
                         </p>
                      </div>
                    </div>

                    <div className="pt-2">
                        <input
                          type="range"
                          min="30"
                          max="90"
                          step="5"
                          value={thresholdConfig.high_threshold * 100}
                          onChange={(e) => updateThresholdConfig('high_threshold', parseInt(e.target.value) / 100)}
                          disabled={!canCustomizeThresholds}
                          className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-emerald-500 hover:accent-emerald-600 transition-all"
                        />
                    </div>
                </div>
              </div>

              {/* Settings Group */}
              <div className="grid md:grid-cols-2 gap-6 pt-2">
                 <div className="flex items-center justify-between p-4 border rounded-xl bg-muted/20">
                    <div className="space-y-1 pr-4">
                      <label className="text-sm font-semibold text-foreground">Belirsizlik Uyarısı</label>
                      <p className="text-xs text-muted-foreground">
                        Sarı bölgeye düşen (şüpheli) sonuçlarda kullanıcıya cevabın kesin olmayabileceğini belirt.
                      </p>
                    </div>
                    <Switch
                      checked={thresholdConfig.show_confidence_warning}
                      onCheckedChange={(v) => updateThresholdConfig('show_confidence_warning', v)}
                    />
                 </div>

                 <div className="space-y-2 p-4 border rounded-xl bg-muted/20">
                    <label className="text-sm font-semibold text-foreground block">Eşleşme Bulunamadığında</label>
                    <Select 
                      value={thresholdConfig.fallback_mode} 
                      onValueChange={(v: 'smart' | 'static' | 'escalate') => updateThresholdConfig('fallback_mode', v)}
                    >
                      <SelectTrigger className="w-full bg-background">
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
                <div className="bg-primary/5 border border-primary/20 rounded-xl p-4 flex items-center gap-4">
                  <div className="p-2.5 bg-primary/10 rounded-full text-primary shrink-0">
                     <Shield className="w-5 h-5" />
                  </div>
                  <div>
                    <span className="font-semibold text-primary block">Pro Özelliği</span>
                    <span className="text-muted-foreground text-sm">Güven eşiklerini özelleştirmek ve gelişmiş ayarları açmak için planınızı yükseltin.</span>
                  </div>
                  <Button variant="outline" size="sm" className="ml-auto border-primary/20 hover:bg-primary/10 text-primary">Yükselt</Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* --- TAB 2: MESSAGES (Yedek Mesajlar) --- */}
        <TabsContent value="messages" className="mt-6 animate-in fade-in-50 duration-300">
          <Card className="border-muted-foreground/20 shadow-sm">
            <CardHeader className="bg-muted/30 pb-6">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-blue-500/10 rounded-xl text-blue-600 ring-1 ring-blue-500/20">
                    <MessageSquareWarning className="w-6 h-6" />
                </div>
                <div>
                    <CardTitle className="text-xl">Yedek Mesajlar</CardTitle>
                    <CardDescription className="text-base mt-1">
                        Beklenmedik durumlarda veya hata anında botun vereceği standart cevapları belirleyin.
                    </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-8 p-6">
              
              <div className="grid gap-8">
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                      <label className="text-sm font-semibold flex items-center gap-2 text-foreground">
                        Bilgi Bulunamadı Mesajı
                        {thresholdConfig.fallback_mode === 'static' && <Badge variant="secondary" className="text-[10px] font-normal">Aktif Mod</Badge>}
                      </label>
                      {!canCustomizeMessages && <Badge variant="outline" className="text-[10px] text-muted-foreground border-primary/20 bg-primary/5">Plan Yükselt</Badge>}
                  </div>
                  <div className="relative">
                      <Textarea
                        placeholder="Üzgünüm, bu konuda bilgim yok."
                        className="min-h-[120px] resize-none bg-background/50 focus:bg-background transition-colors text-base"
                        disabled={!canCustomizeMessages}
                        value={fallbackMessages?.no_info_found || ''}
                        onChange={(e) => updateFallbackMessage('no_info_found', e.target.value)}
                      />
                      <div className="absolute bottom-3 right-3 text-xs text-muted-foreground pointer-events-none">
                         {(fallbackMessages?.no_info_found || '').length} karakter
                      </div>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Bu mesaj, bot kullanıcının sorusuna veritabanında yeterli bir cevap bulamadığında ("Sabit Mesaj" modu seçiliyse) gösterilir.
                  </p>
                </div>
                
                <div className="border-t border-border/50" />
                
                <div className="space-y-3">
                  <label className="text-sm font-semibold text-foreground">Hata Mesajı</label>
                  <div className="relative">
                      <Textarea
                        placeholder="Bir hata oluştu, lütfen tekrar deneyin."
                        className="min-h-[100px] resize-none bg-background/50 focus:bg-background transition-colors text-base"
                        disabled={!canCustomizeMessages}
                        value={fallbackMessages?.error_message || ''}
                        onChange={(e) => updateFallbackMessage('error_message', e.target.value)}
                      />
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Sunucu hatası veya teknik bir aksaklık durumunda gösterilecek genel hata mesajı.
                  </p>
                </div>
              </div>

            </CardContent>
          </Card>
        </TabsContent>

        {/* --- TAB 3: RESTRICTIONS (Konu Kısıtlamaları) --- */}
        <TabsContent value="restrictions" className="mt-6 animate-in fade-in-50 duration-300">
           <Card className="border-muted-foreground/20 shadow-sm">
            <CardHeader className="bg-muted/30 pb-6">
              <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 sm:gap-0">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 bg-rose-500/10 rounded-xl text-rose-600 ring-1 ring-rose-500/20">
                     <Ban className="w-6 h-6" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                        <CardTitle className="text-xl">Konu Sınırlamaları</CardTitle>
                        {!canManageTopics && <Badge variant="secondary" className="text-[10px]">Pro</Badge>}
                    </div>
                    <CardDescription className="text-base mt-1">
                        Botun konuşabileceği ve konuşamayacağı konuları kesin olarak belirleyin.
                    </CardDescription>
                  </div>
                </div>
                <div className="flex items-center justify-between sm:justify-start w-full sm:w-auto gap-3 bg-background px-4 py-2 rounded-xl sm:rounded-full border shadow-sm">
                    <span className="text-sm font-medium">Kısıtlamalar</span>
                    <Switch
                    checked={!!topicRestrictions}
                    onCheckedChange={toggleTopicRestrictions}
                    disabled={!canManageTopics}
                    />
                </div>
              </div>
            </CardHeader>
            <CardContent className="p-6">
              {!topicRestrictions ? (
                <div className="flex flex-col items-center justify-center py-12 text-center border-2 border-dashed rounded-2xl bg-muted/10 space-y-4">
                  <div className="p-4 bg-muted rounded-full">
                     <Shield className="w-8 h-8 text-muted-foreground/50" />
                  </div>
                  <div className="max-w-md space-y-2">
                    <h3 className="text-lg font-semibold">Kısıtlama Yok</h3>
                    <p className="text-muted-foreground">
                        Şu anda konu sınırlaması aktif değil. Bot, bilgi bankasındaki içeriklere dayanarak her türlü soruya cevap vermeye çalışır.
                    </p>
                  </div>
                  <Button onClick={() => toggleTopicRestrictions(true)} disabled={!canManageTopics} className="mt-4">
                    Kısıtlamaları Aktifleştir
                  </Button>
                </div>
              ) : (
                <div className="space-y-8 animate-in slide-in-from-top-4 duration-300">
                  {/* Blocked Topics */}
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                         <label className="text-base font-semibold text-foreground flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-red-500"></span>
                            Yasaklı Konular (Blacklist)
                         </label>
                         <span className="text-xs text-muted-foreground">Bu konularda asla cevap verilmez.</span>
                    </div>
                    
                    <div className="flex flex-col sm:flex-row gap-2">
                      <Input 
                        className="h-11 bg-background/50"
                        placeholder="Örn: Siyaset, Rakip Firmalar, Fiyat Listesi..." 
                        value={newBlockedTopic}
                        onChange={(e) => setNewBlockedTopic(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addTopic('blocked'))}
                      />
                      <Button type="button" size="lg" onClick={() => addTopic('blocked')} disabled={!newBlockedTopic.trim()} className="px-6 w-full sm:w-auto">
                        <Plus className="w-5 h-5" />
                      </Button>
                    </div>

                    <div className="bg-muted/30 rounded-xl p-4 min-h-[80px] border border-border/50">
                        {(!topicRestrictions.blocked_topics || topicRestrictions.blocked_topics.length === 0) ? (
                          <div className="flex items-center justify-center h-full text-sm text-muted-foreground italic">
                            Henüz yasaklı konu eklenmedi.
                          </div>
                        ) : (
                          <div className="flex flex-wrap gap-2">
                            {topicRestrictions.blocked_topics?.map((topic, i) => (
                                <Badge key={i} variant="destructive" className="pl-3 pr-1.5 py-1.5 text-sm hover:bg-red-600 transition-colors shadow-sm">
                                {topic}
                                <button type="button" onClick={() => removeTopic('blocked', i)} className="hover:bg-red-700/50 rounded-full p-0.5 ml-2 transition-colors">
                                    <X className="w-3.5 h-3.5" />
                                </button>
                                </Badge>
                            ))}
                          </div>
                        )}
                    </div>
                  </div>

                  <div className="relative">
                    <div className="absolute inset-0 flex items-center">
                        <span className="w-full border-t border-border" />
                    </div>
                    <div className="relative flex justify-center text-xs uppercase">
                        <span className="bg-card px-2 text-muted-foreground">Ve / Veya</span>
                    </div>
                  </div>

                  {/* Allowed Topics (Whitelist) */}
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                       <label className="text-base font-semibold text-foreground flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
                            İzin Verilen Konular (Whitelist)
                       </label>
                       <Badge variant="secondary" className="text-[10px] font-normal">Opsiyonel</Badge>
                    </div>
                    
                    <div className="flex flex-col sm:flex-row gap-2">
                      <Input 
                        className="h-11 bg-background/50"
                        placeholder="Örn: Ürün Özellikleri, Destek, Kargo..." 
                        value={newAllowedTopic}
                        onChange={(e) => setNewAllowedTopic(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addTopic('allowed'))}
                      />
                      <Button type="button" size="lg" variant="outline" onClick={() => addTopic('allowed')} disabled={!newAllowedTopic.trim()} className="px-6 w-full sm:w-auto">
                        <Plus className="w-5 h-5" />
                      </Button>
                    </div>

                    <div className="bg-muted/30 rounded-xl p-4 min-h-[80px] border border-border/50">
                        {(!topicRestrictions.allowed_topics || topicRestrictions.allowed_topics.length === 0) ? (
                          <div className="flex items-center justify-center h-full text-sm text-muted-foreground italic">
                            Liste boş (Tüm konulara izin verilir).
                          </div>
                        ) : (
                          <div className="flex flex-wrap gap-2">
                            {topicRestrictions.allowed_topics?.map((topic, i) => (
                                <Badge key={i} variant="secondary" className="pl-3 pr-1.5 py-1.5 text-sm bg-background border shadow-sm hover:bg-accent transition-colors text-foreground">
                                {topic}
                                <button type="button" onClick={() => removeTopic('allowed', i)} className="hover:bg-muted-foreground/20 rounded-full p-0.5 ml-2 transition-colors">
                                    <X className="w-3.5 h-3.5" />
                                </button>
                                </Badge>
                            ))}
                          </div>
                        )}
                    </div>
                    <p className="text-xs text-muted-foreground mt-2">
                        * Eğer "İzin Verilen Konular" listesine ekleme yaparsanız, bot <strong>SADECE</strong> bu konular hakkında konuşur, diğer her şeyi reddeder.
                    </p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
