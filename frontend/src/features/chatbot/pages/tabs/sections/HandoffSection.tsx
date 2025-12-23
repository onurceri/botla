import { Headphones, Mail, CheckCircle2, AlertCircle, Info, Lock } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { useChatbotContext } from '../../../context/ChatbotContext'
import { HandoffConfig } from '../../../hooks/useChatbotForm'

export default function HandoffSection() {
  const { handoffEnabled, setHandoffEnabled, handoffConfig, setHandoffConfig, planConfig } =
    useChatbotContext()

  const canUseHandoff = planConfig?.guardrails?.can_use_escalate_fallback ?? false

  const handleConfigChange = (field: keyof HandoffConfig, value: string) => {
    // @ts-ignore
    setHandoffConfig({
      ...(handoffConfig || {}),
      [field]: value,
    })
  }

  return (
    <div className="bg-white rounded-[24px] border border-slate-200/60 shadow-sm overflow-hidden flex flex-col h-full group transition-all hover:shadow-md">
      {/* Header */}
      <div className="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-orange-500/10 text-orange-600 ring-1 ring-orange-500/20 shadow-sm">
            <Headphones className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-bold tracking-tight text-slate-900 uppercase">
              İnsan Desteği
            </h3>
            <p className="text-[11px] text-slate-500 font-medium">
              Botun yetersiz kaldığı durumlarda devreye girer
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          {!canUseHandoff && (
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-100 text-slate-500 border border-slate-200">
              <Lock className="w-3 h-3" />
              <span className="text-[10px] font-bold uppercase tracking-wider">Plan Yükselt</span>
            </div>
          )}
          <Switch
            checked={handoffEnabled}
            onCheckedChange={setHandoffEnabled}
            disabled={!canUseHandoff}
            className="data-[state=checked]:bg-orange-500"
          />
        </div>
      </div>

      <div className="p-6 lg:p-8 space-y-8 flex-1">
        {/* Status State */}
        {!handoffEnabled && (
          <div className="flex flex-col items-center justify-center py-12 text-center rounded-2xl bg-slate-50 border border-dashed border-slate-200">
            <div className="p-4 bg-white rounded-full shadow-sm mb-4">
              <Headphones className="w-8 h-8 text-slate-300" />
            </div>
            <h3 className="text-sm font-bold text-slate-900">İnsan Desteği Kapalı</h3>
            <p className="text-xs text-slate-500 mt-1 max-w-[250px]">
              Kullanıcılar sadece bot ile görüşebilir, bir temsilciye bağlanamazlar.
            </p>
          </div>
        )}

        {handoffEnabled && (
          <div className="space-y-8 animate-in fade-in slide-in-from-bottom-2 duration-300">
            {/* Channels */}
            <div className="space-y-3">
              <label className="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">
                İletişim Kanalı
              </label>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Email - Active */}
                <div className="relative group cursor-pointer">
                  <div className="absolute inset-0 bg-orange-50 rounded-2xl border-2 border-orange-500 opacity-100 transition-all"></div>
                  <div className="relative p-4 flex items-center gap-4">
                    <div className="p-3 bg-white rounded-full text-orange-600 shadow-sm border border-orange-100">
                      <Mail className="w-5 h-5" />
                    </div>
                    <div>
                      <p className="text-sm font-bold text-slate-900">E-posta Bildirimi</p>
                      <p className="text-[11px] text-slate-500 font-medium">
                        Talepler e-posta olarak iletilir.
                      </p>
                    </div>
                    <div className="absolute top-4 right-4 text-orange-500">
                      <CheckCircle2 className="w-4 h-4 fill-orange-100" />
                    </div>
                  </div>
                </div>

                {/* Live Chat - Disabled */}
                <div className="relative group cursor-not-allowed opacity-60 grayscale">
                  <div className="absolute inset-0 bg-slate-50 rounded-2xl border border-slate-200"></div>
                  <div className="relative p-4 flex items-center gap-4">
                    <div className="p-3 bg-white rounded-full text-slate-400 shadow-sm border border-slate-100">
                      <Headphones className="w-5 h-5" />
                    </div>
                    <div>
                      <p className="text-sm font-bold text-slate-900">Canlı Destek</p>
                      <p className="text-[11px] text-slate-500 font-medium">
                        Intercom, Zendesk vb. (Yakında)
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Config Form */}
            <div className="space-y-6 pt-4 border-t border-slate-100">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-3">
                  <label className="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">
                    Alıcı E-posta <span className="text-rose-500">*</span>
                  </label>
                  <Input
                    type="email"
                    placeholder="destek@sirket.com"
                    className="h-11 rounded-xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-orange-500/20 transition-all font-medium text-slate-900"
                    value={handoffConfig?.email_to || ''}
                    onChange={(e) => handleConfigChange('email_to', e.target.value)}
                  />
                </div>
                <div className="space-y-3">
                  <label className="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">
                    Konu Başlığı
                  </label>
                  <Input
                    type="text"
                    placeholder="[Botla] Yeni Destek Talebi"
                    className="h-11 rounded-xl bg-slate-50 border-slate-200 focus:bg-white focus:ring-2 focus:ring-orange-500/20 transition-all font-medium text-slate-900"
                    value={handoffConfig?.email_subject || ''}
                    onChange={(e) => handleConfigChange('email_subject', e.target.value)}
                  />
                </div>
              </div>

              {!handoffConfig?.email_to && (
                <div className="flex items-start gap-3 p-4 rounded-xl bg-rose-50 border border-rose-100 text-rose-700">
                  <AlertCircle className="w-5 h-5 shrink-0" />
                  <p className="text-xs font-medium leading-relaxed">
                    Lütfen bildirimlerin gönderileceği e-posta adresini girin. Aksi takdirde insan
                    desteği çalışmayacaktır.
                  </p>
                </div>
              )}
            </div>

            {/* Info Tip */}
            <div className="flex items-start gap-4 p-4 rounded-xl bg-slate-50 border border-slate-100">
              <div className="p-2 bg-white rounded-lg border border-slate-200 shadow-sm text-slate-400">
                <Info className="w-4 h-4" />
              </div>
              <div className="space-y-1">
                <p className="text-xs font-bold text-slate-900">Nasıl Çalışır?</p>
                <p className="text-[11px] text-slate-500 leading-relaxed">
                  Kullanıcı <strong>"İnsan Desteği"</strong> istediğinde veya bot cevap veremeyip
                  yönlendirme yaptığında, konuşma geçmişi ve kullanıcı bilgileri yukarıdaki adrese
                  e-posta olarak gönderilir.
                </p>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
