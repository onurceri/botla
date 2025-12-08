import { Link2, Zap, Clock, Ban } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'

type DiscoveryMode = 'auto' | 'pending' | 'disabled'

interface DiscoveryModeSectionProps {
  discoveryMode: DiscoveryMode
  setDiscoveryMode: (mode: DiscoveryMode) => void
}

const modes: { value: DiscoveryMode; label: string; description: string; icon: typeof Zap }[] = [
  {
    value: 'auto',
    label: 'Otomatik',
    description: 'Keşfedilen URL\'ler otomatik olarak kaynak olarak eklenir.',
    icon: Zap,
  },
  {
    value: 'pending',
    label: 'Onay Bekle',
    description: 'Keşfedilen URL\'ler size sunulur ve onayınızı bekler.',
    icon: Clock,
  },
  {
    value: 'disabled',
    label: 'Kapalı',
    description: 'Alt sayfa keşfi yapılmaz, sadece eklediğiniz URL işlenir.',
    icon: Ban,
  },
]

export default function DiscoveryModeSection({ discoveryMode, setDiscoveryMode }: DiscoveryModeSectionProps) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center gap-2">
          <Link2 className="w-5 h-5 text-primary" />
          <CardTitle className="text-base">URL Keşif Modu</CardTitle>
        </div>
        <CardDescription>
          Bir URL eklediğinizde, sayfadaki bağlantıların nasıl işleneceğini belirleyin.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {modes.map((mode) => {
          const Icon = mode.icon
          const isSelected = discoveryMode === mode.value
          return (
            <label
              key={mode.value}
              className={`
                flex items-start gap-3 p-4 rounded-xl cursor-pointer transition-all
                border-2
                ${isSelected 
                  ? 'border-primary bg-primary/5' 
                  : 'border-border hover:border-primary/50 hover:bg-muted/50'
                }
              `}
            >
              <input
                type="radio"
                name="discoveryMode"
                value={mode.value}
                checked={isSelected}
                onChange={() => setDiscoveryMode(mode.value)}
                className="sr-only"
              />
              <div className={`
                flex-shrink-0 w-5 h-5 rounded-full border-2 flex items-center justify-center mt-0.5
                ${isSelected ? 'border-primary' : 'border-muted-foreground'}
              `}>
                {isSelected && <div className="w-2.5 h-2.5 rounded-full bg-primary" />}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <Icon className={`w-4 h-4 ${isSelected ? 'text-primary' : 'text-muted-foreground'}`} />
                  <span className={`font-medium ${isSelected ? 'text-primary' : 'text-foreground'}`}>
                    {mode.label}
                  </span>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  {mode.description}
                </p>
              </div>
            </label>
          )
        })}
      </CardContent>
    </Card>
  )
}

