import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { HardDrive, Globe, Cpu, Zap, Shield, Check, X, LayoutDashboard } from 'lucide-react'
import { api } from '@/api/client'

interface PlanConfig {
  branding?: {
    can_hide_branding: boolean
    can_custom_branding: boolean
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
  files_count: number
  max_files_count_in_one_bot: number
  storage_used_mb: number
  urls_count: number
  max_urls_count_in_one_bot: number
  tokens_used: number
  refresh_count?: number
}

const PlanPage = () => {
  const [userPlan, setUserPlan] = useState<string>('free')
  const [planPrice, setPlanPrice] = useState<number>(0)
  const [planCurrency, setPlanCurrency] = useState<string>('TRY')
  const [planConfig, setPlanConfig] = useState<PlanConfig | null>(null)
  const [usage, setUsage] = useState<Usage | null>(null)
  const [rateLimit, setRateLimit] = useState<{ limit: number | null; remaining: number | null }>({ limit: null, remaining: null })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/api/v1/me')
      .then((res) => {
        const plan = res.data?.plan_code ?? 'free'
        setUserPlan(plan)
        setPlanPrice(res.data?.plan_price ?? 0)
        setPlanCurrency(res.data?.plan_currency ?? 'TRY')
        if (res.data?.config) {
          setPlanConfig(res.data.config)
        }
        if (res.data?.usage) {
          setUsage(res.data.usage)
        }
        const limit = parseInt(res.headers['x-ratelimit-limit'] || '', 10)
        const remaining = parseInt(res.headers['x-ratelimit-remaining'] || '', 10)
        setRateLimit({
          limit: Number.isFinite(limit) ? limit : null,
          remaining: Number.isFinite(remaining) ? remaining : null,
        })
      })
      .catch(() => {
        setUserPlan('free')
        setPlanConfig(null)
        setRateLimit({ limit: null, remaining: null })
      })
      .finally(() => setLoading(false))
  }, [])

  const getPlanBadgeColor = (plan: string) => {
    switch (plan.toLowerCase()) {
      case 'ultra': return 'default' // Primary color
      case 'pro': return 'secondary'
      default: return 'outline'
    }
  }

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('tr-TR', { style: 'currency', currency: currency }).format(amount)
  }

  const calculatePercentage = (used: number, total: number) => {
    if (!total || total === 0) return 0
    return Math.min(100, Math.max(0, (used / total) * 100))
  }

  const rateLimitPercentage = rateLimit.limit && rateLimit.remaining !== null
    ? ((rateLimit.limit - rateLimit.remaining) / rateLimit.limit) * 100
    : 0

  const InactiveBadge = () => (
    <Tooltip>
      <TooltipTrigger asChild>
        <span tabIndex={0} className="inline-flex cursor-help">
          <Badge variant="destructive" className="hover:bg-destructive/90">
            <X className="h-3 w-3 mr-1"/> Pasif
          </Badge>
        </span>
      </TooltipTrigger>
      <TooltipContent>
        <p>Bu özellik üst paketlerde mevcuttur.</p>
      </TooltipContent>
    </Tooltip>
  )

  if (loading) {
    return <div className="p-8 text-center text-muted-foreground">Yükleniyor...</div>
  }

  return (
    <TooltipProvider>
    <div className="max-w-5xl space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Plan ve Faturalandırma</h1>
        <p className="text-muted-foreground mt-1">Mevcut planınızı ve kullanım detaylarını inceleyin.</p>
      </div>

      <div className="space-y-6">
        {/* Current Plan Overview */}
        <Card className="overflow-hidden border-primary/20">
          <div className="bg-primary/5 p-6 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2 mb-1">
                 <h2 className="text-2xl font-bold capitalize">{userPlan} Plan</h2>
                 <Badge variant={getPlanBadgeColor(userPlan) as any} className="uppercase">{userPlan}</Badge>
              </div>
              <p className="text-muted-foreground">
                {planPrice > 0 
                  ? `${formatCurrency(planPrice, planCurrency)} / ay` 
                  : 'Ücretsiz'}
              </p>
            </div>
            <Button variant="default" disabled>Planı Yönet</Button>
          </div>
          <CardContent className="p-6">
             <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <div className="flex items-center gap-3 p-3 rounded-lg border bg-card text-card-foreground shadow-sm">
                   <div className="p-2 bg-primary/10 rounded-full text-primary">
                     <HardDrive className="h-4 w-4" />
                   </div>
                   <div className="flex-1">
                      <p className="text-xs text-muted-foreground mb-1">Depolama</p>
                      <div className="space-y-1">
                        <div className="flex justify-between text-xs font-medium">
                          <span>{usage?.storage_used_mb ?? 0} MB</span>
                          <span className="text-muted-foreground">/ {planConfig?.files.total_storage_mb ?? 0} MB</span>
                        </div>
                        <Progress value={calculatePercentage(usage?.storage_used_mb ?? 0, planConfig?.files.total_storage_mb ?? 0)} className="h-1.5" />
                      </div>
                   </div>
                </div>
                <div className="flex items-center gap-3 p-3 rounded-lg border bg-card text-card-foreground shadow-sm">
                   <div className="p-2 bg-primary/10 rounded-full text-primary">
                     <Globe className="h-4 w-4" />
                   </div>
                   <div>
                      <p className="text-xs text-muted-foreground">URL Kullanımı</p>
                      <p className="font-medium">{usage?.urls_count ?? 0} adet</p>
                   </div>
                </div>
                <div className="flex items-center gap-3 p-3 rounded-lg border bg-card text-card-foreground shadow-sm">
                   <div className="p-2 bg-primary/10 rounded-full text-primary">
                     <Zap className="h-4 w-4" />
                   </div>
                   <div className="flex-1">
                      <p className="text-xs text-muted-foreground mb-1">Token Kullanımı</p>
                      <div className="space-y-1">
                        <div className="flex justify-between text-xs font-medium">
                          <span>{(usage?.tokens_used ?? 0).toLocaleString()}</span>
                          <span className="text-muted-foreground">/ {(planConfig?.chat.max_monthly_tokens || 0).toLocaleString()}</span>
                        </div>
                        <Progress value={calculatePercentage(usage?.tokens_used ?? 0, planConfig?.chat.max_monthly_tokens || 0)} className="h-1.5" />
                      </div>
                   </div>
                </div>
                <div className="flex items-center gap-3 p-3 rounded-lg border bg-card text-card-foreground shadow-sm">
                   <div className="p-2 bg-primary/10 rounded-full text-primary">
                     <Shield className="h-4 w-4" />
                   </div>
                   <div>
                      <p className="text-xs text-muted-foreground">Güvenli Embed</p>
                      <div className="mt-1">{userPlan !== 'free' ? <Badge variant="default" className="bg-green-600"><Check className="h-3 w-3 mr-1"/> Aktif</Badge> : <InactiveBadge />}</div>
                   </div>
                </div>
             </div>
          </CardContent>
        </Card>

        {/* Detailed Limits */}
        <div className="grid gap-6 md:grid-cols-2">
          {/* Files & Storage */}
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <HardDrive className="h-5 w-5 text-primary" />
                <CardTitle className="text-lg">Dosya ve Depolama</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
               <div className="space-y-2 py-2 border-b last:border-0">
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Toplam Depolama Kullanımı</span>
                    <span className="font-medium">{usage?.storage_used_mb ?? 0} / {planConfig?.files.total_storage_mb ?? 0} MB</span>
                  </div>
                  <Progress value={calculatePercentage(usage?.storage_used_mb ?? 0, planConfig?.files.total_storage_mb ?? 0)} className="h-2" />
               </div>
               <div className="space-y-2 py-2 border-b last:border-0">
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Toplam Yüklenen Dosya</span>
                    <span className="font-medium">{usage?.files_count ?? 0} / {planConfig?.files.max_files_total ?? 0} adet</span>
                  </div>
                  <Progress value={calculatePercentage(usage?.files_count ?? 0, planConfig?.files.max_files_total ?? 0)} className="h-2" />
               </div>
               <div className="flex justify-between items-center py-2 border-b last:border-0">
                  <span className="text-sm">Maks. Dosya Boyutu</span>
                  <span className="font-medium">{planConfig?.files.max_size_mb} MB</span>
               </div>
               <div className="space-y-2 py-2 border-b last:border-0">
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Bot Başına Dosya</span>
                    <span className="font-medium">{usage?.max_files_count_in_one_bot ?? 0} / {planConfig?.files.max_files_per_bot} adet</span>
                  </div>
                  <Progress value={calculatePercentage(usage?.max_files_count_in_one_bot ?? 0, planConfig?.files.max_files_per_bot ?? 0)} className="h-2" />
               </div>
               <div className="flex justify-between items-center py-2 border-b last:border-0">
                  <span className="text-sm">OCR (Görsel Tarama)</span>
                  {planConfig?.files.ocr_enabled 
                    ? <Badge variant="default" className="bg-green-600"><Check className="h-3 w-3 mr-1"/> Aktif</Badge> 
                    : <InactiveBadge />}
               </div>
            </CardContent>
          </Card>

          {/* Web Scraping */}
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <Globe className="h-5 w-5 text-primary" />
                <CardTitle className="text-lg">Web Tarama</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
               <div className="flex justify-between items-center py-2 border-b last:border-0">
                  <span className="text-sm">Toplam Taranan URL</span>
                  <span className="font-medium">{usage?.urls_count ?? 0} adet</span>
               </div>
               <div className="space-y-2 py-2 border-b last:border-0">
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Bot Başına URL</span>
                    <span className="font-medium">{usage?.max_urls_count_in_one_bot ?? 0} / {planConfig?.scraping.max_urls_per_bot} adet</span>
                  </div>
                  <Progress value={calculatePercentage(usage?.max_urls_count_in_one_bot ?? 0, planConfig?.scraping.max_urls_per_bot ?? 0)} className="h-2" />
               </div>
               <div className="flex justify-between items-center py-2 border-b last:border-0">
                  <span className="text-sm">Alt Sayfa Tarama (Crawler)</span>
                  <span className="font-medium">{planConfig?.scraping.max_pages_per_crawl} sayfa</span>
               </div>
               <div className="flex justify-between items-center py-2 border-b last:border-0">
                  <span className="text-sm">Dinamik (JS) Tarama</span>
                  {planConfig?.scraping.dynamic_enabled 
                    ? <Badge variant="default" className="bg-green-600"><Check className="h-3 w-3 mr-1"/> Aktif</Badge> 
                    : <InactiveBadge />}
               </div>
               <div className="flex justify-between items-center py-2 border-b last:border-0">
                  <span className="text-sm">Kaynak Yenileme</span>
                  {planConfig?.refresh?.enabled 
                    ? <Badge variant="default" className="bg-green-600"><Check className="h-3 w-3 mr-1"/> Aktif</Badge> 
                    : <InactiveBadge />}
               </div>
               {planConfig?.refresh?.enabled && (
                 <div className="space-y-2 py-2 border-b last:border-0">
                   <div className="flex justify-between items-center">
                     <span className="text-sm">Aylık Yenileme Hakkı</span>
                     <span className="font-medium">{usage?.refresh_count ?? 0} / {planConfig?.refresh?.max_monthly ?? 0} kullanıldı</span>
                   </div>
                   <Progress value={calculatePercentage(usage?.refresh_count ?? 0, planConfig?.refresh?.max_monthly ?? 0)} className="h-2" />
                 </div>
               )}
            </CardContent>
          </Card>

          {/* AI & Chat */}
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <Cpu className="h-5 w-5 text-primary" />
                <CardTitle className="text-lg">Yapay Zeka Modelleri</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
               <div className="space-y-2 py-2 border-b last:border-0">
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Aylık Token Kullanımı</span>
                    <span className="font-medium">{(usage?.tokens_used ?? 0).toLocaleString()} / {(planConfig?.chat.max_monthly_tokens || 0).toLocaleString()}</span>
                  </div>
                  <Progress value={calculatePercentage(usage?.tokens_used ?? 0, planConfig?.chat.max_monthly_tokens || 0)} className="h-2" />
               </div>
               <div className="flex justify-between items-start py-2 border-b last:border-0">
                  <span className="text-sm mt-1">Erişilebilir Modeller</span>
                  <div className="flex flex-wrap gap-1 justify-end max-w-[180px]">
                     {planConfig?.chat.allowed_models.map(m => (
                       <Badge key={m} variant="outline" className="text-xs">{m}</Badge>
                     ))}
                  </div>
               </div>
               <div className="flex justify-between items-center py-2 border-b last:border-0">
                  <span className="text-sm">RAG Bağlamı (Context)</span>
                  <span className="font-medium">{planConfig?.chat.rag.max_context_tokens} token</span>
               </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <LayoutDashboard className="h-5 w-5 text-primary" />
                <CardTitle className="text-lg">Marka ve Özelleştirme</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between items-center py-2 border-b last:border-0">
                <span className="text-sm">‘Powered by Botla’ kaldırma</span>
                {planConfig?.branding?.can_hide_branding
                  ? <Badge variant="default" className="bg-green-600"><Check className="h-3 w-3 mr-1"/> Aktif</Badge>
                  : <InactiveBadge />}
              </div>
              <div className="flex justify-between items-center py-2 border-b last:border-0">
                <span className="text-sm">Özel Branding (Logo/Bağlantı)</span>
                {planConfig?.branding?.can_custom_branding
                  ? <Badge variant="default" className="bg-green-600"><Check className="h-3 w-3 mr-1"/> Aktif</Badge>
                  : <InactiveBadge />}
              </div>
            </CardContent>
          </Card>

           {/* System Usage */}
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <LayoutDashboard className="h-5 w-5 text-primary" />
                <CardTitle className="text-lg">Sistem Kullanımı</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
               <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>API Hız Limiti (Dakikalık)</span>
                    <span className="text-muted-foreground">
                       {rateLimit.remaining ?? 0} / {rateLimit.limit ?? 0} kalan
                    </span>
                  </div>
                  <Progress value={rateLimitPercentage} className="h-2" />
                  <p className="text-xs text-muted-foreground">
                    Dakikalık API istek limitinizin %{Math.round(rateLimitPercentage)} kadarını kullandınız.
                  </p>
               </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
    </TooltipProvider>
  )
}

export default PlanPage
