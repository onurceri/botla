import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Shield, MessageSquareWarning, Ban, Plus, X } from 'lucide-react'

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

interface GuardrailsSettingsProps {
  confidenceThreshold: number
  setConfidenceThreshold: (v: number) => void
  fallbackMessages: FallbackMessages | null
  setFallbackMessages: (v: FallbackMessages | null) => void
  topicRestrictions: TopicConfig | null
  setTopicRestrictions: (v: TopicConfig | null) => void
}

export default function GuardrailsSettings({
  confidenceThreshold,
  setConfidenceThreshold,
  fallbackMessages,
  setFallbackMessages,
  topicRestrictions,
  setTopicRestrictions
}: GuardrailsSettingsProps) {
  
  const [newAllowedTopic, setNewAllowedTopic] = useState('')
  const [newBlockedTopic, setNewBlockedTopic] = useState('')

  const handleConfidenceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setConfidenceThreshold(parseFloat(e.target.value))
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
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="w-5 h-5 text-primary" />
            <CardTitle>Güvenlik ve Sınırlar (Guardrails)</CardTitle>
          </div>
          <CardDescription>
            Botun yanıt verme davranışını ve sınırlarını belirleyin.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-8">
          
          {/* Confidence Threshold */}
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <label className="text-sm font-medium">Güven Eşiği (Confidence Threshold)</label>
              <span className="text-sm font-mono bg-muted px-2 py-1 rounded">{confidenceThreshold}</span>
            </div>
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              value={confidenceThreshold}
              onChange={handleConfidenceChange}
              className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-primary"
            />
            <p className="text-xs text-muted-foreground">
              Botun yanıt vermesi için gereken minimum RAG skoru. Yüksek değerler daha az ama daha kesin yanıtlar sağlar.
            </p>
          </div>

          <div className="border-t border-border" />

          {/* Fallback Messages */}
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <MessageSquareWarning className="w-4 h-4 text-primary" />
              <h3 className="font-medium">Yedek Mesajlar (Fallback Messages)</h3>
            </div>
            
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <label className="text-xs font-medium uppercase text-muted-foreground">Bilgi Bulunamadı Mesajı</label>
                <Textarea
                  placeholder="Üzgünüm, bu konuda bilgim yok."
                  value={fallbackMessages?.no_info_found || ''}
                  onChange={(e) => updateFallbackMessage('no_info_found', e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium uppercase text-muted-foreground">Hata Mesajı</label>
                <Textarea
                  placeholder="Bir hata oluştu, lütfen tekrar deneyin."
                  value={fallbackMessages?.error_message || ''}
                  onChange={(e) => updateFallbackMessage('error_message', e.target.value)}
                />
              </div>
            </div>
          </div>

          <div className="border-t border-border" />

          {/* Topic Restrictions */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Ban className="w-4 h-4 text-primary" />
                <h3 className="font-medium">Konu Kısıtlamaları</h3>
              </div>
              <Switch
                checked={!!topicRestrictions}
                onCheckedChange={toggleTopicRestrictions}
              />
            </div>

            {topicRestrictions && (
              <div className="space-y-6 animate-in slide-in-from-top-2 duration-200">
                {/* Blocked Topics */}
                <div className="space-y-2">
                  <label className="text-xs font-medium uppercase text-muted-foreground">Yasaklı Konular</label>
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
                  <div className="flex flex-wrap gap-2 min-h-[32px]">
                    {topicRestrictions.blocked_topics?.map((topic, i) => (
                      <Badge key={i} variant="destructive" className="gap-1 pr-1">
                        {topic}
                        <button type="button" onClick={() => removeTopic('blocked', i)} className="hover:bg-red-600 rounded-full p-0.5">
                          <X className="w-3 h-3" />
                        </button>
                      </Badge>
                    ))}
                  </div>
                </div>

                {/* Allowed Topics (Whitelist) */}
                <div className="space-y-2">
                  <label className="text-xs font-medium uppercase text-muted-foreground">İzin Verilen Konular (Whitelist - Opsiyonel)</label>
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
                  <div className="flex flex-wrap gap-2 min-h-[32px]">
                    {topicRestrictions.allowed_topics?.map((topic, i) => (
                      <Badge key={i} variant="secondary" className="gap-1 pr-1">
                        {topic}
                        <button type="button" onClick={() => removeTopic('allowed', i)} className="hover:bg-secondary-foreground/20 rounded-full p-0.5">
                          <X className="w-3 h-3" />
                        </button>
                      </Badge>
                    ))}
                  </div>
                  <p className="text-[10px] text-muted-foreground">
                    Eğer izin verilen konular belirtilirse, bot SADECE bu konular hakkında konuşur.
                  </p>
                </div>

                {/* Blocked Message */}
                <div className="space-y-2">
                  <label className="text-xs font-medium uppercase text-muted-foreground">Engelleme Mesajı</label>
                  <Input
                    placeholder="Üzgünüm, bu konuda konuşamıyorum."
                    value={topicRestrictions.blocked_message || ''}
                    onChange={(e) => setTopicRestrictions({...topicRestrictions, blocked_message: e.target.value})}
                  />
                </div>
              </div>
            )}
          </div>

        </CardContent>
      </Card>
    </div>
  )
}
