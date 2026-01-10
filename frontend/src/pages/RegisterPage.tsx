import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { 
  Mail, 
  Lock, 
  User, 
  ArrowRight, 
  Zap, 
  Shield, 
  Clock,
  Sparkles,
  CheckCircle2,
  Database,
  Globe
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'
import { getTurkishErrorMessage } from '@/lib/errorMessages'
import { useAuth } from '@/contexts/AuthContext'
import { motion } from 'framer-motion'

const RegisterPage = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const { refetch } = useAuth()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !email || !password) {
      toast('Lütfen tüm alanları doldurun.', 'error')
      return
    }

    setErrorMsg(null)
    setIsLoading(true)
    try {
      // Register the user - backend sets HttpOnly cookies automatically
      await api.post('/api/v1/auth/register', { full_name: name, email, password })
      
      // Clear any stale org selection
      window.localStorage.removeItem('botla_last_org_id')

      // Refetch user to update AuthContext state
      refetch()

      // Fetch onboarding status to determine where to redirect
      const { data: onboardingState } = await api.get('/api/v1/me/onboarding')

      toast('Hesabınız oluşturuldu! Hadi başlayalım.', 'success')

      // Redirect based on onboarding status
      if (onboardingState.completed || onboardingState.skipped) {
        navigate('/dashboard')
      } else {
        navigate('/onboarding')
      }
    } catch (err: any) {
      const errorMessage = getTurkishErrorMessage(err, 'Kayıt başarısız. Lütfen tekrar deneyin.')
      toast(errorMessage, 'error')
      setErrorMsg(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  const features = [
    {
      icon: Zap,
      title: 'Dakikalar İçinde Kurulum',
      description: 'Bot oluşturma sürecinizi sadece birkaç dakikada tamamlayın.',
    },
    {
      icon: Shield,
      title: 'Güvenli & Gizli',
      description: 'Verileriniz şifrelenir ve güvenle saklanır.',
    },
    {
      icon: Clock,
      title: '7/24 Aktif',
      description: 'Botunuz asla uyumaz, her zaman müşterilerinize hizmet eder.',
    },
  ]

  const stats = [
    { value: '10+', label: 'Kaynak Türü', icon: Database },
    { value: '50+', label: 'Dil Desteği', icon: Globe },
  ]

  return (
    <div className="min-h-screen flex bg-background relative overflow-hidden">
      {/* Background Effects */}
      <div className="absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,rgba(245,158,11,0.12),transparent)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(251,146,60,0.08),transparent_50%)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_80%_80%,rgba(245,158,11,0.05),transparent_50%)]" />
        
        {/* Grid Pattern */}
        <div 
          className="absolute inset-0 opacity-[0.02]"
          style={{
            backgroundImage: `linear-gradient(to right, currentColor 1px, transparent 1px),
                             linear-gradient(to bottom, currentColor 1px, transparent 1px)`,
            backgroundSize: '64px 64px'
          }}
        />
      </div>

      {/* Floating Glow Effects */}
      <div className="absolute top-40 right-1/4 w-[500px] h-[500px] bg-primary/10 rounded-full blur-[120px] pointer-events-none" />
      <div className="absolute bottom-20 left-1/4 w-[400px] h-[400px] bg-orange-500/8 rounded-full blur-[100px] pointer-events-none" />

      {/* Left Side - Registration Form */}
      <div className="flex-1 flex items-center justify-center p-6 lg:p-8 xl:p-12 relative z-10">
        <div className="w-full max-w-[420px]">
          {/* Mobile Logo */}
          <motion.div 
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="lg:hidden flex items-center justify-center gap-3 mb-10"
          >
            <div className="relative">
              <div className="absolute inset-0 bg-primary/20 blur-lg rounded-xl" />
              <img
                src="/logo-128.png"
                alt="Botla Logo"
                className="relative w-9 h-9 rounded-xl shadow-lg"
              />
            </div>
            <span className="text-xl font-bold text-foreground">botla.app</span>
          </motion.div>

          {/* Form Card */}
          <motion.div
            initial={{ opacity: 0, y: 30, scale: 0.97 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            transition={{ duration: 0.6 }}
            className="relative"
          >
            {/* Card Glow */}
            <div className="absolute -inset-1 bg-gradient-to-r from-primary/20 via-orange-500/10 to-primary/20 rounded-3xl blur-xl opacity-60" />
            
            <div 
              className="relative p-6 sm:p-8 rounded-2xl bg-card/80 backdrop-blur-xl border border-border/50 shadow-xl"
              data-testid="register-page"
            >
              <div className="text-center mb-6">
                <motion.h2
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.1 }}
                  className="text-xl sm:text-2xl font-bold text-foreground mb-1"
                  data-testid="register-page-title"
                >
                  Hesap Oluştur
                </motion.h2>
                <motion.p
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.15 }}
                  className="text-sm text-muted-foreground"
                >
                  Hemen başlayın, ücretsiz deneyin
                </motion.p>
              </div>

              {/* Error Message */}
              {errorMsg && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="mb-5 p-3 rounded-xl bg-destructive/10 border border-destructive/20
                           text-destructive text-sm"
                  role="alert"
                  data-testid="register-page-error-message"
                >
                  {errorMsg}
                </motion.div>
              )}

              <form onSubmit={handleRegister} className="space-y-4">
                {/* Name Field */}
                <motion.div 
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.2 }}
                  className="space-y-1.5"
                >
                  <label className="text-sm font-medium text-foreground" htmlFor="name">
                    Ad Soyad
                  </label>
                  <div className="relative group">
                    <User className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
                    <Input
                      id="name"
                      placeholder="Adınız Soyadınız"
                      data-testid="register-page-name-input"
                      className="pl-10 h-11 rounded-xl border-border/50 bg-background/50 backdrop-blur-sm
                               focus:bg-background focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                               transition-all duration-200"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                    />
                  </div>
                </motion.div>

                {/* Email Field */}
                <motion.div 
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.25 }}
                  className="space-y-1.5"
                >
                  <label className="text-sm font-medium text-foreground" htmlFor="email">
                    Email
                  </label>
                  <div className="relative group">
                    <Mail className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
                    <Input
                      id="email"
                      placeholder="ornek@sirket.com"
                      type="email"
                      data-testid="register-page-email-input"
                      className="pl-10 h-11 rounded-xl border-border/50 bg-background/50 backdrop-blur-sm
                               focus:bg-background focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                               transition-all duration-200"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                    />
                  </div>
                </motion.div>

                {/* Password Field */}
                <motion.div 
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.3 }}
                  className="space-y-1.5"
                >
                  <label className="text-sm font-medium text-foreground" htmlFor="password">
                    Şifre
                  </label>
                  <div className="relative group">
                    <Lock className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
                    <Input
                      id="password"
                      type="password"
                      placeholder="Güçlü şifre oluşturun"
                      data-testid="register-page-password-input"
                      className="pl-10 h-11 rounded-xl border-border/50 bg-background/50 backdrop-blur-sm
                               focus:bg-background focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                               transition-all duration-200"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                    />
                  </div>
                  <p className="text-xs text-muted-foreground">
                    En az 8 karakter, büyük harf, küçük harf, rakam ve özel karakter
                  </p>
                </motion.div>

                {/* Submit Button */}
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.35 }}
                  className="pt-2"
                >
                  <Button
                    className="w-full h-11 rounded-xl text-sm font-semibold
                             shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30
                             hover:scale-[1.02] active:scale-[0.98]
                             transition-all duration-300 group"
                    type="submit"
                    isLoading={isLoading}
                    data-testid="register-page-submit-button"
                  >
                    <span>Kayıt Ol</span>
                    {!isLoading && (
                      <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
                    )}
                  </Button>
                </motion.div>
              </form>

              {/* Divider */}
              <motion.div 
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.5, delay: 0.4 }}
                className="relative my-6"
              >
                <div className="absolute inset-0 flex items-center">
                  <span className="w-full border-t border-border/50" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="px-4 text-muted-foreground bg-card/80 rounded-full">
                    veya
                  </span>
                </div>
              </motion.div>

              {/* Login Link */}
              <motion.p 
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.45 }}
                className="text-center text-sm text-muted-foreground"
              >
                Zaten hesabınız var mı?{' '}
                <Link
                  to="/login"
                  className="font-semibold text-primary hover:text-primary/80 
                           transition-colors duration-200"
                >
                  Giriş Yapın
                </Link>
              </motion.p>

              {/* Trust Indicators */}
              <motion.div
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.5 }}
                className="mt-6 pt-5 border-t border-border/50"
              >
                <div className="flex items-center justify-center gap-4 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                    <span>Ücretsiz Plan</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                    <span>Kredi Kartı Gerekmez</span>
                  </div>
                </div>
              </motion.div>
            </div>
          </motion.div>
        </div>
      </div>

      {/* Right Side - Features (Hidden on mobile) */}
      <div className="hidden lg:flex lg:w-1/2 flex-col p-8 xl:p-12 relative z-10">
        {/* Logo - Fixed at top */}
        <motion.div 
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="flex items-center gap-3"
        >
          <div className="relative">
            <div className="absolute inset-0 bg-primary/20 blur-xl rounded-2xl" />
            <img
              src="/logo-128.png"
              alt="Botla Logo"
              className="relative w-10 h-10 rounded-xl shadow-lg"
            />
          </div>
          <span className="font-bold text-xl tracking-tight">botla.app</span>
        </motion.div>

        {/* Features Content - Centered */}
        <div className="flex-1 flex flex-col justify-center max-w-lg">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.1 }}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/5 border border-primary/10 mb-6 w-fit"
          >
            <Sparkles className="w-4 h-4 text-primary" />
            <span className="text-sm font-semibold text-primary">Ücretsiz Başlayın</span>
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="text-3xl xl:text-4xl 2xl:text-5xl font-bold tracking-tight leading-[1.15] text-foreground mb-5"
          >
            Müşteri desteğinizi{' '}
            <span className="bg-gradient-to-r from-primary via-orange-500 to-amber-500 bg-clip-text text-transparent">
              otomatize edin
            </span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            className="text-base xl:text-lg text-muted-foreground leading-relaxed mb-8"
          >
            Yapay zeka destekli chatbot ile işletmenizi bir üst seviyeye taşıyın.
          </motion.p>

          {/* Stats */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="flex items-center gap-6 mb-8"
          >
            {stats.map((stat, index) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.5, delay: 0.5 + index * 0.1 }}
                className="text-center"
              >
                <div className="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-primary/10 mb-2">
                  <stat.icon className="w-4 h-4 text-primary" />
                </div>
                <div className="text-xl font-bold text-foreground">{stat.value}</div>
                <div className="text-xs text-muted-foreground">{stat.label}</div>
              </motion.div>
            ))}
          </motion.div>

          {/* Feature Cards */}
          <div className="space-y-3">
            {features.map((feature, index) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.5, delay: 0.6 + index * 0.1 }}
                className="flex items-center gap-4 p-3.5 rounded-xl bg-card/50 border border-border/50 backdrop-blur-sm hover:border-primary/20 transition-colors"
              >
                <div className="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
                  <feature.icon className="w-4 h-4 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold text-foreground text-sm">{feature.title}</h3>
                  <p className="text-xs text-muted-foreground">{feature.description}</p>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Footer - Fixed at bottom */}
        <motion.div 
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.9 }}
          className="text-sm text-muted-foreground"
        >
          © {new Date().getFullYear()} botla.app. Tüm hakları saklıdır.
        </motion.div>
      </div>
    </div>
  )
}

export default RegisterPage
