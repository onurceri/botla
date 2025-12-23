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
  CreditCard,
  Sparkles,
  RefreshCcw,
  FileText,
  ArrowUpRight,
  Clock,
  Database,
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
  <div className={cn('animate-pulse rounded-md bg-muted/60', className)} />
)

// Modern stat card component
const StatCard = ({
  icon: Icon,
  label,
  value,
  max,
  subtext,
  showProgress = false,
  gradient = 'from-primary/10 to-primary/5',
}: {
  icon: React.ElementType
  label: string
  value: string | number
  max?: string | number
  subtext?: React.ReactNode
  showProgress?: boolean
  gradient?: string
}) => {
  const percentage =
    showProgress && typeof value === 'number' && typeof max === 'number' && max > 0
      ? Math.min(100, Math.max(0, (value / max) * 100))
      : 0

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl border border-border/50 p-5',
        'bg-gradient-to-br backdrop-blur-sm',
        'transition-all duration-300 hover:shadow-lg hover:shadow-primary/5 hover:border-primary/20',
        'group',
        gradient,
      )}
    >
      {/* Background decoration */}
      <div className="absolute right-0 top-0 -mr-4 -mt-4 opacity-[0.07] group-hover:opacity-[0.12] transition-opacity">
        <Icon className="w-24 h-24" />
      </div>

      <div className="relative z-10">
        <div className="flex items-center gap-3 mb-3">
          <div className="p-2.5 rounded-xl bg-background/80 shadow-sm border border-border/50">
            <Icon className="w-4 h-4 text-primary" />
          </div>
          <span className="text-sm font-medium text-muted-foreground">{label}</span>
        </div>

        <div className="space-y-2">
          <div className="flex items-baseline gap-1">
            <span className="text-2xl font-bold tracking-tight">{value}</span>
            {max !== undefined && <span className="text-sm text-muted-foreground">/ {max}</span>}
          </div>

          {showProgress && (
            <div className="space-y-1.5">
              <Progress value={percentage} className="h-2" />
              <p className="text-xs text-muted-foreground">%{Math.round(percentage)} kullanıldı</p>
            </div>
          )}

          {subtext && !showProgress && (
            <div className="text-sm text-muted-foreground">{subtext}</div>
          )}
        </div>
      </div>
    </div>
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
    <div className={cn('py-3 border-b border-border/50 last:border-0', className)}>
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{label}</span>
        {enabled !== undefined ? (
          enabled ? (
            <Badge className="bg-emerald-500/10 text-emerald-600 border-emerald-500/20 hover:bg-emerald-500/20">
              <Check className="w-3 h-3 mr-1" />
              Aktif
            </Badge>
          ) : (
            <Tooltip>
              <TooltipTrigger asChild>
                <span tabIndex={0} className="inline-flex cursor-help">
                  <Badge
                    variant="secondary"
                    className="bg-muted text-muted-foreground border-border hover:bg-muted"
                  >
                    <X className="w-3 h-3 mr-1" />
                    Pasif
                  </Badge>
                </span>
              </TooltipTrigger>
              <TooltipContent>
                <p>Bu özellik üst paketlerde mevcuttur.</p>
              </TooltipContent>
            </Tooltip>
          )
        ) : showProgress ? (
          <span className="text-sm font-medium">
            {current?.toLocaleString()} / {max?.toLocaleString()}
          </span>
        ) : (
          <span className="text-sm font-medium">{value}</span>
        )}
      </div>
      {showProgress && current !== undefined && max !== undefined && (
        <div className="mt-2">
          <Progress value={percentage} className="h-1.5" />
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
  gradient = 'from-background to-muted/20',
}: {
  icon: React.ElementType
  title: string
  description?: string
  children: React.ReactNode
  gradient?: string
}) => (
  <Card
    className={cn(
      'overflow-hidden border-border/50 shadow-sm',
      'bg-gradient-to-br transition-all duration-300 hover:shadow-md',
      gradient,
    )}
  >
    <CardHeader className="pb-4">
      <div className="flex items-center gap-3">
        <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
          <Icon className="w-5 h-5" />
        </div>
        <div>
          <CardTitle className="text-lg">{title}</CardTitle>
          {description && <CardDescription className="mt-0.5">{description}</CardDescription>}
        </div>
      </div>
    </CardHeader>
    <CardContent className="pt-0">{children}</CardContent>
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
    return new Intl.NumberFormat('tr-TR', { style: 'currency', currency: currency }).format(amount)
  }

  const rateLimitPercentage =
    rateLimit.limit && rateLimit.remaining !== null
      ? ((rateLimit.limit - rateLimit.remaining) / rateLimit.limit) * 100
      : 0

  // Loading skeleton
  if (loading) {
    return (
      <TooltipProvider>
        <div className="max-w-6xl mx-auto space-y-8 animate-in fade-in duration-500">
          {/* Header Skeleton */}
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-3">
              <Skeleton className="w-12 h-12 rounded-xl" />
              <div className="space-y-2">
                <Skeleton className="h-8 w-64" />
                <Skeleton className="h-4 w-96" />
              </div>
            </div>
          </div>

          {/* Plan Card Skeleton */}
          <Skeleton className="h-48 rounded-2xl" />

          {/* Stats Grid Skeleton */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} className="h-36 rounded-2xl" />
            ))}
          </div>

          {/* Section Cards Skeleton */}
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
      <div className="max-w-6xl mx-auto space-y-8 animate-in fade-in duration-500">
        {/* Modern Header */}
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-3">
            <div className="p-3 rounded-xl bg-gradient-to-br from-primary/20 to-primary/10 text-primary shadow-sm">
              <CreditCard className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-2xl md:text-3xl font-bold tracking-tight">
                Plan ve Faturalandırma
              </h1>
              <p className="text-muted-foreground">
                Mevcut planınızı ve kullanım detaylarınızı görüntüleyin.
              </p>
            </div>
          </div>
        </div>

        {/* Current Plan Hero Card */}
        <Card className="overflow-hidden border-border/50 shadow-lg bg-gradient-to-br from-primary/5 via-background to-primary/10 relative">
          {/* Background decorations */}
          <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full blur-3xl -mr-32 -mt-32" />
          <div className="absolute bottom-0 left-0 w-48 h-48 bg-primary/5 rounded-full blur-3xl -ml-24 -mb-24" />

          <div className="relative z-10">
            <div className="p-6 md:p-8 flex flex-col lg:flex-row items-start lg:items-center justify-between gap-6">
              <div className="flex items-center gap-4">
                <div className="p-4 rounded-2xl bg-gradient-to-br from-primary to-primary/80 text-primary-foreground shadow-lg shadow-primary/25">
                  <Sparkles className="w-8 h-8" />
                </div>
                <div>
                  <div className="flex items-center gap-3 mb-1">
                    <h2 className="text-2xl md:text-3xl font-bold capitalize">{userPlan} Plan</h2>
                    <PlanBadge plan={userPlan as PlanTier} size="md" />
                  </div>
                  <p className="text-lg text-muted-foreground">
                    {planPrice > 0
                      ? `${formatCurrency(planPrice, planCurrency)} / ay`
                      : 'Ücretsiz Plan'}
                  </p>
                </div>
              </div>

              <div className="flex flex-col sm:flex-row gap-3 w-full lg:w-auto">
                <Button variant="outline" className="gap-2" disabled>
                  <RefreshCcw className="w-4 h-4" />
                  Yenile
                </Button>
                <Button className="gap-2 shadow-lg" disabled>
                  <ArrowUpRight className="w-4 h-4" />
                  Planı Yükselt
                </Button>
              </div>
            </div>
          </div>
        </Card>

        {/* Quick Stats Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatCard
            icon={LayoutDashboard}
            label="Chatbot Sayısı"
            value={usage?.chatbots_count ?? 0}
            max={planLimits?.max_chatbots ?? 1}
            showProgress
            gradient="from-violet-500/10 to-violet-500/5"
          />
          <StatCard
            icon={FileText}
            label="Aylık Kaynak Ekleme"
            value={usage?.ingestions_used ?? 0}
            max={planLimits?.max_monthly_ingestions ?? 0}
            showProgress
            gradient="from-emerald-500/10 to-emerald-500/5"
          />
          <StatCard
            icon={Zap}
            label="Aylık AI Token"
            value={(usage?.tokens_used ?? 0).toLocaleString()}
            max={(planFeatures?.chat.max_monthly_tokens || 0).toLocaleString()}
            showProgress
            gradient="from-amber-500/10 to-amber-500/5"
          />
          <StatCard
            icon={HardDrive}
            label="Depolama Alanı"
            value={`${usage?.storage_used_mb ?? 0} MB`}
            max={`${planFeatures?.files.total_storage_mb ?? 0} MB`}
            showProgress
            gradient="from-blue-500/10 to-blue-500/5"
          />
        </div>

        {/* Detailed Sections */}
        <div className="grid gap-6 md:grid-cols-2">
          {/* Files & Storage */}
          <SectionCard
            icon={Database}
            title="Dosya ve Depolama"
            description="Veri depolama limitleri ve kullanımı"
          >
            <div className="divide-y divide-border/50">
              <FeatureRow
                label="Toplam Depolama"
                showProgress
                current={usage?.storage_used_mb ?? 0}
                max={planFeatures?.files.total_storage_mb ?? 0}
              />
              <FeatureRow
                label="Yüklenen Dosya"
                showProgress
                current={usage?.files_count ?? 0}
                max={planFeatures?.files.max_files_total ?? 0}
              />
              <FeatureRow
                label="Maks. Dosya Boyutu"
                value={`${planFeatures?.files.max_size_mb} MB`}
              />
              <FeatureRow
                label="Bot Başına Dosya"
                showProgress
                current={usage?.max_files_count_in_one_bot ?? 0}
                max={planFeatures?.files.max_files_per_bot ?? 0}
              />
              <FeatureRow label="OCR (Görsel Tarama)" enabled={planFeatures?.files.ocr_enabled} />
            </div>
          </SectionCard>

          {/* Web Scraping */}
          <SectionCard
            icon={Globe}
            title="Web İçerik Kaynakları"
            description="URL tarama limitleri ve özellikleri"
          >
            <div className="divide-y divide-border/50">
              <FeatureRow
                label="Tüm Botlardaki URL Sayısı"
                value={`${usage?.urls_count ?? 0} adet`}
              />
              <div className="py-3 border-b border-border/50">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Bot Başına URL Limiti</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="text-xs text-muted-foreground/60 cursor-help">ⓘ</span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-[250px]">
                        <p className="text-xs">
                          Her bot için ekleyebileceğiniz ana URL + alt sayfa tarama sonucu eklenen
                          URL'lerin toplam limiti.
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </div>
                  <span className="text-sm font-medium">
                    {usage?.max_urls_count_in_one_bot ?? 0} /{' '}
                    {(planFeatures?.scraping.max_urls_per_bot ?? 0) +
                      (planFeatures?.scraping.max_pages_per_crawl ?? 0)}
                  </span>
                </div>
                <div className="mt-2">
                  <Progress
                    value={
                      ((usage?.max_urls_count_in_one_bot ?? 0) /
                        ((planFeatures?.scraping.max_urls_per_bot ?? 0) +
                          (planFeatures?.scraping.max_pages_per_crawl ?? 0))) *
                      100
                    }
                    className="h-1.5"
                  />
                </div>
              </div>
              <FeatureRow
                label="Dinamik Site Tarama (JavaScript)"
                enabled={planFeatures?.scraping.dynamic_enabled}
              />
              <FeatureRow label="İçerik Yenileme" enabled={planFeatures?.refresh?.enabled} />
              {planFeatures?.refresh?.enabled && (
                <FeatureRow
                  label="Aylık Yenileme Hakkı"
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
            title="Yapay Zeka Modelleri"
            description="AI model erişimi ve token limitleri"
          >
            <div className="divide-y divide-border/50">
              <FeatureRow
                label="Aylık Token Kullanımı"
                showProgress
                current={usage?.tokens_used ?? 0}
                max={planFeatures?.chat.max_monthly_tokens ?? 0}
              />
              <div className="py-3 border-b border-border/50">
                <div className="flex items-start justify-between">
                  <span className="text-sm text-muted-foreground mt-1">Erişilebilir Modeller</span>
                  <div className="flex flex-wrap gap-1.5 justify-end max-w-[200px]">
                    {planFeatures?.chat.allowed_models.map((m) => (
                      <Badge key={m} variant="outline" className="text-xs bg-background/50">
                        {m}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>
              <FeatureRow
                label="RAG Bağlamı"
                value={`${planFeatures?.chat.rag.max_context_tokens} token`}
              />
            </div>
          </SectionCard>

          {/* Security & Branding */}
          <SectionCard
            icon={Shield}
            title="Güvenlik ve Branding"
            description="Embed güvenliği ve marka özelleştirmeleri"
          >
            <div className="divide-y divide-border/50">
              <FeatureRow
                label="Güvenli Embed (Domain Kilidi)"
                enabled={planFeatures?.security?.secure_embed_enabled}
              />
              <FeatureRow
                label="'Powered by Botla' Kaldırma"
                enabled={planFeatures?.branding?.can_hide_branding}
              />
              <FeatureRow
                label="Özel Branding (Logo/Link)"
                enabled={planFeatures?.branding?.can_custom_branding}
              />
            </div>
          </SectionCard>
        </div>

        {/* System Usage Section */}
        <SectionCard
          icon={Clock}
          title="Sistem Kullanımı"
          description="API limitleri ve kaynak ekleme kotaları"
          gradient="from-background to-muted/30"
        >
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <div className="space-y-3 p-4 rounded-xl bg-muted/30 border border-border/50">
              <div className="flex items-center gap-2 text-sm font-medium">
                <FileText className="w-4 h-4 text-primary" />
                Aylık Kaynak Ekleme
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Kullanılan</span>
                  <span className="font-medium">
                    {usage?.ingestions_used ?? 0} / {planLimits?.max_monthly_ingestions ?? 0}
                  </span>
                </div>
                <Progress
                  value={
                    planLimits?.max_monthly_ingestions
                      ? ((usage?.ingestions_used ?? 0) / planLimits.max_monthly_ingestions) * 100
                      : 0
                  }
                  className="h-2"
                />
              </div>
            </div>

            <div className="space-y-3 p-4 rounded-xl bg-muted/30 border border-border/50">
              <div className="flex items-center gap-2 text-sm font-medium">
                <Database className="w-4 h-4 text-primary" />
                Embedding Token
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Kullanılan</span>
                  <span className="font-medium">
                    {(usage?.ingestion_embedding_tokens ?? 0).toLocaleString()} /{' '}
                    {(planLimits?.max_monthly_embedding_tokens ?? 0).toLocaleString()}
                  </span>
                </div>
                <Progress
                  value={
                    planLimits?.max_monthly_embedding_tokens
                      ? ((usage?.ingestion_embedding_tokens ?? 0) /
                          planLimits.max_monthly_embedding_tokens) *
                        100
                      : 0
                  }
                  className="h-2"
                />
              </div>
            </div>

            <div className="space-y-3 p-4 rounded-xl bg-muted/30 border border-border/50">
              <div className="flex items-center gap-2 text-sm font-medium">
                <Zap className="w-4 h-4 text-primary" />
                API Hız Limiti (Dakika)
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Kalan</span>
                  <span className="font-medium">
                    {rateLimit.remaining ?? 0} / {rateLimit.limit ?? 0}
                  </span>
                </div>
                <Progress value={rateLimitPercentage} className="h-2" />
                <p className="text-xs text-muted-foreground">
                  %{Math.round(rateLimitPercentage)} kullanıldı
                </p>
              </div>
            </div>
          </div>
        </SectionCard>
      </div>
    </TooltipProvider>
  )
}

export default PlanPage
