import { Globe, EyeOff } from 'lucide-react'
import { Switch } from '@/components/ui/switch'
import { useChatbotContext } from '../../../context/ChatbotContext'

export default function VisibilitySection() {
  const { isPublic, setIsPublic } = useChatbotContext()

  return (
    <div className="bg-white rounded-[24px] border border-slate-200/60 shadow-sm overflow-hidden">
      <div className="px-6 py-5 flex items-center justify-between gap-4 bg-slate-50/50">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-emerald-500/10 text-emerald-600 ring-1 ring-emerald-500/20 shadow-sm">
            {isPublic ? <Globe className="w-5 h-5" /> : <EyeOff className="w-5 h-5" />}
          </div>
          <div>
            <h3
              id="visibility-label"
              className="text-sm font-bold tracking-tight text-slate-900 uppercase"
            >
              Herkese Açık
            </h3>
            <p className="text-[11px] text-slate-500 font-medium">
              {isPublic
                ? 'Widget ve herkese açık sohbet bağlantısı bu botu sunar.'
                : 'Bot gizli: widget ve herkese açık uç noktalar 404 döner.'}
            </p>
          </div>
        </div>
        <Switch
          checked={isPublic}
          onCheckedChange={setIsPublic}
          aria-labelledby="visibility-label"
          className="data-[state=checked]:bg-emerald-500"
        />
      </div>
    </div>
  )
}
