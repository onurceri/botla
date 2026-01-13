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
  RotateCcw,
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
    initial={{ opacity: 0, scale: 0.9 }}
    whileInView={{ opacity: 1, scale: 1 }}
    viewport={{ once: true }}
    className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/5 text-primary text-xs font-bold uppercase tracking-wider mb-8 border border-primary/10 shadow-sm select-none"
  >
    {Icon && <Icon className="w-3.5 h-3.5" />}
    {children}
  </motion.div>
)

const SectionTitle = ({ children, highlight }: { children: React.ReactNode; highlight?: string }) => (
  <motion.h2
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay: 0.1 }}
    className="text-4xl sm:text-5xl md:text-6xl font-bold tracking-tighter mb-6 text-foreground"
  >
    {children} {highlight && <span className="text-primary">{highlight}</span>}
  </motion.h2>
)

const SectionDescription = ({ children }: { children: React.ReactNode }) => (
  <motion.p
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay: 0.2 }}
    className="text-xl text-muted-foreground/80 leading-relaxed max-w-2xl mx-auto"
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
        scrolled ? 'bg-background/80 backdrop-blur-md border-b border-border/40 py-4 shadow-sm' : 'py-6 bg-transparent',
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
              <div className="absolute inset-0 bg-primary/20 blur-xl rounded-full group-hover:bg-primary/40 transition-all duration-500" />
              <img
                src="/logo-128.png"
                alt="Botla Logo"
                className="relative w-10 h-10 rounded-xl shadow-lg shadow-primary/20 group-hover:shadow-primary/40 transition-all duration-500 group-hover:scale-105"
              />
            </div>
            <span className="font-bold text-xl tracking-tight text-foreground/90 group-hover:text-foreground transition-colors">botla.app</span>
          </div>

          {/* Desktop Nav */}
          <div className="hidden lg:flex items-center gap-1 bg-background/50 backdrop-blur-sm px-4 py-2 rounded-full border border-border/40 shadow-sm">
            {links.map((link) => (
              <a
                key={link.name}
                href={link.href}
                onClick={(e) => handleScroll(e, link.href)}
                className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-all duration-300 rounded-full hover:bg-muted/80"
              >
                {link.name}
              </a>
            ))}
          </div>

          {/* CTA Buttons */}
          <div className="hidden lg:flex items-center gap-3">
            {authenticated ? (
              <Link to="/dashboard">
                <Button className="rounded-full px-6 font-semibold shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all hover:scale-105">{t.nav.dashboard}</Button>
              </Link>
            ) : (
              <>
                <Link to="/login">
                  <Button variant="ghost" className="rounded-full px-5 font-medium hover:bg-muted/50">
                    {t.nav.login}
                  </Button>
                </Link>
                <Link to="/register">
                  <Button className="rounded-full px-6 font-semibold shadow-lg shadow-primary/25 hover:shadow-primary/40 transition-all hover:scale-105 bg-gradient-to-r from-primary to-orange-600 border-none">
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
            className="lg:hidden p-2 rounded-lg hover:bg-muted/50 transition-colors text-foreground/80"
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
            className="lg:hidden border-t border-border/50 bg-background/95 backdrop-blur-xl absolute inset-x-0 top-full shadow-2xl"
          >
            <div className="max-w-7xl mx-auto px-6 py-8 space-y-6">
              {links.map((link, i) => (
                <motion.a
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.05 }}
                  key={link.name}
                  href={link.href}
                  onClick={(e) => handleScroll(e, link.href)}
                  className="block text-lg font-medium text-foreground/80 hover:text-primary transition-colors"
                >
                  {link.name}
                </motion.a>
              ))}
              <div className="pt-6 border-t border-border/50 flex flex-col gap-4">
                {authenticated ? (
                  <Link to="/dashboard" onClick={() => setIsOpen(false)}>
                    <Button className="w-full h-12 rounded-xl text-base shadow-lg shadow-primary/20">
                      {t.nav.goToDashboard}
                    </Button>
                  </Link>
                ) : (
                  <>
                    <Link to="/login" onClick={() => setIsOpen(false)}>
                      <Button variant="outline" className="w-full h-12 rounded-xl text-base border-primary/20 text-primary hover:bg-primary/5">
                        {t.nav.login}
                      </Button>
                    </Link>
                    <Link to="/register" onClick={() => setIsOpen(false)}>
                      <Button className="w-full h-12 rounded-xl text-base bg-gradient-to-r from-primary to-orange-600 shadow-lg shadow-primary/25">
                         {t.nav.register}
                      </Button>
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
  const [widgetOpen, setWidgetOpen] = useState(false)

  useEffect(() => {
    // Sequence: Browser appears -> Launcher appears -> Widget opens automatically
    const timer = setTimeout(() => {
      setWidgetOpen(true)
    }, 2000) // 2 seconds delay before opening
    return () => clearTimeout(timer)
  }, [])

  return (
    <section className="relative min-h-[110vh] flex items-center justify-center overflow-hidden pt-32 pb-32">
      {/* Dynamic Background */}
      <div className="absolute inset-0 -z-10 bg-background">
        <div className="absolute top-0 left-0 right-0 h-[800px] bg-[radial-gradient(circle_at_50%_-20%,rgba(245,158,11,0.15),transparent_70%)]" />
        <div className="absolute top-[20%] right-[-10%] w-[600px] h-[600px] bg-primary/5 rounded-full blur-[100px] animate-pulse" style={{ animationDuration: '4s' }} />
        <div className="absolute bottom-[-10%] left-[-10%] w-[500px] h-[500px] bg-orange-500/5 rounded-full blur-[80px]" />
        
        {/* Animated Grid Pattern */}
        <div 
          className="absolute inset-0 opacity-[0.03]"
          style={{
            backgroundImage: `linear-gradient(to right, currentColor 1px, transparent 1px),
                             linear-gradient(to bottom, currentColor 1px, transparent 1px)`,
            backgroundSize: '40px 40px',
            maskImage: 'radial-gradient(ellipse at center, black 40%, transparent 80%)'
          }}
        />
      </div>

      <div className="max-w-7xl mx-auto px-6 relative w-full">
        <div className="text-center max-w-5xl mx-auto relative z-10">
          {/* Badge */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, ease: "easeOut" }}
            className="inline-flex items-center gap-3 px-4 py-2 rounded-full bg-white/50 backdrop-blur-md border border-white/60 shadow-sm mb-10 hover:shadow-md transition-all cursor-default"
          >
            <div className="flex items-center gap-2">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
              </span>
              <span className="text-xs font-bold text-emerald-600 uppercase tracking-wide">{t.hero.badgeLive}</span>
            </div>
            <div className="h-3 w-px bg-border/80" />
            <div className="flex items-center gap-1.5">
              <Sparkles className="w-3.5 h-3.5 text-primary" />
              <span className="text-sm font-semibold text-foreground/80">{t.hero.badge}</span>
            </div>
          </motion.div>

          {/* Headline */}
          <motion.h1
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.1, ease: "easeOut" }}
            className="text-6xl sm:text-7xl md:text-8xl font-bold tracking-tighter leading-[1] mb-8"
          >
            <span className="text-foreground drop-shadow-sm">{t.hero.title.line1}</span>
            <br />
            <span className="relative inline-block">
              <span className="absolute -inset-2 bg-gradient-to-r from-primary/20 to-orange-500/20 blur-xl opacity-50" />
              <span className="relative bg-gradient-to-r from-primary via-orange-500 to-amber-500 bg-clip-text text-transparent pb-2">
                {t.hero.title.highlight}
              </span>
            </span>
            <br />
            <span className="text-foreground/90">{t.hero.title.line2}</span>
          </motion.h1>

          {/* Subtitle */}
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="mt-6 text-xl sm:text-2xl text-muted-foreground max-w-3xl mx-auto leading-relaxed font-light"
          >
            {t.hero.subtitle}
          </motion.p>

          {/* CTA Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.3 }}
            className="mt-10 flex flex-col sm:flex-row gap-5 justify-center items-center"
          >
            <Link to={authenticated ? '/dashboard' : '/register'}>
              <Button
                size="lg"
                className="h-14 px-10 text-lg font-bold rounded-full shadow-lg shadow-primary/25 hover:shadow-primary/40 hover:scale-[1.02] transition-all duration-300 bg-gradient-to-r from-primary to-orange-600 border-0"
              >
                {authenticated ? t.hero.cta.primaryAuth : t.hero.cta.primary}
                <ArrowRight className="ml-2 w-5 h-5" />
              </Button>
            </Link>
            <a href="#how-it-works">
              <Button
                variant="outline"
                size="lg"
                className="h-14 px-10 text-lg font-medium rounded-full border-border/60 bg-white/50 backdrop-blur-sm hover:bg-white/80 hover:scale-[1.02] transition-all duration-300"
              >
                <Play className="mr-2 w-5 h-5 text-primary fill-primary/20" />
                {t.hero.cta.secondary}
              </Button>
            </a>
          </motion.div>

          {/* Stats Component - Refined */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5, duration: 1 }}
            className="mt-16 pt-8 border-t border-border/40 flex justify-center gap-12 sm:gap-20"
          >
            {[
              { value: t.hero.stats.sourcesValue, label: t.hero.stats.sources, icon: Database },
              { value: t.hero.stats.securityValue, label: t.hero.stats.security, icon: ShieldCheck },
              { value: t.hero.stats.languagesValue, label: t.hero.stats.languages, icon: Globe },
            ].map((stat, i) => (
               <div key={i} className="flex flex-col items-center gap-2 group cursor-default">
                  <div className="text-3xl font-bold text-foreground tracking-tight group-hover:scale-110 transition-transform duration-300 ease-spring">{stat.value}</div>
                  <div className="flex items-center gap-1.5 text-sm font-medium text-muted-foreground/80">
                    <stat.icon className="w-3.5 h-3.5" />
                    {stat.label}
                  </div>
               </div>
            ))}
          </motion.div>
        </div>

        {/* Hero Visual - Premium Browser Mockup - Hidden on mobile */}
        <motion.div
           initial={{ opacity: 0, y: 100, rotateX: 20 }}
           animate={{ opacity: 1, y: 0, rotateX: 0 }}
           transition={{ duration: 1.2, delay: 0.4, type: "spring", bounce: 0.2 }}
           style={{ perspective: '1200px' }}
           className="mt-24 relative max-w-5xl mx-auto hidden md:block"
        >
          {/* Main Glow */}
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[120%] h-[120%] bg-gradient-to-b from-primary/10 via-orange-500/5 to-transparent blur-3xl opacity-60 rounded-[4rem] -z-10" />

          {/* Browser Frame */}
          <div className="relative bg-white/90 backdrop-blur-2xl rounded-2xl shadow-[0_30px_60px_-10px_rgba(0,0,0,0.12)] border border-white/60 overflow-hidden ring-1 ring-black/5">
            {/* Browser Header */}
            <div className="flex items-center gap-4 px-4 lg:px-6 py-3 lg:py-4 bg-white/80 border-b border-black/[0.03]">
              <div className="flex gap-1.5 lg:gap-2">
                <div className="w-2.5 h-2.5 lg:w-3 lg:h-3 rounded-full bg-[#FF5F57] shadow-inner" />
                <div className="w-2.5 h-2.5 lg:w-3 lg:h-3 rounded-full bg-[#FEBC2E] shadow-inner" />
                <div className="w-2.5 h-2.5 lg:w-3 lg:h-3 rounded-full bg-[#28C840] shadow-inner" />
              </div>
              <div className="flex-1 flex justify-center">
                <div className="flex items-center gap-2 px-4 lg:px-6 py-1 lg:py-1.5 rounded-lg bg-gray-100/80 text-[10px] lg:text-xs text-gray-500 font-medium border border-gray-200/50 shadow-inner">
                  <Lock className="w-2.5 h-2.5 lg:w-3 lg:h-3 text-gray-400" />
                  <span>botla.app/demo</span>
                </div>
              </div>
              <div className="w-12 lg:w-16" />
            </div>

            {/* Website Content Mockup */}
            <div className="relative h-[350px] lg:h-[420px] bg-[#FAFAFA] overflow-hidden flex">
               {/* Sidebar Mockup */}
               <div className="hidden lg:flex flex-col w-56 border-r border-black/[0.03] bg-white p-5 gap-5">
                  <div className="h-7 w-20 bg-gray-100 rounded-lg animate-pulse" />
                  <div className="space-y-2.5">
                    <div className="h-3.5 w-full bg-gray-50 rounded-md" />
                    <div className="h-3.5 w-3/4 bg-gray-50 rounded-md" />
                    <div className="h-3.5 w-5/6 bg-gray-50 rounded-md" />
                  </div>
                  <div className="mt-auto h-10 w-full bg-primary/5 rounded-xl border border-primary/10" />
               </div>

               {/* Main Content Mockup */}
               <div className="flex-1 p-6 lg:p-10 relative">
                  <div className="max-w-xl">
                    <div className="h-8 lg:h-9 w-2/3 bg-gray-200/50 rounded-xl mb-5" />
                    <div className="space-y-3">
                      <div className="h-3.5 w-full bg-gray-100 rounded-lg" />
                      <div className="h-3.5 w-full bg-gray-100 rounded-lg" />
                      <div className="h-3.5 w-3/4 bg-gray-100 rounded-lg" />
                    </div>
                    
                    <div className="hidden lg:grid grid-cols-2 gap-5 mt-8">
                       <div className="aspect-video rounded-xl bg-gradient-to-br from-gray-100 to-gray-50 shadow-sm border border-gray-100" />
                       <div className="aspect-video rounded-xl bg-gradient-to-br from-gray-100 to-gray-50 shadow-sm border border-gray-100" />
                    </div>
                  </div>
               </div>

               {/* The Chat Widget - Animated Showcase */}
               <div className="absolute bottom-4 right-4 lg:bottom-5 lg:right-5 z-20 flex flex-col items-end pointer-events-none">
                 <AnimatePresence mode="wait">
                   {!widgetOpen ? (
                     <motion.div
                       key="launcher"
                       initial={{ scale: 0, opacity: 0 }}
                       animate={{ scale: 1, opacity: 1 }}
                       exit={{ scale: 0, opacity: 0, transition: { duration: 0.2 } }}
                       transition={{ 
                         type: "spring", 
                         stiffness: 260, 
                         damping: 20, 
                         delay: 1.2 
                       }}
                       className="relative"
                     >
                       <div className="absolute -inset-1 bg-primary/20 rounded-full blur-md opacity-60" />
                       <div className="relative w-11 h-11 lg:w-12 lg:h-12 bg-gradient-to-r from-primary to-orange-500 rounded-full flex items-center justify-center shadow-lg shadow-primary/30 text-white">
                          <MessageCircle className="w-5 h-5 lg:w-6 lg:h-6 fill-white/20" />
                          <span className="absolute -top-0.5 -right-0.5 w-3 h-3 lg:w-3.5 lg:h-3.5 bg-emerald-500 border-2 border-white rounded-full">
                            <span className="absolute inset-0 rounded-full bg-emerald-500 animate-ping opacity-75" />
                          </span>
                       </div>
                     </motion.div>
                   ) : (
                     <motion.div
                        key="widget"
                        initial={{ opacity: 0, scale: 0.8, y: 20, transformOrigin: "bottom right" }}
                        animate={{ opacity: 1, scale: 1, y: 0 }}
                        exit={{ opacity: 0, scale: 0.8, y: 20 }}
                        transition={{ 
                          type: "spring", 
                          stiffness: 200, 
                          damping: 25 
                        }}
                        className="w-[260px] lg:w-[280px]"
                     >
                       <div className="bg-white rounded-[20px] shadow-[0_16px_32px_-4px_rgba(0,0,0,0.12)] border border-white/50 overflow-hidden ring-1 ring-black/5 origin-bottom-right">
                          {/* Header */}
                          <div className="bg-gradient-to-r from-primary to-orange-500 px-4 py-3 flex items-center justify-between">
                             <div className="flex items-center gap-2.5">
                                <div className="w-8 h-8 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center border border-white/10">
                                   <Bot className="w-4 h-4 text-white" />
                                </div>
                                <span className="font-bold text-white text-sm">Chatbot</span>
                             </div>
                             <div className="flex items-center gap-1">
                                <div className="p-1.5 rounded-full hover:bg-white/10">
                                   <RotateCcw className="w-3.5 h-3.5 text-white/80" />
                                </div>
                                <div className="p-1.5 rounded-full hover:bg-white/10">
                                   <X className="w-3.5 h-3.5 text-white/80" />
                                </div>
                             </div>
                          </div>

                          {/* Chat Area */}
                          <div className="p-3 h-[180px] lg:h-[200px] bg-[#FAF9F6] flex flex-col gap-3 overflow-hidden relative">
                             {/* Bot Message */}
                             <motion.div 
                                initial={{ opacity: 0, y: 10 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: 0.3 }}
                                className="flex flex-col"
                             >
                                <div className="bg-white p-3 rounded-2xl shadow-sm text-[11px] text-gray-700 leading-relaxed max-w-[85%]">
                                   Merhaba! Size nasıl yardımcı olabilirim?
                                   <div className="text-[9px] text-gray-400 mt-1.5 text-right">23:09</div>
                                </div>
                                <div className="w-5 h-5 rounded-full bg-white border border-gray-100 flex items-center justify-center mt-1.5 shadow-sm">
                                   <Bot className="w-2.5 h-2.5 text-primary" />
                                </div>
                             </motion.div>

                             {/* Suggested Questions Carousel */}
                             <motion.div 
                                initial={{ opacity: 0, y: 10 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: 0.8 }}
                                className="mt-auto"
                             >
                                <div className="flex items-center justify-between mb-2">
                                   <div className="flex items-center gap-1">
                                      <Sparkles className="w-3 h-3 text-primary" />
                                      <span className="text-[9px] font-semibold text-primary uppercase tracking-wide">Önerilen Sorular</span>
                                   </div>
                                   <span className="text-[9px] text-gray-400 font-medium">1 / 3</span>
                                </div>
                                <div className="flex items-center gap-2">
                                   <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center shrink-0">
                                      <ChevronRight className="w-3 h-3 text-gray-400 rotate-180" />
                                   </div>
                                   <div className="flex-1 bg-white px-3 py-2 rounded-xl text-[10px] text-gray-700 font-medium shadow-sm border border-gray-100">
                                      Bu asistan nasıl çalışır?
                                   </div>
                                   <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center shrink-0">
                                      <ChevronRight className="w-3 h-3 text-gray-400" />
                                   </div>
                                </div>
                             </motion.div>
                          </div>

                          {/* Input Footer */}
                          <div className="p-3 bg-white border-t border-gray-100">
                            <div className="flex items-center gap-2 bg-gray-50 border border-gray-200 rounded-full px-3 py-2">
                              <span className="flex-1 text-[10px] text-gray-400">Mesaj yazın...</span>
                              <div className="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center">
                                 <ArrowRight className="w-3 h-3 text-gray-400 rotate-[-45deg]" />
                              </div>
                            </div>
                            <div className="flex items-center justify-between mt-2 px-1">
                               <span className="text-[8px] text-gray-400">0 / 1000</span>
                               <span className="text-[8px] text-gray-400">Powered by <span className="font-semibold text-primary">Botla</span></span>
                            </div>
                          </div>
                       </div>
                     </motion.div>
                   )}
                 </AnimatePresence>
               </div>

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
  <section className="py-20 border-y border-border/40 bg-white relative overflow-hidden">
    <div className="absolute inset-0 bg-gradient-to-b from-transparent to-gray-50/50" />
    <div className="max-w-7xl mx-auto px-6 relative">
      <p className="text-center text-sm font-semibold text-muted-foreground/70 uppercase tracking-widest mb-10">
        Global Teknoloji Platformları İle Tam Uyumlu
      </p>
      
      <div className="relative flex overflow-x-hidden group">
        <div className="animate-marquee flex gap-16 items-center whitespace-nowrap py-4">
          {[...Array(2)].map((_, setIndex) => (
             <React.Fragment key={setIndex}>
              {['OpenAI', 'Anthropic', 'Google Cloud', 'AWS', 'Vercel', 'Stripe', 'Supabase'].map((name) => (
                <div
                  key={name}
                  className="flex items-center gap-3 text-gray-400 group-hover:text-gray-600 transition-colors cursor-default"
                >
                  <div className="w-10 h-10 rounded-xl bg-gray-100 flex items-center justify-center shadow-inner">
                     <Cpu className="w-5 h-5 opacity-50" />
                  </div>
                  <span className="text-xl font-bold tracking-tight opacity-70">{name}</span>
                </div>
              ))}
             </React.Fragment>
          ))}
        </div>
        
        {/* Fade masks for infinite scroll illusion */}
        <div className="absolute inset-y-0 left-0 w-32 bg-gradient-to-r from-white to-transparent pointer-events-none" />
        <div className="absolute inset-y-0 right-0 w-32 bg-gradient-to-l from-white to-transparent pointer-events-none" />
      </div>
    </div>
    
    <style>{`
      @keyframes marquee {
        0% { transform: translateX(0); }
        100% { transform: translateX(-50%); }
      }
      .animate-marquee {
        animation: marquee 30s linear infinite;
      }
    `}</style>
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
  className,
}: {
  icon: React.ElementType
  title: string
  description: string
  index: number
  className?: string
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ delay: index * 0.1 }}
    className={cn(
      "group relative p-8 rounded-[2.5rem] bg-white border border-gray-100 hover:border-primary/20 hover:shadow-2xl hover:shadow-primary/5 transition-all duration-500 flex flex-col h-full overflow-hidden",
      className
    )}
  >
    <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-primary/5 to-transparent rounded-bl-[100%] -mr-16 -mt-16 transition-transform group-hover:scale-125 duration-700" />
    
    <div className="w-14 h-14 rounded-2xl bg-white border border-gray-100 flex items-center justify-center mb-6 group-hover:bg-primary group-hover:border-primary group-hover:text-white transition-all duration-300 shadow-sm relative z-10 group-hover:rotate-3 group-hover:scale-110">
      <Icon className="w-7 h-7 text-gray-900 group-hover:text-white transition-colors duration-300" />
    </div>

    <h3 className="text-xl font-bold mb-3 text-foreground tracking-tight relative z-10">{title}</h3>
    <p className="text-muted-foreground leading-relaxed text-[15px] relative z-10 font-medium opacity-80">{description}</p>
  </motion.div>
)

const Features = () => {
  const features = [
    { icon: Search, ...t.features.items.rag, className: "md:col-span-2 bg-gradient-to-br from-white to-gray-50" },
    { icon: Database, ...t.features.items.sources },
    { icon: Palette, ...t.features.items.widget },
    { icon: Zap, ...t.features.items.actions },
    { icon: ShieldCheck, ...t.features.items.guardrails },
    { icon: BarChart3, ...t.features.items.analytics, className: "md:col-span-2 bg-gradient-to-br from-white to-gray-50" },
    { icon: Headphones, ...t.features.items.handoff },
    { icon: Building2, ...t.features.items.multiTenant },
  ]

  return (
    <section id="features" className="py-24 sm:py-32 scroll-mt-20 bg-gray-50/50">
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center mb-16 max-w-3xl mx-auto">
          <SectionBadge icon={Sparkles}>{t.features.badge}</SectionBadge>
          <SectionTitle highlight={t.features.titleHighlight}>{t.features.title}</SectionTitle>
          <SectionDescription>{t.features.subtitle}</SectionDescription>
        </div>

        <div className="grid md:grid-cols-3 gap-6">
          {features.map((feature, i) => (
            <FeatureCard
              key={feature.title}
              icon={feature.icon}
              title={feature.title}
              description={feature.description}
              index={i}
              className={feature.className}
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
    className="group relative p-8 sm:p-10 rounded-[2.5rem] bg-white border border-gray-100 hover:border-primary/20 hover:shadow-2xl hover:shadow-primary/5 transition-all duration-500 h-full flex flex-col"
  >
    <div className="absolute top-0 right-0 w-40 h-40 bg-gradient-to-br from-primary/5 to-transparent rounded-bl-full -mr-10 -mt-10 transition-transform group-hover:scale-125 duration-700" />

    {/* Icon */}
    <div className="w-16 h-16 rounded-2xl bg-white border border-gray-100 flex items-center justify-center mb-8 group-hover:bg-primary group-hover:border-primary group-hover:text-white transition-all duration-300 shadow-sm relative z-10 group-hover:-translate-y-1">
      <Icon className="w-8 h-8 text-primary group-hover:text-white transition-colors duration-300" />
    </div>

    {/* Content */}
    <h3 className="text-2xl font-bold mb-4 text-foreground relative z-10">{title}</h3>
    <p className="text-muted-foreground leading-relaxed mb-8 relative z-10 text-[15px]">{description}</p>

    {/* Feature Tags */}
    <div className="flex flex-wrap gap-2.5 mt-auto relative z-10">
      {features.map((feature) => (
        <span
          key={feature}
          className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-gray-50 text-gray-700 text-xs font-semibold border border-gray-100 group-hover:bg-white group-hover:border-primary/20 group-hover:text-primary transition-all duration-300"
        >
          <CheckCircle2 className="w-3.5 h-3.5 text-primary" />
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
    <section id="use-cases" className="py-24 sm:py-32 bg-white relative overflow-hidden scroll-mt-20">
      {/* Decorative blobs */}
      <div className="absolute top-1/2 left-0 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[100px] -translate-x-1/2 -translate-y-1/2 -z-10 pointer-events-none" />
      <div className="absolute bottom-0 right-0 w-[500px] h-[500px] bg-orange-500/5 rounded-full blur-[100px] translate-x-1/4 translate-y-1/4 -z-10 pointer-events-none" />

      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center mb-20 max-w-3xl mx-auto">
          <SectionBadge icon={Users}>{t.useCases.badge}</SectionBadge>
          <SectionTitle highlight={t.useCases.titleHighlight}>{t.useCases.title}</SectionTitle>
          <SectionDescription>{t.useCases.subtitle}</SectionDescription>
        </div>

        <div className="grid md:grid-cols-2 gap-8 lg:gap-10">
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
      'relative p-8 sm:p-10 rounded-[2.5rem] flex flex-col h-full transition-all duration-500',
      recommended
        ? 'border-2 border-primary bg-white shadow-2xl shadow-primary/10 scale-105 z-10'
        : 'border border-gray-100 bg-white hover:border-border/80'
    )}
  >
    {recommended && (
      <div className="absolute -top-5 left-1/2 -translate-x-1/2">
        <span className="px-5 py-2 rounded-full bg-primary text-white text-sm font-bold shadow-lg shadow-primary/30 flex items-center gap-2">
          <Star className="w-4 h-4 fill-white" />
          {t.pricing.mostPopular}
        </span>
      </div>
    )}

    {/* Header */}
    <div className="text-center mb-10">
      <div className="flex justify-center mb-6">
        <PlanBadge plan={title.toLowerCase() as PlanTier} size="lg" variant={recommended ? "default" : "soft"} />
      </div>
      {description && <p className="text-muted-foreground mb-6 font-medium">{description}</p>}
      <div className="flex items-baseline justify-center gap-1">
        <span className="text-5xl sm:text-6xl font-bold tracking-tighter text-foreground">{price}</span>
        <span className="text-lg text-muted-foreground font-medium">{t.pricing.perMonth}</span>
      </div>
    </div>

    {/* Features */}
    <ul className="space-y-4 flex-1 mb-10">
      {features.map((f, i) => (
        <li key={i} className="flex items-center gap-3.5 text-sm">
          {f.included ? (
            <div className={`w-6 h-6 rounded-full flex items-center justify-center shrink-0 ${recommended ? 'bg-primary/10 text-primary' : 'bg-emerald-500/10 text-emerald-500'}`}>
              <Check className="w-3.5 h-3.5" />
            </div>
          ) : (
            <div className="w-6 h-6 rounded-full bg-gray-50 flex items-center justify-center shrink-0">
              <Minus className="w-3.5 h-3.5 text-gray-300" />
            </div>
          )}
          <span className={f.included ? 'text-gray-700 font-medium' : 'text-gray-400'}>
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
            'w-full h-14 font-bold rounded-2xl text-base transition-all hover:scale-105',
            recommended 
              ? 'bg-primary hover:bg-primary/90 text-white shadow-lg shadow-primary/25 hover:shadow-primary/40' 
              : 'border-2 text-foreground hover:bg-gray-50'
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
          'w-full h-14 font-bold rounded-2xl text-base transition-all',
          recommended && 'bg-primary text-white shadow-lg shadow-primary/25',
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
