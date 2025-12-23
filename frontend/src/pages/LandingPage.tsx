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

// --- Components ---

const Navbar = () => {
  const [isOpen, setIsOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  const links = [
    { name: 'Özellikler', href: '#features' },
    { name: 'Nasıl Çalışır?', href: '#how-it-works' },
    { name: 'Fiyatlandırma', href: '#pricing' },
    { name: 'SSS', href: '#faq' },
  ]

  const handleScroll = (e: React.MouseEvent<HTMLAnchorElement>, href: string) => {
    e.preventDefault()

    // If we are not on the home page (landing page), navigate to home with hash
    if (location.pathname !== '/') {
      navigate('/' + href)
      return
    }

    // If we are on home page, smooth scroll and update URL
    const targetId = href.replace('#', '')
    const element = document.getElementById(targetId)
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' })
      // Update URL without jumping
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
    // Clear hash from URL when going to top
    window.history.pushState(null, '', '/')
  }

  return (
    <nav className="fixed inset-x-0 top-0 z-50 bg-background/70 backdrop-blur-xl border-b border-border/60">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center gap-2 cursor-pointer" onClick={handleLogoClick}>
            <div className="bg-primary/10 p-2 rounded-xl border border-primary/15">
              <Bot className="w-5 h-5 text-primary" />
            </div>
            <span className="font-semibold text-lg tracking-tight text-foreground">botla.app</span>
          </div>

          {/* Desktop Nav */}
          <div className="hidden md:flex items-center gap-8">
            {links.map((link) => (
              <a
                key={link.name}
                href={link.href}
                onClick={(e) => handleScroll(e, link.href)}
                className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
              >
                {link.name}
              </a>
            ))}
          </div>

          <div className="hidden md:flex items-center gap-4">
            <Link to="/login">
              <Button variant="ghost" className="font-medium">
                Giriş Yap
              </Button>
            </Link>
            <Link to="/register">
              <Button className="font-semibold shadow-sm">Ücretsiz Dene</Button>
            </Link>
          </div>

          {/* Mobile Menu Button */}
          <div className="md:hidden">
            <button onClick={() => setIsOpen(!isOpen)} className="text-foreground p-2">
              {isOpen ? <X /> : <Menu />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Nav */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="md:hidden border-t border-border bg-background"
          >
            <div className="px-4 py-6 space-y-4">
              {links.map((link) => (
                <a
                  key={link.name}
                  href={link.href}
                  onClick={(e) => handleScroll(e, link.href)}
                  className="block text-base font-medium text-foreground hover:text-primary"
                >
                  {link.name}
                </a>
              ))}
              <div className="pt-4 flex flex-col gap-3">
                <Link to="/login" onClick={() => setIsOpen(false)}>
                  <Button variant="outline" className="w-full justify-center">
                    Giriş Yap
                  </Button>
                </Link>
                <Link to="/register" onClick={() => setIsOpen(false)}>
                  <Button className="w-full justify-center">Ücretsiz Dene</Button>
                </Link>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </nav>
  )
}

const Hero = () => {
  const [heroSrc, setHeroSrc] = useState('/assets/landing-hero.png')
  const [useVideo, setUseVideo] = useState(true)

  return (
    <section className="relative pt-28 pb-16 lg:pt-36 lg:pb-24 overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-b from-secondary/70 via-background to-background" />
      <div className="absolute -top-24 right-[-120px] w-[520px] h-[520px] bg-primary/10 blur-[120px] rounded-full" />
      <div className="absolute -bottom-24 left-[-160px] w-[560px] h-[560px] bg-primary/5 blur-[140px] rounded-full" />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="grid lg:grid-cols-2 gap-12 lg:gap-8 items-center">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6 }}
          >
            <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-background/70 border border-border text-foreground text-sm font-semibold mb-6">
              <Sparkles className="w-4 h-4 text-primary" />
              <span>Yapay zeka ile destek hattınızı otomatikleştirin</span>
            </div>

            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-semibold tracking-tight text-foreground mb-6 leading-[1.08]">
              Verilerinizi
              <span className="block bg-clip-text text-transparent bg-gradient-to-r from-primary to-orange-600">
                Akıllı sohbetlere dönüştürün
              </span>
            </h1>

            <p className="text-base sm:text-lg text-muted-foreground mb-8 max-w-xl leading-relaxed">
              PDF dosyalarınızı yükleyin, web sitenizi taratın veya metin ekleyin. 1 dakika içinde
              size özel, 7/24 çalışan ve satış yapan bir yapay zeka asistanı oluşturun.
            </p>

            <div className="flex flex-col sm:flex-row gap-3 sm:items-center mb-10">
              <Link to="/register">
                <Button size="lg" className="h-12 px-6 text-base font-semibold shadow-sm">
                  Hemen Başla
                  <ArrowRight className="ml-2 w-4 h-4" />
                </Button>
              </Link>
              <a href="#how-it-works">
                <Button
                  variant="outline"
                  size="lg"
                  className="h-12 px-6 text-base font-semibold bg-background/60"
                >
                  Nasıl Çalışır?
                </Button>
              </a>
            </div>

            <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-6 text-sm text-muted-foreground font-medium">
              <span className="flex items-center gap-2">
                <CheckCircle2 className="w-5 h-5 text-green-500" /> Kredi Kartı Gerekmez
              </span>
              <span className="flex items-center gap-2">
                <CheckCircle2 className="w-5 h-5 text-green-500" /> Türkçe Destek
              </span>
            </div>

            <div className="mt-10 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <span className="rounded-full border border-border bg-background/60 px-3 py-1">
                PDF
              </span>
              <span className="rounded-full border border-border bg-background/60 px-3 py-1">
                Website
              </span>
              <span className="rounded-full border border-border bg-background/60 px-3 py-1">
                Metin
              </span>
              <span className="rounded-full border border-border bg-background/60 px-3 py-1">
                Widget
              </span>
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            className="relative w-full flex items-center justify-center"
          >
            {/* 3D Asset Container */}
            <div
              className={`relative w-full rounded-2xl bg-gradient-to-b from-muted/50 to-background border border-border/60 shadow-2xl shadow-primary/5 overflow-hidden ${useVideo ? 'aspect-video shadow-primary/10' : 'aspect-square max-w-[500px] shadow-primary/5'}`}
            >
              {useVideo ? (
                <video
                  className="w-full h-full object-cover"
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
                  alt="Data to Chat Pipeline"
                  className="w-full h-full object-cover"
                  onError={() => setHeroSrc('/assets/landing-hero-final.png')}
                />
              )}
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  )
}

const FeatureCard = ({ icon: Icon, title, desc, delay }: any) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    whileHover={{ y: -6 }}
    viewport={{ once: true }}
    transition={{ delay, duration: 0.5 }}
    className="group p-8 rounded-3xl bg-card border border-border hover:border-primary/30 hover:shadow-xl hover:shadow-primary/5 transition-all duration-300 relative overflow-hidden"
  >
    <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-br from-primary/10 to-transparent rounded-bl-full opacity-0 group-hover:opacity-100 transition-opacity" />

    <div className="w-14 h-14 rounded-2xl bg-primary/10 flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-300">
      <Icon className="w-7 h-7 text-primary" />
    </div>

    <h3 className="text-xl font-bold mb-3 text-foreground">{title}</h3>
    <p className="text-muted-foreground leading-relaxed">{desc}</p>
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
    <section id="features" className="scroll-mt-24 py-20 sm:py-24 bg-secondary/40 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <Badge>Özellikler</Badge>
          <h2 className="text-3xl md:text-5xl font-bold mt-4 mb-6 text-foreground">
            Sadece Bir Chatbot Değil,
            <br />
            Tam Donanımlı Bir Asistan
          </h2>
          <p className="text-lg text-muted-foreground">
            İşletmenizin ihtiyaçlarına göre özelleştirilebilen güçlü altyapı.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
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
    className={`relative p-8 rounded-3xl border ${recommended ? 'border-primary shadow-2xl shadow-primary/10 bg-card' : 'border-border bg-card/50'} flex flex-col`}
  >
    {recommended && (
      <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 rounded-full bg-primary text-primary-foreground text-sm font-bold shadow-lg">
        En Popüler
      </div>
    )}

    <div className="mb-8">
      <h3 className="text-xl font-bold mb-2">{title}</h3>
      <div className="flex items-baseline gap-1">
        <span className="text-4xl font-bold">{price}</span>
        <span className="text-muted-foreground text-sm">/ay</span>
      </div>
    </div>

    <ul className="space-y-4 mb-8 flex-1">
      {features.map((f: any, i: number) => (
        <li key={i} className="flex items-start gap-3 text-sm text-muted-foreground">
          {f.included ? (
            <CheckCircle2 className="w-5 h-5 text-green-500 shrink-0" />
          ) : (
            <X className="w-5 h-5 text-gray-300 shrink-0" />
          )}
          <span className={f.included ? 'text-foreground font-medium' : 'text-muted-foreground/60'}>
            {f.text}
          </span>
        </li>
      ))}
    </ul>

    <Link to={cta.href} className="w-full">
      <Button
        variant={recommended ? 'default' : 'outline'}
        className="w-full h-12 text-base font-semibold"
      >
        {cta.text}
      </Button>
    </Link>
  </div>
)

const Pricing = () => {
  const plans = [
    {
      title: 'Başlangıç (Free)',
      price: '0 TL',
      cta: { text: 'Ücretsiz Başla', href: '/register' },
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
      cta: { text: "Pro'ya Geç", href: '/register' },
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
    <section id="pricing" className="scroll-mt-24 py-24">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <Badge>Fiyatlandırma</Badge>
          <h2 className="text-3xl md:text-5xl font-bold mt-4 mb-6 text-foreground">
            Şeffaf ve Esnek Paketler
          </h2>
          <p className="text-lg text-muted-foreground">
            İster küçük bir blog, ister büyük bir e-ticaret sitesi olun. Size uygun bir planımız
            var.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
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
    <section id="how-it-works" className="scroll-mt-24 py-20 sm:py-24">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto mb-14">
          <Badge>Nasıl Çalışır?</Badge>
          <h2 className="text-3xl md:text-5xl font-bold mt-4 mb-5 text-foreground">
            3 adımda yayına alın
          </h2>
          <p className="text-lg text-muted-foreground">
            Dakikalar içinde kurulum, günlerce süren geliştirme yok.
          </p>
        </div>

        <div className="grid lg:grid-cols-12 gap-8 items-start">
          <div className="lg:col-span-5 space-y-3">
            {steps.map((step, i) => {
              const isActive = i === activeStep
              const Icon = step.icon
              return (
                <button
                  key={step.num}
                  type="button"
                  onClick={() => setActiveStep(i)}
                  onMouseEnter={() => setActiveStep(i)}
                  onFocus={() => setActiveStep(i)}
                  className={`w-full text-left relative rounded-2xl border transition-all duration-200 p-5 ${isActive ? 'border-primary/40 bg-primary/5 shadow-sm' : 'border-border bg-card hover:bg-muted/30'}`}
                >
                  <div className="flex items-start gap-4">
                    <div
                      className={`w-10 h-10 rounded-xl flex items-center justify-center border ${isActive ? 'bg-primary/10 text-primary border-primary/20' : 'bg-muted text-muted-foreground border-border'}`}
                    >
                      <Icon className="w-5 h-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-3">
                        <div className="font-semibold text-foreground">{step.title}</div>
                        <div
                          className={`text-sm font-bold tabular-nums ${isActive ? 'text-primary' : 'text-muted-foreground'}`}
                        >
                          {step.num}
                        </div>
                      </div>
                      <div className="mt-1 text-sm text-muted-foreground leading-relaxed">
                        {step.desc}
                      </div>
                    </div>
                  </div>
                  {isActive && (
                    <motion.div
                      layoutId="how-it-works-active"
                      className="absolute inset-0 rounded-2xl ring-1 ring-primary/15"
                      transition={{ type: 'spring', bounce: 0.2, duration: 0.55 }}
                    />
                  )}
                </button>
              )
            })}
          </div>

          <div className="lg:col-span-7 lg:sticky lg:top-24">
            <AnimatePresence mode="wait">
              <motion.div
                key={activeStep}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: 10 }}
                transition={{ duration: 0.28, ease: 'easeOut' }}
                className="rounded-3xl bg-card border border-border shadow-sm p-8"
              >
                <div className="flex items-start gap-4">
                  <div className="w-12 h-12 rounded-2xl bg-primary/10 border border-primary/15 flex items-center justify-center text-primary">
                    <ActiveIcon className="w-6 h-6" />
                  </div>
                  <div className="flex-1">
                    <div className="text-xs font-semibold text-muted-foreground">
                      Adım {active.num}
                    </div>
                    <div className="mt-1 text-2xl font-bold tracking-tight text-foreground">
                      {active.title}
                    </div>
                    <div className="mt-2 text-muted-foreground leading-relaxed">{active.desc}</div>
                  </div>
                </div>

                <div className="mt-6 grid sm:grid-cols-2 gap-3">
                  {active.bullets.map((b: string) => (
                    <div
                      key={b}
                      className="rounded-2xl border border-border bg-background/60 px-4 py-3 flex items-center gap-2 text-sm text-foreground"
                    >
                      <CheckCircle2 className="w-4 h-4 text-green-600" />
                      <span className="font-medium">{b}</span>
                    </div>
                  ))}
                </div>

                <div className="mt-8 rounded-2xl border border-border bg-secondary/40 p-5">
                  <div className="text-sm font-semibold text-foreground">İpucu</div>
                  <div className="mt-1 text-sm text-muted-foreground leading-relaxed">
                    URL eklediğinizde sistem sayfaları keşfedebilir; isterseniz sadece belirli
                    yolları dahil/haric tutarak taramayı kontrol edebilirsiniz.
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
  <span className="inline-flex items-center rounded-full bg-primary/10 px-3 py-1 text-sm font-medium text-primary ring-1 ring-inset ring-primary/20">
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
    <section id="faq" className="scroll-mt-24 py-24 bg-secondary/20">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <Badge>SSS</Badge>
          <h2 className="text-3xl md:text-4xl font-bold mt-4 mb-6 text-foreground">
            Sıkça Sorulan Sorular
          </h2>
        </div>

        <div className="space-y-4">
          {faqs.map((faq, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 10 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.1 }}
              className="border border-border rounded-2xl bg-card overflow-hidden"
            >
              <details className="group">
                <summary className="flex items-center justify-between p-6 cursor-pointer list-none">
                  <span className="font-semibold text-lg">{faq.q}</span>
                  <span className="transition-transform group-open:rotate-180">
                    <ArrowRight className="w-5 h-5 rotate-90" />
                  </span>
                </summary>
                <div className="px-6 pb-6 pt-0 text-muted-foreground leading-relaxed border-t border-border/50">
                  <div className="pt-4">{faq.a}</div>
                </div>
              </details>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}

const Footer = () => (
  <footer className="bg-foreground text-background py-16 border-t border-border">
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="grid md:grid-cols-12 gap-12">
        <div className="col-span-12 md:col-span-4">
          <div className="flex items-center gap-2 mb-6">
            <div className="bg-primary p-2 rounded-lg">
              <Bot className="w-6 h-6 text-primary-foreground" />
            </div>
            <span className="font-bold text-2xl text-background">botla.app</span>
          </div>
          <p className="text-background/70 mb-6">
            Yeni nesil web siteleri için geliştirilmiş, verilerinizle eğitilen akıllı müşteri
            asistanı.
          </p>
        </div>

        <div className="col-span-6 md:col-span-2 md:col-start-7">
          <h4 className="font-semibold text-background mb-4">Ürün</h4>
          <ul className="space-y-3 text-sm text-background/70">
            <li>
              <a href="#features" className="hover:text-background transition-colors">
                Özellikler
              </a>
            </li>
            <li>
              <a href="#pricing" className="hover:text-background transition-colors">
                Fiyatlandırma
              </a>
            </li>
            <li>
              <Link to="/login" className="hover:text-background transition-colors">
                Giriş Yap
              </Link>
            </li>
            <li>
              <Link to="/register" className="hover:text-background transition-colors">
                Kayıt Ol
              </Link>
            </li>
          </ul>
        </div>

        <div className="col-span-6 md:col-span-2">
          <h4 className="font-semibold text-background mb-4">Şirket</h4>
          <ul className="space-y-3 text-sm text-background/70">
            <li>
              <a href="#" className="hover:text-background transition-colors">
                Hakkımızda
              </a>
            </li>
            <li>
              <a href="#" className="hover:text-background transition-colors">
                Blog
              </a>
            </li>
            <li>
              <a href="#" className="hover:text-background transition-colors">
                İletişim
              </a>
            </li>
          </ul>
        </div>

        <div className="col-span-6 md:col-span-2">
          <h4 className="font-semibold text-background mb-4">Yasal</h4>
          <ul className="space-y-3 text-sm text-background/70">
            <li>
              <a href="#" className="hover:text-background transition-colors">
                Gizlilik Politikası
              </a>
            </li>
            <li>
              <a href="#" className="hover:text-background transition-colors">
                Kullanım Koşulları
              </a>
            </li>
          </ul>
        </div>
      </div>

      <div className="border-t border-background/15 mt-12 pt-8 text-center text-sm text-background/60">
        &copy; {new Date().getFullYear()} botla.app. Tüm hakları saklıdır.
      </div>
    </div>
  </footer>
)

export default function LandingPage() {
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
      <Navbar />
      <main>
        <Hero />
        <Features />
        <HowItWorks />
        <Pricing />
        <FAQ />
      </main>
      <Footer />
    </div>
  )
}
