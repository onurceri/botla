import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Settings, CreditCard, FileText, Bot, Zap, Globe, Shield, RefreshCw, MessageSquare, Clock, ChevronRight } from 'lucide-react'
import * as adminApi from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/components/ui/toast'

const formatPrice = (price: number, currency: string) => {
  return new Intl.NumberFormat('tr-TR', {
    style: 'currency',
    currency: currency,
  }).format(price)
}

const formatNumber = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

export function AdminPlansPage() {
  const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null)
  const [editingLimits, setEditingLimits] = useState<adminApi.UpdatePlanLimitsRequest>({})

  const queryClient = useQueryClient()
  const { toast } = useToast()

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin', 'plans'],
    queryFn: () => adminApi.listPlans(),
  })

  const { data: planDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['admin', 'plans', selectedPlanId],
    queryFn: () => selectedPlanId ? adminApi.getPlan(selectedPlanId) : null,
    enabled: !!selectedPlanId,
  })

  const invalidateCacheMutation = useMutation({
    mutationFn: (planId: string) => adminApi.invalidatePlanCache(planId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'plans'] })
      toast('Önbellek temizlendi.', 'success')
    },
    onError: () => {
      toast('Önbellek temizlenirken hata oluştu.', 'error')
    },
  })

  const saveLimitsMutation = useMutation({
    mutationFn: async (updates: adminApi.UpdatePlanLimitsRequest) => {
      if (!selectedPlanId) throw new Error('No plan selected')
      await adminApi.updatePlanLimits(selectedPlanId, updates)
    },
    onSuccess: () => {
      setEditingLimits({})
      queryClient.invalidateQueries({ queryKey: ['admin', 'plans'] })
      queryClient.invalidateQueries({ queryKey: ['admin', 'plans', selectedPlanId] })
      toast('Plan limitleri güncellendi.', 'success')
    },
    onError: () => {
      toast('Limitler güncellenirken hata oluştu.', 'error')
    },
  })

  const handleSelectPlan = (planId: string) => {
    setSelectedPlanId(planId)
    setEditingLimits({})
  }

  const handleToggle = (field: keyof adminApi.UpdatePlanLimitsRequest) => {
    if (!planDetail) return
    const currentValue = editingLimits[field] ?? planDetail.limits[field as keyof adminApi.PlanLimitsDetail]
    const newValue = !currentValue
    setEditingLimits((prev) => ({
      ...prev,
      [field]: newValue,
    }))
  }

  const handleNumberChange = (field: keyof adminApi.UpdatePlanLimitsRequest, value: string) => {
    const numValue = parseInt(value, 10)
    if (!isNaN(numValue)) {
      setEditingLimits((prev) => ({
        ...prev,
        [field]: numValue,
      }))
    }
  }

  const handleSave = () => {
    if (Object.keys(editingLimits).length > 0) {
      saveLimitsMutation.mutate(editingLimits)
    }
  }

  const handleCancel = () => {
    setEditingLimits({})
  }

  const getValue = <T extends keyof adminApi.PlanLimitsDetail>(field: T): adminApi.PlanLimitsDetail[T] => {
    if (field in editingLimits) {
      return editingLimits[field as keyof adminApi.UpdatePlanLimitsRequest] as adminApi.PlanLimitsDetail[T]
    }
    if (planDetail) {
      return planDetail.limits[field]
    }
    return 0 as adminApi.PlanLimitsDetail[T]
  }

  const plans = data?.plans ?? []
  const total = data?.total ?? 0
  const hasChanges = Object.keys(editingLimits).length > 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Planlar</h1>
        <p className="text-muted-foreground">
          Abonelik planlarını ve özelliklerini yönet. Toplam: {total}
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Panel - Plan List */}
        <div className="lg:col-span-1 space-y-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <CreditCard className="w-5 h-5" />
            Planlar
          </h2>
          
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <Card key={i} className="animate-pulse">
                  <CardHeader className="pb-3">
                    <div className="h-5 bg-muted rounded w-1/3" />
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <div className="h-4 bg-muted rounded w-2/3" />
                    <div className="h-4 bg-muted rounded w-1/2" />
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : error ? (
            <Card className="border-destructive">
              <CardContent className="py-8 text-center text-destructive">
                Hata: {(error as Error).message}
              </CardContent>
            </Card>
          ) : plans.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                Plan bulunamadı.
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-3">
              {plans.map((plan) => (
                <PlanListItem
                  key={plan.id}
                  plan={plan}
                  isSelected={selectedPlanId === plan.id}
                  onClick={() => handleSelectPlan(plan.id)}
                />
              ))}
            </div>
          )}
        </div>

        {/* Right Panel - Plan Details */}
        <div className="lg:col-span-2">
          {!selectedPlanId ? (
            <Card className="h-full flex items-center justify-center min-h-[400px]">
              <div className="text-center text-muted-foreground">
                <Settings className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p className="text-lg font-medium">Plan seçin</p>
                <p className="text-sm">Düzenlemek için bir plan seçin</p>
              </div>
            </Card>
          ) : detailLoading ? (
            <Card className="h-full flex items-center justify-center min-h-[400px]">
              <div className="animate-pulse space-y-4 w-full max-w-md">
                <div className="h-6 bg-muted rounded w-1/3" />
                <div className="space-y-3">
                  <div className="h-4 bg-muted rounded w-full" />
                  <div className="h-4 bg-muted rounded w-2/3" />
                  <div className="h-4 bg-muted rounded w-1/2" />
                </div>
              </div>
            </Card>
          ) : planDetail ? (
            <PlanDetailPanel
              plan={planDetail}
              getValue={getValue}
              hasChanges={hasChanges}
              isSaving={saveLimitsMutation.isPending}
              isInvalidating={invalidateCacheMutation.isPending}
              onToggle={handleToggle}
              onNumberChange={handleNumberChange}
              onSave={handleSave}
              onCancel={handleCancel}
              onInvalidateCache={() => invalidateCacheMutation.mutate(selectedPlanId)}
            />
          ) : null}
        </div>
      </div>
    </div>
  )
}

interface PlanListItemProps {
  plan: adminApi.AdminPlanSummary
  isSelected: boolean
  onClick: () => void
}

function PlanListItem({ plan, isSelected, onClick }: PlanListItemProps) {
  const isPopular = plan.code === 'pro'

  return (
    <Card
      className={`cursor-pointer transition-all duration-200 hover:shadow-md ${
        isSelected
          ? 'ring-2 ring-primary border-primary'
          : 'hover:border-primary/50'
      }`}
      onClick={onClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CardTitle className="text-lg font-bold capitalize">{plan.code}</CardTitle>
            {isPopular && (
              <Badge variant="default" className="text-xs">
                Popüler
              </Badge>
            )}
          </div>
          <span className={`px-2 py-1 text-xs font-medium rounded-full ${
            plan.status === 'active'
              ? 'bg-green-500/10 text-green-600'
              : 'bg-muted text-muted-foreground'
          }`}>
            {plan.status === 'active' ? 'Aktif' : plan.status}
          </span>
        </div>
        <div className="mt-1">
          <span className="text-xl font-bold">{formatPrice(plan.price, plan.currency)}</span>
          <span className="text-muted-foreground text-sm ml-1">
            /{plan.billing_cycle === 'monthly' ? 'ay' : plan.billing_cycle}
          </span>
        </div>
      </CardHeader>

      <CardContent className="pt-0">
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <Bot className="w-3.5 h-3.5" />
            <span>Chatbot:</span>
            <span className="font-medium">{plan.max_chatbots}</span>
          </div>
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <Zap className="w-3.5 h-3.5" />
            <span>Index:</span>
            <span className="font-medium">{formatNumber(plan.max_monthly_ingestions)}</span>
          </div>
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <FileText className="w-3.5 h-3.5" />
            <span>Dosya:</span>
            <span className="font-medium">{plan.files_max_size_mb} MB</span>
          </div>
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <CreditCard className="w-3.5 h-3.5" />
            <span>Modeller:</span>
            <span className="font-medium">{plan.chat_allowed_models_count}</span>
          </div>
        </div>

        {plan.trial_days > 0 && (
          <div className="mt-3 text-xs text-muted-foreground bg-muted/50 rounded-lg px-3 py-2">
            {plan.trial_days} gün ücretsiz deneme
          </div>
        )}

        <div className="mt-3 flex items-center gap-1 text-xs text-muted-foreground">
          <ChevronRight className={`w-4 h-4 transition-transform ${isSelected ? 'rotate-90' : ''}`} />
          <span>{isSelected ? 'Seçili' : 'Detayları görüntüle'}</span>
        </div>
      </CardContent>
    </Card>
  )
}

interface PlanDetailPanelProps {
  plan: adminApi.AdminPlanDetail
  getValue: <T extends keyof adminApi.PlanLimitsDetail>(field: T) => adminApi.PlanLimitsDetail[T]
  hasChanges: boolean
  isSaving: boolean
  isInvalidating: boolean
  onToggle: (field: keyof adminApi.UpdatePlanLimitsRequest) => void
  onNumberChange: (field: keyof adminApi.UpdatePlanLimitsRequest, value: string) => void
  onSave: () => void
  onCancel: () => void
  onInvalidateCache: () => void
}

function PlanDetailPanel({
  plan,
  getValue,
  hasChanges,
  isSaving,
  isInvalidating,
  onToggle,
  onNumberChange,
  onSave,
  onCancel,
  onInvalidateCache,
}: PlanDetailPanelProps) {
  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <Settings className="w-5 h-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-xl capitalize flex items-center gap-2">
                {plan.plan.code} Planı
                <Badge variant={plan.plan.status === 'active' ? 'default' : 'secondary'}>
                  {plan.plan.status === 'active' ? 'Aktif' : plan.plan.status}
                </Badge>
              </CardTitle>
              <p className="text-sm text-muted-foreground">
                {formatPrice(plan.plan.price, plan.plan.currency)}/{plan.plan.billing_cycle === 'monthly' ? 'ay' : plan.plan.billing_cycle}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={onInvalidateCache}
              disabled={isInvalidating}
            >
              <Zap className="w-4 h-4 mr-1" />
              Önbelleği Temizle
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <Tabs defaultValue="core" className="w-full">
          <TabsList className="grid w-full grid-cols-5 lg:grid-cols-9 mb-4">
            <TabsTrigger value="core" className="text-xs">
              <Zap className="w-3 h-3 mr-1" />
              Core
            </TabsTrigger>
            <TabsTrigger value="files" className="text-xs">
              <FileText className="w-3 h-3 mr-1" />
              Dosya
            </TabsTrigger>
            <TabsTrigger value="scraping" className="text-xs">
              <Globe className="w-3 h-3 mr-1" />
              Scraping
            </TabsTrigger>
            <TabsTrigger value="chat" className="text-xs">
              <MessageSquare className="w-3 h-3 mr-1" />
              Chat
            </TabsTrigger>
            <TabsTrigger value="guardrails" className="text-xs">
              <Shield className="w-3 h-3 mr-1" />
              Guardrails
            </TabsTrigger>
            <TabsTrigger value="branding" className="text-xs">
              <Bot className="w-3 h-3 mr-1" />
              Markalama
            </TabsTrigger>
            <TabsTrigger value="refresh" className="text-xs">
              <RefreshCw className="w-3 h-3 mr-1" />
              Yenileme
            </TabsTrigger>
            <TabsTrigger value="security" className="text-xs">
              <Shield className="w-3 h-3 mr-1" />
              Güvenlik
            </TabsTrigger>
            <TabsTrigger value="rate-limits" className="text-xs">
              <Clock className="w-3 h-3 mr-1" />
              Rate Limits
            </TabsTrigger>
          </TabsList>

          {/* Core Limits */}
          <TabsContent value="core" className="space-y-4">
            <Section title="Temel Limitler" description="Planın temel kullanım limitleri">
              <div className="grid grid-cols-2 gap-4">
                <NumberField
                  label="Max Chatbot"
                  value={getValue('max_chatbots')}
                  onChange={(v) => onNumberChange('max_chatbots', v)}
                />
                <NumberField
                  label="Aylık Index"
                  value={getValue('max_monthly_ingestions')}
                  onChange={(v) => onNumberChange('max_monthly_ingestions', v)}
                />
                <NumberField
                  label="Aylık Embedding Token"
                  value={getValue('max_monthly_embedding_tokens')}
                  onChange={(v) => onNumberChange('max_monthly_embedding_tokens', v)}
                />
                <NumberField
                  label="Re-add Cooldown (dk)"
                  value={getValue('min_readd_cooldown_minutes')}
                  onChange={(v) => onNumberChange('min_readd_cooldown_minutes', v)}
                />
              </div>
            </Section>
          </TabsContent>

          {/* Files */}
          <TabsContent value="files" className="space-y-4">
            <Section title="Dosya & Depolama" description="Dosya yükleme ve depolama limitleri">
              <div className="grid grid-cols-2 gap-4">
                <NumberField
                  label="Max Dosya Boyutu (MB)"
                  value={getValue('files_max_size_mb')}
                  onChange={(v) => onNumberChange('files_max_size_mb', v)}
                />
                <NumberField
                  label="Max Dosya/Bot"
                  value={getValue('files_max_files_per_bot')}
                  onChange={(v) => onNumberChange('files_max_files_per_bot', v)}
                />
                <NumberField
                  label="Toplam Dosya Sayısı"
                  value={getValue('files_max_files_total')}
                  onChange={(v) => onNumberChange('files_max_files_total', v)}
                />
                <NumberField
                  label="Toplam Depolama (MB)"
                  value={getValue('files_total_storage_mb')}
                  onChange={(v) => onNumberChange('files_total_storage_mb', v)}
                />
                <NumberField
                  label="Max Metin Uzunluğu"
                  value={getValue('files_max_text_length')}
                  onChange={(v) => onNumberChange('files_max_text_length', v)}
                />
              </div>
            </Section>
          </TabsContent>

          {/* Scraping */}
          <TabsContent value="scraping" className="space-y-4">
            <Section title="Web Scraping" description="Web sitesi tarama ve içerik çıkarma ayarları">
              <div className="space-y-4">
                <ToggleField
                  label="Dinamik Scraping"
                  description="Dinamik içerikli siteleri tarayabilme"
                  enabled={getValue('scraping_dynamic_enabled')}
                  onToggle={() => onToggle('scraping_dynamic_enabled')}
                />
                <div className="grid grid-cols-2 gap-4">
                  <NumberField
                    label="Max URL/Bot"
                    value={getValue('scraping_max_urls_per_bot')}
                    onChange={(v) => onNumberChange('scraping_max_urls_per_bot', v)}
                  />
                  <NumberField
                    label="Max Sayfa/Crawl"
                    value={getValue('scraping_max_pages_per_crawl')}
                    onChange={(v) => onNumberChange('scraping_max_pages_per_crawl', v)}
                  />
                </div>
              </div>
            </Section>
          </TabsContent>

          {/* Chat */}
          <TabsContent value="chat" className="space-y-4">
            <Section title="Chatbot Ayarları" description="Chatbot yanıt ve model yapılandırması">
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <NumberField
                    label="Aylık Token Limiti"
                    value={getValue('chat_max_monthly_tokens')}
                    onChange={(v) => onNumberChange('chat_max_monthly_tokens', v)}
                  />
                  <NumberField
                    label="Varsayılan Model"
                    value={getValue('chat_default_model')}
                    onChange={(v) => onNumberChange('chat_default_model', v)}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <NumberField
                    label="RAG Top K"
                    value={getValue('chat_rag_top_k')}
                    onChange={(v) => onNumberChange('chat_rag_top_k', v)}
                  />
                  <NumberField
                    label="Max Context Token"
                    value={getValue('chat_rag_max_context_tokens')}
                    onChange={(v) => onNumberChange('chat_rag_max_context_tokens', v)}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <NumberField
                    label="Max Önerilen Soru"
                    value={getValue('chat_max_suggested_questions')}
                    onChange={(v) => onNumberChange('chat_max_suggested_questions', v)}
                  />
                  <NumberField
                    label="Max Manuel Soru"
                    value={getValue('chat_max_manual_questions')}
                    onChange={(v) => onNumberChange('chat_max_manual_questions', v)}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <NumberField
                    label="Min Yanıt Token"
                    value={getValue('chat_min_response_token_limit')}
                    onChange={(v) => onNumberChange('chat_min_response_token_limit', v)}
                  />
                  <NumberField
                    label="Max Yanıt Token"
                    value={getValue('chat_max_response_token_limit')}
                    onChange={(v) => onNumberChange('chat_max_response_token_limit', v)}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium text-muted-foreground mb-2 block">
                    İzin Verilen Modeller
                  </label>
                  <div className="p-3 bg-muted/50 rounded-lg text-sm font-mono">
                    {getValue('chat_allowed_models').join(', ') || 'Model yok'}
                  </div>
                </div>
              </div>
            </Section>
          </TabsContent>

          {/* Guardrails */}
          <TabsContent value="guardrails" className="space-y-4">
            <Section title="Guardrails" description="İçerik filtreleme ve güvenlik kuralları">
              <div className="grid grid-cols-1 gap-3">
                <ToggleField
                  label="Eşikleri Özelleştir"
                  description="İçerik filtreleme eşiklerini özelleştirebilme"
                  enabled={getValue('guardrails_can_customize_thresholds')}
                  onToggle={() => onToggle('guardrails_can_customize_thresholds')}
                />
                <ToggleField
                  label="Akıllı Fallback"
                  description="Akıllı fallback yanıtları kullanabilme"
                  enabled={getValue('guardrails_can_use_smart_fallback')}
                  onToggle={() => onToggle('guardrails_can_use_smart_fallback')}
                />
                <ToggleField
                  label="Eskalasyon Fallback"
                  description="Eskalasyon fallback mekanizması"
                  enabled={getValue('guardrails_can_use_escalate_fallback')}
                  onToggle={() => onToggle('guardrails_can_use_escalate_fallback')}
                />
                <ToggleField
                  label="Konuları Yönet"
                  description="Özel konular oluşturabilme ve yönetebilme"
                  enabled={getValue('guardrails_can_manage_topics')}
                  onToggle={() => onToggle('guardrails_can_manage_topics')}
                />
                <ToggleField
                  label="Mesajları Özelleştir"
                  description="Filtreleme mesajlarını özelleştirebilme"
                  enabled={getValue('guardrails_can_customize_messages')}
                  onToggle={() => onToggle('guardrails_can_customize_messages')}
                />
              </div>
            </Section>
          </TabsContent>

          {/* Branding */}
          <TabsContent value="branding" className="space-y-4">
            <Section title="Markalama" description="Chatbot görünümü ve markalama ayarları">
              <div className="grid grid-cols-1 gap-3">
                <ToggleField
                  label="Markalamayı Gizle"
                  description="Botla markalamasını gizleyebilme"
                  enabled={getValue('branding_can_hide_branding')}
                  onToggle={() => onToggle('branding_can_hide_branding')}
                />
                <ToggleField
                  label="Özel Markalama"
                  description="Özel renkler ve logoyla markalama"
                  enabled={getValue('branding_can_custom_branding')}
                  onToggle={() => onToggle('branding_can_custom_branding')}
                />
              </div>
            </Section>
          </TabsContent>

          {/* Refresh */}
          <TabsContent value="refresh" className="space-y-4">
            <Section title="Yenileme" description="Otomatik içerik yenileme ayarları">
              <div className="space-y-4">
                <ToggleField
                  label="Yenileme Aktif"
                  description="Otomatik içerik yenileme özelliği"
                  enabled={getValue('refresh_enabled')}
                  onToggle={() => onToggle('refresh_enabled')}
                />
                <NumberField
                  label="Max Aylık Yenileme"
                  value={getValue('refresh_max_monthly')}
                  onChange={(v) => onNumberChange('refresh_max_monthly', v)}
                />
              </div>
            </Section>
          </TabsContent>

          {/* Security */}
          <TabsContent value="security" className="space-y-4">
            <Section title="Güvenlik" description="Güvenlik ve embed ayarları">
              <ToggleField
                label="Güvenli Embed"
                description="Güvenli embed iframe yapılandırması"
                enabled={getValue('security_secure_embed_enabled')}
                onToggle={() => onToggle('security_secure_embed_enabled')}
              />
            </Section>
          </TabsContent>

          {/* Rate Limits */}
          <TabsContent value="rate-limits" className="space-y-4">
            <Section title="Rate Limits" description="API istek limitleri ve pencereleri">
              <div className="grid grid-cols-2 gap-4">
                <NumberField
                  label="İstek/Dakika"
                  value={getValue('rate_limits_requests_per_minute')}
                  onChange={(v) => onNumberChange('rate_limits_requests_per_minute', v)}
                />
                <NumberField
                  label="Pencere (saniye)"
                  value={getValue('rate_limits_window_seconds')}
                  onChange={(v) => onNumberChange('rate_limits_window_seconds', v)}
                />
                <NumberField
                  label="Chat RPM"
                  value={getValue('rate_limits_chat_rpm')}
                  onChange={(v) => onNumberChange('rate_limits_chat_rpm', v)}
                />
                <NumberField
                  label="Chat Pencere (saniye)"
                  value={getValue('rate_limits_chat_window')}
                  onChange={(v) => onNumberChange('rate_limits_chat_window', v)}
                />
                <NumberField
                  label="Sources RPM"
                  value={getValue('rate_limits_sources_rpm')}
                  onChange={(v) => onNumberChange('rate_limits_sources_rpm', v)}
                />
                <NumberField
                  label="Sources Pencere (saniye)"
                  value={getValue('rate_limits_sources_window')}
                  onChange={(v) => onNumberChange('rate_limits_sources_window', v)}
                />
              </div>
            </Section>
          </TabsContent>
        </Tabs>

        {/* Action Buttons */}
        {hasChanges && (
          <div className="flex items-center justify-end gap-3 mt-6 pt-4 border-t">
            <Button variant="outline" onClick={onCancel} disabled={isSaving}>
              İptal
            </Button>
            <Button onClick={onSave} disabled={isSaving}>
              {isSaving ? 'Kaydediliyor...' : 'Değişiklikleri Kaydet'}
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

interface SectionProps {
  title: string
  description: string
  children: React.ReactNode
}

function Section({ title, description, children }: SectionProps) {
  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-base font-semibold">{title}</h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <div className="bg-muted/30 rounded-lg p-4">
        {children}
      </div>
    </div>
  )
}

interface NumberFieldProps {
  label: string
  value: number | string
  onChange: (value: string) => void
}

function NumberField({ label, value, onChange }: NumberFieldProps) {
  return (
    <div className="space-y-1.5">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <Input
        type="number"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-9"
      />
    </div>
  )
}

interface ToggleFieldProps {
  label: string
  description: string
  enabled: boolean
  onToggle: () => void
}

function ToggleField({ label, description, enabled, onToggle }: ToggleFieldProps) {
  return (
    <div className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-muted/50 transition-colors">
      <div className="space-y-0.5">
        <label className="text-sm font-medium cursor-pointer" onClick={onToggle}>
          {label}
        </label>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      <Switch
        checked={enabled}
        onCheckedChange={onToggle}
      />
    </div>
  )
}