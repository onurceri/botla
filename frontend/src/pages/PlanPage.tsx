import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { PlanBadge, PlanTier } from '@/components/ui/plan-badge'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  HardDrive,
  Globe,
  Cpu,
  Zap,
  Shield,
  Check,
  X,
  LayoutDashboard,
  RefreshCcw,
  FileText,
  Clock,
  Database,
  ExternalLink,
} from 'lucide-react'
import { usePlan, useUsage } from '@/hooks/queries/useProfile'
import { useRateLimit } from '@/hooks/useRateLimit'
import { cn } from '@/lib/utils'

interface PlanLimits {
  max_chatbots: number
  max_monthly_ingestions: number
  max_monthly_embedding_tokens: number
  min_readd_cooldown_minutes: number
}

interface PlanFeatures {
  branding?: {
    can_hide_branding: boolean
    can_custom_branding: boolean
  }
  security?: {
    secure_embed_enabled: boolean
  }
  scraping: {
    dynamic_enabled: boolean
    max_urls_per_bot: number
    max_pages_per_crawl: number
  }
  files: {
    ocr_enabled: boolean
    max_size_mb: number
    max_files_per_bot: number
    max_files_total: number
    total_storage_mb: number
  }
  chat: {
    allowed_models: string[]
    max_monthly_tokens: number
    rag: {
      top_k: number
      max_context_tokens: number
    }
  }
  refresh?: {
    enabled: boolean
    max_monthly: number
  }
}

interface Usage {
  chatbots_count: number
  files_count: number
  max_files_count_in_one_bot: number
  storage_used_mb: number
  urls_count: number
  max_urls_count_in_one_bot: number
  tokens_used: number
  ingestions_used: number
  ingestion_embedding_tokens: number
  refresh_count?: number
}

// Skeleton component for loading state
const Skeleton = ({ className }: { className?: string }) => (
  <div className={cn('animate-pulse rounded-xl bg-muted/60', className)} />
)

// Stat card component with a cleaner, more professional look
const StatCard = ({
  icon: Icon,
  label,
  value,
  max,
  showProgress = false,
  accentColor = 'primary',
}: {
  icon: React.ElementType
  label: string
  value: string | number
  max?: string | number
  showProgress?: boolean
  accentColor?: 'primary' | 'amber' | 'blue' | 'violet' | 'emerald'
}) => {
  const percentage =
    showProgress && typeof value === 'number' && typeof max === 'number' && max > 0
      ? Math.min(100, Math.max(0, (value / max) * 100))
      : 0

  const accentClasses = {
    primary: 'text-primary bg-primary/10',
    amber: 'text-amber-600 bg-amber-500/10',
    blue: 'text-blue-600 bg-blue-500/10',
    violet: 'text-violet-600 bg-violet-500/10',
    emerald: 'text-emerald-600 bg-emerald-500/10',
  }

  return (
    <Card className="overflow-hidden border-border/50 bg-card/50 backdrop-blur-sm transition-all duration-200 hover:border-border hover:shadow-sm">
      <CardContent className="p-5">
        <div className="flex items-center gap-3 mb-4">
          <div className={cn('p-2 rounded-lg shrink-0', accentClasses[accentColor])}>
            <Icon className="w-4 h-4" />
          </div>
          <span className="text-sm font-medium text-muted-foreground truncate">{label}</span>
        </div>

        <div className="space-y-4">
          <div className="flex items-baseline gap-1.5">
            <span className="text-2xl font-bold tracking-tight">{value}</span>
            {max !== undefined && (
              <span className="text-sm text-muted-foreground font-medium">/ {max}</span>
            )}
          </div>

          {showProgress && (
            <div className="space-y-2">
              <Progress value={percentage} className="h-1.5 bg-muted/50" />
              <div className="flex justify-between items-center text-[10px] uppercase tracking-wider font-bold text-muted-foreground/70">
                <span>Kullanım</span>
                <span>%{Math.round(percentage)}</span>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

// Feature row component
const FeatureRow = ({
  label,
  value,
  enabled,
  showProgress = false,
  current,
  max,
  className,
}: {
  label: string
  value?: string | number
  enabled?: boolean
  showProgress?: boolean
  current?: number
  max?: number
  className?: string
}) => {
  const percentage =
    showProgress && current !== undefined && max !== undefined && max > 0
      ? Math.min(100, Math.max(0, (current / max) * 100))
      : 0

  return (
    <div className={cn('py-3.5 first:pt-0 last:pb-0', className)}>
      <div className="flex items-center justify-between gap-4">
        <span className="text-sm text-muted-foreground font-medium">{label}</span>
        <div className="flex items-center gap-2 shrink-0">
          {enabled !== undefined ? (
            enabled ? (
              <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-600 border border-emerald-500/20 text-[11px] font-bold uppercase tracking-wider">
                <Check className="w-3 h-3" />
                Aktif
              </div>
            ) : (
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-muted text-muted-foreground border border-border text-[11px] font-bold uppercase tracking-wider cursor-help">
                    <X className="w-3 h-3" />
                    Pasif
                  </div>
                </TooltipTrigger>
                <TooltipContent side="left">
                  <p className="text-xs">Bu özellik üst paketlerde mevcuttur.</p>
                </TooltipContent>
              </Tooltip>
            )
          ) : showProgress ? (
            <span className="text-sm font-bold tracking-tight">
              {current?.toLocaleString()} <span className="text-muted-foreground font-medium">/</span> {max?.toLocaleString()}
            </span>
          ) : (
            <span className="text-sm font-bold tracking-tight">{value}</span>
          )}
        </div>
      </div>
      {showProgress && current !== undefined && max !== undefined && (
        <div className="mt-2.5">
          <Progress value={percentage} className="h-1 bg-muted/50" />
        </div>
      )}
    </div>
  )
}

// Section card component
const SectionCard = ({
  icon: Icon,
  title,
  description,
  children,
}: {
  icon: React.ElementType
  title: string
  description?: string
  children: React.ReactNode
}) => (
  <Card className="border-border/50 bg-card/50 backdrop-blur-sm overflow-hidden">
    <CardHeader className="pb-5">
      <div className="flex items-center gap-4">
        <div className="p-2.5 rounded-xl bg-muted border border-border/50 text-foreground/80">
          <Icon className="w-5 h-5" />
        </div>
        <div className="space-y-0.5">
          <CardTitle className="text-base font-bold tracking-tight">{title}</CardTitle>
          {description && <CardDescription className="text-xs">{description}</CardDescription>}
        </div>
      </div>
    </CardHeader>
    <CardContent className="pt-0 border-t border-border/10 mt-1">{children}</CardContent>
  </Card>
)

const PlanPage = () => {
  const { data: planData, isLoading: planLoading } = usePlan()
  const { data: usageData, isLoading: usageLoading } = useUsage()
  const rateLimit = useRateLimit()

  const userPlan = planData?.code || 'free'
  const planPrice = planData?.price || 0
  const planCurrency = planData?.currency || 'TRY'
  const planLimits = planData?.limits as PlanLimits | null
  const planFeatures = planData?.features as PlanFeatures | null
  const usage = usageData as Usage | null
  const loading = planLoading || usageLoading

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('tr-TR', { 
      style: 'currency', 
      currency: currency,
      maximumFractionDigits: 0
    }).format(amount)
  }

  const rateLimitPercentage =
    rateLimit.limit && rateLimit.remaining !== null
      ? ((rateLimit.limit - rateLimit.remaining) / rateLimit.limit) * 100
      : 0

  // Loading skeleton
  if (loading) {
    return (
      <TooltipProvider>
        <div className="max-w-6xl mx-auto space-y-8 pb-12">
          <div className="flex flex-col gap-2">
            <Skeleton className="h-9 w-64" />
            <Skeleton className="h-4 w-96" />
          </div>
          <Skeleton className="h-44 rounded-2xl" />
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} className="h-32 rounded-xl" />
            ))}
          </div>
          <div className="grid gap-6 md:grid-cols-2">
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} className="h-64 rounded-xl" />
            ))}
          </div>
        </div>
      </TooltipProvider>
    )
  }

  return (
    <TooltipProvider>
      <div className="max-w-6xl mx-auto space-y-8 pb-12 animate-in fade-in duration-700">
        {/* Simplified Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl md:text-3xl font-extrabold tracking-tight">
              Plan ve Kullanım
            </h1>
            <p className="text-muted-foreground text-sm font-medium">
              Mevcut paketinizin detayları ve kaynak kullanım durumunuz.
            </p>
          </div>
        </div>

        {/* Current Plan Card - More Professional & Structured */}
        <Card className="overflow-hidden border-border/60 bg-gradient-to-br from-background via-muted/5 to-muted/10 shadow-sm relative">
          <div className="p-6 md:p-10 flex flex-col lg:flex-row items-stretch lg:items-center gap-8">
            <div className="flex-1 space-y-4 text-center md:text-left">
              <div className="flex flex-wrap items-center justify-center md:justify-start gap-3">
                <h2 className="text-3xl md:text-4xl font-extrabold tracking-tight capitalize">{userPlan}</h2>
                <PlanBadge plan={userPlan as PlanTier} size="md" variant="solid" />
              </div>
              <p className="text-muted-foreground font-medium max-w-lg text-lg">
                {planPrice > 0
                  ? `Aylık ${formatCurrency(planPrice, planCurrency)} karşılığında profesyonel özelliklere erişiminiz var.`
                  : "Ücretsiz plan ile Botla'nın temel özelliklerini kullanıyorsunuz."}
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-3 min-w-[320px]">
              <div className="flex-1 p-5 rounded-2xl bg-background/80 border border-border/50 flex flex-col justify-center gap-1.5 shadow-sm">
                <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground/80">Plan Ücreti</span>
                <span className="text-2xl font-black">
                  {planPrice > 0 ? formatCurrency(planPrice, planCurrency) : 'Ücretsiz'}
                  {planPrice > 0 && <span className="text-sm font-bold text-muted-foreground ml-1">/ ay</span>}
                </span>
              </div>
            </div>
          </div>
        </Card>

        {/* Refined Quick Stats Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 font-inter">
          <StatCard
            icon={LayoutDashboard}
            label="Mevcut Chatbot"
            value={usage?.chatbots_count ?? 0}
            max={planLimits?.max_chatbots ?? 1}
            showProgress
            accentColor="violet"
          />
          <StatCard
            icon={FileText}
            label="Aylık Kaynak Ekleme"
            value={usage?.ingestions_used ?? 0}
            max={planLimits?.max_monthly_ingestions ?? 0}
            showProgress
            accentColor="emerald"
          />
          <StatCard
            icon={Zap}
            label="AI Token Kullanımı"
            value={(usage?.tokens_used ?? 0).toLocaleString()}
            max={(planFeatures?.chat.max_monthly_tokens || 0).toLocaleString()}
            showProgress
            accentColor="amber"
          />
          <StatCard
            icon={HardDrive}
            label="Depolama Alanı"
            value={`${usage?.storage_used_mb ?? 0} MB`}
            max={`${planFeatures?.files.total_storage_mb ?? 0} MB`}
            showProgress
            accentColor="blue"
          />
        </div>

        {/* Detailed Sections with cleaner layout */}
        <div className="grid gap-6 md:grid-cols-2">
          {/* Files & Storage */}
          <SectionCard
            icon={Database}
            title="Veri Kaynakları & Depolama"
            description="Dosya yükleme ve depolama limitleri"
          >
            <div className="divide-y divide-border/10">
              <FeatureRow
                label="Toplam Depolama"
                showProgress
                current={usage?.storage_used_mb ?? 0}
                max={planFeatures?.files.total_storage_mb ?? 0}
              />
              <FeatureRow
                label="Toplam Yüklenen Dosya"
                showProgress
                current={usage?.files_count ?? 0}
                max={planFeatures?.files.max_files_total ?? 0}
              />
              <FeatureRow
                label="Maksimum Dosya Boyutu"
                value={`${planFeatures?.files.max_size_mb} MB`}
              />
              <FeatureRow
                label="Bot Başına Dosya Limiti"
                showProgress
                current={usage?.max_files_count_in_one_bot ?? 0}
                max={planFeatures?.files.max_files_per_bot ?? 0}
              />
              <FeatureRow label="OCR (Görsel İçerik Tarama)" enabled={planFeatures?.files.ocr_enabled} />
            </div>
          </SectionCard>

          {/* Web Scraping */}
          <SectionCard
            icon={Globe}
            title="Web Sitesi Denetimi"
            description="URL kotaları ve tarama özellikleri"
          >
            <div className="divide-y divide-border/10">
              <FeatureRow
                label="Toplam Kayıtlı URL"
                value={`${usage?.urls_count ?? 0} adet`}
              />
              <FeatureRow
                label="Bot Başına URL Limiti"
                showProgress
                current={usage?.max_urls_count_in_one_bot ?? 0}
                max={(planFeatures?.scraping.max_urls_per_bot ?? 0) + (planFeatures?.scraping.max_pages_per_crawl ?? 0)}
              />
              <FeatureRow
                label="Dinamik Web Taraması (JS)"
                enabled={planFeatures?.scraping.dynamic_enabled}
              />
              <FeatureRow label="İçerik Otomatik Yenileme" enabled={planFeatures?.refresh?.enabled} />
              {planFeatures?.refresh?.enabled && (
                <FeatureRow
                  label="Aylık Yenileme Kotası"
                  showProgress
                  current={usage?.refresh_count ?? 0}
                  max={planFeatures?.refresh?.max_monthly ?? 0}
                />
              )}
            </div>
          </SectionCard>

          {/* AI Models */}
          <SectionCard
            icon={Cpu}
            title="Zeka & Model Erişimi"
            description="Yapay zeka modelleri ve RAG ayarları"
          >
            <div className="divide-y divide-border/10">
              <FeatureRow
                label="AI Token Kotası"
                showProgress
                current={usage?.tokens_used ?? 0}
                max={planFeatures?.chat.max_monthly_tokens ?? 0}
              />
              <div className="py-4 last:pb-0">
                <div className="flex items-start justify-between gap-4">
                  <span className="text-sm text-muted-foreground font-medium">Kullanılabilir Modeller</span>
                  <div className="flex flex-wrap gap-1.5 justify-end">
                    {planFeatures?.chat.allowed_models.map((m) => (
                      <Badge key={m} variant="secondary" className="px-2 py-0 text-[10px] font-bold h-5 uppercase tracking-wide bg-muted border border-border/50">
                        {m}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>
              <FeatureRow
                label="RAG Context (Bağlam) Penceresi"
                value={`${planFeatures?.chat.rag.max_context_tokens.toLocaleString()} token`}
              />
            </div>
          </SectionCard>

          {/* Security & Branding */}
          <SectionCard
            icon={Shield}
            title="Marka & Güvenlik"
            description="Widget özelleştirme ve erişim denetimi"
          >
            <div className="divide-y divide-border/10">
              <FeatureRow
                label="Güvenli Embed (Domain Kilidi)"
                enabled={planFeatures?.security?.secure_embed_enabled}
              />
              <FeatureRow
                label="Botla Markasını Kaldırma"
                enabled={planFeatures?.branding?.can_hide_branding}
              />
              <FeatureRow
                label="Özel Logo ve Link Ayarları"
                enabled={planFeatures?.branding?.can_custom_branding}
              />
            </div>
          </SectionCard>
        </div>

        {/* Refined API Usage Section */}
        <SectionCard
          icon={Clock}
          title="Sistem API & Kaynak Kullanımı"
          description="Altyapı kotaları ve teknik limitler"
        >
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 pt-4">
            <UsageResourceCard 
              icon={FileText} 
              label="Kaynak Ekleme" 
              current={usage?.ingestions_used ?? 0} 
              max={planLimits?.max_monthly_ingestions ?? 0} 
            />
            <UsageResourceCard 
              icon={Database} 
              label="Embedding Token" 
              current={usage?.ingestion_embedding_tokens ?? 0} 
              max={planLimits?.max_monthly_embedding_tokens ?? 0} 
              isNumber 
            />
            <UsageResourceCard 
              icon={Zap} 
              label="API Limit (Dakikalık)" 
              current={rateLimit.limit ? (rateLimit.limit - (rateLimit.remaining ?? 0)) : 0} 
              max={rateLimit.limit ?? 0} 
              progressValue={rateLimitPercentage}
              progressLabel={`%${Math.round(rateLimitPercentage)} kullanıldı`}
            />
          </div>
        </SectionCard>
      </div>
    </TooltipProvider>
  )
}

const UsageResourceCard = ({ 
  icon: Icon, 
  label, 
  current, 
  max, 
  isNumber = false,
  progressValue,
  progressLabel
}: { 
  icon: any, 
  label: string, 
  current: number, 
  max: number,
  isNumber?: boolean,
  progressValue?: number,
  progressLabel?: string
}) => {
  const percentage = max > 0 ? (current / max) * 100 : 0
  const actualProgress = progressValue !== undefined ? progressValue : percentage
  
  return (
    <div className="space-y-4 p-5 rounded-2xl bg-muted/20 border border-border/50">
      <div className="flex items-center gap-2.5 text-xs font-bold uppercase tracking-wider text-muted-foreground">
        <Icon className="w-3.5 h-3.5 text-foreground/70" />
        {label}
      </div>
      <div className="space-y-2.5">
        <div className="flex justify-between items-baseline">
          <span className="text-xl font-black">
            {isNumber ? current.toLocaleString() : current}
          </span>
          <span className="text-xs font-bold text-muted-foreground/60">
            /{isNumber ? max.toLocaleString() : max}
          </span>
        </div>
        <Progress
          value={actualProgress}
          className="h-1.5 bg-muted"
        />
        {progressLabel && (
          <p className="text-[10px] font-bold text-muted-foreground/70 uppercase tracking-tight">
            {progressLabel}
          </p>
        )}
      </div>
    </div>
  )
}

export default PlanPage
