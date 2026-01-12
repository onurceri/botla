import React, { useEffect, useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '@/contexts/AuthContext'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Bot,
  Zap,
  ShieldCheck,
  Database,
  ArrowRight,
  CheckCircle2,
  Menu,
  X,
  Sparkles,
  Search,
  BarChart3,
  Lock,
  Palette,
  Headphones,
  Building2,
  Code2,
  Play,
  ChevronRight,
  Shield,
  Key,
  Eye,
  Activity,
  Clock,
  Globe,
  Upload,
  Cpu,
  MessageCircle,
  Users,
  Star,
  Check,
  Minus,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PlanBadge, PlanTier } from '@/components/ui/plan-badge'
import { cn } from '@/lib/utils'
import { usePlans } from '@/hooks/queries/usePlans'
import { landing } from '@/i18n/landing'

const t = landing

// ============================================
// DESIGN SYSTEM COMPONENTS
// ============================================

const SectionBadge = ({ children, icon: Icon }: { children: React.ReactNode; icon?: React.ElementType }) => (
  <motion.div
    initial={{ opacity: 0, y: 10 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/5 border border-primary/10 mb-6"
  >
    {Icon && <Icon className="w-4 h-4 text-primary" />}
    <span className="text-sm font-semibold text-primary">{children}</span>
  </motion.div>
)

const SectionTitle = ({ children, highlight }: { children: React.ReactNode; highlight?: string }) => (
  <motion.h2
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay: 0.1 }}
    className="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight text-foreground"
  >
    {children}
    {highlight && <span className="text-primary"> {highlight}</span>}
  </motion.h2>
)

const SectionDescription = ({ children }: { children: React.ReactNode }) => (
  <motion.p
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay: 0.2 }}
    className="mt-6 text-lg sm:text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed"
  >
    {children}
  </motion.p>
)

// ============================================
// NAVBAR
// ============================================

const Navbar = ({ authenticated }: { authenticated: boolean }) => {
  const [isOpen, setIsOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 20)
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const links = [
    { name: t.nav.features, href: '#features' },
    { name: t.nav.useCases, href: '#use-cases' },
    { name: t.nav.howItWorks, href: '#how-it-works' },
    { name: t.nav.pricing, href: '#pricing' },
    { name: t.nav.faq, href: '#faq' },
  ]

  const handleScroll = (e: React.MouseEvent<HTMLAnchorElement>, href: string) => {
    e.preventDefault()
    if (location.pathname !== '/') {
      navigate('/' + href)
      return
    }
    const targetId = href.replace('#', '')
    const element = document.getElementById(targetId)
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' })
      window.history.pushState(null, '', href)
    }
    setIsOpen(false)
  }

  const handleLogoClick = () => {
    if (location.pathname !== '/') {
      navigate('/')
      return
    }
    window.scrollTo({ top: 0, behavior: 'smooth' })
    window.history.pushState(null, '', '/')
  }

  return (
    <nav
      className={cn(
        'fixed inset-x-0 top-0 z-50 transition-all duration-500',
        scrolled ? 'bg-background/80 backdrop-blur-xl border-b border-border/50 py-4' : 'py-6',
      )}
    >
      <div className="max-w-7xl mx-auto px-6">
        <div className="flex items-center justify-between">
          {/* Logo */}
          <div
            className="flex items-center gap-3 cursor-pointer group"
            onClick={handleLogoClick}
          >
            <div className="relative">
              <div className="absolute inset-0 bg-primary/20 blur-xl rounded-full group-hover:bg-primary/30 transition-colors" />
              <div className="relative bg-gradient-to-br from-primary to-orange-600 p-2.5 rounded-xl">
                <Bot className="w-5 h-5 text-white" />
              </div>
            </div>
            <span className="font-bold text-xl tracking-tight">botla.app</span>
          </div>

          {/* Desktop Nav */}
          <div className="hidden lg:flex items-center gap-1">
            {links.map((link) => (
              <a
                key={link.name}
                href={link.href}
                onClick={(e) => handleScroll(e, link.href)}
                className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors rounded-lg hover:bg-muted/50"
              >
                {link.name}
              </a>
            ))}
          </div>

          {/* CTA Buttons */}
          <div className="hidden lg:flex items-center gap-3">
            {authenticated ? (
              <Link to="/dashboard">
                <Button className="rounded-full px-6 font-semibold">{t.nav.dashboard}</Button>
              </Link>
            ) : (
              <>
                <Link to="/login">
                  <Button variant="ghost" className="rounded-full px-5 font-medium">
                    {t.nav.login}
                  </Button>
                </Link>
                <Link to="/register">
                  <Button className="rounded-full px-6 font-semibold shadow-lg shadow-primary/25 hover:shadow-primary/40 transition-shadow">
                    {t.nav.register}
                    <ArrowRight className="ml-2 w-4 h-4" />
                  </Button>
                </Link>
              </>
            )}
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsOpen(!isOpen)}
            className="lg:hidden p-2 rounded-lg hover:bg-muted/50 transition-colors"
          >
            {isOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>
      </div>

      {/* Mobile Nav */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="lg:hidden border-t border-border/50 bg-background/95 backdrop-blur-xl"
          >
            <div className="max-w-7xl mx-auto px-6 py-6 space-y-4">
              {links.map((link) => (
                <a
                  key={link.name}
                  href={link.href}
                  onClick={(e) => handleScroll(e, link.href)}
                  className="block py-3 text-lg font-medium text-foreground hover:text-primary transition-colors"
                >
                  {link.name}
                </a>
              ))}
              <div className="pt-4 border-t border-border/50 flex flex-col gap-3">
                {authenticated ? (
                  <Link to="/dashboard" onClick={() => setIsOpen(false)}>
                    <Button className="w-full h-12 rounded-xl text-base">
                      {t.nav.goToDashboard}
                    </Button>
                  </Link>
                ) : (
                  <>
                    <Link to="/login" onClick={() => setIsOpen(false)}>
                      <Button variant="outline" className="w-full h-12 rounded-xl text-base">
                        {t.nav.login}
                      </Button>
                    </Link>
                    <Link to="/register" onClick={() => setIsOpen(false)}>
                      <Button className="w-full h-12 rounded-xl text-base">{t.nav.register}</Button>
                    </Link>
                  </>
                )}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </nav>
  )
}

// ============================================
// HERO SECTION
// ============================================

const Hero = ({ authenticated }: { authenticated: boolean }) => {
  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden pt-24 pb-20">
      {/* Background */}
      <div className="absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,rgba(245,158,11,0.15),transparent)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_80%_20%,rgba(251,146,60,0.08),transparent_50%)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_80%,rgba(245,158,11,0.05),transparent_50%)]" />
        
        {/* Grid Pattern */}
        <div 
          className="absolute inset-0 opacity-[0.015]"
          style={{
            backgroundImage: `linear-gradient(to right, currentColor 1px, transparent 1px),
                             linear-gradient(to bottom, currentColor 1px, transparent 1px)`,
            backgroundSize: '64px 64px'
          }}
        />
      </div>

      <div className="max-w-7xl mx-auto px-6 relative">
        <div className="text-center max-w-5xl mx-auto">
          {/* Badge */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="inline-flex items-center gap-3 px-5 py-2.5 rounded-full bg-gradient-to-r from-primary/10 to-orange-500/10 border border-primary/20 mb-8"
          >
            <div className="flex items-center gap-1.5">
              <Cpu className="w-4 h-4 text-primary" />
              <span className="text-sm font-semibold text-foreground">{t.hero.badge}</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-1.5">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
              </span>
              <span className="text-sm font-medium text-emerald-600">{t.hero.badgeLive}</span>
            </div>
          </motion.div>

          {/* Headline */}
          <motion.h1
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.1 }}
            className="text-5xl sm:text-6xl md:text-7xl lg:text-8xl font-bold tracking-tight leading-[1.1]"
          >
            <span className="text-foreground">{t.hero.title.line1}</span>
            <br />
            <span className="relative">
              <span className="bg-gradient-to-r from-primary via-orange-500 to-amber-500 bg-clip-text text-transparent">
                {t.hero.title.highlight}
              </span>
            </span>
            <br />
            <span className="text-foreground">{t.hero.title.line2}</span>
          </motion.h1>

          {/* Subtitle */}
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="mt-8 text-lg sm:text-xl text-muted-foreground max-w-3xl mx-auto leading-relaxed"
          >
            {t.hero.subtitle}
          </motion.p>

          {/* CTA Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.3 }}
            className="mt-10 flex flex-col sm:flex-row gap-4 justify-center items-center"
          >
            <Link to={authenticated ? '/dashboard' : '/register'}>
              <Button
                size="lg"
                className="h-14 px-8 text-base font-semibold rounded-full shadow-xl shadow-primary/30 hover:shadow-primary/50 hover:scale-105 transition-all duration-300 group"
              >
                {authenticated ? t.hero.cta.primaryAuth : t.hero.cta.primary}
                <ArrowRight className="ml-2 w-5 h-5 group-hover:translate-x-1 transition-transform" />
              </Button>
            </Link>
            <a href="#how-it-works">
              <Button
                variant="outline"
                size="lg"
                className="h-14 px-8 text-base font-medium rounded-full border-2 hover:bg-muted/50 group"
              >
                <Play className="mr-2 w-5 h-5 text-primary group-hover:scale-110 transition-transform" />
                {t.hero.cta.secondary}
              </Button>
            </a>
          </motion.div>

          {/* Stats */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.4 }}
            className="mt-16 pt-10 border-t border-border/50"
          >
            <div className="grid grid-cols-3 gap-8 max-w-2xl mx-auto">
              {[
                { value: t.hero.stats.sourcesValue, label: t.hero.stats.sources, icon: Database },
                { value: t.hero.stats.securityValue, label: t.hero.stats.security, icon: ShieldCheck },
                { value: t.hero.stats.languagesValue, label: t.hero.stats.languages, icon: Globe },
              ].map((stat, i) => (
                <div key={i} className="text-center group">
                  <div className="inline-flex items-center justify-center w-12 h-12 rounded-2xl bg-primary/5 mb-3 group-hover:bg-primary/10 transition-colors">
                    <stat.icon className="w-5 h-5 text-primary" />
                  </div>
                  <div className="text-3xl sm:text-4xl font-bold text-foreground">{stat.value}</div>
                  <div className="text-sm text-muted-foreground mt-1">{stat.label}</div>
                </div>
              ))}
            </div>
          </motion.div>
        </div>

        {/* Hero Visual - Chat Widget Demo */}
        <motion.div
          initial={{ opacity: 0, y: 60 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, delay: 0.5 }}
          className="mt-20 relative max-w-4xl mx-auto"
        >
          {/* Glow Effect */}
          <div className="absolute -inset-4 bg-gradient-to-r from-orange-500/20 via-primary/20 to-orange-500/20 rounded-[2rem] blur-3xl opacity-60" />
          
          {/* Browser Frame */}
          <div className="relative bg-white/80 backdrop-blur-xl rounded-2xl shadow-2xl border border-white/50 overflow-hidden">
            {/* Browser Header */}
            <div className="flex items-center gap-2 px-6 py-4 bg-white/50 border-b border-black/5">
              <div className="flex gap-2">
                <div className="w-3 h-3 rounded-full bg-[#FF5F57]" />
                <div className="w-3 h-3 rounded-full bg-[#FEBC2E]" />
                <div className="w-3 h-3 rounded-full bg-[#28C840]" />
              </div>
              <div className="flex-1 flex justify-center">
                <div className="flex items-center gap-2 px-3 py-1 rounded-md text-sm text-gray-400 font-medium">
                  <Lock className="w-3 h-3" />
                  <span>yourwebsite.com</span>
                </div>
              </div>
            </div>

            {/* Website Content Area */}
            <div className="relative h-[400px] sm:h-[500px] bg-[#F8F9FC] p-8 sm:p-12 overflow-hidden">
              <div className="flex flex-col gap-6 opacity-60">
                {/* Header placeholders */}
                <div className="w-1/4 h-8 bg-slate-200/80 rounded-full" />
                <div className="w-full h-4 bg-slate-200/80 rounded-full" />
                
                {/* Hero text placeholders */}
                <div className="space-y-3 mt-4">
                  <div className="w-3/4 h-4 bg-slate-200/80 rounded-full" />
                  <div className="w-5/6 h-4 bg-slate-200/80 rounded-full" />
                </div>

                {/* Content blocks */}
                <div className="grid grid-cols-2 gap-6 mt-8">
                  <div className="h-32 bg-slate-200/80 rounded-3xl" />
                  <div className="h-32 bg-slate-200/80 rounded-3xl" />
                </div>
              </div>

              {/* Floating Chat Widget */}
              <motion.div
                initial={{ scale: 0.9, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: 1, duration: 0.5 }}
                className="absolute bottom-8 right-8 w-[340px] shadow-2xl rounded-[2rem]"
              >
                {/* Chat Window */}
                <div className="bg-white rounded-[2rem] overflow-hidden shadow-[0_8px_30px_rgba(0,0,0,0.12)] border border-gray-100">
                  {/* Chat Header */}
                  <div className="bg-[#FF8800] px-6 py-5 flex items-center gap-4">
                    <div className="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center backdrop-blur-sm">
                      <Bot className="w-5 h-5 text-white" />
                    </div>
                    <div>
                      <h3 className="text-white font-bold text-base leading-tight">Destek Asistanı</h3>
                      <div className="flex items-center gap-1.5 text-white/90 text-xs mt-0.5 font-medium">
                        <span className="w-2 h-2 rounded-full bg-[#4ADE80]" />
                        Çevrimiçi
                      </div>
                    </div>
                  </div>

                  {/* Chat Messages */}
                  <div className="p-6 space-y-5 bg-white">
                    {/* Bot Greeting */}
                    <div className="flex gap-3">
                      <div className="w-8 h-8 rounded-full bg-[#FF8800]/10 flex items-center justify-center shrink-0 mt-1">
                        <Bot className="w-4 h-4 text-[#FF8800]" />
                      </div>
                      <div className="bg-gray-50 rounded-2xl rounded-tl-none px-4 py-3 text-sm text-gray-700 shadow-sm border border-gray-100">
                        Merhaba! 👋 Size nasıl yardımcı olabilirim?
                      </div>
                    </div>

                    {/* User Question */}
                    <div className="flex justify-end">
                      <div className="bg-[#FF8800] text-white rounded-2xl rounded-tr-none px-4 py-3 text-sm font-medium shadow-md shadow-orange-500/20">
                        Çalışma saatleriniz nedir?
                      </div>
                    </div>

                    {/* Bot Answer */}
                    <div className="flex gap-3">
                      <div className="w-8 h-8 rounded-full bg-[#FF8800]/10 flex items-center justify-center shrink-0 mt-1">
                        <Bot className="w-4 h-4 text-[#FF8800]" />
                      </div>
                      <div className="bg-gray-50 rounded-2xl rounded-tl-none px-4 py-3 text-sm text-gray-700 shadow-sm border border-gray-100">
                        Hafta içi 09:00-18:00 arası hizmetinizdeyiz! 🕐
                      </div>
                    </div>
                  </div>

                  {/* Input Area */}
                  <div className="p-4 pt-2 pb-6 bg-white">
                    <div className="relative">
                      <input 
                        type="text" 
                        disabled
                        placeholder="Mesajınızı yazın..." 
                        className="w-full bg-gray-100 text-gray-500 text-sm rounded-full pl-5 pr-12 py-3.5 focus:outline-none"
                      />
                      <div className="absolute right-1.5 top-1.5 bottom-1.5 aspect-square rounded-full bg-[#FF8800] flex items-center justify-center shadow-lg shadow-orange-500/20">
                        <ArrowRight className="w-4 h-4 text-white" />
                      </div>
                    </div>
                  </div>
                </div>
              </motion.div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

// ============================================
// LOGOS / SOCIAL PROOF
// ============================================

const LogoCloud = () => (
  <section className="py-16 border-y border-border/50 bg-muted/30">
    <div className="max-w-7xl mx-auto px-6">
      <motion.p
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true }}
        className="text-center text-sm font-medium text-muted-foreground mb-8"
      >
        Güvenilir teknolojilerle destekleniyor
      </motion.p>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        className="flex flex-wrap justify-center items-center gap-x-12 gap-y-6"
      >
        {['OpenAI', 'PostgreSQL', 'Qdrant', 'Redis', 'Cloudflare'].map((name) => (
          <div
            key={name}
            className="flex items-center gap-2 text-muted-foreground/60 hover:text-muted-foreground transition-colors"
          >
            <div className="w-8 h-8 rounded-lg bg-muted flex items-center justify-center">
              <Cpu className="w-4 h-4" />
            </div>
            <span className="font-semibold">{name}</span>
          </div>
        ))}
      </motion.div>
    </div>
  </section>
)

// ============================================
// FEATURES SECTION
// ============================================

const FeatureCard = ({
  icon: Icon,
  title,
  description,
  index,
}: {
  icon: React.ElementType
  title: string
  description: string
  index: number
}) => (
  <motion.div
    initial={{ opacity: 0, y: 30 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay: index * 0.1 }}
    className="group relative"
  >
    <div className="relative p-6 sm:p-8 rounded-2xl bg-card border border-border/50 h-full hover:border-primary/30 hover:shadow-xl hover:shadow-primary/5 transition-all duration-500">
      {/* Icon */}
      <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary/10 to-primary/5 flex items-center justify-center mb-5 group-hover:scale-110 group-hover:from-primary/20 group-hover:to-primary/10 transition-all duration-500">
        <Icon className="w-6 h-6 text-primary" />
      </div>

      {/* Content */}
      <h3 className="text-lg font-bold mb-2 text-foreground group-hover:text-primary transition-colors">
        {title}
      </h3>
      <p className="text-sm text-muted-foreground leading-relaxed">{description}</p>

      {/* Hover Effect */}
      <div className="absolute inset-0 rounded-2xl bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none" />
    </div>
  </motion.div>
)

const Features = () => {
  const features = [
    { icon: Search, ...t.features.items.rag },
    { icon: Database, ...t.features.items.sources },
    { icon: Palette, ...t.features.items.widget },
    { icon: Zap, ...t.features.items.actions },
    { icon: ShieldCheck, ...t.features.items.guardrails },
    { icon: BarChart3, ...t.features.items.analytics },
    { icon: Headphones, ...t.features.items.handoff },
    { icon: Building2, ...t.features.items.multiTenant },
  ]

  return (
    <section id="features" className="py-24 sm:py-32 scroll-mt-20">
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center mb-16">
          <SectionBadge icon={Sparkles}>{t.features.badge}</SectionBadge>
          <SectionTitle highlight={t.features.titleHighlight}>{t.features.title}</SectionTitle>
          <SectionDescription>{t.features.subtitle}</SectionDescription>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {features.map((feature, i) => (
            <FeatureCard
              key={feature.title}
              icon={feature.icon}
              title={feature.title}
              description={feature.description}
              index={i}
            />
          ))}
        </div>
      </div>
    </section>
  )
}

// ============================================
// USE CASES SECTION
// ============================================

const UseCaseCard = ({
  icon: Icon,
  title,
  description,
  features,
  index,
}: {
  icon: React.ElementType
  title: string
  description: string
  features: string[]
  index: number
}) => (
  <motion.div
    initial={{ opacity: 0, y: 30 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay: index * 0.1 }}
    className="group relative p-8 rounded-3xl bg-card border border-border/50 hover:border-primary/30 hover:shadow-2xl hover:shadow-primary/5 transition-all duration-500"
  >
    {/* Icon */}
    <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-primary/20 to-orange-500/10 flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-500">
      <Icon className="w-7 h-7 text-primary" />
    </div>

    {/* Content */}
    <h3 className="text-2xl font-bold mb-3 text-foreground">{title}</h3>
    <p className="text-muted-foreground leading-relaxed mb-6">{description}</p>

    {/* Feature Tags */}
    <div className="flex flex-wrap gap-2">
      {features.map((feature) => (
        <span
          key={feature}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-primary/5 text-primary text-xs font-medium border border-primary/10"
        >
          <CheckCircle2 className="w-3 h-3" />
          {feature}
        </span>
      ))}
    </div>
  </motion.div>
)

const UseCases = () => {
  const cases = [
    { icon: Globe, ...t.useCases.items.website },
    { icon: Database, ...t.useCases.items.docs },
    { icon: Headphones, ...t.useCases.items.support },
    { icon: Palette, ...t.useCases.items.brand },
  ]

  return (
    <section id="use-cases" className="py-24 sm:py-32 bg-muted/30 scroll-mt-20">
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center mb-16">
          <SectionBadge icon={Users}>{t.useCases.badge}</SectionBadge>
          <SectionTitle highlight={t.useCases.titleHighlight}>{t.useCases.title}</SectionTitle>
          <SectionDescription>{t.useCases.subtitle}</SectionDescription>
        </div>

        <div className="grid md:grid-cols-2 gap-8">
          {cases.map((useCase, i) => (
            <UseCaseCard
              key={useCase.title}
              icon={useCase.icon}
              title={useCase.title}
              description={useCase.description}
              features={[...useCase.features]}
              index={i}
            />
          ))}
        </div>
      </div>
    </section>
  )
}

// ============================================
// HOW IT WORKS SECTION
// ============================================

const HowItWorks = () => {
  const [activeStep, setActiveStep] = useState(0)
  const steps = [
    { ...t.howItWorks.steps.step1, icon: Upload },
    { ...t.howItWorks.steps.step2, icon: ShieldCheck },
    { ...t.howItWorks.steps.step3, icon: Code2 },
  ]

  return (
    <section id="how-it-works" className="py-24 sm:py-32 scroll-mt-20">
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center mb-16">
          <SectionBadge icon={Zap}>{t.howItWorks.badge}</SectionBadge>
          <SectionTitle highlight={t.howItWorks.titleHighlight}>{t.howItWorks.title}</SectionTitle>
          <SectionDescription>{t.howItWorks.subtitle}</SectionDescription>
        </div>

        <div className="grid lg:grid-cols-12 gap-8 lg:gap-12">
          {/* Steps Navigation */}
          <div className="lg:col-span-5 space-y-4">
            {steps.map((step, i) => {
              const isActive = i === activeStep
              const Icon = step.icon
              return (
                <motion.button
                  key={step.number}
                  initial={{ opacity: 0, x: -20 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.1 }}
                  type="button"
                  onClick={() => setActiveStep(i)}
                  className={cn(
                    'w-full text-left p-6 rounded-2xl border transition-all duration-300',
                    isActive
                      ? 'border-primary/50 bg-primary/5 shadow-lg shadow-primary/10'
                      : 'border-border/50 bg-card hover:border-border hover:bg-muted/50',
                  )}
                >
                  <div className="flex items-start gap-4">
                    <div
                      className={cn(
                        'w-12 h-12 rounded-xl flex items-center justify-center shrink-0 transition-all duration-300',
                        isActive
                          ? 'bg-primary text-white shadow-lg shadow-primary/30'
                          : 'bg-muted text-muted-foreground',
                      )}
                    >
                      <Icon className="w-5 h-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-2 mb-1">
                        <span className="font-bold text-foreground">{step.title}</span>
                        <span
                          className={cn(
                            'text-xs font-bold tracking-wider',
                            isActive ? 'text-primary' : 'text-muted-foreground/50',
                          )}
                        >
                          {step.number}
                        </span>
                      </div>
                      <p className="text-sm text-muted-foreground line-clamp-2">{step.description}</p>
                    </div>
                  </div>
                </motion.button>
              )
            })}
          </div>

          {/* Step Content */}
          <div className="lg:col-span-7">
            <AnimatePresence mode="wait">
              <motion.div
                key={activeStep}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.3 }}
                className="p-8 rounded-3xl bg-gradient-to-br from-card to-muted/30 border border-border/50 shadow-xl"
              >
                <div className="flex items-center gap-4 mb-8">
                  <div className="w-14 h-14 rounded-2xl bg-primary/10 flex items-center justify-center">
                    {React.createElement(steps[activeStep].icon, {
                      className: 'w-7 h-7 text-primary',
                    })}
                  </div>
                  <div>
                    <span className="text-xs font-bold text-primary uppercase tracking-wider">
                      Adım {steps[activeStep].number}
                    </span>
                    <h3 className="text-2xl font-bold text-foreground">{steps[activeStep].title}</h3>
                  </div>
                </div>

                <div className="space-y-3 mb-8">
                  {steps[activeStep].bullets.map((bullet: string, i: number) => (
                    <motion.div
                      key={bullet}
                      initial={{ opacity: 0, x: 10 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: i * 0.1 }}
                      className="flex items-center gap-3 p-4 rounded-xl bg-background border border-border/50"
                    >
                      <CheckCircle2 className="w-5 h-5 text-emerald-500 shrink-0" />
                      <span className="text-sm font-medium text-foreground">{bullet}</span>
                    </motion.div>
                  ))}
                </div>

                <div className="p-5 rounded-2xl bg-primary/5 border border-primary/10">
                  <div className="flex items-center gap-2 mb-2">
                    <Sparkles className="w-4 h-4 text-primary" />
                    <span className="text-sm font-bold text-foreground">Profesyonel İpucu</span>
                  </div>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    {steps[activeStep].tip}
                  </p>
                </div>
              </motion.div>
            </AnimatePresence>
          </div>
        </div>
      </div>
    </section>
  )
}

// ============================================
// PRICING SECTION
// ============================================

const PricingCard = ({
  title,
  price,
  features,
  recommended,
  cta,
  description,
}: {
  title: string
  price: string
  features: { text: string; included: boolean }[]
  recommended?: boolean
  cta: { text: string; href?: string }
  description?: string
}) => (
  <motion.div
    initial={{ opacity: 0, y: 30 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    className={cn(
      'relative p-8 rounded-3xl border flex flex-col h-full transition-all duration-500',
      recommended
        ? 'border-primary bg-gradient-to-b from-primary/5 to-transparent shadow-xl shadow-primary/10 scale-105 z-10'
        : 'border-border/50 bg-card hover:border-border',
    )}
  >
    {recommended && (
      <div className="absolute -top-4 left-1/2 -translate-x-1/2">
        <span className="px-4 py-1.5 rounded-full bg-primary text-white text-xs font-bold shadow-lg shadow-primary/30">
          {t.pricing.mostPopular}
        </span>
      </div>
    )}

    {/* Header */}
    <div className="text-center mb-8">
      <div className="flex justify-center mb-4">
        <PlanBadge plan={title.toLowerCase() as PlanTier} size="lg" variant="soft" />
      </div>
      {description && <p className="text-sm text-muted-foreground mb-4">{description}</p>}
      <div className="flex items-baseline justify-center gap-1">
        <span className="text-4xl sm:text-5xl font-bold tracking-tight">{price}</span>
        <span className="text-muted-foreground">{t.pricing.perMonth}</span>
      </div>
    </div>

    {/* Features */}
    <ul className="space-y-3 flex-1 mb-8">
      {features.map((f, i) => (
        <li key={i} className="flex items-center gap-3 text-sm">
          {f.included ? (
            <div className="w-5 h-5 rounded-full bg-emerald-500/10 flex items-center justify-center shrink-0">
              <Check className="w-3 h-3 text-emerald-500" />
            </div>
          ) : (
            <div className="w-5 h-5 rounded-full bg-muted flex items-center justify-center shrink-0">
              <Minus className="w-3 h-3 text-muted-foreground/50" />
            </div>
          )}
          <span className={f.included ? 'text-foreground' : 'text-muted-foreground/50'}>
            {f.text}
          </span>
        </li>
      ))}
    </ul>

    {/* CTA */}
    {cta.href ? (
      <Link to={cta.href} className="w-full">
        <Button
          variant={recommended ? 'default' : 'outline'}
          className={cn(
            'w-full h-12 font-semibold rounded-xl transition-all',
            recommended && 'shadow-lg shadow-primary/25 hover:shadow-primary/40',
          )}
        >
          {cta.text}
          <ArrowRight className="ml-2 w-4 h-4" />
        </Button>
      </Link>
    ) : (
      <Button
        variant={recommended ? 'default' : 'outline'}
        disabled
        className={cn(
          'w-full h-12 font-semibold rounded-xl transition-all',
          recommended && 'shadow-lg shadow-primary/25 hover:shadow-primary/40',
        )}
      >
        {cta.text}
      </Button>
    )}
  </motion.div>
)

const Pricing = ({ authenticated }: { authenticated: boolean }) => {
  const { data: apiPlans, isLoading } = usePlans()

  const plans =
    apiPlans?.map((p) => {
      const isFree = p.code === 'free'
      const isUltra = p.code === 'ultra'
      const isPro = p.code === 'pro'

      return {
        title: p.name || p.code.charAt(0).toUpperCase() + p.code.slice(1),
        price: p.price === 0 ? '0 TL' : `${p.price} ${p.currency}`,
        recommended: isFree,
        description: isFree
          ? t.pricing.plans.free.description
          : isPro
            ? t.pricing.plans.pro.description
            : t.pricing.plans.ultra.description,
        cta: isFree
          ? {
              text: authenticated
                ? t.pricing.cta.authenticated
                : t.pricing.cta.free,
              href: authenticated ? '/dashboard' : '/register',
            }
          : {
              text: isPro ? t.pricing.cta.pro : t.pricing.cta.ultra,
            },
        features: [
          { text: `${p.limits.max_chatbots} ${t.pricing.features.chatbots}`, included: true },
          {
            text: `${p.limits.max_monthly_ingestions.toLocaleString('tr-TR')} ${t.pricing.features.tokens}`,
            included: true,
          },
          { text: `${p.features.scraping.max_urls_per_bot} ${t.pricing.features.sites}`, included: true },
          { text: `${p.features.files.max_files_per_bot} ${t.pricing.features.pdfs}`, included: true },
          {
            text: isUltra ? 'GPT-4o & GPT-5' : isPro ? 'GPT-4o & GPT-4o Mini' : 'GPT-4o Mini',
            included: true,
          },
          { text: t.pricing.features.dynamicScraping, included: isPro || isUltra },
          { text: t.pricing.features.guardrails, included: isPro || isUltra },
          { text: t.pricing.features.smartFallback, included: isPro || isUltra },
          { text: t.pricing.features.handoff, included: isUltra },
          { text: t.pricing.features.branding, included: isUltra },
        ],
      }
    }) || []

  if (isLoading) {
    return (
      <section id="pricing" className="py-24 sm:py-32 bg-muted/30 scroll-mt-20">
        <div className="max-w-7xl mx-auto px-6 text-center">
          <SectionBadge>{t.pricing.badge}</SectionBadge>
          <div className="mt-16 grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-[500px] bg-card animate-pulse rounded-3xl" />
            ))}
          </div>
        </div>
      </section>
    )
  }

  return (
    <section id="pricing" className="py-24 sm:py-32 bg-muted/30 scroll-mt-20">
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center mb-16">
          <SectionBadge icon={Star}>{t.pricing.badge}</SectionBadge>
          <SectionTitle highlight={t.pricing.titleHighlight}>{t.pricing.title}</SectionTitle>
          <SectionDescription>{t.pricing.subtitle}</SectionDescription>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto items-stretch">
          {plans.map((plan, i) => (
            <PricingCard key={i} {...plan} />
          ))}
        </div>
      </div>
    </section>
  )
}

// ============================================
// SECURITY SECTION
// ============================================

const Security = () => {
  const securityItems = [
    { icon: Key, ...t.security.items.jwt },
    { icon: Shield, ...t.security.items.ssrf },
    { icon: Lock, ...t.security.items.encryption },
    { icon: Eye, ...t.security.items.rls },
    { icon: Activity, ...t.security.items.rateLimit },
    { icon: Clock, ...t.security.items.audit },
  ]

  return (
    <section className="py-24 sm:py-32">
      <div className="max-w-7xl mx-auto px-6">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          {/* Content */}
          <div>
            <SectionBadge icon={ShieldCheck}>{t.security.badge}</SectionBadge>
            <motion.h2
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              className="text-4xl sm:text-5xl font-bold tracking-tight mb-6"
            >
              {t.security.title} <span className="text-primary">{t.security.titleHighlight}</span>
            </motion.h2>
            <motion.p
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: 0.1 }}
              className="text-lg text-muted-foreground mb-10"
            >
              {t.security.subtitle}
            </motion.p>

            <div className="grid sm:grid-cols-2 gap-6">
              {securityItems.map((item, i) => (
                <motion.div
                  key={item.title}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.1 }}
                  className="flex items-start gap-4"
                >
                  <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                    <item.icon className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <h4 className="font-bold text-foreground mb-1">{item.title}</h4>
                    <p className="text-sm text-muted-foreground leading-relaxed">{item.description}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>

          {/* Visual */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            className="relative"
          >
            <div className="absolute -inset-4 bg-gradient-to-r from-primary/10 via-transparent to-primary/10 rounded-3xl blur-2xl" />
            <div className="relative p-8 rounded-3xl bg-card border border-border/50 shadow-xl">
              <div className="flex items-center gap-4 mb-6">
                <div className="w-12 h-12 rounded-2xl bg-emerald-500/10 flex items-center justify-center">
                  <ShieldCheck className="w-6 h-6 text-emerald-500" />
                </div>
                <div>
                  <div className="font-bold text-foreground">Güvenlik Durumu</div>
                  <div className="text-sm text-emerald-500 flex items-center gap-1.5">
                    <span className="relative flex h-2 w-2">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                      <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                    </span>
                    Tüm sistemler aktif
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                {['KVKK Uyumlu', 'GDPR Uyumlu', 'SSL/TLS Şifreleme', 'SOC 2 Ready'].map((item, i) => (
                  <motion.div
                    key={item}
                    initial={{ opacity: 0, x: -10 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    viewport={{ once: true }}
                    transition={{ delay: 0.3 + i * 0.1 }}
                    className="flex items-center gap-3 p-3 rounded-xl bg-muted/50"
                  >
                    <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                    <span className="text-sm font-medium text-foreground">{item}</span>
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  )
}

// ============================================
// FAQ SECTION
// ============================================

const FAQ = () => {
  const [openIndex, setOpenIndex] = useState<number | null>(0)

  return (
    <section id="faq" className="py-24 sm:py-32 bg-muted/30 scroll-mt-20">
      <div className="max-w-4xl mx-auto px-6">
        <div className="text-center mb-16">
          <SectionBadge icon={MessageCircle}>{t.faq.badge}</SectionBadge>
          <SectionTitle highlight={t.faq.titleHighlight}>{t.faq.title}</SectionTitle>
          <SectionDescription>{t.faq.subtitle}</SectionDescription>
        </div>

        <div className="space-y-4">
          {t.faq.items.map((faq, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 10 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.05 }}
              className="border border-border/50 rounded-2xl bg-card overflow-hidden hover:border-primary/20 transition-colors"
            >
              <button
                type="button"
                onClick={() => setOpenIndex(openIndex === i ? null : i)}
                className="w-full flex items-center justify-between p-6 text-left"
              >
                <span className="font-bold text-lg pr-4">{faq.question}</span>
                <div
                  className={cn(
                    'w-8 h-8 rounded-full bg-muted flex items-center justify-center shrink-0 transition-transform duration-300',
                    openIndex === i && 'rotate-45 bg-primary/10',
                  )}
                >
                  <ChevronRight
                    className={cn(
                      'w-4 h-4 rotate-90 transition-colors',
                      openIndex === i ? 'text-primary' : 'text-muted-foreground',
                    )}
                  />
                </div>
              </button>
              <AnimatePresence>
                {openIndex === i && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.3 }}
                    className="overflow-hidden"
                  >
                    <div className="px-6 pb-6 pt-0 text-muted-foreground leading-relaxed border-t border-border/50 mt-0 pt-4">
                      {faq.answer}
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}

// ============================================
// CTA SECTION
// ============================================

const CTASection = ({ authenticated }: { authenticated: boolean }) => (
  <section className="py-24 sm:py-32 relative overflow-hidden">
    {/* Background */}
    <div className="absolute inset-0 -z-10">
      <div className="absolute inset-0 bg-gradient-to-b from-background via-primary/5 to-background" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[400px] bg-primary/10 blur-[150px] rounded-full" />
    </div>

    <div className="max-w-4xl mx-auto px-6 text-center">
      <motion.div
        initial={{ opacity: 0, y: 30 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
      >
        <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight mb-6">
          {t.cta.title}
          <br />
          <span className="text-primary">{t.cta.titleHighlight}</span>
        </h2>
        <p className="text-lg sm:text-xl text-muted-foreground mb-10 max-w-2xl mx-auto">
          {t.cta.subtitle}
        </p>

        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <Link to={authenticated ? '/dashboard' : '/register'}>
            <Button
              size="lg"
              className="h-14 px-10 text-lg font-bold rounded-full shadow-xl shadow-primary/30 hover:shadow-primary/50 hover:scale-105 transition-all duration-300 group"
            >
              {authenticated ? t.cta.buttonAuth : t.cta.button}
              <ArrowRight className="ml-2 w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </Button>
          </Link>
          <span className="text-sm text-muted-foreground">{t.cta.note}</span>
        </div>
      </motion.div>
    </div>
  </section>
)

// ============================================
// FOOTER
// ============================================

const Footer = ({ authenticated }: { authenticated: boolean }) => (
  <footer className="bg-foreground text-background py-16 sm:py-20">
    <div className="max-w-7xl mx-auto px-6">
      <div className="grid md:grid-cols-12 gap-12 mb-12">
        {/* Brand */}
        <div className="md:col-span-5">
          <div className="flex items-center gap-3 mb-6">
            <div className="bg-primary p-2.5 rounded-xl">
              <Bot className="w-5 h-5 text-white" />
            </div>
            <span className="font-bold text-2xl">botla.app</span>
          </div>
          <p className="text-background/60 leading-relaxed max-w-sm">{t.footer.description}</p>
        </div>

        {/* Links */}
        <div className="md:col-span-2 md:col-start-7">
          <h4 className="font-bold mb-5">{t.footer.product.title}</h4>
          <ul className="space-y-3 text-background/60 text-sm">
            <li>
              <a href="#features" className="hover:text-primary transition-colors">
                {t.footer.product.features}
              </a>
            </li>
            <li>
              <a href="#pricing" className="hover:text-primary transition-colors">
                {t.footer.product.pricing}
              </a>
            </li>
            {authenticated ? (
              <li>
                <Link to="/dashboard" className="hover:text-primary transition-colors">
                  {t.footer.product.dashboard}
                </Link>
              </li>
            ) : (
              <>
                <li>
                  <Link to="/login" className="hover:text-primary transition-colors">
                    {t.footer.product.login}
                  </Link>
                </li>
                <li>
                  <Link to="/register" className="hover:text-primary transition-colors">
                    {t.footer.product.register}
                  </Link>
                </li>
              </>
            )}
          </ul>
        </div>

        <div className="md:col-span-2">
          <h4 className="font-bold mb-5">{t.footer.company.title}</h4>
          <ul className="space-y-3 text-background/60 text-sm">
            <li>
              <a href="#" className="hover:text-primary transition-colors">
                {t.footer.company.about}
              </a>
            </li>
            <li>
              <a href="#" className="hover:text-primary transition-colors">
                {t.footer.company.blog}
              </a>
            </li>
            <li>
              <a href="#" className="hover:text-primary transition-colors">
                {t.footer.company.contact}
              </a>
            </li>
          </ul>
        </div>

        <div className="md:col-span-2">
          <h4 className="font-bold mb-5">{t.footer.legal.title}</h4>
          <ul className="space-y-3 text-background/60 text-sm">
            <li>
              <a href="#" className="hover:text-primary transition-colors">
                {t.footer.legal.privacy}
              </a>
            </li>
            <li>
              <a href="#" className="hover:text-primary transition-colors">
                {t.footer.legal.terms}
              </a>
            </li>
            <li>
              <a href="#" className="hover:text-primary transition-colors">
                {t.footer.legal.kvkk}
              </a>
            </li>
          </ul>
        </div>
      </div>

      {/* Bottom */}
      <div className="pt-8 border-t border-background/10 flex flex-col md:flex-row justify-between items-center gap-4 text-background/40 text-sm">
        <p>{t.footer.copyright.replace('{year}', new Date().getFullYear().toString())}</p>
        <p>{t.footer.madeWith}</p>
      </div>
    </div>
  </footer>
)

// ============================================
// MAIN COMPONENT
// ============================================

export default function LandingPage() {
  const { isAuthenticated: authenticated } = useAuth()

  useEffect(() => {
    const hash = window.location.hash
    if (!hash) return
    const id = hash.startsWith('#') ? hash.slice(1) : hash
    if (!id) return
    const el = document.getElementById(id)
    if (!el) return
    const t = window.setTimeout(() => {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }, 50)
    return () => window.clearTimeout(t)
  }, [])

  useEffect(() => {
    // Don't show widget for authenticated users
    if (authenticated) {
      // Clean up widget if user becomes authenticated
      const existingScript = document.querySelector('script[data-bot]')
      if (existingScript) existingScript.remove()
      const widgetHost = document.getElementById('chatbot-widget-host')
      if (widgetHost) widgetHost.remove()
      return
    }

    const chatbotId = import.meta.env.VITE_LANDING_CHATBOT_ID
    const widgetUrl = import.meta.env.VITE_WIDGET_SCRIPT_URL

    if (!chatbotId || !widgetUrl) return

    const existingScript = document.querySelector(`script[data-bot="${chatbotId}"]`)
    if (existingScript) return

    const script = document.createElement('script')
    script.src = widgetUrl
    script.type = 'module'
    script.setAttribute('data-bot', chatbotId)
    script.async = true
    document.body.appendChild(script)

    // No cleanup - widget persists while unauthenticated
  }, [authenticated])

  return (
    <div className="relative min-h-screen bg-background font-sans antialiased">
      <Navbar authenticated={authenticated} />
      <main>
        <Hero authenticated={authenticated} />
        <LogoCloud />
        <Features />
        <UseCases />
        <HowItWorks />
        <Pricing authenticated={authenticated} />
        <Security />
        <FAQ />
        <CTASection authenticated={authenticated} />
      </main>
      <Footer authenticated={authenticated} />
    </div>
  )
}
