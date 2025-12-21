import { ChevronDown, ChevronUp, Headphones } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useState } from 'react'
import HandoffSettings from '../../../components/HandoffSettings'
import { useChatbotContext } from '../../../context/ChatbotContext'
import { useAutoSave } from '../../../hooks/useAutoSave'
import { useUpdateHandoff } from '@/hooks/mutations/useChatbotMutations'

interface HandoffSectionProps {
  chatbotId: string
}

export default function HandoffSection({ chatbotId }: HandoffSectionProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const {
    handoffEnabled, setHandoffEnabled,
    handoffType, setHandoffType,
    handoffConfig, setHandoffConfig,
    planConfig,
    buildHandoffPayload,
  } = useChatbotContext()

  const { mutateAsync: updateHandoff } = useUpdateHandoff(chatbotId)

  useAutoSave({
    payload: buildHandoffPayload(),
    saveFn: (_, payload) => updateHandoff(payload),
  })

  const canUseHandoff = planConfig?.guardrails?.can_use_escalate_fallback

  // Summary for collapsed view
  const getSummaryText = () => {
    if (!canUseHandoff) return 'Pro+ planı gerektirir'
    if (!handoffEnabled) return 'Devre dışı'
    if (handoffType === 'email') return 'E-posta bildirimi aktif'
    if (handoffType === 'dashboard') return 'Dashboard bildirimi aktif'
    return 'Aktif'
  }

  return (
    <Card className="border-muted-foreground/20 shadow-sm hover:shadow-md transition-shadow">
      <CardHeader 
        className="cursor-pointer select-none"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-600">
              <Headphones className="w-5 h-5" />
            </div>
            <div>
              <CardTitle className="text-lg flex items-center gap-2">
                İnsan Desteği
                {handoffEnabled && canUseHandoff && (
                  <span className="px-2 py-0.5 text-xs font-medium bg-green-500/10 text-green-600 rounded-full">
                    Aktif
                  </span>
                )}
              </CardTitle>
              <CardDescription className="mt-0.5">
                {isExpanded ? 'Konuşmaları insana yönlendirme' : getSummaryText()}
              </CardDescription>
            </div>
          </div>
          <Button variant="ghost" size="icon" className="shrink-0">
            {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </Button>
        </div>
      </CardHeader>
      
      {isExpanded && (
        <CardContent className="pt-0 animate-in slide-in-from-top-2 duration-200">
          <HandoffSettings
            handoffEnabled={handoffEnabled}
            setHandoffEnabled={setHandoffEnabled}
            handoffType={handoffType}
            setHandoffType={setHandoffType}
            handoffConfig={handoffConfig}
            setHandoffConfig={setHandoffConfig}
            canUseHandoff={canUseHandoff}
          />
        </CardContent>
      )}
    </Card>
  )
}
