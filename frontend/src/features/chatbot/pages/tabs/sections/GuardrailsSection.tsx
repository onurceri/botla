import { useState } from 'react'
import {
  Shield,
  MessageSquareWarning,
  Ban,
  Plus,
  X,
  Gauge,
  Sparkles,
  AlertTriangle,
  FileText,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useChatbotContext } from '../../../context/ChatbotContext'
import { ThresholdConfig, FallbackMessages } from '../../../hooks/useChatbotForm'
import { cn } from '@/lib/utils'

export default function GuardrailsSection() {
  const {
    thresholdConfig,
    setThresholdConfig,
    fallbackMessages,
    setFallbackMessages,
    topicRestrictions,
    setTopicRestrictions,
    planConfig,
  } = useChatbotContext()

  // Tab State
  const [activeTab, setActiveTab] = useState<'thresholds' | 'messages' | 'restrictions'>(
    'thresholds',
  )

  // Local State for Topics
  const [newAllowedTopic, setNewAllowedTopic] = useState('')
  const [newBlockedTopic, setNewBlockedTopic] = useState('')

  // Feature Flags
  const canCustomizeThresholds = planConfig?.guardrails?.can_customize_thresholds ?? false
  const canUseSmartFallback = planConfig?.guardrails?.can_use_smart_fallback ?? true
  const canUseEscalateFallback = planConfig?.guardrails?.can_use_escalate_fallback ?? false
  const canManageTopics = planConfig?.guardrails?.can_manage_topics ?? false
  const canCustomizeMessages = planConfig?.guardrails?.can_customize_messages ?? false

  // Handlers
  const updateThresholdConfig = (key: keyof ThresholdConfig, value: any) => {
    setThresholdConfig({ ...thresholdConfig, [key]: value })
  }

  const updateFallbackMessage = (key: keyof FallbackMessages, value: string) => {
    // @ts-ignore
    const current = fallbackMessages || {}
    setFallbackMessages({ ...current, [key]: value })
  }

  const toggleTopicRestrictions = (enabled: boolean) => {
    if (enabled) {
      setTopicRestrictions({
        allowed_topics: [],
        blocked_topics: [],
        blocked_message: 'Üzgünüm, bu konuda konuşamıyorum.',
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
        allowed_topics: [...(topicRestrictions.allowed_topics || []), newAllowedTopic.trim()],
      })
      setNewAllowedTopic('')
    } else if (type === 'blocked' && newBlockedTopic.trim()) {
      setTopicRestrictions({
        ...topicRestrictions,
        blocked_topics: [...(topicRestrictions.blocked_topics || []), newBlockedTopic.trim()],
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
    <div className="bg-white rounded-[24px] border border-slate-200/60 shadow-sm overflow-hidden flex flex-col h-full group transition-all hover:shadow-md">
      {/* Header */}
      <div className="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-violet-500/10 text-violet-600 ring-1 ring-violet-500/20 shadow-sm">
            <Shield className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-bold tracking-tight text-slate-900 uppercase">
              Güvenlik ve Kısıtlamalar
            </h3>
            <p className="text-[11px] text-slate-500 font-medium">
              Botun neye cevap vereceğini ve güven sınırlarını kontrol edin
            </p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="px-6 pt-6 pb-2">
        <div className="flex p-1 bg-slate-100 rounded-xl gap-1">
          <button
            onClick={() => setActiveTab('thresholds')}
            className={cn(
              'flex-1 flex items-center justify-center gap-2 py-2.5 text-xs font-bold rounded-lg transition-all',
              activeTab === 'thresholds'
                ? 'bg-white text-slate-900 shadow-sm ring-1 ring-black/5'
                : 'text-slate-500 hover:text-slate-700 hover:bg-slate-200/50',
            )}
          >
            <Gauge className="w-3.5 h-3.5" />
            Eşleşme & Güven
          </button>
          <button
            onClick={() => setActiveTab('messages')}
            className={cn(
              'flex-1 flex items-center justify-center gap-2 py-2.5 text-xs font-bold rounded-lg transition-all',
              activeTab === 'messages'
                ? 'bg-white text-slate-900 shadow-sm ring-1 ring-black/5'
                : 'text-slate-500 hover:text-slate-700 hover:bg-slate-200/50',
            )}
          >
            <MessageSquareWarning className="w-3.5 h-3.5" />
            Yedek Mesajlar
          </button>
          <button
            onClick={() => setActiveTab('restrictions')}
            className={cn(
              'flex-1 flex items-center justify-center gap-2 py-2.5 text-xs font-bold rounded-lg transition-all',
              activeTab === 'restrictions'
                ? 'bg-white text-slate-900 shadow-sm ring-1 ring-black/5'
                : 'text-slate-500 hover:text-slate-700 hover:bg-slate-200/50',
            )}
          >
            <Ban className="w-3.5 h-3.5" />
            Konu Kontrolü
          </button>
        </div>
      </div>

      <div className="p-6 lg:p-8 space-y-8 flex-1">
        {/* TAB: THRESHOLDS */}
        {activeTab === 'thresholds' && (
          <div className="space-y-8 animate-in fade-in slide-in-from-bottom-2 duration-300">
            {/* Visualization */}
            <div className="space-y-4">
              <div className="flex flex-col gap-1 mb-2">
                <label className="flex items-center gap-2 text-[11px] font-bold text-slate-500 uppercase tracking-widest">
                  <Gauge className="w-3.5 h-3.5 text-blue-500" />
                  Güven Skoru Analizi
                </label>
                <p className="text-xs text-slate-500">
                  Yapay zekanın cevaplarının güvenilirliğini ölçer ve belirlediğiniz risk seviyesine
                  göre cevapları yönetir.
                </p>
              </div>

              <div className="relative h-16 w-full bg-slate-50 rounded-2xl border border-slate-200/50 overflow-hidden flex shadow-inner">
                {/* Zones */}
                <div
                  style={{ width: `${thresholdConfig.medium_threshold * 100}%` }}
                  className="h-full bg-rose-500/10 flex flex-col items-center justify-center border-r border-rose-500/20 relative group transition-all duration-500"
                >
                  <span className="text-[10px] font-bold text-rose-600 bg-rose-50 px-2 py-0.5 rounded-full ring-1 ring-rose-500/20 mb-1">
                    RED
                  </span>
                  <span className="text-[9px] text-rose-500/80 font-medium">Cevap Yok</span>
                </div>

                <div
                  style={{
                    width: `${(thresholdConfig.high_threshold - thresholdConfig.medium_threshold) * 100}%`,
                  }}
                  className="h-full bg-amber-500/10 flex flex-col items-center justify-center border-r border-amber-500/20 relative group transition-all duration-500"
                >
                  <span className="text-[10px] font-bold text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full ring-1 ring-amber-500/20 mb-1">
                    ŞÜPHELİ
                  </span>
                  <span className="text-[9px] text-amber-500/80 font-medium">Kontrol Edilir</span>
                </div>

                <div
                  style={{ width: `${(1 - thresholdConfig.high_threshold) * 100}%` }}
                  className="h-full bg-emerald-500/10 flex flex-col items-center justify-center relative group transition-all duration-500"
                >
                  <span className="text-[10px] font-bold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full ring-1 ring-emerald-500/20 mb-1">
                    GÜVENİLİR
                  </span>
                  <span className="text-[9px] text-emerald-500/80 font-medium">Direkt Cevap</span>
                </div>
              </div>

              {/* Sliders */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
                {/* Low Threshold */}
                <div className="space-y-4 p-5 rounded-2xl bg-slate-50 border border-slate-200/60">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold text-slate-700 flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-rose-500"></span>
                      Alt Sınır (Red)
                    </span>
                    <span className="text-xs font-mono font-medium text-slate-500 bg-white px-2 py-1 rounded-md border border-slate-200">
                      {(thresholdConfig.medium_threshold * 100).toFixed(0)}%
                    </span>
                  </div>
                  <input
                    type="range"
                    min="10"
                    max="50"
                    step="5"
                    value={thresholdConfig.medium_threshold * 100}
                    onChange={(e) =>
                      updateThresholdConfig('medium_threshold', parseInt(e.target.value) / 100)
                    }
                    disabled={!canCustomizeThresholds}
                    className="w-full h-1.5 bg-slate-200 rounded-lg appearance-none cursor-pointer accent-rose-500 hover:accent-rose-600 transition-all"
                  />
                  <p className="text-[11px] text-slate-500 leading-relaxed">
                    Bu skorun altındaki tüm eşleşmeler reddedilir.
                  </p>
                </div>

                {/* High Threshold */}
                <div className="space-y-4 p-5 rounded-2xl bg-slate-50 border border-slate-200/60">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold text-slate-700 flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
                      Üst Sınır (Kabul)
                    </span>
                    <span className="text-xs font-mono font-medium text-slate-500 bg-white px-2 py-1 rounded-md border border-slate-200">
                      {(thresholdConfig.high_threshold * 100).toFixed(0)}%
                    </span>
                  </div>
                  <input
                    type="range"
                    min="30"
                    max="90"
                    step="5"
                    value={thresholdConfig.high_threshold * 100}
                    onChange={(e) =>
                      updateThresholdConfig('high_threshold', parseInt(e.target.value) / 100)
                    }
                    disabled={!canCustomizeThresholds}
                    className="w-full h-1.5 bg-slate-200 rounded-lg appearance-none cursor-pointer accent-emerald-500 hover:accent-emerald-600 transition-all"
                  />
                  <p className="text-[11px] text-slate-500 leading-relaxed">
                    Bu skorun üzerindeki eşleşmeler güvenilir kabul edilir.
                  </p>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 border-t border-slate-100 pt-8">
              {/* Uncertainty Warning */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-semibold text-slate-900">
                    Belirsizlik Uyarısı
                  </label>
                  <Switch
                    checked={thresholdConfig.show_confidence_warning}
                    onCheckedChange={(v) => updateThresholdConfig('show_confidence_warning', v)}
                  />
                </div>
                <p className="text-xs text-slate-500 leading-relaxed">
                  Şüpheli (sarı) bölgedeki cevaplar için kullanıcıya cevabın kesin olmayabileceğini
                  belirtir.
                </p>
              </div>

              {/* No Match Action */}
              <div className="space-y-3">
                <label className="text-sm font-semibold text-slate-900 block">
                  Eşleşme Bulunamadığında
                </label>
                <Select
                  value={thresholdConfig.fallback_mode}
                  onValueChange={(v: 'smart' | 'static' | 'escalate') =>
                    updateThresholdConfig('fallback_mode', v)
                  }
                >
                  <SelectTrigger className="w-full h-10 rounded-xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-violet-500/20 text-sm font-medium">
                    <SelectValue placeholder="Seçim yapın" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="static">Sabit Mesaj Göster</SelectItem>
                    <SelectItem value="smart" disabled={!canUseSmartFallback}>
                      Akıllı Yönlendirme (AI) {!canUseSmartFallback && '(Pro)'}
                    </SelectItem>
                    <SelectItem value="escalate" disabled={!canUseEscalateFallback}>
                      İnsan Desteğine Aktar {!canUseEscalateFallback && '(Ent)'}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {!canCustomizeThresholds && (
              <div className="flex items-center gap-3 p-4 rounded-xl bg-violet-50 text-violet-700 border border-violet-100">
                <Sparkles className="w-5 h-5 flex-shrink-0" />
                <div className="flex-1">
                  <p className="text-xs font-bold">Pro Özelliği</p>
                  <p className="text-[10px] opacity-80">
                    Eşikleri özelleştirmek için planınızı yükseltin.
                  </p>
                </div>
              </div>
            )}
          </div>
        )}

        {/* TAB: MESSAGES */}
        {activeTab === 'messages' && (
          <div className="space-y-8 animate-in fade-in slide-in-from-bottom-2 duration-300">
            <div className="space-y-4">
              <div className="flex items-start justify-between gap-4">
                <div className="space-y-1">
                  <label className="flex items-center gap-2 text-sm font-bold text-slate-900">
                    <FileText className="w-4 h-4 text-slate-400" />
                    Bilgi Bulunamadı Mesajı
                  </label>
                  <p className="text-xs text-slate-500">
                    Sadece "Sabit Mesaj" modu seçiliyse kullanılır.
                  </p>
                </div>
                {thresholdConfig.fallback_mode === 'static' && (
                  <span className="text-[10px] font-bold bg-green-100 text-green-700 px-2 py-1 rounded-full border border-green-200">
                    Aktif
                  </span>
                )}
              </div>

              <div className="relative group">
                <Textarea
                  className="min-h-[120px] w-full rounded-2xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-violet-500/20 transition-all text-sm leading-relaxed p-4 resize-none"
                  value={fallbackMessages?.no_info_found || ''}
                  onChange={(e) => updateFallbackMessage('no_info_found', e.target.value)}
                  placeholder="Üzgünüm, bu konuda bilgim yok."
                  disabled={!canCustomizeMessages}
                />
                <div className="absolute bottom-4 right-4 text-[10px] font-medium text-slate-400">
                  {(fallbackMessages?.no_info_found || '').length} karakter
                </div>
              </div>
            </div>

            <div className="space-y-4 pt-4 border-t border-slate-100">
              <div className="space-y-1">
                <label className="flex items-center gap-2 text-sm font-bold text-slate-900">
                  <AlertTriangle className="w-4 h-4 text-slate-400" />
                  Genel Hata Mesajı
                </label>
                <p className="text-xs text-slate-500">Teknik bir sorun oluştuğunda gösterilir.</p>
              </div>
              <Textarea
                className="min-h-[100px] w-full rounded-2xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-violet-500/20 transition-all text-sm leading-relaxed p-4 resize-none"
                value={fallbackMessages?.error_message || ''}
                onChange={(e) => updateFallbackMessage('error_message', e.target.value)}
                placeholder="Bir hata oluştu, lütfen tekrar deneyin."
                disabled={!canCustomizeMessages}
              />
            </div>
          </div>
        )}

        {/* TAB: RESTRICTIONS */}
        {activeTab === 'restrictions' && (
          <div className="space-y-6 animate-in fade-in slide-in-from-bottom-2 duration-300">
            <div className="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-200/60">
              <div>
                <h4 className="text-sm font-bold text-slate-900">Kısıtlamalar</h4>
                <p className="text-xs text-slate-500 mt-1">
                  Belirli konuları yasaklayın veya sadece belirli konulara izin verin.
                </p>
              </div>
              <Switch
                checked={!!topicRestrictions}
                onCheckedChange={toggleTopicRestrictions}
                disabled={!canManageTopics}
              />
            </div>

            {!!topicRestrictions && (
              <div className="space-y-8 pt-4">
                {/* Blocked Topics */}
                <div className="space-y-3">
                  <label className="flex items-center justify-between text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">
                    <span>Yasaklı Konular (Blacklist)</span>
                    <span className="text-rose-500">Asla Cevaplamaz</span>
                  </label>

                  <div className="flex gap-2">
                    <Input
                      placeholder="Örn: Siyaset, Rakip..."
                      className="flex-1 h-11 rounded-xl bg-white border-slate-200 focus:ring-2 focus:ring-rose-500/20"
                      value={newBlockedTopic}
                      onChange={(e) => setNewBlockedTopic(e.target.value)}
                      onKeyDown={(e) =>
                        e.key === 'Enter' && (e.preventDefault(), addTopic('blocked'))
                      }
                    />
                    <Button
                      onClick={() => addTopic('blocked')}
                      disabled={!newBlockedTopic.trim()}
                      className="h-11 w-11 rounded-xl bg-slate-900 text-white hover:bg-slate-800 p-0 flex items-center justify-center shrink-0"
                    >
                      <Plus className="w-5 h-5" />
                    </Button>
                  </div>

                  <div className="min-h-[60px] p-2">
                    {!topicRestrictions.blocked_topics?.length ? (
                      <p className="text-xs text-slate-400 italic pl-1">Henüz yasaklı konu yok.</p>
                    ) : (
                      <div className="flex flex-wrap gap-2">
                        {topicRestrictions.blocked_topics.map((t, i) => (
                          <span
                            key={i}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-rose-50 border border-rose-100 text-rose-700 text-xs font-semibold"
                          >
                            {t}
                            <button
                              onClick={() => removeTopic('blocked', i)}
                              className="hover:text-rose-900 transition-colors"
                            >
                              <X className="w-3.5 h-3.5" />
                            </button>
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                <div className="relative">
                  <div className="absolute inset-0 flex items-center">
                    <div className="w-full border-t border-slate-100"></div>
                  </div>
                  <div className="relative flex justify-center">
                    <span className="bg-white px-2 text-[10px] text-slate-400 font-bold uppercase tracking-widest">
                      VE / VEYA
                    </span>
                  </div>
                </div>

                {/* Allowed Topics */}
                <div className="space-y-3">
                  <label className="flex items-center justify-between text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">
                    <span>İzin Verilen Konular (Whitelist)</span>
                    <span className="text-emerald-500">Sadece Bunları Cevaplar</span>
                  </label>

                  <div className="flex gap-2">
                    <Input
                      placeholder="Örn: Ürünler, Destek..."
                      className="flex-1 h-11 rounded-xl bg-white border-slate-200 focus:ring-2 focus:ring-emerald-500/20"
                      value={newAllowedTopic}
                      onChange={(e) => setNewAllowedTopic(e.target.value)}
                      onKeyDown={(e) =>
                        e.key === 'Enter' && (e.preventDefault(), addTopic('allowed'))
                      }
                    />
                    <Button
                      onClick={() => addTopic('allowed')}
                      disabled={!newAllowedTopic.trim()}
                      className="h-11 w-11 rounded-xl bg-slate-100 text-slate-600 hover:bg-slate-200 p-0 flex items-center justify-center shrink-0 border border-slate-200"
                    >
                      <Plus className="w-5 h-5" />
                    </Button>
                  </div>

                  <div className="min-h-[60px] p-2">
                    {!topicRestrictions.allowed_topics?.length ? (
                      <p className="text-xs text-slate-400 italic pl-1">
                        Tüm konulara izin veriliyor.
                      </p>
                    ) : (
                      <div className="flex flex-wrap gap-2">
                        {topicRestrictions.allowed_topics.map((t, i) => (
                          <span
                            key={i}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-50 border border-emerald-100 text-emerald-700 text-xs font-semibold"
                          >
                            {t}
                            <button
                              onClick={() => removeTopic('allowed', i)}
                              className="hover:text-emerald-900 transition-colors"
                            >
                              <X className="w-3.5 h-3.5" />
                            </button>
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                  <p className="text-[10px] text-slate-400 pl-1">
                    * Eğer buraya konu eklerseniz, bot <strong>sadece</strong> bu konular hakkında
                    konuşur.
                  </p>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
