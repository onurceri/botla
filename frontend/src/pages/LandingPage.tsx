import React, { useEffect, useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Bot,
  Zap,
  ShieldCheck,
  Globe,
  Database,
  ArrowRight,
  CheckCircle2,
  Menu,
  X,
  Sparkles,
  Cpu,
  MessagesSquare,
  Search,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

// --- Components ---

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
    { name: 'Özellikler', href: '#features' },
    { name: 'Nasıl Çalışır?', href: '#how-it-works' },
    { name: 'Fiyatlandırma', href: '#pricing' },
    { name: 'SSS', href: '#faq' },
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
          <div className="hidden md:flex items-center gap-1 bg-muted/40 p-1 rounded-full border border-border/40">
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

          <div className="hidden md:flex items-center gap-3">
            {authenticated ? (
              <Link to="/dashboard">
                <Button className="rounded-full px-6 font-semibold shadow-sm">Dashboard</Button>
              </Link>
            ) : (
              <>
                <Link to="/login">
                  <Button variant="ghost" className="rounded-full px-6 font-medium text-muted-foreground hover:text-foreground">
                    Giriş Yap
                  </Button>
                </Link>
                <Link to="/register">
                  <Button className="rounded-full px-6 font-semibold shadow-glow transition-all hover:scale-105">Ücretsiz Dene</Button>
                </Link>
              </>
            )}
          </div>

          {/* Mobile Menu Button */}
          <div className="md:hidden">
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
            className="md:hidden fixed top-[80px] inset-x-4 glass p-8 rounded-3xl shadow-xl z-50 border border-white/20"
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
                    <Button className="w-full h-14 rounded-2xl text-lg">Dashboard'a Git</Button>
                  </Link>
                ) : (
                  <>
                    <Link to="/login" onClick={() => setIsOpen(false)}>
                      <Button variant="outline" className="w-full h-14 rounded-2xl text-lg border-border/40">
                        Giriş Yap
                      </Button>
                    </Link>
                    <Link to="/register" onClick={() => setIsOpen(false)}>
                      <Button className="w-full h-14 rounded-2xl text-lg shadow-glow">Ücretsiz Dene</Button>
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

const Hero = ({ authenticated }: { authenticated: boolean }) => {
  const [heroSrc, setHeroSrc] = useState('/assets/landing-hero.png')
  const [useVideo, setUseVideo] = useState(true)

  return (
    <section className="relative pt-32 pb-20 lg:pt-48 lg:pb-32 overflow-hidden">
      {/* Background Orbs */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-full -z-10 pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[600px] h-[600px] bg-primary/10 blur-[120px] rounded-full animate-pulse" />
        <div className="absolute bottom-[20%] right-[-10%] w-[500px] h-[500px] bg-primary/5 blur-[100px] rounded-full animate-pulse" style={{ animationDelay: '1s' }} />
      </div>

      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10 relative">
        <div className="flex flex-col items-center text-center max-w-4xl mx-auto mb-16 px-4">
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass border border-white/20 text-muted-foreground text-sm font-medium mb-8"
          >
            <Sparkles className="w-4 h-4 text-primary" />
            <span>Yapay zeka ile müşteri desteğinde yeni çağ</span>
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1, duration: 0.8 }}
            className="text-5xl sm:text-6xl lg:text-7xl font-bold tracking-tight text-foreground mb-8 leading-[1.05]"
          >
            Verilerinizi <br />
            <span className="bg-clip-text text-transparent bg-gradient-to-b from-primary to-orange-600">
              Akıllı Asistanlara
            </span>
            <br /> Dönüştürün.
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2, duration: 0.8 }}
            className="text-lg sm:text-xl text-muted-foreground mb-10 max-w-2xl leading-relaxed"
          >
            1 dakika içinde PDF dosyalarınızı birer uzmana dönüştürün. 
            Web sitenizi taratın, dokümanlarınızı yükleyin ve 7/24 çalışan asistanınızı hemen yayınlayın.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3, duration: 0.8 }}
            className="flex flex-col sm:flex-row gap-4 items-center"
          >
            <Link to={authenticated ? "/dashboard" : "/register"}>
              <Button size="lg" className="h-14 px-10 text-lg font-semibold rounded-full shadow-glow group transition-all hover:scale-[1.02]">
                {authenticated ? "Dashboard'a Dön" : "Hemen Ücretsiz Başla"}
                <ArrowRight className="ml-2 w-5 h-5 transition-transform group-hover:translate-x-1" />
              </Button>
            </Link>
            <a href="#how-it-works">
              <Button
                variant="outline"
                size="lg"
                className="h-14 px-10 text-lg font-semibold rounded-full border-border/40 hover:bg-muted/50 transition-all hover:scale-[1.02]"
              >
                Nasıl Çalışır?
              </Button>
            </a>
          </motion.div>

          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
            className="mt-12 flex items-center gap-8 text-sm text-muted-foreground border-t border-border/10 pt-8"
          >
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-5 h-5 text-success" />
              <span>Kredi Kartı Gerekmez</span>
            </div>
            <div className="flex items-center gap-2 border-l border-border/10 pl-8">
              <CheckCircle2 className="w-5 h-5 text-success" />
              <span>Saniyeler İçinde Kurulum</span>
            </div>
          </motion.div>
        </div>

        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 40 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          transition={{ duration: 1, ease: [0.16, 1, 0.3, 1] }}
          className="relative max-w-5xl mx-auto"
        >
          <div className="glass shadow-2xl rounded-[2.5rem] p-3 border border-white/30 backdrop-blur-2xl">
            <div className="relative rounded-[2rem] overflow-hidden bg-card border border-border/40 aspect-video group">
              {useVideo ? (
                <video
                  className="w-full h-full object-cover transition-transform duration-700 group-hover:scale-[1.01]"
                  autoPlay
                  muted
                  loop
                  playsInline
                  preload="metadata"
                  poster={heroSrc}
                  onError={() => setUseVideo(false)}
                >
                  <source src="/assets/landing_hero.mp4" type="video/mp4" />
                </video>
              ) : (
                <img
                  src={heroSrc}
                  alt="Dashboard Preview"
                  className="w-full h-full object-cover"
                  onError={() => setHeroSrc('/assets/landing-hero-final.png')}
                />
              )}
              <div className="absolute inset-x-0 bottom-0 h-40 bg-gradient-to-t from-background/40 via-transparent to-transparent opacity-60 pointer-events-none" />
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

const FeatureCard = ({ icon: Icon, title, desc, delay }: any) => (
  <motion.div
    initial={{ opacity: 0, y: 24 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true, margin: "-100px" }}
    transition={{ delay, duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
    className="group relative p-10 rounded-[2.5rem] bg-card border border-border/50 hover:border-primary/40 transition-all duration-500 hover:shadow-xl hover:shadow-primary/5"
  >
    <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mb-10 group-hover:scale-110 transition-transform duration-500 shadow-glow">
      <Icon className="w-8 h-8 text-primary" />
    </div>

    <h3 className="text-2xl font-bold mb-4 text-foreground tracking-tight">{title}</h3>
    <p className="text-muted-foreground leading-relaxed text-lg">{desc}</p>
    
    <div className="mt-8 flex items-center text-primary text-sm font-semibold opacity-0 group-hover:opacity-100 transition-opacity translate-y-2 group-hover:translate-y-0 duration-500 cursor-pointer">
      Daha fazlası <ArrowRight className="ml-1 w-4 h-4" />
    </div>
  </motion.div>
)

const Features = () => {
  const features = [
    {
      icon: Database,
      title: 'Verilerinizi Bağlayın',
      desc: 'PDF yükleyin, web sitenizi kaynak olarak ekleyin veya metin içeriği girin. AI modelimiz verilerinizi analiz eder ve öğrenir.',
    },
    {
      icon: Cpu,
      title: 'Gelişmiş RAG Teknolojisi',
      desc: 'Vektör tabanlı arama ve GPT-4o ile, botunuz her zaman bağlama uygun ve doğru cevaplar verir. Halüsinasyon görmez.',
    },
    {
      icon: Globe,
      title: 'Dinamik Web Tarama',
      desc: 'Web siteniz güncellendiğinde botunuz da güncellensin. Dinamik site tarayıcımızla içeriğinizi otomatik senkronize edin.',
    },
    {
      icon: ShieldCheck,
      title: 'Güvenlik & Guardrails',
      desc: 'Botunuzun konuşacağı konuları sınırlayın. Rakip firmalardan bahsetmesini veya uygunsuz içerik üretmesini engelleyin.',
    },
    {
      icon: Zap,
      title: 'Hızlı Entegrasyon',
      desc: 'Tek satır JavaScript kodu ile sitenize ekleyin. Güvenli embed seçenekleriyle alan adlarınızı kısıtlayın.',
    },
    {
      icon: Search,
      title: 'Analitik ve İçgörü',
      desc: 'Kullanıcılarınızın ne sorduğunu görün. Başarısız konuşmaları analiz edin ve botunuzu sürekli iyileştirin.',
    },
  ]

  return (
    <section id="features" className="scroll-mt-24 py-32 bg-secondary/30 relative overflow-hidden">
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-primary/5 blur-[160px] rounded-full -z-10" />
      
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center max-w-3xl mx-auto mb-20 px-4">
          <Badge>Özellikler</Badge>
          <h2 className="text-4xl md:text-6xl font-bold mt-6 mb-8 text-foreground tracking-tight leading-[1.1]">
            Yapay Zekanın Gücünü <br />
            <span className="text-primary tracking-tighter">İşinize Taşıyın</span>
          </h2>
          <p className="text-xl text-muted-foreground leading-relaxed">
            İşletmenizin her adımında size eşlik eden, tamamen verilerinizle eğitilmiş profesyonel bir asistan.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((f, i) => (
            <FeatureCard key={i} {...f} delay={i * 0.1} />
          ))}
        </div>
      </div>
    </section>
  )
}

const PricingCard = ({ title, price, features, recommended, cta }: any) => (
  <div
    className={cn(
      "relative p-10 rounded-[2.5rem] border transition-all duration-500 flex flex-col items-center text-center",
      recommended 
        ? "border-primary/50 shadow-2xl shadow-primary/10 bg-card scale-105 z-10" 
        : "border-border/50 bg-card/50 hover:bg-card hover:border-border transition-colors"
    )}
  >
    {recommended && (
      <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-6 py-1.5 rounded-full bg-primary text-primary-foreground text-sm font-bold shadow-glow">
        En Popüler
      </div>
    )}

    <div className="mb-10 w-full">
      <h3 className="text-2xl font-bold mb-4">{title}</h3>
      <div className="flex items-baseline justify-center gap-1">
        <span className="text-5xl font-bold tracking-tight">{price}</span>
        <span className="text-muted-foreground text-base">/ay</span>
      </div>
    </div>

    <ul className="space-y-4 mb-10 flex-1 w-full">
      {features.map((f: any, i: number) => (
        <li key={i} className="flex items-center justify-center gap-3 text-sm">
          {f.included ? (
            <CheckCircle2 className="w-5 h-5 text-success shrink-0" />
          ) : (
            <X className="w-5 h-5 text-muted-foreground/30 shrink-0" />
          )}
          <span className={f.included ? 'text-foreground font-medium text-base' : 'text-muted-foreground/60 text-base line-through opacity-50'}>
            {f.text}
          </span>
        </li>
      ))}
    </ul>

    <Link to={cta.href} className="w-full">
      <Button
        variant={recommended ? 'default' : 'outline'}
        className={cn(
          "w-full h-14 text-lg font-semibold rounded-2xl transition-all",
          recommended ? "shadow-glow hover:scale-[1.02]" : "border-border/40 hover:bg-muted/50"
        )}
      >
        {cta.text}
      </Button>
    </Link>
  </div>
)

const Pricing = ({ authenticated }: { authenticated: boolean }) => {
  const plans = [
    {
      title: 'Başlangıç (Free)',
      price: '0 TL',
      cta: {
        text: authenticated ? "Dashboard'a Git" : 'Ücretsiz Başla',
        href: authenticated ? '/dashboard' : '/register',
      },
      features: [
        { text: '1 Adet Chatbot', included: true },
        { text: 'Aylık 100.000 Token', included: true },
        { text: '1 Web Sitesi Kaynağı', included: true },
        { text: '1 PDF Dosyası (Max 5MB)', included: true },
        { text: 'GPT-4o Mini Modeli', included: true },
        { text: 'botla.app İmzası', included: true },
        { text: 'Gelişmiş Guardrails', included: false },
        { text: 'Dinamik Web Tarama', included: false },
      ],
    },
    {
      title: 'Profesyonel (Pro)',
      price: '199 TL',
      recommended: true,
      cta: {
        text: authenticated ? "Dashboard'a Git" : "Pro'ya Geç",
        href: authenticated ? '/dashboard' : '/register',
      },
      features: [
        { text: '10 Adet Chatbot', included: true },
        { text: 'Aylık 1.000.000 Token', included: true },
        { text: '10 Web Sitesi & 20 PDF', included: true },
        { text: 'GPT-4o & GPT-4o Mini', included: true },
        { text: 'OCR (Görsel Tarama)', included: true },
        { text: 'Dinamik Web Tarama (Crawler)', included: true },
        { text: 'Gelişmiş Guardrails', included: true },
        { text: 'Öncelikli Destek', included: true },
      ],
    },
    {
      title: 'Ultra',
      price: '999 TL',
      cta: { text: 'İletişime Geç', href: 'mailto:sales@botla.app' },
      features: [
        { text: '100 Adet Chatbot', included: true },
        { text: 'Aylık 5.000.000 Token', included: true },
        { text: '50 Site & 100 PDF', included: true },
        { text: 'GPT-4o & Claude 3.5 Sonnet', included: true },
        { text: 'Tam Özelleştirme & Whitelabel', included: true },
        { text: 'İnsan Desteğine Aktarma', included: true },
        { text: 'Özel Entegrasyonlar', included: true },
        { text: '7/24 Özel Temsilci', included: true },
      ],
    },
  ]

  return (
    <section id="pricing" className="scroll-mt-24 py-32 bg-background relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center max-w-3xl mx-auto mb-20 px-4">
          <Badge>Fiyatlandırma</Badge>
          <h2 className="text-4xl md:text-6xl font-bold mt-6 mb-8 text-foreground tracking-tight leading-[1.1]">
            Şeffaf ve <span className="text-primary italic">Esnek</span> Paketler
          </h2>
          <p className="text-xl text-muted-foreground">
            İster küçük bir blog, ister büyük bir e-ticaret sitesi olun. Size uygun bir planımız var.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto items-center">
          {plans.map((plan, i) => (
            <PricingCard key={i} {...plan} />
          ))}
        </div>
      </div>
    </section>
  )
}

const HowItWorks = () => {
  const [activeStep, setActiveStep] = useState(0)
  const steps = [
    {
      num: '01',
      title: 'Kaynak Ekleyin',
      desc: 'PDF, web sitesi URL’i veya metin ekleyerek botunuzu eğitin.',
      icon: Database,
      bullets: ['PDF yükleyin', 'Web sitenizi taratın', 'Metin ekleyin'],
    },
    {
      num: '02',
      title: 'Kuralları Belirleyin',
      desc: 'Botunuzun tonunu, güvenlik kurallarını ve sınırlarını ayarlayın.',
      icon: ShieldCheck,
      bullets: ['Guardrails', 'Güvenli konuşma sınırları', 'Gerekirse insan desteğine yönlendirme'],
    },
    {
      num: '03',
      title: 'Widget ile Yayınlayın',
      desc: 'Tek satır kodla sitenize ekleyin, isterseniz domain kısıtlayın.',
      icon: MessagesSquare,
      bullets: ['Tek satır embed kodu', 'İzinli domain listesi', 'Güvenli embed (planına göre)'],
    },
  ]

  const active = steps[activeStep]
  const ActiveIcon = active.icon

  return (
    <section id="how-it-works" className="scroll-mt-24 py-32 bg-secondary/20 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center max-w-3xl mx-auto mb-20 px-4">
          <Badge>Nasıl Çalışır?</Badge>
          <h2 className="text-4xl md:text-6xl font-bold mt-6 mb-8 text-foreground tracking-tight leading-[1.1]">
            3 Adımda Yayına Alın
          </h2>
          <p className="text-xl text-muted-foreground leading-relaxed">
            Dakikalar içinde kurulum, günlerce süren geliştirme süreçlerine son.
          </p>
        </div>

        <div className="grid lg:grid-cols-12 gap-12 items-center">
          <div className="lg:col-span-5 space-y-4">
            {steps.map((step, i) => {
              const isActive = i === activeStep
              const Icon = step.icon
              return (
                <button
                  key={step.num}
                  type="button"
                  onClick={() => setActiveStep(i)}
                  onMouseEnter={() => setActiveStep(i)}
                  className={cn(
                    "w-full text-left relative rounded-[2rem] border transition-all duration-300 p-8",
                    isActive 
                      ? "border-primary/40 bg-card shadow-xl shadow-primary/5 -translate-x-2" 
                      : "border-border/40 bg-card/40 hover:bg-card hover:border-border"
                  )}
                >
                  <div className="flex items-center gap-6">
                    <div
                      className={cn(
                        "w-14 h-14 rounded-2xl flex items-center justify-center border transition-all duration-300",
                        isActive 
                          ? "bg-primary text-primary-foreground border-primary shadow-glow" 
                          : "bg-muted text-muted-foreground border-border"
                      )}
                    >
                      <Icon className="w-7 h-7" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-3 mb-1">
                        <div className="font-bold text-xl text-foreground">{step.title}</div>
                        <div className={cn("text-sm font-bold opacity-30 tracking-widest", isActive && "text-primary opacity-100")}>
                          {step.num}
                        </div>
                      </div>
                      <div className="text-muted-foreground leading-relaxed">
                        {step.desc}
                      </div>
                    </div>
                  </div>
                </button>
              )
            })}
          </div>

          <div className="lg:col-span-7">
            <AnimatePresence mode="wait">
              <motion.div
                key={activeStep}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.4, ease: "easeOut" }}
                className="rounded-[3rem] glass shadow-2xl p-12 border border-white/30"
              >
                <div className="flex items-center gap-6 mb-10">
                  <div className="w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary">
                    <ActiveIcon className="w-8 h-8" />
                  </div>
                  <div>
                    <div className="text-sm font-bold text-primary uppercase tracking-[0.2em]">Adım {active.num}</div>
                    <div className="text-3xl font-bold tracking-tight text-foreground">{active.title}</div>
                  </div>
                </div>

                <div className="space-y-4 mb-10">
                  {active.bullets.map((b: string) => (
                    <div
                      key={b}
                      className="rounded-2xl border border-border/40 bg-background/50 px-6 py-4 flex items-center gap-4 text-base font-medium text-foreground transition-all hover:bg-background hover:border-primary/30"
                    >
                      <CheckCircle2 className="w-6 h-6 text-success" />
                      {b}
                    </div>
                  ))}
                </div>

                <div className="rounded-[2rem] bg-muted/40 p-6 border border-border/30">
                  <div className="flex items-center gap-3 text-sm font-bold text-foreground mb-2">
                    <Sparkles className="w-4 h-4 text-primary" />
                    Profesyonel İpucu
                  </div>
                  <div className="text-muted-foreground leading-relaxed">
                    URL eklediğinizde sistem sayfaları otomatik keşfedebilir; isterseniz sadece belirli yolları dahil/haric tutarak taramayı kontrol edebilirsiniz.
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

// Minimal Badge Component
const Badge = ({ children }: { children: React.ReactNode }) => (
  <span className="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-xs font-bold text-primary tracking-widest uppercase ring-1 ring-inset ring-primary/20 shadow-sm shadow-primary/5">
    {children}
  </span>
)

const FAQ = () => {
  const faqs = [
    {
      q: 'Botu sitelerime nasıl eklerim?',
      a: 'Size verdiğimiz tek satırlık JavaScript kodunu sitenizin <head> veya <body> etiketleri arasına yapıştırmanız yeterlidir. Wordpress, Shopify, Wix gibi tüm altyapılarla uyumludur.',
    },
    {
      q: 'Hangi dosya formatlarını destekliyorsunuz?',
      a: "Şu an için PDF, metin (TXT / kopyala-yapıştır) ve doğrudan web sitesi URL'lerini destekliyoruz.",
    },
    {
      q: 'Ücretsiz planda kredi kartı gerekiyor mu?',
      a: 'Hayır, Başlangıç (Free) planımızı kullanmak için kredi kartı gerekmez. Sonsuza kadar ücretsiz kullanabilirsiniz.',
    },
    {
      q: 'Botum yanlış cevap verirse ne olur?',
      a: "Guardrails (Güvenlik Önlemleri) özelliğimiz sayesinde botun cevap veremediği veya emin olamadığı durumlarda 'Bunu bilmiyorum' demesini veya insan temsilciye yönlendirmesini sağlayabilirsiniz.",
    },
    {
      q: 'Verilerim güvende mi?',
      a: 'Evet, tüm verileriniz şifrelenerek saklanır ve sadece sizin botunuzun eğitimi için kullanılır. Başka hiçbir amaçla kullanılmaz veya paylaşılmaz.',
    },
  ]

  return (
    <section id="faq" className="scroll-mt-24 py-32 bg-secondary/20">
      <div className="max-w-4xl mx-auto px-6 sm:px-8 lg:px-10">
        <div className="text-center mb-20 px-4">
          <Badge>SSS</Badge>
          <h2 className="text-4xl md:text-5xl font-bold mt-6 mb-8 text-foreground tracking-tight">
            Sıkça Sorulan Sorular
          </h2>
          <p className="text-lg text-muted-foreground">Merak ettiğiniz her şey burada.</p>
        </div>

        <div className="space-y-4">
          {faqs.map((faq, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 10 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="border border-border/40 rounded-[2rem] bg-card overflow-hidden hover:border-primary/30 transition-colors"
            >
              <details className="group">
                <summary className="flex items-center justify-between p-8 cursor-pointer list-none">
                  <span className="font-bold text-xl pr-8">{faq.q}</span>
                  <span className="transition-transform group-open:rotate-180 bg-muted/50 p-2 rounded-full">
                    <ArrowRight className="w-5 h-5 rotate-90" />
                  </span>
                </summary>
                <div className="px-8 pb-8 pt-0 text-muted-foreground leading-relaxed">
                  <div className="pt-4 border-t border-border/10 text-lg">{faq.a}</div>
                </div>
              </details>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}

const Footer = ({ authenticated }: { authenticated: boolean }) => (
  <footer className="bg-foreground text-background py-24 border-t border-border/10">
    <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
      <div className="grid md:grid-cols-12 gap-16">
        <div className="col-span-12 md:col-span-5">
          <div className="flex items-center gap-3 mb-8">
            <div className="bg-primary p-3 rounded-2xl shadow-glow">
              <Bot className="w-7 h-7 text-primary-foreground" />
            </div>
            <span className="font-bold text-3xl tracking-tight text-background">botla.app</span>
          </div>
          <p className="text-background/60 text-lg mb-8 leading-relaxed max-w-sm">
            Yeni nesil web siteleri için geliştirilmiş, verilerinizle eğitilen akıllı müşteri asistanı. 
            Müşterilerinize 7/24 kesintisiz destek sunun.
          </p>
        </div>

        <div className="col-span-6 md:col-span-2 md:col-start-7">
          <h4 className="font-bold text-background text-lg mb-6">Ürün</h4>
          <ul className="space-y-4 text-background/60 text-base">
            <li><a href="#features" className="hover:text-primary transition-colors">Özellikler</a></li>
            <li><a href="#pricing" className="hover:text-primary transition-colors">Fiyatlandırma</a></li>
            {authenticated ? (
              <li><Link to="/dashboard" className="hover:text-primary transition-colors">Dashboard</Link></li>
            ) : (
              <>
                <li><Link to="/login" className="hover:text-primary transition-colors">Giriş Yap</Link></li>
                <li><Link to="/register" className="hover:text-primary transition-colors">Kayıt Ol</Link></li>
              </>
            )}
          </ul>
        </div>

        <div className="col-span-6 md:col-span-2">
          <h4 className="font-bold text-background text-lg mb-6">Şirket</h4>
          <ul className="space-y-4 text-background/60 text-base">
            <li><a href="#" className="hover:text-primary transition-colors">Hakkımızda</a></li>
            <li><a href="#" className="hover:text-primary transition-colors">Blog</a></li>
            <li><a href="#" className="hover:text-primary transition-colors">İletişim</a></li>
          </ul>
        </div>

        <div className="col-span-6 md:col-span-2">
          <h4 className="font-bold text-background text-lg mb-6">Yasal</h4>
          <ul className="space-y-4 text-background/60 text-base">
            <li><a href="#" className="hover:text-primary transition-colors">Gizlilik</a></li>
            <li><a href="#" className="hover:text-primary transition-colors">Koşullar</a></li>
          </ul>
        </div>
      </div>

      <div className="border-t border-background/10 mt-20 pt-10 flex flex-col md:flex-row justify-between items-center gap-4 text-background/40">
        <p className="text-sm font-medium">
          &copy; {new Date().getFullYear()} botla.app. Tüm hakları saklıdır.
        </p>
        <div className="flex items-center gap-6 text-sm">
          <span>Made with ❤️ in Istanbul</span>
        </div>
      </div>
    </div>
  </footer>
)

export default function LandingPage() {
  const [authenticated, setAuthenticated] = useState(false)

  useEffect(() => {
    const token = window.localStorage.getItem('botla_token')
    const isValid =
      token !== null && token !== 'undefined' && token !== 'null' && token.length > 0
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
    // Check if user is authenticated - don't show widget for logged in users
    const token = window.localStorage.getItem('botla_token')
    const isAuthenticated = token !== null && token !== 'undefined' && token !== 'null' && token.length > 0
    if (isAuthenticated) return

    const chatbotId = import.meta.env.VITE_LANDING_CHATBOT_ID
    const widgetUrl = import.meta.env.VITE_WIDGET_SCRIPT_URL

    if (!chatbotId || !widgetUrl) return // Skip if not configured

    // Check if script already exists
    const existingScript = document.querySelector(`script[data-bot="${chatbotId}"]`)
    if (existingScript) return

    const script = document.createElement('script')
    // Add reset-session=1 to URL for fresh demo experience each visit
    const widgetUrlWithReset = new URL(widgetUrl)
    widgetUrlWithReset.searchParams.set('reset-session', '1')
    script.src = widgetUrlWithReset.toString()
    script.type = 'module'
    script.setAttribute('data-bot', chatbotId)
    script.async = true
    document.body.appendChild(script)

    return () => {
      // Cleanup on unmount
      script.remove()
      // Also remove widget container if present
      const widgetHost = document.getElementById('chatbot-widget-host')
      if (widgetHost) widgetHost.remove()
    }
  }, [])

  return (
    <div className="min-h-screen bg-background font-sans selection:bg-primary/20 text-foreground">
      <Navbar authenticated={authenticated} />
      <main>
        <Hero authenticated={authenticated} />
        <Features />
        <HowItWorks />
        <Pricing authenticated={authenticated} />
        <FAQ />
      </main>
      <Footer authenticated={authenticated} />
    </div>
  )
}
