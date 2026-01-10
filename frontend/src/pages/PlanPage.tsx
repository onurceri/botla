import { useRef } from 'react'
import { motion, Variants } from 'framer-motion'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { PlanBadge, PlanTier } from '@/components/ui/plan-badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  HardDrive,
  Globe,
  Cpu,
  Zap,
  Shield,
  Check,
  FileText,
  Clock,
  Database,
  Lock,
  Bot,
  Sparkles,
  TrendingUp,
} from 'lucide-react'
import { usePlan, useUsage } from '@/hooks/queries/useProfile'
import { useRateLimit } from '@/hooks/useRateLimit'
import { cn } from '@/lib/utils'

// Data Interfaces
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

// ------------------------------------------------------------------
// Custom Visual Components
// ------------------------------------------------------------------

const EASE_OUT_EXPO = [0.16, 1, 0.3, 1] as const

const CircularProgress = ({
  value,
  max,
  size = 60,
  strokeWidth = 4,
  color = 'var(--color-primary)',
  trackColor = 'rgba(0,0,0,0.1)',
  label,
  subLabel,
}: {
  value: number
  max: number
  size?: number
  strokeWidth?: number
  color?: string
  trackColor?: string
  label?: string
  subLabel?: string
}) => {
  const radius = (size - strokeWidth) / 2
  const circumference = radius * 2 * Math.PI
  const percentage = max > 0 ? Math.min(100, Math.max(0, (value / max) * 100)) : 0
  const offset = circumference - (percentage / 100) * circumference

  return (
    <div className="flex flex-col items-center justify-center gap-2">
      <div className="relative flex items-center justify-center" style={{ width: size, height: size }}>
        <svg width={size} height={size} className="transform -rotate-90">
          {/* Track */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke={trackColor}
            strokeWidth={strokeWidth}
            strokeLinecap="round"
          />
          {/* Indicator */}
          <motion.circle
            initial={{ strokeDashoffset: circumference }}
            animate={{ strokeDashoffset: offset }}
            transition={{ duration: 1.5, ease: EASE_OUT_EXPO }}
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke={color}
            strokeWidth={strokeWidth}
            strokeLinecap="round"
            strokeDasharray={circumference}
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center flex-col">
          <span className="text-[10px] font-bold text-muted-foreground">{Math.round(percentage)}%</span>
        </div>
      </div>
      {(label || subLabel) && (
        <div className="text-center space-y-0.5">
          {label && <div className="text-xs font-bold">{label}</div>}
          {subLabel && <div className="text-[10px] text-muted-foreground">{subLabel}</div>}
        </div>
      )}
    </div>
  )
}

const FeatureTile = ({
  label,
  enabled,
  value,
  icon: Icon,
  tooltip,
}: {
  label: string
  enabled?: boolean
  value?: string
  icon: any
  tooltip?: string
}) => {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <div
          className={cn(
            'group relative flex flex-col items-center justify-center p-3 rounded-xl border transition-all duration-300',
            enabled === false
              ? 'bg-muted/30 border-dashed border-border/60 text-muted-foreground opacity-70'
              : 'bg-gradient-to-br from-card to-card/50 border-border/50 hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5 hover:-translate-y-0.5'
          )}
        >
          <div
            className={cn(
              'p-2 rounded-full mb-2 transition-colors duration-300',
              enabled === false
                ? 'bg-muted text-muted-foreground'
                : 'bg-primary/10 text-primary group-hover:bg-primary/20'
            )}
          >
            {enabled === false ? <Lock className="w-4 h-4" /> : <Icon className="w-4 h-4" />}
          </div>
          <span className="text-[11px] font-bold uppercase tracking-wider text-center text-muted-foreground/80 mb-0.5">
            {label}
          </span>
          <span
            className={cn(
              'text-sm font-bold text-center',
              enabled === false ? 'text-muted-foreground line-through' : 'text-foreground'
            )}
          >
            {value || (enabled ? 'Aktif' : 'Pasif')}
          </span>
        </div>
      </TooltipTrigger>
      {tooltip && <TooltipContent>{tooltip}</TooltipContent>}
    </Tooltip>
  )
}

const LimitBar = ({
  label,
  current,
  max,
  icon: Icon,
  format = (v: number) => v.toLocaleString(),
}: {
  label: string
  current: number
  max: number
  icon: any
  format?: (v: number) => string
}) => {
  const percentage = max > 0 ? (current / max) * 100 : 0

  return (
    <div className="space-y-2">
      <div className="flex justify-between items-end">
        <div className="flex items-center gap-2">
          <Icon className="w-4 h-4 text-primary" />
          <span className="text-sm font-medium text-foreground">{label}</span>
        </div>
        <div className="text-xs font-bold">
          <span className={cn(percentage > 90 ? 'text-red-500' : 'text-foreground')}>
            {format(current)}
          </span>
          <span className="text-muted-foreground"> / {format(max)}</span>
        </div>
      </div>
      <div className="h-2 w-full bg-muted/50 rounded-full overflow-hidden">
        <motion.div
          className={cn(
            'h-full rounded-full',
            percentage > 90
              ? 'bg-gradient-to-r from-red-500 to-red-400'
              : 'bg-gradient-to-r from-primary to-amber-400'
          )}
          initial={{ width: 0 }}
          animate={{ width: `${Math.min(100, percentage)}%` }}
          transition={{ duration: 1, delay: 0.2, ease: EASE_OUT_EXPO }}
        />
      </div>
    </div>
  )
}

// ------------------------------------------------------------------
// Main Page Component
// ------------------------------------------------------------------

const PlanPage = () => {
  const { data: planData, isLoading: planLoading } = usePlan()
  const { data: usageData, isLoading: usageLoading } = useUsage()
  const rateLimit = useRateLimit()

  // Use refs for potential scroll animations
  const containerRef = useRef(null)

  const userPlan = planData?.code || 'free'
  const planPrice = planData?.price || 0
  const planCurrency = planData?.currency || 'TRY'
  const planLimits = planData?.limits as PlanLimits | null
  const planFeatures = planData?.features as PlanFeatures | null
  const usage = usageData as Usage | null
  const loading = planLoading || usageLoading

  // Constants
  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('tr-TR', {
      style: 'currency',
      currency: currency,
      maximumFractionDigits: 0,
    }).format(amount)
  }

  const rateLimitUsed = rateLimit.limit ? (rateLimit.limit - (rateLimit.remaining ?? 0)) : 0
  const rateLimitMax = rateLimit.limit ?? 0

  // Animation variants
  const containerVariants: Variants = {
    hidden: { opacity: 0 },
    show: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
        delayChildren: 0.2,
      },
    },
  }

  const itemVariants: Variants = {
    hidden: { opacity: 0, y: 20 },
    show: { opacity: 1, y: 0, transition: { duration: 0.5, ease: EASE_OUT_EXPO } },
  }

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto p-4 space-y-8 animate-pulse">
        <div className="h-48 bg-muted/20 rounded-3xl" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-6">
          <div className="h-64 bg-muted/20 rounded-3xl" />
          <div className="h-64 bg-muted/20 rounded-3xl" />
        </div>
      </div>
    )
  }

  return (
    <TooltipProvider>
      <div className="min-h-screen pb-20 overflow-x-hidden" ref={containerRef}>
        <motion.div
          className="max-w-7xl mx-auto space-y-8"
          variants={containerVariants}
          initial="hidden"
          animate="show"
        >
          {/* HEADER SECTION */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 px-1">
            <motion.div variants={itemVariants} className="space-y-1">
              <h1 className="text-3xl font-extrabold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/60">
                Plan &amp; Limitler
              </h1>
              <p className="text-muted-foreground font-medium text-base">
                Hesap özelliklerinizi yönetin ve kullanımınızı takip edin.
              </p>
            </motion.div>
            
            <motion.div variants={itemVariants}>
               {/* Optional Header Action */}
            </motion.div>
          </div>

          {/* TOTAL OVERVIEW GRID */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            
            {/* 1. CURRENT PLAN CARD */}
            <motion.div variants={itemVariants} className="lg:col-span-2">
              <div className="relative h-full overflow-hidden rounded-3xl border border-orange-100/50 bg-gradient-to-br from-white via-orange-50/40 to-white shadow-2xl shadow-orange-500/5">
                {/* Background decorative blobs */}
                <div className="absolute top-0 right-0 w-[400px] h-[400px] bg-orange-400/10 blur-[120px] rounded-full translate-x-1/3 -translate-y-1/3" />
                <div className="absolute bottom-0 left-0 w-[300px] h-[300px] bg-amber-300/10 blur-[100px] rounded-full -translate-x-1/3 translate-y-1/3" />

                <div className="relative z-10 p-8 h-full flex flex-col justify-between">
                  <div className="flex justify-between items-start">
                    <div className="space-y-4">
                      <div className="flex items-center gap-3">
                        <PlanBadge plan={userPlan as PlanTier} variant="soft" className="scale-110 origin-left" />
                        <span className="flex items-center gap-1.5 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                          Aktif Paket
                        </span>
                      </div>
                      <div>
                        <h2 className="text-4xl md:text-5xl font-black tracking-tighter text-foreground capitalize mb-2">
                          {userPlan} Plan
                        </h2>
                        <p className="text-muted-foreground font-medium max-w-md text-lg leading-relaxed">
                          {planPrice > 0 
                            ? 'İşletmeniz için tam profesyonel özellikler paketi.'
                            : 'Bireysel kullanım ve deneme için başlangıç paketi.'
                          }
                        </p>
                      </div>
                    </div>
                    {/* Price Tag */}
                    <div className="text-right hidden sm:block">
                      <div className="text-4xl font-black text-foreground tracking-tight">
                        {planPrice > 0 ? formatCurrency(planPrice, planCurrency) : 'Ücretsiz'}
                      </div>
                      {planPrice > 0 && <div className="text-sm font-bold text-muted-foreground/60 uppercase tracking-wide mt-1">aylık faturalandırma</div>}
                    </div>
                  </div>

                  {/* Quick Features */}
                  <div className="mt-10 grid grid-cols-2 sm:grid-cols-4 gap-6">
                     <div className="flex items-center gap-3.5 group">
                       <div className="p-2.5 rounded-xl bg-orange-50 text-orange-600 ring-1 ring-orange-100 group-hover:scale-110 transition-transform duration-300"><Bot className="w-5 h-5" /></div>
                       <div className="flex flex-col gap-0.5">
                         <span className="text-[10px] text-muted-foreground/70 uppercase font-black tracking-wider">Botlar</span>
                         <span className="font-bold text-foreground text-sm">{planLimits?.max_chatbots || 1} Adet</span>
                       </div>
                     </div>
                     <div className="flex items-center gap-3.5 group">
                       <div className="p-2.5 rounded-xl bg-blue-50 text-blue-600 ring-1 ring-blue-100 group-hover:scale-110 transition-transform duration-300"><Database className="w-5 h-5" /></div>
                       <div className="flex flex-col gap-0.5">
                         <span className="text-[10px] text-muted-foreground/70 uppercase font-black tracking-wider">Depolama</span>
                         <span className="font-bold text-foreground text-sm">{planFeatures?.files.total_storage_mb || 0} MB</span>
                       </div>
                     </div>
                     <div className="flex items-center gap-3.5 group">
                       <div className="p-2.5 rounded-xl bg-purple-50 text-purple-600 ring-1 ring-purple-100 group-hover:scale-110 transition-transform duration-300"><Sparkles className="w-5 h-5" /></div>
                       <div className="flex flex-col gap-0.5">
                         <span className="text-[10px] text-muted-foreground/70 uppercase font-black tracking-wider">AI Model</span>
                         <span className="font-bold text-foreground text-sm truncate max-w-[100px]">{planFeatures?.chat.allowed_models[0] || 'GPT-3.5'}</span>
                       </div>
                     </div>
                  </div>
                </div>
              </div>
            </motion.div>

            {/* 2. HEALTH & STATUS CARD */}
            <motion.div variants={itemVariants} className="lg:col-span-1">
              <Card className="h-full border-border/50 shadow-lg bg-card/40 backdrop-blur-xl flex flex-col">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <TrendingUp className="w-5 h-5 text-primary" />
                    Kullanım Özeti
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 flex flex-col justify-center items-center gap-6 py-2">
                   <div className="grid grid-cols-2 gap-x-8 gap-y-6 w-full px-4">
                      <CircularProgress 
                        value={usage?.chatbots_count ?? 0} 
                        max={planLimits?.max_chatbots ?? 1} 
                        label="Chatbots"
                        subLabel={`${usage?.chatbots_count} / ${planLimits?.max_chatbots}`}
                        color="#f59e0b"
                      />
                      <CircularProgress 
                        value={usage?.storage_used_mb ?? 0} 
                        max={planFeatures?.files.total_storage_mb ?? 0} 
                        label="Depolama"
                        subLabel={`${usage?.storage_used_mb}MB`}
                        color="#3b82f6"
                      />
                      <CircularProgress 
                        value={usage?.tokens_used ?? 0} 
                        max={planFeatures?.chat.max_monthly_tokens ?? 0} 
                        label="AI Token"
                        subLabel="Bu Ay"
                        color="#8b5cf6"
                      />
                      <CircularProgress 
                        value={rateLimitUsed} 
                        max={rateLimitMax} 
                        label="İstek Limiti"
                        subLabel="Kullanım"
                        color="#ef4444"
                      />
                   </div>
                </CardContent>
              </Card>
            </motion.div>
          </div>

          {/* DETAILED FEATURES GRID */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            
            {/* AI CAPABILITIES */}
            <motion.div variants={itemVariants}>
              <Card className="h-full hover:shadow-lg transition-shadow duration-500 border-border/60">
                <CardHeader className="pb-4 border-b border-border/5">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="p-2.5 bg-violet-500/10 rounded-xl text-violet-600">
                        <Cpu className="w-6 h-6" />
                      </div>
                      <div>
                        <CardTitle className="text-lg">Yapay Zeka &amp; Modeller</CardTitle>
                        <CardDescription>Erişilebilir modeller ve zeka kapasitesi</CardDescription>
                      </div>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-6 space-y-6">
                  {/* Token Usage Bar */}
                  <LimitBar 
                    label="Aylık Token Tüketimi" 
                    icon={Zap} 
                    current={usage?.tokens_used || 0} 
                    max={planFeatures?.chat.max_monthly_tokens || 0} 
                  />

                  {/* Feature Tiles */}
                  <div className="grid grid-cols-3 gap-3 pt-2">
                    {planFeatures?.chat.allowed_models.map((model) => (
                      <FeatureTile 
                        key={model} 
                        label="Model" 
                        value={model} 
                        icon={Bot} 
                        enabled={true} 
                      />
                    ))}
                    <FeatureTile 
                      label="RAG Context" 
                      value={`${(planFeatures?.chat.rag.max_context_tokens || 0) / 1000}k Token`} 
                      icon={Database} 
                      enabled={true} 
                    />
                  </div>
                </CardContent>
              </Card>
            </motion.div>

            {/* DATA & SCRAPING */}
            <motion.div variants={itemVariants}>
              <Card className="h-full hover:shadow-lg transition-shadow duration-500 border-border/60">
                 <CardHeader className="pb-4 border-b border-border/5">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="p-2.5 bg-blue-500/10 rounded-xl text-blue-600">
                        <Globe className="w-6 h-6" />
                      </div>
                      <div>
                        <CardTitle className="text-lg">Veri &amp; Web Tarama</CardTitle>
                        <CardDescription>Web kazıma ve veri işleme limitleri</CardDescription>
                      </div>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-6 space-y-6">
                  <div className="space-y-4">
                    <LimitBar 
                      label="Web Kaynakları (URL)" 
                      icon={Globe} 
                      current={usage?.urls_count || 0} 
                      max={(planFeatures?.scraping.max_urls_per_bot || 0) + (planFeatures?.scraping.max_pages_per_crawl || 0)} 
                    />
                    <LimitBar 
                      label="Yüklenen Dosyalar" 
                      icon={FileText} 
                      current={usage?.files_count || 0} 
                      max={planFeatures?.files.max_files_total || 0} 
                    />
                  </div>

                  <div className="grid grid-cols-3 gap-3 pt-2">
                     <FeatureTile 
                       label="Dinamik Tarama" 
                       enabled={planFeatures?.scraping.dynamic_enabled} 
                       icon={Zap} 
                       tooltip="Javascript tabanlı siteleri tarayabilme (Puppeteer)"
                     />
                     <FeatureTile 
                       label="Max Dosya Boyutu" 
                       value={`${planFeatures?.files.max_size_mb} MB`} 
                       icon={HardDrive} 
                       enabled={true}
                     />
                     <FeatureTile 
                       label="Oto Yenileme" 
                       enabled={planFeatures?.refresh?.enabled} 
                       icon={Clock} 
                       tooltip="Veri kaynaklarını otomatik güncelleme"
                     />
                  </div>
                </CardContent>
              </Card>
            </motion.div>

            {/* SECURITY & INFRASTRUCTURE */}
            <motion.div variants={itemVariants} className="md:col-span-2">
              <Card className="hover:shadow-lg transition-shadow duration-500 border-border/60 bg-gradient-to-r from-card to-muted/20">
                <CardContent className="p-6 flex flex-col md:flex-row items-center justify-between gap-6">
                  
                  <div className="flex items-center gap-4 flex-1">
                    <div className="p-3 bg-emerald-500/10 rounded-2xl text-emerald-600 shrink-0">
                      <Shield className="w-8 h-8" />
                    </div>
                    <div>
                      <h3 className="text-lg font-bold">Güvenlik &amp; Marka</h3>
                      <p className="text-sm text-muted-foreground">Widget güvenliği ve özelleştirme seçenekleriniz</p>
                    </div>
                  </div>

                  <div className="flex flex-wrap justify-center gap-3">
                     <div className="flex items-center gap-2 px-4 py-2 bg-background border rounded-lg shadow-sm">
                        {planFeatures?.branding?.can_hide_branding ? <Check className="w-4 h-4 text-emerald-500" /> : <Lock className="w-4 h-4 text-orange-500" />}
                        <span className="text-sm font-medium">No-Branding</span>
                     </div>
                     <div className="flex items-center gap-2 px-4 py-2 bg-background border rounded-lg shadow-sm">
                        {planFeatures?.branding?.can_custom_branding ? <Check className="w-4 h-4 text-emerald-500" /> : <Lock className="w-4 h-4 text-orange-500" />}
                        <span className="text-sm font-medium">Özel Logo</span>
                     </div>
                     <div className="flex items-center gap-2 px-4 py-2 bg-background border rounded-lg shadow-sm">
                        {planFeatures?.security?.secure_embed_enabled ? <Check className="w-4 h-4 text-emerald-500" /> : <Lock className="w-4 h-4 text-orange-500" />}
                        <span className="text-sm font-medium">Domain Kilidi</span>
                     </div>
                  </div>

                </CardContent>
              </Card>
            </motion.div>

          </div>
        </motion.div>
      </div>
    </TooltipProvider>
  )
}

export default PlanPage
