import { Clock, RefreshCw, Calendar, CalendarDays } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface RefreshSettingsProps {
  refreshPolicy: 'manual' | 'auto'
  refreshFrequency: 'daily' | 'weekly' | 'monthly' | null
  nextRefreshAt: string | null
  lastRefreshAt: string | null
  onRefreshPolicyChange: (policy: 'manual' | 'auto') => void
  onRefreshFrequencyChange: (frequency: 'daily' | 'weekly' | 'monthly') => void
}

export default function RefreshSettings({
  refreshPolicy,
  refreshFrequency,
  nextRefreshAt,
  lastRefreshAt,
  onRefreshPolicyChange,
  onRefreshFrequencyChange,
}: RefreshSettingsProps) {
  
  const frequencyOptions: { value: 'daily' | 'weekly' | 'monthly'; label: string; icon: React.ReactNode; description: string }[] = [
    { value: 'daily', label: 'Günlük', icon: <Clock className="w-4 h-4" />, description: 'Her gün gece yarısı' },
    { value: 'weekly', label: 'Haftalık', icon: <CalendarDays className="w-4 h-4" />, description: 'Her Pazar gece yarısı' },
    { value: 'monthly', label: 'Aylık', icon: <Calendar className="w-4 h-4" />, description: 'Her ayın 1\'i' },
  ]

  const formatDate = (dateStr: string | null): string => {
    if (!dateStr) return '-'
    try {
      const date = new Date(dateStr)
      return date.toLocaleDateString('tr-TR', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    } catch {
      return '-'
    }
  }

  return (
    <Card className="border-none shadow-sm bg-white/80 backdrop-blur-sm">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold flex items-center gap-2">
          <RefreshCw className="w-4 h-4 text-blue-500" />
          Otomatik Yenileme
        </CardTitle>
        <p className="text-sm text-muted-foreground mt-1">
          URL kaynaklarının otomatik olarak güncellenme ayarları
        </p>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Policy Selection */}
        <div className="space-y-3">
          <label className="text-sm font-medium text-gray-700">Yenileme Modu</label>
          <div className="grid grid-cols-2 gap-3">
            <button
              type="button"
              onClick={() => onRefreshPolicyChange('manual')}
              className={`flex flex-col items-center justify-center p-4 rounded-xl border-2 transition-all ${
                refreshPolicy === 'manual'
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 bg-white hover:border-gray-300 text-gray-600'
              }`}
            >
              <RefreshCw className={`w-5 h-5 mb-2 ${refreshPolicy === 'manual' ? 'text-blue-500' : 'text-gray-400'}`} />
              <span className="font-medium text-sm">Manuel</span>
              <span className="text-xs text-gray-500 mt-1">Sadece elle tetikleme</span>
            </button>

            <button
              type="button"
              onClick={() => {
                onRefreshPolicyChange('auto')
                if (!refreshFrequency) {
                  onRefreshFrequencyChange('weekly') // Default to weekly
                }
              }}
              className={`flex flex-col items-center justify-center p-4 rounded-xl border-2 transition-all ${
                refreshPolicy === 'auto'
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 bg-white hover:border-gray-300 text-gray-600'
              }`}
            >
              <Clock className={`w-5 h-5 mb-2 ${refreshPolicy === 'auto' ? 'text-blue-500' : 'text-gray-400'}`} />
              <span className="font-medium text-sm">Otomatik</span>
              <span className="text-xs text-gray-500 mt-1">Zamana bağlı yenileme</span>
            </button>
          </div>
        </div>

        {/* Frequency Selection - Only show when auto is selected */}
        {refreshPolicy === 'auto' && (
          <div className="space-y-3 animate-in fade-in slide-in-from-top-2 duration-200">
            <label className="text-sm font-medium text-gray-700">Yenileme Sıklığı</label>
            <div className="grid grid-cols-3 gap-2">
              {frequencyOptions.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => onRefreshFrequencyChange(option.value)}
                  className={`flex flex-col items-center p-3 rounded-lg border transition-all ${
                    refreshFrequency === option.value
                      ? 'border-blue-500 bg-blue-50 text-blue-700'
                      : 'border-gray-200 bg-white hover:border-gray-300 text-gray-600'
                  }`}
                >
                  <div className={`mb-1 ${refreshFrequency === option.value ? 'text-blue-500' : 'text-gray-400'}`}>
                    {option.icon}
                  </div>
                  <span className="font-medium text-sm">{option.label}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Refresh Status Info */}
        {refreshPolicy === 'auto' && (
          <div className="bg-gray-50 rounded-lg p-4 space-y-3 animate-in fade-in slide-in-from-top-2 duration-200">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Son Yenileme</span>
              <Badge variant="outline" className="text-xs">
                {formatDate(lastRefreshAt)}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Sonraki Yenileme</span>
              <Badge variant="secondary" className="text-xs">
                {formatDate(nextRefreshAt)}
              </Badge>
            </div>
          </div>
        )}

        {/* Info Note */}
        <div className="flex items-start gap-2 p-3 bg-blue-50/50 rounded-lg border border-blue-100">
          <RefreshCw className="w-4 h-4 text-blue-500 mt-0.5 flex-shrink-0" />
          <p className="text-xs text-blue-700">
            Otomatik yenileme, URL kaynaklarınızın içeriğini belirtilen aralıklarda kontrol eder 
            ve değişiklik varsa güncellenmiş içeriği chatbot'unuza aktarır. Aylık yenileme limitiniz 
            planınıza göre belirlenir.
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
