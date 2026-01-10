import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { 
  Mail, 
  Lock, 
  ArrowRight, 
  Sparkles,
  ShieldCheck,
  Zap,
  Clock,
  CheckCircle2
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'
import { useAuth } from '@/contexts/AuthContext'
import { motion } from 'framer-motion'

const LoginPage = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const { refetch } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || !password) {
      toast('Lütfen tüm alanları doldurun.', 'error')
      return
    }

    setIsLoading(true)
    try {
      // Login - backend sets HttpOnly cookies automatically
      await api.post('/api/v1/auth/login', { email, password })
      
      // Refetch user to update AuthContext state and wait for it
      const { data: user } = await refetch()
      
      if (user) {
        toast('Giriş başarılı! Yönlendiriliyorsunuz...', 'success')
        navigate('/dashboard')
      } else {
        toast('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.', 'error')
      }
    } catch {
      toast('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  const features = [
    {
      icon: Zap,
      title: 'Hızlı Kurulum',
      description: 'Dakikalar içinde chatbotunuzu oluşturun',
    },
    {
      icon: ShieldCheck,
      title: 'Güvenli Altyapı',
      description: 'Verileriniz şifrelenmiş olarak korunur',
    },
    {
      icon: Clock,
      title: '7/24 Aktif',
      description: 'Kesintisiz müşteri desteği sunun',
    },
  ]

  return (
    <div className="min-h-screen flex bg-background relative overflow-hidden">
      {/* Background Effects */}
      <div className="absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,rgba(245,158,11,0.12),transparent)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_80%_20%,rgba(251,146,60,0.08),transparent_50%)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_80%,rgba(245,158,11,0.05),transparent_50%)]" />
        
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
      <div className="absolute top-20 left-1/4 w-[500px] h-[500px] bg-primary/10 rounded-full blur-[120px] pointer-events-none" />
      <div className="absolute bottom-20 right-1/4 w-[400px] h-[400px] bg-orange-500/8 rounded-full blur-[100px] pointer-events-none" />

      {/* Left Side - Branding Section (Hidden on mobile) */}
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

        {/* Hero Content - Centered */}
        <div className="flex-1 flex flex-col justify-center max-w-lg">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.1 }}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/5 border border-primary/10 mb-6 w-fit"
          >
            <Sparkles className="w-4 h-4 text-primary" />
            <span className="text-sm font-semibold text-primary">Yapay Zeka Destekli</span>
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="text-3xl xl:text-4xl 2xl:text-5xl font-bold tracking-tight leading-[1.15] text-foreground mb-5"
          >
            Kendi verilerinizle eğitilmiş{' '}
            <span className="bg-gradient-to-r from-primary via-orange-500 to-amber-500 bg-clip-text text-transparent">
              akıllı chatbotlar
            </span>{' '}
            oluşturun.
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            className="text-base xl:text-lg text-muted-foreground leading-relaxed mb-8"
          >
            PDF, URL veya metin dosyalarınızı yükleyin, dakikalar içinde sitenize entegre edin.
          </motion.p>

          {/* Feature Pills */}
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="space-y-3"
          >
            {features.map((feature, index) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.5, delay: 0.5 + index * 0.1 }}
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
          </motion.div>
        </div>

        {/* Footer - Fixed at bottom */}
        <motion.div 
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.8 }}
          className="text-sm text-muted-foreground"
        >
          © {new Date().getFullYear()} botla.app. Tüm hakları saklıdır.
        </motion.div>
      </div>

      {/* Right Side - Login Form */}
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
            transition={{ duration: 0.6, delay: 0.1 }}
            className="relative"
          >
            {/* Card Glow */}
            <div className="absolute -inset-1 bg-gradient-to-r from-primary/20 via-orange-500/10 to-primary/20 rounded-3xl blur-xl opacity-60" />
            
            <div 
              className="relative p-6 sm:p-8 rounded-2xl bg-card/80 backdrop-blur-xl border border-border/50 shadow-xl"
              data-testid="login-page"
            >
              <div className="text-center mb-6">
                <motion.h2
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.2 }}
                  className="text-xl sm:text-2xl font-bold text-foreground mb-1"
                  data-testid="login-page-title"
                >
                  Hoş Geldiniz
                </motion.h2>
                <motion.p
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.25 }}
                  className="text-sm text-muted-foreground"
                >
                  Hesabınıza giriş yapın
                </motion.p>
              </div>

              <form onSubmit={handleLogin} className="space-y-4">
                {/* Email Field */}
                <motion.div 
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.3 }}
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
                      data-testid="login-page-email-input"
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
                  transition={{ duration: 0.5, delay: 0.35 }}
                  className="space-y-1.5"
                >
                  <div className="flex items-center justify-between">
                    <label className="text-sm font-medium text-foreground" htmlFor="password">
                      Şifre
                    </label>
                    <Link
                      to="/forgot-password"
                      className="text-xs font-medium text-primary hover:text-primary/80
                               transition-colors duration-200"
                      data-testid="login-page-forgot-password-link"
                    >
                      Şifremi unuttum?
                    </Link>
                  </div>
                  <div className="relative group">
                    <Lock className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
                    <Input
                      id="password"
                      type="password"
                      placeholder="••••••••"
                      data-testid="login-page-password-input"
                      className="pl-10 h-11 rounded-xl border-border/50 bg-background/50 backdrop-blur-sm
                               focus:bg-background focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                               transition-all duration-200"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                    />
                  </div>
                </motion.div>

                {/* Submit Button */}
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.4 }}
                  className="pt-2"
                >
                  <Button
                    className="w-full h-11 rounded-xl text-sm font-semibold
                             shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30
                             hover:scale-[1.02] active:scale-[0.98]
                             transition-all duration-300 group"
                    type="submit"
                    isLoading={isLoading}
                    data-testid="login-page-submit-button"
                  >
                    <span>Giriş Yap</span>
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
                transition={{ duration: 0.5, delay: 0.45 }}
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

              {/* Register Link */}
              <motion.p 
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.5 }}
                className="text-center text-sm text-muted-foreground"
              >
                Hesabınız yok mu?{' '}
                <Link
                  to="/register"
                  className="font-semibold text-primary hover:text-primary/80 
                           transition-colors duration-200"
                >
                  Kayıt Olun
                </Link>
              </motion.p>

              {/* Trust Indicators */}
              <motion.div
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.55 }}
                className="mt-6 pt-5 border-t border-border/50"
              >
                <div className="flex items-center justify-center gap-4 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                    <span>SSL Korumalı</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                    <span>KVKK Uyumlu</span>
                  </div>
                </div>
              </motion.div>
            </div>
          </motion.div>
        </div>
      </div>
    </div>
  )
}

export default LoginPage
