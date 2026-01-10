import { useState } from 'react'
import * as Dialog from '@radix-ui/react-dialog'
import { X, Zap, Bot, FileText, Shield, RefreshCw, Settings } from 'lucide-react'
import * as adminApi from '@/api/admin'
import { Button } from '@/components/ui/button'

interface PlanFeatureModalProps {
  isOpen: boolean
  onClose: () => void
  plan: adminApi.AdminPlanDetail
  onSave: (updates: adminApi.UpdatePlanLimitsRequest) => Promise<void>
}

export function PlanFeatureModal({ isOpen, onClose, plan, onSave }: PlanFeatureModalProps) {
  const [updates, setUpdates] = useState<adminApi.UpdatePlanLimitsRequest>({})
  const [isSaving, setIsSaving] = useState(false)

  const handleToggle = (field: keyof adminApi.UpdatePlanLimitsRequest) => {
    const currentValue = updates[field] ?? plan.limits[field as keyof adminApi.PlanLimitsDetail]
    setUpdates((prev) => ({
      ...prev,
      [field]: !currentValue,
    }))
  }

  const handleNumberChange = (field: keyof adminApi.UpdatePlanLimitsRequest, value: string) => {
    const numValue = parseInt(value, 10)
    if (!isNaN(numValue)) {
      setUpdates((prev) => ({
        ...prev,
        [field]: numValue,
      }))
    }
  }

  const handleSave = async () => {
    setIsSaving(true)
    try {
      await onSave(updates)
    } finally {
      setIsSaving(false)
    }
  }

  const getValue = <T extends keyof adminApi.PlanLimitsDetail>(field: T): adminApi.PlanLimitsDetail[T] => {
    if (field in updates) {
      return updates[field as keyof adminApi.UpdatePlanLimitsRequest] as adminApi.PlanLimitsDetail[T]
    }
    return plan.limits[field]
  }

  return (
    <Dialog.Root open={isOpen} onOpenChange={onClose}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50" />
        <Dialog.Content className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-2xl max-h-[85vh] bg-background rounded-xl shadow-2xl z-50 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b">
            <div>
              <Dialog.Title className="text-lg font-bold capitalize flex items-center gap-2">
                <Settings className="w-5 h-5" />
                {plan.plan.code} Planı Özellikleri
              </Dialog.Title>
              <Dialog.Description className="text-sm text-muted-foreground">
                Plan özelliklerini ve limitlerini düzenle
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <Button variant="ghost" size="icon" className="rounded-full">
                <X className="w-4 h-4" />
              </Button>
            </Dialog.Close>
          </div>

          {/* Content */}
          <div className="overflow-y-auto max-h-[calc(85vh-140px)] px-6 py-4 space-y-6">
            {/* Core Limits */}
            <section>
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                <Zap className="w-4 h-4" />
                Temel Limitler
              </h3>
              <div className="grid grid-cols-3 gap-4">
                <FeatureInput
                  label="Max Chatbot"
                  value={getValue('max_chatbots')}
                  onChange={(v) => handleNumberChange('max_chatbots', v)}
                />
                <FeatureInput
                  label="Aylık İndex"
                  value={getValue('max_monthly_ingestions')}
                  onChange={(v) => handleNumberChange('max_monthly_ingestions', v)}
                />
                <FeatureInput
                  label="Embedding Token"
                  value={getValue('max_monthly_embedding_tokens')}
                  onChange={(v) => handleNumberChange('max_monthly_embedding_tokens', v)}
                />
              </div>
            </section>

            {/* Files */}
            <section>
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                <FileText className="w-4 h-4" />
                Dosya & Depolama
              </h3>
              <div className="grid grid-cols-3 gap-4">
                <FeatureInput
                  label="Max Dosya Boyutu (MB)"
                  value={getValue('files_max_size_mb')}
                  onChange={(v) => handleNumberChange('files_max_size_mb', v)}
                />
                <FeatureInput
                  label="Max Dosya/Bot"
                  value={getValue('files_max_files_per_bot')}
                  onChange={(v) => handleNumberChange('files_max_files_per_bot', v)}
                />
                <FeatureInput
                  label="Toplam Depolama (MB)"
                  value={getValue('files_total_storage_mb')}
                  onChange={(v) => handleNumberChange('files_total_storage_mb', v)}
                />
              </div>
            </section>

            {/* Guardrails Features */}
            <section>
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                <Shield className="w-4 h-4" />
                Güvenlik & Kurallar
              </h3>
              <div className="grid grid-cols-2 gap-3">
                <FeatureToggle
                  label="Eşikleri Özelleştir"
                  enabled={getValue('guardrails_can_customize_thresholds')}
                  onToggle={() => handleToggle('guardrails_can_customize_thresholds')}
                />
                <FeatureToggle
                  label="Akıllı Fallback"
                  enabled={getValue('guardrails_can_use_smart_fallback')}
                  onToggle={() => handleToggle('guardrails_can_use_smart_fallback')}
                />
                <FeatureToggle
                  label="Eskalasyon Fallback"
                  enabled={getValue('guardrails_can_use_escalate_fallback')}
                  onToggle={() => handleToggle('guardrails_can_use_escalate_fallback')}
                />
                <FeatureToggle
                  label="Konuları Yönet"
                  enabled={getValue('guardrails_can_manage_topics')}
                  onToggle={() => handleToggle('guardrails_can_manage_topics')}
                />
                <FeatureToggle
                  label="Mesajları Özelleştir"
                  enabled={getValue('guardrails_can_customize_messages')}
                  onToggle={() => handleToggle('guardrails_can_customize_messages')}
                />
              </div>
            </section>

            {/* Branding Features */}
            <section>
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                <Bot className="w-4 h-4" />
                Markalama
              </h3>
              <div className="grid grid-cols-2 gap-3">
                <FeatureToggle
                  label="Markalamayı Gizle"
                  enabled={getValue('branding_can_hide_branding')}
                  onToggle={() => handleToggle('branding_can_hide_branding')}
                />
                <FeatureToggle
                  label="Özel Markalama"
                  enabled={getValue('branding_can_custom_branding')}
                  onToggle={() => handleToggle('branding_can_custom_branding')}
                />
              </div>
            </section>

            {/* Refresh Feature */}
            <section>
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                <RefreshCw className="w-4 h-4" />
                Yenileme
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <FeatureToggle
                  label="Yenileme Aktif"
                  enabled={getValue('refresh_enabled')}
                  onToggle={() => handleToggle('refresh_enabled')}
                />
                <FeatureInput
                  label="Max Aylık Yenileme"
                  value={getValue('refresh_max_monthly')}
                  onChange={(v) => handleNumberChange('refresh_max_monthly', v)}
                />
              </div>
            </section>

            {/* Security */}
            <section>
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                <Shield className="w-4 h-4" />
                Güvenlik
              </h3>
              <FeatureToggle
                label="Güvenli Embed"
                enabled={getValue('security_secure_embed_enabled')}
                onToggle={() => handleToggle('security_secure_embed_enabled')}
              />
            </section>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-3 px-6 py-4 border-t bg-muted/30">
            <Dialog.Close asChild>
              <Button variant="outline">İptal</Button>
            </Dialog.Close>
            <Button onClick={handleSave} disabled={isSaving}>
              {isSaving ? 'Kaydediliyor...' : 'Kaydet'}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}

interface FeatureToggleProps {
  label: string
  enabled: boolean
  onToggle: () => void
}

function FeatureToggle({ label, enabled, onToggle }: FeatureToggleProps) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className="flex items-center justify-between p-3 rounded-lg border hover:bg-muted/50 transition-colors cursor-pointer"
    >
      <span className="text-sm font-medium">{label}</span>
      <div
        className={`relative w-11 h-6 rounded-full transition-colors ${
          enabled ? 'bg-primary' : 'bg-muted'
        }`}
      >
        <div
          className={`absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow-sm transition-transform ${
            enabled ? 'translate-x-5' : ''
          }`}
        />
      </div>
    </button>
  )
}

interface FeatureInputProps {
  label: string
  value: number
  onChange: (value: string) => void
}

function FeatureInput({ label, value, onChange }: FeatureInputProps) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full px-3 py-2 text-sm border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
      />
    </div>
  )
}
