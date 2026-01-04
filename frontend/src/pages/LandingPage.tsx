import React, { useEffect, useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
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
  MessagesSquare,
  Search,
  BarChart3,
  Lock,
  Palette,
  Headphones,
  Building2,
  ShoppingCart,
  Code2,
  BookOpen,
  Play,
  ChevronRight,
  Shield,
  Key,
  Eye,
  Activity,
  Clock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PlanBadge, PlanTier } from '@/components/ui/plan-badge'
import { cn } from '@/lib/utils'
import { usePlans } from '@/hooks/queries/usePlans'
import { landing } from '@/i18n/landing'

const t = landing

// --- Premium Visual Primitives ---

const Noise = () => (
  <div className="fixed inset-0 w-full h-full pointer-events-none opacity-[0.02] z-[100] mix-blend-overlay">
    <svg viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg">
      <filter id="noise">
        <feTurbulence type="fractalNoise" baseFrequency="0.65" numOctaves="3" stitchTiles="stitch" />
      </filter>
      <rect width="100%" height="100%" filter="url(#noise)" />
    </svg>
  </div>
)

const MouseHighlight = () => {
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 })
  useEffect(() => {
    const handleMove = (e: MouseEvent) => setMousePos({ x: e.clientX, y: e.clientY })
    window.addEventListener('mousemove', handleMove)
    return () => window.removeEventListener('mousemove', handleMove)
  }, [])

  return (
    <div 
      className="pointer-events-none fixed inset-0 z-30 transition-opacity duration-300"
      style={{
        background: `radial-gradient(600px circle at ${mousePos.x}px ${mousePos.y}px, rgba(245, 158, 11, 0.04), transparent 80%)`
      }}
    />
  )
}

// --- Navbar ---

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
    <nav className={cn(
      "fixed inset-x-0 top-0 z-50 transition-all duration-300 border-b",
      scrolled 
        ? "bg-background/80 backdrop-blur-xl border-border/40 py-3" 
        : "bg-transparent border-transparent py-5"
    )}>
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center gap-3 cursor-pointer group" onClick={handleLogoClick}>
            <div className="bg-primary p-2.5 rounded-xl shadow-glow transition-transform group-hover:scale-105">
              <Bot className="w-5 h-5 text-primary-foreground" />
            </div>
            <span className="font-bold text-xl tracking-tight text-foreground">botla.app</span>
          </div>

          {/* Desktop Nav */}
          <div className="hidden lg:flex items-center gap-1 bg-muted/40 p-1 rounded-full border border-border/40">
            {links.map((link) => (
              <a
                key={link.name}
                href={link.href}
                onClick={(e) => handleScroll(e, link.href)}
                className="px-4 py-1.5 text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-background/80 rounded-full transition-all"
                >
                {link.name}
              </a>
            ))}
          </div>

          <div className="hidden lg:flex items-center gap-3">
            {authenticated ? (
              <Link to="/dashboard">
                <Button className="rounded-full px-6 font-semibold shadow-sm">{t.nav.dashboard}</Button>
              </Link>
            ) : (
              <>
                <Link to="/login">
                  <Button variant="ghost" className="rounded-full px-6 font-medium text-muted-foreground hover:text-foreground">
                    {t.nav.login}
                  </Button>
                </Link>
                <Link to="/register">
                  <Button className="rounded-full px-6 font-semibold shadow-glow transition-all hover:scale-105">{t.nav.register}</Button>
                </Link>
              </>
            )}
          </div>

          {/* Mobile Menu Button */}
          <div className="lg:hidden">
            <button onClick={() => setIsOpen(!isOpen)} className="text-foreground p-2 rounded-full hover:bg-muted transition-colors">
              {isOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Nav */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: -10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: -10 }}
            className="lg:hidden fixed top-[80px] inset-x-4 glass p-8 rounded-3xl shadow-xl z-50 border border-white/20"
          >
            <div className="space-y-6">
              {links.map((link) => (
                <a
                  key={link.name}
                  href={link.href}
                  onClick={(e) => handleScroll(e, link.href)}
                  className="block text-xl font-semibold text-foreground hover:text-primary transition-colors"
                >
                  {link.name}
                </a>
              ))}
              <hr className="border-border/10" />
              <div className="flex flex-col gap-4">
                {authenticated ? (
                  <Link to="/dashboard" onClick={() => setIsOpen(false)}>
                    <Button className="w-full h-14 rounded-2xl text-lg">{t.nav.goToDashboard}</Button>
                  </Link>
                ) : (
                  <>
                    <Link to="/login" onClick={() => setIsOpen(false)}>
                      <Button variant="outline" className="w-full h-14 rounded-2xl text-lg border-border/40">
                        {t.nav.login}
                      </Button>
                    </Link>
                    <Link to="/register" onClick={() => setIsOpen(false)}>
                      <Button className="w-full h-14 rounded-2xl text-lg shadow-glow">{t.nav.register}</Button>
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

// --- Hero Section ---

const Hero = ({ authenticated }: { authenticated: boolean }) => {
  return (
    <section className="relative pt-32 pb-24 lg:pt-48 lg:pb-32 overflow-hidden">
      {/* Background Elements */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-full -z-10 pointer-events-none">
        <div className="absolute top-[-10%] left-1/2 -translate-x-1/2 w-[600px] h-[600px] bg-primary/10 blur-[150px] rounded-full animate-pulse" />
        <div className="absolute top-[20%] left-[30%] w-[300px] h-[300px] bg-orange-600/5 blur-[100px] rounded-full animate-bounce" style={{ animationDuration: '8s' }} />
      </div>

      {/* Floating Elements */}
      <div className="absolute inset-0 -z-5 pointer-events-none overflow-hidden opacity-15">
        {[...Array(5)].map((_, i) => (
          <motion.div
            key={i}
            initial={{ opacity: 0, scale: 0 }}
            animate={{ 
              opacity: [0, 1, 0], 
              scale: [0.5, 1, 0.5],
              y: [0, -80, 0],
            }}
            transition={{ duration: 8 + i * 2, repeat: Infinity, delay: i * 1.5 }}
            className="absolute rounded-full border border-primary/20 bg-primary/5 flex items-center justify-center p-2"
            style={{ 
              top: `${25 + i * 12}%`, 
              left: `${10 + i * 16}%`,
            }}
          >
            <div className="w-1.5 h-1.5 rounded-full bg-primary" />
          </motion.div>
        ))}
      </div>

      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10 relative">
        <div className="flex flex-col items-center text-center max-w-5xl mx-auto">
          {/* Badge */}
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            className="inline-flex items-center gap-3 px-5 py-2.5 rounded-full glass border border-white/30 text-foreground text-xs font-bold uppercase tracking-widest mb-10 shadow-xl shadow-primary/5"
          >
            <div className="flex -space-x-2">
              <div className="w-5 h-5 rounded-full border-2 border-background bg-emerald-400" />
              <div className="w-5 h-5 rounded-full border-2 border-background bg-primary" />
              <div className="w-5 h-5 rounded-full border-2 border-background bg-violet-400" />
            </div>
            <span className="opacity-80">{t.hero.badge}</span>
            <div className="flex items-center gap-1.5">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              <span className="text-emerald-500 text-[10px]">{t.hero.badgeLive}</span>
            </div>
          </motion.div>

          {/* Title */}
          <motion.h1
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1, duration: 1, ease: [0.16, 1, 0.3, 1] }}
            className="text-5xl sm:text-6xl lg:text-8xl font-bold tracking-tight text-foreground mb-8 leading-[1.05]"
          >
            {t.hero.title.line1} <br />
            <span className="relative inline-block">
              <span className="relative z-10 bg-clip-text text-transparent bg-gradient-to-r from-primary via-orange-500 to-primary">
                {t.hero.title.highlight}
              </span>
            </span>
            <br />{t.hero.title.line2}
          </motion.h1>

          {/* Subtitle */}
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2, duration: 1, ease: [0.16, 1, 0.3, 1] }}
            className="text-lg sm:text-xl text-muted-foreground/80 mb-12 max-w-3xl leading-relaxed"
          >
            {t.hero.subtitle}
          </motion.p>

          {/* CTA Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3, duration: 1 }}
            className="flex flex-col sm:flex-row gap-4 items-center mb-16"
          >
            <Link to={authenticated ? "/dashboard" : "/register"}>
              <Button size="lg" className="h-14 px-10 text-lg font-bold rounded-2xl shadow-glow group hover:scale-105 transition-all">
                {authenticated ? t.hero.cta.primaryAuth : t.hero.cta.primary}
                <ArrowRight className="ml-2 w-5 h-5 transition-transform group-hover:translate-x-1" />
              </Button>
            </Link>
            <a href="#how-it-works">
              <Button
                variant="outline"
                size="lg"
                className="h-14 px-10 text-lg font-medium rounded-2xl border-border/40 hover:bg-muted/50 transition-all group"
              >
                <Play className="mr-2 w-5 h-5 text-primary" />
                {t.hero.cta.secondary}
              </Button>
            </a>
          </motion.div>

          {/* Stats */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4, duration: 1 }}
            className="grid grid-cols-3 gap-8 sm:gap-16 pt-8 border-t border-border/30"
          >
            <div className="text-center">
              <div className="text-3xl sm:text-4xl font-bold text-foreground mb-1">{t.hero.stats.sourcesValue}</div>
              <div className="text-sm text-muted-foreground">{t.hero.stats.sources}</div>
            </div>
            <div className="text-center">
              <div className="text-3xl sm:text-4xl font-bold text-primary mb-1">{t.hero.stats.securityValue}</div>
              <div className="text-sm text-muted-foreground">{t.hero.stats.security}</div>
            </div>
            <div className="text-center">
              <div className="text-3xl sm:text-4xl font-bold text-foreground mb-1">{t.hero.stats.languagesValue}</div>
              <div className="text-sm text-muted-foreground">{t.hero.stats.languages}</div>
            </div>
          </motion.div>
        </div>

        {/* Chat Widget Demo */}
        <motion.div
          initial={{ opacity: 0, scale: 0.98, y: 60 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          transition={{ delay: 0.5, duration: 1.2, ease: [0.16, 1, 0.3, 1] }}
          className="relative max-w-lg mx-auto mt-16 group"
        >
          <div className="absolute -inset-4 bg-gradient-to-r from-primary/20 via-orange-500/10 to-primary/20 rounded-[3rem] blur-3xl opacity-50 group-hover:opacity-80 transition-opacity duration-1000" />
          
          {/* Chat Widget Container */}
          <div className="relative rounded-3xl shadow-2xl shadow-primary/20 overflow-hidden border border-border/50 bg-card">
            {/* Widget Header */}
            <div className="bg-gradient-to-r from-primary to-orange-500 px-6 py-4 flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center">
                <Bot className="w-5 h-5 text-white" />
              </div>
              <div className="flex-1">
                <div className="text-white font-semibold text-sm">Müşteri Asistanı</div>
                <div className="text-white/70 text-xs flex items-center gap-1.5">
                  <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                  Çevrimiçi
                </div>
              </div>
              <div className="flex gap-1.5">
                <div className="w-2.5 h-2.5 rounded-full bg-white/30" />
                <div className="w-2.5 h-2.5 rounded-full bg-white/30" />
              </div>
            </div>

            {/* Chat Messages */}
            <div className="p-5 space-y-4 bg-gradient-to-b from-muted/30 to-background min-h-[320px]">
              {/* Bot Welcome */}
              <motion.div 
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.8 }}
                className="flex gap-3"
              >
                <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
                  <Bot className="w-4 h-4 text-primary" />
                </div>
                <div className="bg-card rounded-2xl rounded-tl-md px-4 py-3 shadow-sm border border-border/50 max-w-[85%]">
                  <p className="text-sm text-foreground">
                    Merhaba! 👋 Size nasıl yardımcı olabilirim?
                  </p>
                </div>
              </motion.div>

              {/* User Question */}
              <motion.div 
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 1.5 }}
                className="flex gap-3 justify-end"
              >
                <div className="bg-primary text-primary-foreground rounded-2xl rounded-tr-md px-4 py-3 shadow-sm max-w-[85%]">
                  <p className="text-sm">Kargo takip numaram nedir?</p>
                </div>
              </motion.div>

              {/* Bot Response with Typing Animation */}
              <motion.div 
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 2.2 }}
                className="flex gap-3"
              >
                <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
                  <Bot className="w-4 h-4 text-primary" />
                </div>
                <div className="bg-card rounded-2xl rounded-tl-md px-4 py-3 shadow-sm border border-border/50 max-w-[85%]">
                  <p className="text-sm text-foreground">
                    Siparişinizin kargo takip numarası: <span className="font-mono font-semibold text-primary">TR1234567890</span>
                  </p>
                  <p className="text-sm text-foreground mt-2">
                    📦 Kargo şu an <span className="font-medium">dağıtımda</span> - Bugün teslim edilecek.
                  </p>
                </div>
              </motion.div>

              {/* Suggestions */}
              <motion.div 
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 2.8 }}
                className="flex gap-2 flex-wrap pt-2"
              >
                {['İade nasıl yapılır?', 'Ödeme seçenekleri', 'İletişim'].map((suggestion) => (
                  <span
                    key={suggestion}
                    className="px-3 py-1.5 text-xs font-medium bg-primary/5 text-primary border border-primary/20 rounded-full cursor-pointer hover:bg-primary/10 transition-colors"
                  >
                    {suggestion}
                  </span>
                ))}
              </motion.div>
            </div>

            {/* Input Area */}
            <div className="px-4 py-3 border-t border-border/50 bg-card">
              <div className="flex items-center gap-2 bg-muted/50 rounded-xl px-4 py-2.5">
                <span className="text-sm text-muted-foreground flex-1">Mesajınızı yazın...</span>
                <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
                  <ArrowRight className="w-4 h-4 text-primary-foreground" />
                </div>
              </div>
            </div>
          </div>

          {/* Decorative elements */}
          <div className="absolute -right-8 top-1/4 w-16 h-16 bg-primary/10 rounded-full blur-2xl" />
          <div className="absolute -left-8 bottom-1/4 w-20 h-20 bg-orange-500/10 rounded-full blur-2xl" />
        </motion.div>
      </div>
    </section>
  )
}

// --- Features Section ---

const Badge = ({ children }: { children: React.ReactNode }) => (
  <span className="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-xs font-bold text-primary tracking-widest uppercase ring-1 ring-inset ring-primary/20 shadow-sm shadow-primary/5">
    {children}
  </span>
)

const FeatureCard = ({ icon: Icon, title, description, delay, gradient }: {
  icon: React.ElementType
  title: string
  description: string
  delay: number
  gradient?: boolean
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay, duration: 0.6 }}
    className={cn(
      "group relative overflow-hidden rounded-3xl bg-card border border-border/50 p-8 hover:border-primary/40 transition-all duration-500",
      gradient && "bg-gradient-to-br from-card to-primary/5"
    )}
  >
    <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none">
      <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/50 to-transparent" />
    </div>

    <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center mb-5 group-hover:scale-110 transition-transform duration-500">
      <Icon className="w-6 h-6 text-primary" />
    </div>
    <h3 className="text-xl font-bold mb-3 tracking-tight">{title}</h3>
    <p className="text-muted-foreground/80 leading-relaxed text-sm">{description}</p>
  </motion.div>
)

const Features = () => {
  const features = [
    { icon: Search, ...t.features.items.rag, gradient: true },
    { icon: Database, ...t.features.items.sources },
    { icon: Palette, ...t.features.items.widget },
    { icon: Zap, ...t.features.items.actions },
    { icon: ShieldCheck, ...t.features.items.guardrails, gradient: true },
    { icon: BarChart3, ...t.features.items.analytics },
    { icon: Headphones, ...t.features.items.handoff },
    { icon: Building2, ...t.features.items.multiTenant },
  ]

  return (
    <section id="features" className="py-32 bg-secondary/30 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center max-w-3xl mx-auto mb-20">
          <Badge>{t.features.badge}</Badge>
          <h2 className="text-4xl md:text-6xl font-bold mt-6 mb-6 tracking-tight">
            {t.features.title} <span className="text-primary">{t.features.titleHighlight}</span>
          </h2>
          <p className="text-lg text-muted-foreground leading-relaxed">
            {t.features.subtitle}
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {features.map((feature, i) => (
            <FeatureCard
              key={feature.title}
              icon={feature.icon}
              title={feature.title}
              description={feature.description}
              delay={i * 0.1}
              gradient={feature.gradient}
            />
          ))}
        </div>
      </div>
    </section>
  )
}

// --- Use Cases Section ---

const UseCaseCard = ({ icon: Icon, title, description, features, delay }: {
  icon: React.ElementType
  title: string
  description: string
  features: string[]
  delay: number
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay, duration: 0.6 }}
    className="group relative rounded-3xl bg-card border border-border/50 p-8 hover:border-primary/40 hover:shadow-xl hover:shadow-primary/5 transition-all duration-500"
  >
    <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-primary/20 to-primary/5 flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-500">
      <Icon className="w-7 h-7 text-primary" />
    </div>
    <h3 className="text-2xl font-bold mb-3 tracking-tight">{title}</h3>
    <p className="text-muted-foreground leading-relaxed mb-6">{description}</p>
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
    { icon: ShoppingCart, ...t.useCases.items.ecommerce },
    { icon: Code2, ...t.useCases.items.saas },
    { icon: Headphones, ...t.useCases.items.support },
    { icon: BookOpen, ...t.useCases.items.internal },
  ]

  return (
    <section id="use-cases" className="py-32 bg-background relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center max-w-3xl mx-auto mb-20">
          <Badge>{t.useCases.badge}</Badge>
          <h2 className="text-4xl md:text-6xl font-bold mt-6 mb-6 tracking-tight">
            {t.useCases.title} <span className="text-primary">{t.useCases.titleHighlight}</span>
          </h2>
          <p className="text-lg text-muted-foreground leading-relaxed">
            {t.useCases.subtitle}
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {cases.map((useCase, i) => (
            <UseCaseCard
              key={useCase.title}
              icon={useCase.icon}
              title={useCase.title}
              description={useCase.description}
              features={useCase.features}
              delay={i * 0.15}
            />
          ))}
        </div>
      </div>
    </section>
  )
}

// --- How It Works Section ---

const HowItWorks = () => {
  const [activeStep, setActiveStep] = useState(0)
  const steps = [
    { ...t.howItWorks.steps.step1, icon: Database },
    { ...t.howItWorks.steps.step2, icon: ShieldCheck },
    { ...t.howItWorks.steps.step3, icon: MessagesSquare },
  ]

  const active = steps[activeStep]
  const ActiveIcon = active.icon

  return (
    <section id="how-it-works" className="scroll-mt-24 py-32 bg-secondary/20 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center max-w-3xl mx-auto mb-20">
          <Badge>{t.howItWorks.badge}</Badge>
          <h2 className="text-4xl md:text-6xl font-bold mt-6 mb-6 tracking-tight">
            {t.howItWorks.title} <span className="text-primary">{t.howItWorks.titleHighlight}</span>
          </h2>
          <p className="text-lg text-muted-foreground leading-relaxed">
            {t.howItWorks.subtitle}
          </p>
        </div>

        <div className="grid lg:grid-cols-12 gap-12 items-start">
          <div className="lg:col-span-5 space-y-4">
            {steps.map((step, i) => {
              const isActive = i === activeStep
              const Icon = step.icon
              return (
                <button
                  key={step.number}
                  type="button"
                  onClick={() => setActiveStep(i)}
                  onMouseEnter={() => setActiveStep(i)}
                  className={cn(
                    "w-full text-left relative rounded-[2rem] border transition-all duration-300 p-6 sm:p-8",
                    isActive 
                      ? "border-primary/40 bg-card shadow-xl shadow-primary/5 -translate-x-2" 
                      : "border-border/40 bg-card/40 hover:bg-card hover:border-border"
                  )}
                >
                  <div className="flex items-start gap-5">
                    <div
                      className={cn(
                        "w-12 h-12 rounded-xl flex items-center justify-center border transition-all duration-300 shrink-0",
                        isActive 
                          ? "bg-primary text-primary-foreground border-primary shadow-glow" 
                          : "bg-muted text-muted-foreground border-border"
                      )}
                    >
                      <Icon className="w-6 h-6" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-3 mb-2">
                        <div className="font-bold text-lg text-foreground">{step.title}</div>
                        <div className={cn("text-sm font-bold opacity-30 tracking-widest", isActive && "text-primary opacity-100")}>
                          {step.number}
                        </div>
                      </div>
                      <div className="text-muted-foreground text-sm leading-relaxed">
                        {step.description}
                      </div>
                    </div>
                  </div>
                </button>
              )
            })}
          </div>

          <div className="lg:col-span-7 lg:sticky lg:top-32">
            <AnimatePresence mode="wait">
              <motion.div
                key={activeStep}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.4, ease: "easeOut" }}
                className="rounded-[2.5rem] glass shadow-2xl p-8 sm:p-10 border border-white/30"
              >
                <div className="flex items-center gap-5 mb-8">
                  <div className="w-14 h-14 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary">
                    <ActiveIcon className="w-7 h-7" />
                  </div>
                  <div>
                    <div className="text-sm font-bold text-primary uppercase tracking-[0.2em]">Adım {active.number}</div>
                    <div className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">{active.title}</div>
                  </div>
                </div>

                <div className="space-y-3 mb-8">
                  {active.bullets.map((bullet: string) => (
                    <div
                      key={bullet}
                      className="rounded-2xl border border-border/40 bg-background/50 px-5 py-4 flex items-center gap-3 text-sm font-medium text-foreground transition-all hover:bg-background hover:border-primary/30"
                    >
                      <CheckCircle2 className="w-5 h-5 text-emerald-500 shrink-0" />
                      {bullet}
                    </div>
                  ))}
                </div>

                <div className="rounded-2xl bg-primary/5 p-5 border border-primary/10">
                  <div className="flex items-center gap-2 text-sm font-bold text-foreground mb-2">
                    <Sparkles className="w-4 h-4 text-primary" />
                    Profesyonel İpucu
                  </div>
                  <div className="text-muted-foreground text-sm leading-relaxed">
                    {active.tip}
                  </div>
                </div>
              </motion.div>
            </AnimatePresence>
          </div>
        </div>
      </div>
    </section>
  )
}

// --- Pricing Section ---

const PricingCard = ({ title, price, features, recommended, cta, description }: {
  title: string
  price: string
  features: { text: string; included: boolean }[]
  recommended?: boolean
  cta: { text: string; href: string }
  description?: string
}) => (
  <div
    className={cn(
      "relative p-8 sm:p-10 rounded-[2.5rem] border transition-all duration-500 flex flex-col",
      recommended 
        ? "border-primary/50 shadow-2xl shadow-primary/10 bg-card scale-100 lg:scale-105 z-10" 
        : "border-border/50 bg-card/50 hover:bg-card hover:border-border transition-colors"
    )}
  >
    {recommended && (
      <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-5 py-1.5 rounded-full bg-primary text-primary-foreground text-xs font-bold shadow-glow">
        {t.pricing.mostPopular}
      </div>
    )}

    <div className="mb-8">
      <div className="flex justify-center mb-4">
        <PlanBadge plan={title.toLowerCase() as PlanTier} size="lg" variant="soft" />
      </div>
      {description && (
        <p className="text-sm text-muted-foreground text-center mb-4">{description}</p>
      )}
      <div className="flex items-baseline justify-center gap-1">
        <span className="text-4xl sm:text-5xl font-bold tracking-tight">{price}</span>
        <span className="text-muted-foreground text-base">{t.pricing.perMonth}</span>
      </div>
    </div>

    <ul className="space-y-3 mb-8 flex-1">
      {features.map((f, i) => (
        <li key={i} className="flex items-start gap-3 text-sm">
          {f.included ? (
            <CheckCircle2 className="w-5 h-5 text-emerald-500 shrink-0 mt-0.5" />
          ) : (
            <X className="w-5 h-5 text-muted-foreground/30 shrink-0 mt-0.5" />
          )}
          <span className={f.included ? 'text-foreground' : 'text-muted-foreground/50 line-through'}>
            {f.text}
          </span>
        </li>
      ))}
    </ul>

    <Link to={cta.href} className="w-full mt-auto">
      <Button
        variant={recommended ? 'default' : 'outline'}
        className={cn(
          "w-full h-12 text-base font-semibold rounded-2xl transition-all",
          recommended ? "shadow-glow hover:scale-[1.02]" : "border-border/40 hover:bg-muted/50"
        )}
      >
        {cta.text}
      </Button>
    </Link>
  </div>
)

const Pricing = ({ authenticated }: { authenticated: boolean }) => {
  const { data: apiPlans, isLoading } = usePlans()

  const plans = apiPlans?.map(p => {
    const isFree = p.code === 'free'
    const isUltra = p.code === 'ultra'
    const isPro = p.code === 'pro'

    return {
      title: p.name || (p.code.charAt(0).toUpperCase() + p.code.slice(1)),
      price: p.price === 0 ? '0 TL' : `${p.price} ${p.currency}`,
      recommended: isPro,
      description: isFree 
        ? t.pricing.plans.free.description 
        : isPro 
          ? t.pricing.plans.pro.description 
          : t.pricing.plans.ultra.description,
      cta: isUltra 
        ? { text: t.pricing.cta.ultra, href: 'mailto:sales@botla.app' }
        : {
            text: authenticated ? t.pricing.cta.authenticated : (isFree ? t.pricing.cta.free : t.pricing.cta.pro),
            href: authenticated ? '/dashboard' : '/register',
          },
      features: [
        { text: `${p.limits.max_chatbots} ${t.pricing.features.chatbots}`, included: true },
        { text: `${p.limits.max_monthly_ingestions.toLocaleString('tr-TR')} ${t.pricing.features.tokens}`, included: true },
        { text: `${p.features.scraping.max_urls_per_bot} ${t.pricing.features.sites}`, included: true },
        { text: `${p.features.files.max_files_per_bot} ${t.pricing.features.pdfs}`, included: true },
        { text: isUltra ? 'GPT-4o & GPT-5' : (isPro ? 'GPT-4o & GPT-4o Mini' : 'GPT-4o Mini'), included: true },
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
      <section id="pricing" className="py-32 bg-background relative overflow-hidden">
        <div className="max-w-7xl mx-auto px-6 text-center">
          <Badge>{t.pricing.badge}</Badge>
          <div className="mt-20 flex flex-col lg:flex-row justify-center gap-8">
            <div className="w-full lg:w-80 h-[500px] bg-muted animate-pulse rounded-3xl" />
            <div className="w-full lg:w-80 h-[550px] bg-muted animate-pulse rounded-3xl" />
            <div className="w-full lg:w-80 h-[500px] bg-muted animate-pulse rounded-3xl" />
          </div>
        </div>
      </section>
    )
  }

  return (
    <section id="pricing" className="scroll-mt-24 py-32 bg-background relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center max-w-3xl mx-auto mb-20">
          <Badge>{t.pricing.badge}</Badge>
          <h2 className="text-4xl md:text-6xl font-bold mt-6 mb-6 tracking-tight">
            {t.pricing.title} <span className="text-primary italic">{t.pricing.titleHighlight}</span>
          </h2>
          <p className="text-lg text-muted-foreground">
            {t.pricing.subtitle}
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-6 lg:gap-8 max-w-5xl mx-auto items-start">
          {plans.map((plan, i) => (
            <PricingCard key={i} {...plan} />
          ))}
        </div>
      </div>
    </section>
  )
}

// --- Security Section ---

const SecurityFeature = ({ icon: Icon, title, description }: {
  icon: React.ElementType
  title: string
  description: string
}) => (
  <div className="flex items-start gap-4">
    <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
      <Icon className="w-5 h-5 text-primary" />
    </div>
    <div>
      <h4 className="font-bold text-foreground mb-1">{title}</h4>
      <p className="text-sm text-muted-foreground leading-relaxed">{description}</p>
    </div>
  </div>
)

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
    <section className="py-32 bg-secondary/30 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          <div>
            <Badge>{t.security.badge}</Badge>
            <h2 className="text-4xl md:text-5xl font-bold mt-6 mb-6 tracking-tight">
              {t.security.title} <span className="text-primary">{t.security.titleHighlight}</span>
            </h2>
            <p className="text-lg text-muted-foreground leading-relaxed mb-8">
              {t.security.subtitle}
            </p>
            <div className="grid sm:grid-cols-2 gap-6">
              {securityItems.map((item) => (
                <SecurityFeature
                  key={item.title}
                  icon={item.icon}
                  title={item.title}
                  description={item.description}
                />
              ))}
            </div>
          </div>

          <div className="relative">
            <div className="absolute -inset-8 bg-gradient-to-r from-primary/10 via-transparent to-primary/10 blur-3xl rounded-full" />
            <div className="relative glass rounded-3xl p-8 border border-white/20">
              <div className="flex items-center gap-4 mb-6">
                <div className="w-12 h-12 rounded-2xl bg-emerald-500/20 flex items-center justify-center">
                  <ShieldCheck className="w-6 h-6 text-emerald-500" />
                </div>
                <div>
                  <div className="font-bold text-foreground">Güvenlik Durumu</div>
                  <div className="text-sm text-emerald-500">Tüm sistemler aktif</div>
                </div>
              </div>
              <div className="space-y-4">
                {['KVKK Uyumlu', 'GDPR Uyumlu', 'SSL/TLS Şifreleme', 'SOC 2 Ready'].map((item) => (
                  <div key={item} className="flex items-center gap-3 text-sm">
                    <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                    <span className="text-foreground">{item}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

// --- FAQ Section ---

const FAQ = () => {
  return (
    <section id="faq" className="scroll-mt-24 py-32 bg-background">
      <div className="max-w-4xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center mb-16">
          <Badge>{t.faq.badge}</Badge>
          <h2 className="text-4xl md:text-5xl font-bold mt-6 mb-6 tracking-tight">
            {t.faq.title} <span className="text-primary">{t.faq.titleHighlight}</span>
          </h2>
          <p className="text-lg text-muted-foreground">{t.faq.subtitle}</p>
        </div>

        <div className="space-y-4">
          {t.faq.items.map((faq, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 10 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="border border-border/40 rounded-2xl bg-card overflow-hidden hover:border-primary/30 transition-colors"
            >
              <details className="group">
                <summary className="flex items-center justify-between p-6 cursor-pointer list-none">
                  <span className="font-bold text-lg pr-6">{faq.question}</span>
                  <span className="transition-transform group-open:rotate-180 bg-muted/50 p-2 rounded-full shrink-0">
                    <ChevronRight className="w-5 h-5 rotate-90" />
                  </span>
                </summary>
                <div className="px-6 pb-6 pt-0 text-muted-foreground leading-relaxed">
                  <div className="pt-4 border-t border-border/10">{faq.answer}</div>
                </div>
              </details>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}

// --- CTA Section ---

const CTASection = ({ authenticated }: { authenticated: boolean }) => (
  <section className="py-32 bg-gradient-to-b from-background to-secondary/30 relative overflow-hidden">
    <div className="absolute inset-0 pointer-events-none">
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-primary/10 blur-[150px] rounded-full" />
    </div>

    <div className="max-w-4xl mx-auto px-6 sm:px-8 lg:px-10 text-center relative">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
      >
        <h2 className="text-4xl md:text-6xl font-bold mb-6 tracking-tight">
          {t.cta.title} <br />
          <span className="text-primary">{t.cta.titleHighlight}</span>
        </h2>
        <p className="text-lg text-muted-foreground mb-10 max-w-2xl mx-auto">
          {t.cta.subtitle}
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <Link to={authenticated ? "/dashboard" : "/register"}>
            <Button size="lg" className="h-14 px-12 text-lg font-bold rounded-2xl shadow-glow group hover:scale-105 transition-all">
              {authenticated ? t.cta.buttonAuth : t.cta.button}
              <ArrowRight className="ml-2 w-5 h-5 transition-transform group-hover:translate-x-1" />
            </Button>
          </Link>
          <span className="text-sm text-muted-foreground">{t.cta.note}</span>
        </div>
      </motion.div>
    </div>
  </section>
)

// --- Footer ---

const Footer = ({ authenticated }: { authenticated: boolean }) => (
  <footer className="bg-foreground text-background py-20 border-t border-border/10">
    <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
      <div className="grid md:grid-cols-12 gap-12">
        <div className="col-span-12 md:col-span-5">
          <div className="flex items-center gap-3 mb-6">
            <div className="bg-primary p-2.5 rounded-xl shadow-glow">
              <Bot className="w-6 h-6 text-primary-foreground" />
            </div>
            <span className="font-bold text-2xl tracking-tight text-background">botla.app</span>
          </div>
          <p className="text-background/60 leading-relaxed max-w-sm">
            {t.footer.description}
          </p>
        </div>

        <div className="col-span-6 md:col-span-2 md:col-start-7">
          <h4 className="font-bold text-background mb-5">{t.footer.product.title}</h4>
          <ul className="space-y-3 text-background/60 text-sm">
            <li><a href="#features" className="hover:text-primary transition-colors">{t.footer.product.features}</a></li>
            <li><a href="#pricing" className="hover:text-primary transition-colors">{t.footer.product.pricing}</a></li>
            {authenticated ? (
              <li><Link to="/dashboard" className="hover:text-primary transition-colors">{t.footer.product.dashboard}</Link></li>
            ) : (
              <>
                <li><Link to="/login" className="hover:text-primary transition-colors">{t.footer.product.login}</Link></li>
                <li><Link to="/register" className="hover:text-primary transition-colors">{t.footer.product.register}</Link></li>
              </>
            )}
          </ul>
        </div>

        <div className="col-span-6 md:col-span-2">
          <h4 className="font-bold text-background mb-5">{t.footer.company.title}</h4>
          <ul className="space-y-3 text-background/60 text-sm">
            <li><a href="#" className="hover:text-primary transition-colors">{t.footer.company.about}</a></li>
            <li><a href="#" className="hover:text-primary transition-colors">{t.footer.company.blog}</a></li>
            <li><a href="#" className="hover:text-primary transition-colors">{t.footer.company.contact}</a></li>
          </ul>
        </div>

        <div className="col-span-6 md:col-span-2">
          <h4 className="font-bold text-background mb-5">{t.footer.legal.title}</h4>
          <ul className="space-y-3 text-background/60 text-sm">
            <li><a href="#" className="hover:text-primary transition-colors">{t.footer.legal.privacy}</a></li>
            <li><a href="#" className="hover:text-primary transition-colors">{t.footer.legal.terms}</a></li>
            <li><a href="#" className="hover:text-primary transition-colors">{t.footer.legal.kvkk}</a></li>
          </ul>
        </div>
      </div>

      <div className="border-t border-background/10 mt-16 pt-8 flex flex-col md:flex-row justify-between items-center gap-4 text-background/40 text-sm">
        <p>{t.footer.copyright.replace('{year}', new Date().getFullYear().toString())}</p>
        <p>{t.footer.madeWith}</p>
      </div>
    </div>
  </footer>
)

// --- Main Component ---

export default function LandingPage() {
  const [authenticated, setAuthenticated] = useState(false)

  useEffect(() => {
    const token = window.localStorage.getItem('botla_token')
    const isValid = token !== null && token !== 'undefined' && token !== 'null' && token.length > 0
    setAuthenticated(isValid)
  }, [])

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

  // Dynamically load chatbot widget for landing page demo
  useEffect(() => {
    const token = window.localStorage.getItem('botla_token')
    const isAuthenticated = token !== null && token !== 'undefined' && token !== 'null' && token.length > 0
    if (isAuthenticated) return

    const chatbotId = import.meta.env.VITE_LANDING_CHATBOT_ID
    const widgetUrl = import.meta.env.VITE_WIDGET_SCRIPT_URL

    if (!chatbotId || !widgetUrl) return

    const existingScript = document.querySelector(`script[data-bot="${chatbotId}"]`)
    if (existingScript) return

    const script = document.createElement('script')
    const widgetUrlWithReset = new URL(widgetUrl)
    widgetUrlWithReset.searchParams.set('reset-session', '1')
    script.src = widgetUrlWithReset.toString()
    script.type = 'module'
    script.setAttribute('data-bot', chatbotId)
    script.async = true
    document.body.appendChild(script)

    return () => {
      script.remove()
      const widgetHost = document.getElementById('chatbot-widget-host')
      if (widgetHost) widgetHost.remove()
    }
  }, [])

  return (
    <div className="relative isolate min-h-screen bg-background font-sans selection:bg-primary/20 text-foreground">
      <Noise />
      <MouseHighlight />
      <Navbar authenticated={authenticated} />
      <main>
        <Hero authenticated={authenticated} />
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
